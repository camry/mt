package mt

import (
    "context"

    clientv3 "go.etcd.io/etcd/client/v3"
    "gorm.io/gorm"
)

type TableOption func(x *tableOptions)

type tableOptions struct {
    db             *gorm.DB
    ec             *clientv3.Client
    ctx            context.Context
    tableKeyPrefix string
    tableNames     []string
}

func WithGORM(db *gorm.DB) func(x *tableOptions) {
    return func(x *tableOptions) { x.db = db }
}

func WithEtcdClient(ec *clientv3.Client) func(x *tableOptions) {
    return func(x *tableOptions) { x.ec = ec }
}

func WithContext(ctx context.Context) func(x *tableOptions) {
    return func(x *tableOptions) { x.ctx = ctx }
}

func WithTableKeyPrefix(tableKeyPrefix string) func(x *tableOptions) {
    return func(x *tableOptions) { x.tableKeyPrefix = tableKeyPrefix }
}

func WithTables(tableNames ...string) func(x *tableOptions) {
    return func(x *tableOptions) { x.tableNames = tableNames }
}
