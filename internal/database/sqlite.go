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
	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{Logger: gormLog})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil{
		return nil, fmt.Errorf("Error initializing db pool : %w", err)
	}
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

	err = db.AutoMigrate(
		&domain.Email{},
		&domain.Chunk{}, 
	)
	if err != nil {
		return nil, fmt.Errorf("auto-migration failed: %w", err)
	}

	log.Info("SQLite database connected", "path", cfg.Path)

	return db, nil
}
