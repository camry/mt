package mt

import (
    "context"

    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/camry/g/v2/glog"
    "github.com/camry/pb/cpb"
    "go.etcd.io/etcd/client/v3"
    "google.golang.org/protobuf/proto"
    "gorm.io/gorm"
)

var tableMap = make(map[string]ITableManager, TableSizeDefault)

func Register(tableManager ITableManager) {
    tableMap[tableManager.TableName()] = tableManager
}

type Manager struct {
    db *gorm.DB
    ec *clientv3.Client
}

func New(db *gorm.DB, ec *clientv3.Client) *Manager {
    return &Manager{
        db: db,
        ec: ec,
    }
}

func (x *Manager) Start(ctx context.Context) error {
    var cTables []*cpb.CTables
    err := x.db.Find(&cTables).Error
    if err != nil {
        return gerror.Wrap(err, "Manager Start DB Find cpb.CTables Failed")
    }
    for _, cTable := range cTables {
        if iTableManager, ok := tableMap[cTable.GetTableName()]; ok {
            err = iTableManager.Load(cTable.GetTableCrc(), cTable.GetTableData())
            if err != nil {
                return gerror.Wrapf(err, "Manager Start ITableManager Load %s Failed", cTable.GetTableName())
            }
        }
    }
    go x.watch(ctx)
    return nil
}

func (x *Manager) Stop(context.Context) error {
    tableMap = make(map[string]ITableManager, TableSizeDefault)
    glog.Debugf("Manager Stop")
    return nil
}

func (x *Manager) watch(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case watchResponse := <-x.ec.Watch(ctx, TableKeyDefault):
            for _, event := range watchResponse.Events {
                switch event.Type {
                case clientv3.EventTypePut:
                    cTableIndex := &cpb.CTableIndex{}
                    err := proto.Unmarshal(event.Kv.Value, cTableIndex)
                    if err != nil {
                        glog.Error(gerror.Wrap(err, "Manager watch proto.Unmarshal Failed"))
                        continue
                    }
                    tableNameList := make([]string, 0, TableSizeDefault)
                    for _, cTableHash := range cTableIndex.GetTableHashMap() {
                        if iTableManager, ok := tableMap[cTableHash.GetTableName()]; ok {
                            if iTableManager.TableCrc() != cTableHash.GetTableCrc() {
                                tableNameList = append(tableNameList, cTableHash.GetTableName())
                            } else {
                                glog.Debugf("Manager watch TableName %s cTable=%d EQ cTableIndexHash=%d", cTableHash.GetTableName(), cTableHash.GetTableCrc(), cTableHash.GetTableCrc())
                            }
                        }
                    }
                    var cTables []*cpb.CTables
                    err = x.db.Where("table_name IN ?", tableNameList).Find(&cTables).Error
                    if err != nil {
                        glog.Error(gerror.Wrap(err, "Manager watch DB Find cbpb.CTables Failed"))
                        continue
                    }
                    for _, cTable := range cTables {
                        iTableManager, ok1 := tableMap[cTable.GetTableName()]
                        cTableHash, ok2 := cTableIndex.GetTableHashMap()[cTable.GetTableName()]
                        if ok1 && ok2 {
                            if cTable.GetTableCrc() == cTableHash.GetTableCrc() {
                                err = iTableManager.Watch(cTable.GetTableCrc(), cTable.GetTableData())
                                if err != nil {
                                    glog.Error(gerror.Wrapf(err, "Manager watch ITableManager Watch %s Failed", cTable.GetTableName()))
                                    continue
                                }
                            } else {
                                glog.Warnf("Manager watch TableName %s cTable=%d NE cTableHash=%d", cTable.GetTableName(), cTable.GetTableCrc(), cTableHash.GetTableCrc())
                            }
                        }
                    }
                }
            }
        }
    }
}
