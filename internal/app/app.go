package app

import (
	"fmt"

	"github.com/M1ngdaXie/go-local-rag-email/internal/config"
	"github.com/M1ngdaXie/go-local-rag-email/internal/database"
	"github.com/M1ngdaXie/go-local-rag-email/pkg/logger"
	"gorm.io/gorm"
)

// App holds all application dependencies (Dependency Injection container)
type App struct {
	config *config.Config
	logger logger.Logger
	sqliteDB *gorm.DB

	// Services will be added here later (lazy-loaded)
	// emailService  email.Service
	// ragService    rag.Service
	// llmService    llm.Service
}

// New initializes the application with all dependencies
func New() (*App, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	log := logger.NewSlog(cfg.Logging.Level)
	log.Info("Application initializing", "name", cfg.App.Name)


	      db, err := database.NewSQLite(cfg.SQLite, log)
	      if err != nil {
	          return nil, fmt.Errorf("failed to initialize SQLite: %w", err)
	      }

	// Create app container
	app := &App{
		config: cfg,
		logger: log,
		sqliteDB: db,
	}

	log.Info("Application initialized successfully")
	return app, nil
}

// Config returns the application configuration
func (a *App) Config() *config.Config {
	return a.config
}

// Logger returns the application logger
func (a *App) Logger() logger.Logger {
	return a.logger
}

func (a *App) SQLiteDB() *gorm.DB {
    return a.sqliteDB
}

// Shutdown performs cleanup when the application exits
func (a *App) Shutdown() {
	a.logger.Info("Application shutting down")

	if a.sqliteDB != nil {
	          if sqlDB, err := a.sqliteDB.DB(); err == nil {
	              sqlDB.Close()
	          }
	      }

	a.logger.Info("Shutdown complete")
}
