package database

import (
	"fmt"

	"github.com/carcheky/mediacheky/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Config holds database configuration
type Config struct {
	Path string // Database file path
}

// Initialize creates and configures the database connection
func Initialize(cfg Config, log *logger.Logger) (*gorm.DB, error) {
	if cfg.Path == "" {
		cfg.Path = "./data/mediacheky.db"
	}

	log.Info("Initializing database", "path", cfg.Path)

	db, err := gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}

// Close closes the database connection
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	return sqlDB.Close()
}
