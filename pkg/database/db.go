package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/Vighnesh-V-H/async/internal/logger"
	"github.com/Vighnesh-V-H/async/internal/models"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
    once sync.Once
    db     *gorm.DB
    sqlDB  *sql.DB
    initErr error
    dbLog  zerolog.Logger
)

func InitDB(ctx context.Context, connection string, logCfg logger.Config) (*gorm.DB, error) {
    once.Do(func() {
        dbLog = logger.New(logCfg)
        dbLog.Info().Msg("Initializing database connection")
        
        var err error
        db, err = gorm.Open(postgres.Open(connection), &gorm.Config{
            Logger: gormlogger.Default.LogMode(gormlogger.Warn),
        })
        if err != nil {
            initErr = fmt.Errorf("gorm open failed: %w", err)
            dbLog.Error().Err(err).Msg("Failed to open database connection")
            return
        }

        sqlDBContainer, err := db.DB()
        if err != nil {
            initErr = fmt.Errorf("get sql.DB failed: %w", err)
            dbLog.Error().Err(err).Msg("Failed to get underlying sql.DB")
            return
        }
        sqlDB = sqlDBContainer

        sqlDB.SetMaxOpenConns(25)
        sqlDB.SetMaxIdleConns(10)
        sqlDB.SetConnMaxLifetime(5 * time.Minute)
        sqlDB.SetConnMaxIdleTime(5 * time.Minute)
        
        dbLog.Info().
            Int("max_open_conns", 25).
            Int("max_idle_conns", 10).
            Dur("conn_max_lifetime", 5*time.Minute).
            Msg("Database connection pool configured")

        dbLog.Info().Msg("Starting database migration")
        
        if err = db.WithContext(ctx).AutoMigrate(
            &models.Workflow{},
            &models.WorkflowInstance{},
            &models.Task{},
            &models.HistoryEntry{},
            &models.WorkflowRegistry{},
        ); err != nil {
            initErr = fmt.Errorf("migration failed: %w", err)
            dbLog.Error().Err(err).Msg("Database migration failed")
            return
        }

        dbLog.Info().
            Strs("models", []string{"Workflow", "WorkflowInstance", "Task", "HistoryEntry", "WorkflowRegistry"}).
            Msg("Database initialized and migrated successfully")
    })

    if initErr != nil {
        return nil, initErr
    }
    return db, nil
}

func GetDB(connection string, logCfg logger.Config) *gorm.DB {
        db ,  err := InitDB(context.Background(), connection, logCfg)
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

    dbLog.Info().Msg("DB connection pool closed")
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