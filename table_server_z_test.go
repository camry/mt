package mt_test

import (
    "context"
    "testing"
    "time"

    "github.com/camry/dove/v2"
    "github.com/camry/g/v2/glog"
    "github.com/stretchr/testify/assert"
    clientv3 "go.etcd.io/etcd/client/v3"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "gitea.meigushi.go/meigushi/mt"
)

func defaultDB() (*gorm.DB, error) {
    db, err := gorm.Open(mysql.New(mysql.Config{
        DSN: "wild:skUbPzoif4K9uA3n@tcp(192.168.11.77:3306)/wild_conf?charset=utf8mb4&parseTime=True&loc=Local",
    }), &gorm.Config{
        SkipDefaultTransaction: true,
        DisableAutomaticPing:   true,
        Logger:                 logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, err
    }
    return db, nil
}

func defaultEtcdClient() (*clientv3.Client, error) {
    cli, err := clientv3.New(clientv3.Config{
        Username:    "wild",
        Password:    "bZbETE3bZP8RvT68",
        Endpoints:   []string{"192.168.11.77:2379", "192.168.11.77:2479"},
        DialTimeout: 5 * time.Second,
    })
    return cli, err
}

func TestTableServer_Start(t *testing.T) {
    db, err := defaultDB()
    assert.Nil(t, err)
    etcdClient, err := defaultEtcdClient()
    assert.Nil(t, err)
    tableServer := mt.New(
        mt.WithGORM(db),
        mt.WithEtcdClient(etcdClient),
        mt.WithTables("base_buff", "base_skill", "base_skill_level"),
        mt.WithAfterStart(func(ctx context.Context) error {
            glog.Debug("TableServer AfterStart")
            return nil
        }),
    )

    app := dove.New(
        dove.Name("dove"),
        dove.Version(dove.Release),
        dove.Context(context.Background()),
        dove.Server(tableServer),
        dove.BeforeStart(func(_ context.Context) error {
            glog.Debug("BeforeStart...")
            return nil
        }),
        dove.BeforeStop(func(_ context.Context) error {
            glog.Debug("BeforeStop...")
            return nil
        }),
        dove.AfterStart(func(_ context.Context) error {
            glog.Debug("AfterStart...")
            return nil
        }),
        dove.AfterStop(func(_ context.Context) error {
            glog.Debug("AfterStop...")
            return nil
        }),
    )
    time.AfterFunc(time.Second, func() {
        _ = app.Stop()
    })
    if err = app.Run(); err != nil {
        glog.Fatal(err)
    }
}
