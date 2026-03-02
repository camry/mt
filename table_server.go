package mt

import (
    "context"

    "gitea.meigushi.go/meigushi/pb/cpb"
    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/camry/g/v2/glog"
    "go.etcd.io/etcd/client/v3"
    "google.golang.org/protobuf/proto"
    "gorm.io/gorm"
)

var tableMap = make(map[string]ITableManager, TableSizeDefault)

func Register(tableManager ITableManager) {
    tableMap[tableManager.TableName()] = tableManager
}

type TableServer struct {
    db     *gorm.DB
    ec     *clientv3.Client
    cancel context.CancelFunc
}

func New(db *gorm.DB, ec *clientv3.Client) *TableServer {
    return &TableServer{
        db: db,
        ec: ec,
    }
}

func (x *TableServer) Start(ctx context.Context) error {
    ctx, x.cancel = context.WithCancel(ctx)
    var cTables []*cpb.CTables
    err := x.db.Find(&cTables).Error
    if err != nil {
        return gerror.Wrap(err, "TableServer Start DB Find cpb.CTables Failed")
    }
    for _, cTable := range cTables {
        if iTableManager, ok := tableMap[cTable.GetTableName()]; ok {
            err = iTableManager.Load(cTable.GetTableCrc(), cTable.GetTableData())
            if err != nil {
                return gerror.Wrapf(err, "TableServer Start ITableManager Load %s Failed", cTable.GetTableName())
            }
            glog.Debugf("TableServer Start ITableManager Load %s Success", cTable.GetTableName())
        }
    }
    go x.watch(ctx)
    return nil
}

func (x *TableServer) Stop(ctx context.Context) error {
    tableMap = make(map[string]ITableManager, TableSizeDefault)
    glog.Debugf("TableServer Stop")
    if x.cancel != nil {
        x.cancel()
    }
    return nil
}

func (x *TableServer) watch(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            glog.Debugf("TableServer watch Stop")
            return
        case watchResponse := <-x.ec.Watch(ctx, TableKeyDefault):
            for _, event := range watchResponse.Events {
                switch event.Type {
                case clientv3.EventTypePut:
                    cTableIndex := &cpb.CTableIndex{}
                    err := proto.Unmarshal(event.Kv.Value, cTableIndex)
                    if err != nil {
                        glog.Error(gerror.Wrap(err, "TableServer watch proto.Unmarshal Failed"))
                        continue
                    }
                    tableNameList := make([]string, 0, TableSizeDefault)
                    for _, cTableHash := range cTableIndex.GetTableHashMap() {
                        if iTableManager, ok := tableMap[cTableHash.GetTableName()]; ok {
                            if iTableManager.TableCrc() != cTableHash.GetTableCrc() {
                                tableNameList = append(tableNameList, cTableHash.GetTableName())
                            } else {
                                glog.Debugf("TableServer watch TableName %s cTable=%d EQ cTableIndexHash=%d", cTableHash.GetTableName(), cTableHash.GetTableCrc(), cTableHash.GetTableCrc())
                            }
                        }
                    }
                    var cTables []*cpb.CTables
                    err = x.db.Where("table_name IN ?", tableNameList).Find(&cTables).Error
                    if err != nil {
                        glog.Error(gerror.Wrap(err, "TableServer watch DB Find cbpb.CTables Failed"))
                        continue
                    }
                    for _, cTable := range cTables {
                        iTableManager, ok1 := tableMap[cTable.GetTableName()]
                        cTableHash, ok2 := cTableIndex.GetTableHashMap()[cTable.GetTableName()]
                        if ok1 && ok2 {
                            if cTable.GetTableCrc() == cTableHash.GetTableCrc() {
                                err = iTableManager.Watch(cTable.GetTableCrc(), cTable.GetTableData())
                                if err != nil {
                                    glog.Error(gerror.Wrapf(err, "TableServer watch ITableManager Watch %s Failed", cTable.GetTableName()))
                                    continue
                                }
                            } else {
                                glog.Warnf("TableServer watch TableName %s cTable=%d NE cTableHash=%d", cTable.GetTableName(), cTable.GetTableCrc(), cTableHash.GetTableCrc())
                            }
                        }
                    }
                }
            }
        }
    }
}
