package mt_test

import (
    "context"
    "log"
    "testing"
    "time"

    "github.com/camry/dove/v2"
    "github.com/camry/g/v2/glog"
    "github.com/camry/pb/cpb"
    "github.com/stretchr/testify/assert"
    "gitlab.walrus.cs/treasure/treasure-pb/sbpb"
    clientv3 "go.etcd.io/etcd/client/v3"
    "google.golang.org/protobuf/proto"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/camry/mt"
    "github.com/camry/mt/tables"
    _ "github.com/camry/mt/tables"
)

func defaultDB() (*gorm.DB, error) {
    db, err := gorm.Open(mysql.New(mysql.Config{
        DSN: "dev:Zqa5Q5pYgvmT@tcp(192.168.110.112:3306)/treasure_base?timeout=10s&parseTime=true&charset=utf8mb4",
    }), &gorm.Config{
        SkipDefaultTransaction: true,
        DisableAutomaticPing:   true,
        Logger:                 logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, err
    }
    return db, nil
}

func defaultEtcdClient() (*clientv3.Client, error) {
    cli, err := clientv3.New(clientv3.Config{
        Endpoints:   []string{"192.168.110.5:2379"},
        DialTimeout: 5 * time.Second,
    })
    return cli, err
}

func TestNew(t *testing.T) {
    db, err := defaultDB()
    assert.Nil(t, err)
    etcdClient, err := defaultEtcdClient()
    assert.Nil(t, err)
    x := mt.New(db, etcdClient)
    err = x.Start(context.Background())
    assert.Nil(t, err)

    t.Log(tables.BaseLevelManager.TableName(), tables.BaseLevelManager.TableCrc())
    tables.BaseLevelManager.ForEach(func(index int, baseLevel *sbpb.BaseLevel) {
        t.Log(baseLevel)
    })
    baseLevel, err := tables.BaseLevelManager.GetByLevel(1)
    assert.Nil(t, err)
    t.Log(baseLevel)
    baseLevelCache, err := tables.BaseLevelManager.GetByLevel(1)
    assert.Nil(t, err)
    t.Log(baseLevelCache)

    time.Sleep(time.Second * 8)

    cTableIndex := &cpb.CTableIndex{}
    cTableIndex.TableHashMap = map[string]*cpb.CTableHash{
        "base_level": {
            TableName: "base_level",
            TableCrc:  495674361,
        },
    }
    cTableIndexBlob, err := proto.Marshal(cTableIndex)
    assert.Nil(t, err)
    _, err = etcdClient.Put(context.Background(), mt.TableKeyDefault, string(cTableIndexBlob))
    assert.Nil(t, err)

    time.Sleep(time.Second * 3)

    err = x.Stop(context.Background())
    assert.Nil(t, err)
}

func TestManager_Start(t *testing.T) {
    db, err := defaultDB()
    assert.Nil(t, err)
    etcdClient, err := defaultEtcdClient()
    assert.Nil(t, err)
    mtServer := mt.New(db, etcdClient)

    app := dove.New(
        dove.Name("dove"),
        dove.Version(dove.Release),
        dove.Context(context.Background()),
        dove.Server(mtServer),
        dove.BeforeStart(func(_ context.Context) error {
            log.Println("BeforeStart...")
            return nil
        }),
        dove.BeforeStop(func(_ context.Context) error {
            log.Println("BeforeStop...")
            return nil
        }),
        dove.AfterStart(func(_ context.Context) error {
            log.Println("AfterStart...")
            return nil
        }),
        dove.AfterStop(func(_ context.Context) error {
            log.Println("AfterStop...")
            return nil
        }),
    )
    time.AfterFunc(time.Second, func() {
        _ = app.Stop()
    })
    if err := app.Run(); err != nil {
        glog.Fatal(err)
    }
}
