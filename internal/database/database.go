package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/carcheky/mediacheky/pkg/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
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

	// Ensure the directory exists
	dir := filepath.Dir(cfg.Path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Configure GORM logger to suppress "record not found" errors
	gormConfig := &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	}

	db, err := gorm.Open(sqlite.Open(cfg.Path), gormConfig)
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

// SeedData seeds initial data (global config and templates)
func SeedData(db *gorm.DB) error {
	// Always update templates to ensure they have the latest version
	if err := UpdateTemplates(db); err != nil {
		return fmt.Errorf("failed to update templates: %w", err)
	}

	// Seed global config if not exists
	if err := seedGlobalConfig(db); err != nil {
		return fmt.Errorf("failed to seed global config: %w", err)
	}

	return nil
}
