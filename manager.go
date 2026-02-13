package mt

import (
    "context"

    "github.com/camry/g/v2/gerrors/gerror"
    "go.etcd.io/etcd/client/v3"
    "gorm.io/gorm"

    "github.com/camry/mt/cpb"
)

var tableMap = make(map[string]ITableManager, 500)

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

func (x *Manager) Start(context.Context) error {
    var cTables []*cpb.CTables
    err := x.db.Find(&cTables).Error
    if err != nil {
        return gerror.Wrap(err, "Manager Start DB Find sbpb.CTables Failed")
    }
    for _, cTable := range cTables {
        if iTableManager, ok := tableMap[cTable.GetTableName()]; ok {
            err = iTableManager.Load(cTable.GetTableCrc(), cTable.GetTableData())
            if err != nil {
                return gerror.Wrapf(err, "Manager Start ITableManager Load %s Failed", cTable.GetTableName())
            }
        }
    }
    go x.watch()
    return nil
}

func (x *Manager) Stop(context.Context) error {
    tableMap = make(map[string]ITableManager, 500)
    return nil
}

func (x *Manager) watch() {
}
