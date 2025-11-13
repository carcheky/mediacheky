package database

import (
	"fmt"
	"log"

	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// RunMigrations executes all database migrations
func RunMigrations(db *gorm.DB) error {
	// Run AutoMigrate for all models
	if err := db.AutoMigrate(
		// Existing models
		&models.Media{},
		&models.Schedule{},
		&models.History{},
		&models.Settings{},
		// New service management models
		&models.Service{},
		&models.Template{},
		&models.GlobalConfig{},
		&models.ServiceLog{},
		// Proxy models
		&models.ProxyConfig{},
		&models.Domain{},
	); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Add performance indices for existing models
	if err := addPerformanceIndices(db); err != nil {
		return fmt.Errorf("failed to add performance indices: %w", err)
	}

	// Seed initial data
	if err := seedGlobalConfig(db); err != nil {
		return fmt.Errorf("failed to seed global config: %w", err)
	}

	if err := seedTemplates(db); err != nil {
		return fmt.Errorf("failed to seed templates: %w", err)
	}

	return nil
}

// addPerformanceIndices creates composite indices for frequently used queries
func addPerformanceIndices(db *gorm.DB) error {
	// Detect database dialect to use appropriate syntax
	dialectName := db.Dialector.Name()

	// PostgreSQL and SQLite both support partial indices with WHERE clause
	if dialectName == "sqlite" || dialectName == "postgres" {
		// Index for healthy files query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_healthy_files 
			ON media(in_jellyfin, in_radarr, in_sonarr, torrent_state) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create healthy files index: %w", err)
		}

		// Index for orphan downloads query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_orphan_downloads 
			ON media(in_q_bittorrent, in_jellyfin, in_radarr, in_sonarr) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create orphan downloads index: %w", err)
		}

		// Index for dead torrents query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_dead_torrents 
			ON media(in_q_bittorrent, torrent_state) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create dead torrents index: %w", err)
		}

		// Index for default sorting
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_default_sort 
			ON media(in_q_bittorrent DESC, in_jellyfin DESC, file_path ASC) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create default sort index: %w", err)
		}
	} else if dialectName == "mysql" {
		// MySQL doesn't support partial indices, create without WHERE clause
		// Note: Errors are logged but not returned for the following reasons:
		// 1. Indices may already exist from previous migrations
		// 2. Index creation is a performance optimization, not critical for functionality
		// 3. MySQL syntax variations across versions make error handling complex
		// If index creation fails, queries will still work but may be slower
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_media_healthy_files ON media(in_jellyfin, in_radarr, in_sonarr, torrent_state)`).Error; err != nil {
			log.Printf("Warning: Failed to create idx_media_healthy_files: %v", err)
		}
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_media_orphan_downloads ON media(in_q_bittorrent, in_jellyfin, in_radarr, in_sonarr)`).Error; err != nil {
			log.Printf("Warning: Failed to create idx_media_orphan_downloads: %v", err)
		}
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_media_dead_torrents ON media(in_q_bittorrent, torrent_state)`).Error; err != nil {
			log.Printf("Warning: Failed to create idx_media_dead_torrents: %v", err)
		}
		if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_media_default_sort ON media(in_q_bittorrent, in_jellyfin, file_path)`).Error; err != nil {
			log.Printf("Warning: Failed to create idx_media_default_sort: %v", err)
		}
	}

	return nil
}

// UpdateTemplates updates existing templates with the latest versions
// This is useful when templates are modified in the code
func UpdateTemplates(db *gorm.DB) error {
	return upsertTemplates(db)
}

// upsertTemplates creates or updates templates from the default templates
// DISABLED: Templates are now loaded from filesystem (./templates/*.yml)
// Database cache was causing stale template issues
func upsertTemplates(db *gorm.DB) error {
	log.Println("Template upsert disabled - using filesystem templates only")
	return nil
}

// seedGlobalConfig creates default global configuration values
func seedGlobalConfig(db *gorm.DB) error {
	defaultConfigs := []models.GlobalConfig{
		{Key: "PUID", Value: "1000", Category: "system"},
		{Key: "PGID", Value: "1000", Category: "system"},
		{Key: "TZ", Value: "UTC", Category: "system"},
		{Key: "BASE_PATH", Value: "/data", Category: "paths"},
		{Key: "CONFIG_PATH", Value: "/config", Category: "paths"},
		{Key: "MEDIA_PATH", Value: "/media", Category: "paths"},
		{Key: "DOWNLOAD_PATH", Value: "/downloads", Category: "paths"},
		{Key: "NETWORK_MODE", Value: "bridge", Category: "network"},
	}

	for _, config := range defaultConfigs {
		// Only create if doesn't exist
		var existing models.GlobalConfig
		result := db.Where("key = ?", config.Key).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&config).Error; err != nil {
				return fmt.Errorf("failed to create config %s: %w", config.Key, err)
			}
		}
	}

	return nil
}

// seedTemplates creates default service templates
// DISABLED: Templates are now loaded from filesystem (./templates/*.yml)
// Database cache was causing stale template issues
func seedTemplates(db *gorm.DB) error {
	log.Println("Template seeding disabled - using filesystem templates only")
	return nil
}

// seedServices creates default service records with their configurations
func seedServices(db *gorm.DB) error {
	// Services are now created via API when enabled
	// No need to pre-seed with default configurations
	log.Println("Service seeding skipped - services created on-demand")
	return nil
}
