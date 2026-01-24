package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/M1ngdaXie/go-local-rag-email/internal/config"
	"github.com/M1ngdaXie/go-local-rag-email/internal/domain"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewSQLite creates a new SQLite database connection
func NewSQLite(cfg config.SQLiteConfig, log logger.Logger) (*gorm.DB, error) {
	gormLog := gormlogger.Default.LogMode(gormlogger.Silent)
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}
	// TODO: Step 2 - Open the database connection
	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{Logger: gormLog})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// TODO: Step 3 - Get the underlying SQL DB for connection pool settings
	sqlDB, err := db.DB()
	if err != nil{
		return nil, fmt.Errorf("Error initializing db pool : %w", err)
	}
	// TODO: Step 4 - Set connection pool settings from config
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if cfg.EnableWAL {
		if err := db.Exec("PRAGMA journal_mode=WAL;").Error; err != nil {
			log.Warn("Failed to enable WAL mode", "error", err)
		}
	}

	// Step 6: Enable foreign keys
	// SQLite 默认是不强制外键约束的，需要手动开启
	if cfg.EnableForeignKeys {
		if err := db.Exec("PRAGMA foreign_keys=ON;").Error; err != nil {
			log.Warn("Failed to enable foreign keys", "error", err)
		}
	}

	// TODO: Step 7 - Auto-migrate all domain models
	err = db.AutoMigrate(
		&domain.Email{},
		// &domain.Chunk{}, // TODO: Uncomment when Chunk model is created
	)
	if err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	// TODO: Step 8 - Log success message
	log.Info("SQLite database connected", "path", cfg.Path)

	// TODO: Step 9 - Return the database connection
	return db, nil
}
