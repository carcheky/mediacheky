package apitest

import (
	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// SeedTestServices creates test services in the database
func SeedTestServices(ta *TestApp) error {
	services := []models.Service{
		{
			Name:        "radarr",
			DisplayName: "Radarr",
			Enabled:     true,
			Status:      "running",
			ContainerID: "radarr-container-123",
			Config: models.ServiceConfig{
				"port":    "7878",
				"api_key": "test-radarr-key",
			},
		},
		{
			Name:        "sonarr",
			DisplayName: "Sonarr",
			Enabled:     true,
			Status:      "running",
			ContainerID: "sonarr-container-456",
			Config: models.ServiceConfig{
				"port":    "8989",
				"api_key": "test-sonarr-key",
			},
		},
		{
			Name:        "jellyfin",
			DisplayName: "Jellyfin",
			Enabled:     false,
			Status:      "stopped",
			ContainerID: "",
			Config: models.ServiceConfig{
				"port": "8096",
			},
		},
		{
			Name:        "prowlarr",
			DisplayName: "Prowlarr",
			Enabled:     true,
			Status:      "running",
			ContainerID: "prowlarr-container-789",
			Config: models.ServiceConfig{
				"port":    "9696",
				"api_key": "test-prowlarr-key",
			},
		},
		{
			Name:        "qbittorrent",
			DisplayName: "qBittorrent",
			Enabled:     true,
			Status:      "running",
			ContainerID: "qbittorrent-container-012",
			Config: models.ServiceConfig{
				"port":     "8080",
				"username": "admin",
				"password": "adminpass",
			},
		},
	}

	for _, service := range services {
		if err := ta.DB.Create(&service).Error; err != nil {
			return err
		}
	}

	return nil
}

// SeedTestGlobalConfig creates test global configuration
func SeedTestGlobalConfig(ta *TestApp) error {
	configs := []models.GlobalConfig{
		{Key: "PUID", Value: "1000", Category: "system"},
		{Key: "PGID", Value: "1000", Category: "system"},
		{Key: "TZ", Value: "Europe/Madrid", Category: "system"},
		{Key: "BASE_PATH", Value: "/media", Category: "paths"},
		{Key: "CONFIG_PATH", Value: "/config", Category: "paths"},
		{Key: "MEDIA_PATH", Value: "/media/movies", Category: "paths"},
		{Key: "DOWNLOAD_PATH", Value: "/downloads", Category: "paths"},
		{Key: "NETWORK_MODE", Value: "bridge", Category: "network"},
	}

	for _, cfg := range configs {
		// Update existing or create new
		var existing models.GlobalConfig
		if err := ta.DB.Where("key = ?", cfg.Key).First(&existing).Error; err == nil {
			// Update existing
			if err := ta.DB.Model(&existing).Updates(cfg).Error; err != nil {
				return err
			}
		} else {
			// Create new
			if err := ta.DB.Create(&cfg).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

// CreateTestService creates a single test service
func CreateTestService(ta *TestApp, name, displayName string, enabled bool) (*models.Service, error) {
	service := &models.Service{
		Name:        name,
		DisplayName: displayName,
		Enabled:     enabled,
		Status:      "running",
		Config:      models.ServiceConfig{},
	}

	if err := ta.DB.Create(service).Error; err != nil {
		return nil, err
	}

	return service, nil
}

// GetTestService retrieves a service by name
func GetTestService(ta *TestApp, name string) (*models.Service, error) {
	var service models.Service
	if err := ta.DB.Where("name = ?", name).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

// GetTestGlobalConfig retrieves a global config by key
func GetTestGlobalConfig(ta *TestApp, key string) (*models.GlobalConfig, error) {
	var config models.GlobalConfig
	if err := ta.DB.Where("key = ?", key).First(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// ClearTestServices removes all services from the test database
func ClearTestServices(ta *TestApp) error {
	return ta.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Service{}).Error
}

// ClearTestGlobalConfig removes all global config from the test database
func ClearTestGlobalConfig(ta *TestApp) error {
	return ta.DB.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.GlobalConfig{}).Error
}
