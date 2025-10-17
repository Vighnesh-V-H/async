package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Vighnesh-V-H/async/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
    once sync.Once
    db     *gorm.DB
    sqlDB  *sql.DB
    initErr error
)

func InitDB(ctx context.Context, dsn string) (*gorm.DB, error) {
    once.Do(func() {
        var err error
        db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
            Logger: logger.Default.LogMode(logger.Warn),
        })
        if err != nil {
            initErr = fmt.Errorf("gorm open failed: %w", err)
            return
        }

        sqlDBContainer, err := db.DB()
        if err != nil {
            initErr = fmt.Errorf("get sql.DB failed: %w", err)
            return
        }
        sqlDB = sqlDBContainer

        sqlDB.SetMaxOpenConns(25)
        sqlDB.SetMaxIdleConns(10)
        sqlDB.SetConnMaxLifetime(5 * time.Minute)
        sqlDB.SetConnMaxIdleTime(5 * time.Minute)

        if err = db.WithContext(ctx).AutoMigrate(
            &models.Workflow{},
            &models.WorkflowInstance{},
            &models.Task{},
            &models.HistoryEntry{},
            &models.WorkflowRegistry{},
        ); err != nil {
            initErr = fmt.Errorf("migration failed: %w", err)
            return
        }

        log.Println("DB initialized and migrated successfully")
    })

    if initErr != nil {
        return nil, initErr
    }
    return db, nil
}

func GetDB(dsn string) *gorm.DB {
        db ,  err := InitDB(context.Background(), dsn);
		if err != nil {
            panic(fmt.Sprintf("Failed to init DB: %v", err))
        }

    return db
}

func CloseDB(ctx context.Context) error {
    if sqlDB == nil {
        return nil
    }

    if err := sqlDB.Close(); err != nil {
        return fmt.Errorf("close sql.DB failed: %w", err)
    }

    log.Println("DB connection pool closed")
    sqlDB = nil
    db = nil
    return nil
}

func DBStats() sql.DBStats {
    if sqlDB == nil {
        return sql.DBStats{}
    }
    return sqlDB.Stats()
}