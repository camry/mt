package mt

import (
    "context"
    "fmt"
    "sync"

    "gitea.meigushi.go/meigushi/pb/cpb"
    "github.com/camry/g/v2/gerrors/gerror"
    "github.com/camry/g/v2/glog"
    "github.com/samber/lo"
    "go.etcd.io/etcd/client/v3"
    "google.golang.org/protobuf/encoding/protojson"
)

var tableMap = make(map[string]ITableManager, TableSizeDefault)

func Register(tableManager ITableManager) {
    tableMap[tableManager.TableName()] = tableManager
}

type TableServer struct {
    opts   *tableOptions
    ctx    context.Context
    cancel context.CancelFunc
}

func New(opts ...TableOption) *TableServer {
    o := &tableOptions{
        ctx:            context.Background(),
        tableKeyPrefix: TableKeyPrefixDefault,
    }
    for _, opt := range opts {
        opt(o)
    }
    ctx, cancel := context.WithCancel(o.ctx)
    return &TableServer{
        opts:   o,
        ctx:    ctx,
        cancel: cancel,
    }
}

func (x *TableServer) Start(_ context.Context) error {
    glog.Info("[TABLE] server starting")
    if x.opts.db == nil {
        return gerror.New("TableServer Start gorm DB is nil")
    }
    if x.opts.ec == nil {
        return gerror.New("TableServer Start etcd Client is nil")
    }
    var (
        cTables []*cpb.CTables
        wg      sync.WaitGroup
    )
    err := x.opts.db.Find(&cTables).Error
    if err != nil {
        return gerror.Wrap(err, "TableServer Start DB Find cpb.CTables Failed")
    }
    chunkCTables := lo.Chunk(cTables, 20)
    for _, cCTables := range chunkCTables {
        wg.Go(func() {
            for _, cTable := range cCTables {
                if lo.Contains(x.opts.tableNames, cTable.GetTableName()) {
                    continue
                }
                if iTableManager, ok := tableMap[cTable.GetTableName()]; ok {
                    err = iTableManager.Load(cTable.GetTableCrc(), cTable.GetTableData())
                    if err != nil {
                        glog.Error(gerror.Wrapf(err, "TableServer Start ITableManager Load %s Failed", cTable.GetTableName()))
                    }
                    glog.Debugf("TableServer Start ITableManager Load %s Success", cTable.GetTableName())
                }
            }
        })
    }
    wg.Go(x.watch)
    wg.Wait()
    return nil
}

func (x *TableServer) Stop(_ context.Context) error {
    glog.Info("[TABLE] server stopping")
    defer x.cancel()
    clear(tableMap)
    return nil
}

func (x *TableServer) watch() {
    for {
        select {
        case <-x.ctx.Done():
            glog.Debugf("TableServer watch context cancel")
            return
        case watchResponse := <-x.opts.ec.Watch(x.ctx, fmt.Sprintf("%s/%s", x.opts.tableKeyPrefix, TableKeyName)):
            for _, event := range watchResponse.Events {
                switch event.Type {
                case clientv3.EventTypePut:
                    cTableIndex := &cpb.CTableIndex{}
                    err := protojson.Unmarshal(event.Kv.Value, cTableIndex)
                    if err != nil {
                        glog.Error(gerror.Wrap(err, "TableServer watch protojson.Unmarshal Failed"))
                        continue
                    }
                    tableNameList := make([]string, 0, TableSizeDefault)
                    for _, cTableHash := range cTableIndex.GetTableHashMap() {
                        if lo.Contains(x.opts.tableNames, cTableHash.GetTableName()) {
                            continue
                        }
                        if iTableManager, ok := tableMap[cTableHash.GetTableName()]; ok {
                            if iTableManager.TableCrc() != cTableHash.GetTableCrc() {
                                tableNameList = append(tableNameList, cTableHash.GetTableName())
                            } else {
                                glog.Debugf("TableServer watch TableName %s cTable=%d EQ cTableIndexHash=%d", cTableHash.GetTableName(), cTableHash.GetTableCrc(), cTableHash.GetTableCrc())
                            }
                        }
                    }
                    cTables := make([]*cpb.CTables, 0, len(tableNameList))
                    if len(tableNameList) > 0 {
                        err = x.opts.db.Where("table_name IN ?", tableNameList).Find(&cTables).Error
                        if err != nil {
                            glog.Error(gerror.Wrap(err, "TableServer watch DB Find cbpb.CTables Failed"))
                            continue
                        }
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
                                glog.Debugf("TableServer watch ITableManager Watch %s Success", cTable.GetTableName())
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
