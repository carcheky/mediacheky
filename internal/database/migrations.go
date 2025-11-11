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
func upsertTemplates(db *gorm.DB) error {
	defaultTemplates := []models.Template{
		{
			Name:    "radarr",
			Version: "1.0.0",
			Content: `version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Umask }}
      - UMASK={{ .Umask }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
      {{- if .Paths.Downloads }}
      - {{ .Paths.Downloads }}:/downloads
      {{- end }}
    ports:
      - "{{ .Port }}:7878"
    {{- if .Network }}
    networks:
      - {{ .Network }}
    {{- end }}
    restart: {{ .RestartPolicy }}
{{- if .Network }}

networks:
  {{ .Network }}:
    external: true
{{- end }}
`,
			Schema: models.JSONSchema{
				"type": "object",
				"properties": map[string]interface{}{
					"Port": map[string]interface{}{
						"type":    "integer",
						"default": 7878,
					},
				},
			},
		},
	}

	for _, template := range defaultTemplates {
		// Find existing template
		var existing models.Template
		result := db.Where("name = ?", template.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new template
			if err := db.Create(&template).Error; err != nil {
				return fmt.Errorf("failed to create template %s: %w", template.Name, err)
			}
			log.Printf("Created template: %s", template.Name)
		} else if result.Error != nil {
			return fmt.Errorf("error checking template %s: %w", template.Name, result.Error)
		} else {
			// Update existing template
			existing.Content = template.Content
			existing.Version = template.Version
			existing.Schema = template.Schema
			if err := db.Save(&existing).Error; err != nil {
				return fmt.Errorf("failed to update template %s: %w", template.Name, err)
			}
			log.Printf("Updated template: %s", template.Name)
		}
	}

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
func seedTemplates(db *gorm.DB) error {
	defaultTemplates := []models.Template{
		{
			Name:    "radarr",
			Version: "2.0.0",
			Content: `version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Umask }}
      - UMASK={{ .Umask }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
      {{- if .Paths.Downloads }}
      - {{ .Paths.Downloads }}:/downloads
      {{- end }}
    {{- if .ExposePort }}
    ports:
      - "{{ if .HostPort }}{{ .HostPort }}{{ else }}{{ .Port }}{{ end }}:7878"
    {{- end }}
    networks:
      - mediacheky-net
    restart: {{ .RestartPolicy }}

networks:
  mediacheky-net:
    external: true
`,
			Schema: models.JSONSchema{
				"type": "object",
				"properties": map[string]interface{}{
					"Port": map[string]interface{}{
						"type":    "integer",
						"default": 7878,
					},
					"ExposePort": map[string]interface{}{
						"type":    "boolean",
						"default": false,
					},
					"HostPort": map[string]interface{}{
						"type":    "integer",
						"default": 0,
					},
				},
			},
		},
		{
			Name:    "sonarr",
			Version: "1.0.0",
			Content: `version: '3.8'
services:
  sonarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Umask }}
      - UMASK={{ .Umask }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.TV }}:/tv
      {{- if .Paths.Downloads }}
      - {{ .Paths.Downloads }}:/downloads
      {{- end }}
    ports:
      - "{{ .Port }}:8989"
    {{- if .Network }}
    networks:
      - {{ .Network }}
    {{- end }}
    restart: {{ .RestartPolicy }}
{{- if .Network }}

networks:
  {{ .Network }}:
    external: true
{{- end }}
`,
			Schema: models.JSONSchema{
				"type": "object",
				"properties": map[string]interface{}{
					"Port": map[string]interface{}{
						"type":    "integer",
						"default": 8989,
					},
				},
			},
		},
		{
			Name:    "jellyfin",
			Version: "1.0.0",
			Content: `version: '3.8'
services:
  jellyfin:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Umask }}
      - UMASK={{ .Umask }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Media }}:/media
    ports:
      - "{{ .Port }}:8096"
    {{- if .Network }}
    networks:
      - {{ .Network }}
    {{- end }}
    restart: {{ .RestartPolicy }}
{{- if .Network }}

networks:
  {{ .Network }}:
    external: true
{{- end }}
`,
			Schema: models.JSONSchema{
				"type": "object",
				"properties": map[string]interface{}{
					"Port": map[string]interface{}{
						"type":    "integer",
						"default": 8096,
					},
				},
			},
		},
	}

	for _, template := range defaultTemplates {
		var existing models.Template
		result := db.Where("name = ?", template.Name).First(&existing)

		if result.Error == gorm.ErrRecordNotFound {
			// Create new template
			if err := db.Create(&template).Error; err != nil {
				return fmt.Errorf("failed to create template %s: %w", template.Name, err)
			}
			log.Printf("Created new template: %s v%s", template.Name, template.Version)
		} else if result.Error == nil {
			// Update if version changed
			if existing.Version != template.Version {
				existing.Version = template.Version
				existing.Content = template.Content
				existing.Schema = template.Schema
				if err := db.Save(&existing).Error; err != nil {
					return fmt.Errorf("failed to update template %s: %w", template.Name, err)
				}
				log.Printf("Updated template: %s from v%s to v%s", template.Name, existing.Version, template.Version)
			}
		} else {
			return fmt.Errorf("error checking template %s: %w", template.Name, result.Error)
		}
	}

	return nil
}

// seedServices creates default service records with their configurations
func seedServices(db *gorm.DB) error {
	// Get the radarr template
	var radarrTemplate models.Template
	if err := db.Where("name = ?", "radarr").First(&radarrTemplate).Error; err != nil {
		return fmt.Errorf("Radarr template not found: %w", err)
	}

	// Define default services
	defaultServices := []models.Service{
		{
			Name:        "radarr",
			DisplayName: "Radarr",
			Icon:        "🎬",
			Enabled:     false,
			Status:      "stopped",
			Image:       "linuxserver/radarr:latest",
			Port:        7878,
			TemplateID:  radarrTemplate.ID,
			Config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:latest",
				"ContainerName": "radarr",
				"Port":          7878,
				"ExposePort":    false, // CRITICAL: Do NOT expose ports by default
				"HostPort":      0,     // 0 means use service Port
				"Paths": map[string]interface{}{
					"Config":    "/data/config/radarr",
					"Movies":    "/data/media/movies",
					"Downloads": "/data/downloads",
				},
				"RestartPolicy": "unless-stopped",
			},
		},
	}

	for _, service := range defaultServices {
		// Only create if doesn't exist
		var existing models.Service
		result := db.Where("name = ?", service.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := db.Create(&service).Error; err != nil {
				return fmt.Errorf("failed to create service %s: %w", service.Name, err)
			}
		}
	}

	return nil
}
