package mt_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "gitlab.walrus.cs/treasure/treasure-pb/sbpb"
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

func TestNew(t *testing.T) {
    db, err := defaultDB()
    assert.Nil(t, err)
    x := mt.New(db, nil)
    err = x.Start(context.Background())
    assert.Nil(t, err)

    t.Log(tables.BaseLevelManager.TableName(), tables.BaseLevelManager.TableCrC())
    tables.BaseLevelManager.ForEach(func(index int, baseLevel *sbpb.BaseLevel) {
        t.Log(baseLevel)
    })

    baseLevel, err := tables.BaseLevelManager.GetByLevel(1)
    assert.Nil(t, err)
    t.Log(baseLevel)
    baseLevelCache, err := tables.BaseLevelManager.GetByLevel(1)
    assert.Nil(t, err)
    t.Log(baseLevelCache)

    err = x.Stop(context.Background())
    assert.Nil(t, err)
}
