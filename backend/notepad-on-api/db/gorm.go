package db

import (
    "log"
    "time"
    
    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "github.com/DanielGregorini/notepad-on/config"
	"github.com/DanielGregorini/notepad-on/model"
)

func Connect(cfg *config.Config) *gorm.DB {
    db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})

    if err != nil {
        log.Fatalf("falha ao conectar no banco: %v", err)
    }

    sqlDB, err := db.DB()

    if err != nil {
        log.Fatalf("falha ao obter sql.DB: %v", err)
    }

    sqlDB.SetMaxOpenConns(10)
    sqlDB.SetConnMaxIdleTime(time.Minute * 5)

    // migrations
    if err := db.AutoMigrate(&model.Page{}); err != nil {
        log.Fatalf("AutoMigrate falhou: %v", err)
    }

    return db
}
