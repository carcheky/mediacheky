package service

import (
	"encoding/json"
	"fmt"

	"github.com/carcheky/mediacheky/internal/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ServiceManager handles multimedia service operations
type ServiceManager struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewServiceManager creates a new service manager
func NewServiceManager(db *gorm.DB, logger *zap.Logger) *ServiceManager {
	return &ServiceManager{
		db:     db,
		logger: logger,
	}
}

// InitializeDefaultServices creates default service configurations
func (sm *ServiceManager) InitializeDefaultServices() error {
	services := []models.Service{
		{
			Name:        "jellyfin",
			DisplayName: "Jellyfin",
			Image:       "jellyfin/jellyfin:latest",
			Port:        "8096",
			WebPort:     "8096",
			Enabled:     false,
		},
		{
			Name:        "sonarr",
			DisplayName: "Sonarr",
			Image:       "linuxserver/sonarr:latest",
			Port:        "8989",
			WebPort:     "8989",
			Enabled:     false,
		},
		{
			Name:        "radarr",
			DisplayName: "Radarr",
			Image:       "linuxserver/radarr:latest",
			Port:        "7878",
			WebPort:     "7878",
			Enabled:     false,
		},
		{
			Name:        "prowlarr",
			DisplayName: "Prowlarr",
			Image:       "linuxserver/prowlarr:latest",
			Port:        "9696",
			WebPort:     "9696",
			Enabled:     false,
		},
		{
			Name:        "qbittorrent",
			DisplayName: "qBittorrent",
			Image:       "linuxserver/qbittorrent:latest",
			Port:        "8080",
			WebPort:     "8080",
			Enabled:     false,
		},
		{
			Name:        "jellyseerr",
			DisplayName: "Jellyseerr",
			Image:       "fallenbagel/jellyseerr:latest",
			Port:        "5055",
			WebPort:     "5055",
			Enabled:     false,
		},
		{
			Name:        "qbitmanager",
			DisplayName: "qBit Manager",
			Image:       "bobokun/qbit_manage:latest",
			Port:        "",
			WebPort:     "",
			Enabled:     false,
		},
		{
			Name:        "dockercontroller",
			DisplayName: "Docker Controller Bot",
			Image:       "carcheky/docker-controller-bot:latest",
			Port:        "",
			WebPort:     "",
			Enabled:     false,
		},
		{
			Name:        "bazarr",
			DisplayName: "Bazarr",
			Image:       "linuxserver/bazarr:latest",
			Port:        "6767",
			WebPort:     "6767",
			Enabled:     false,
		},
		{
			Name:        "jellystat",
			DisplayName: "Jellystat",
			Image:       "cyfershepard/jellystat:latest",
			Port:        "3000",
			WebPort:     "3000",
			Enabled:     false,
		},
	}

	for _, service := range services {
		var existing models.Service
		result := sm.db.Where("name = ?", service.Name).First(&existing)
		if result.Error == gorm.ErrRecordNotFound {
			if err := sm.db.Create(&service).Error; err != nil {
				return fmt.Errorf("failed to create service %s: %w", service.Name, err)
			}
			sm.logger.Info("Created default service", zap.String("name", service.Name))
		}
	}

	return nil
}

// GetAllServices returns all configured services
func (sm *ServiceManager) GetAllServices() ([]models.Service, error) {
	var services []models.Service
	if err := sm.db.Find(&services).Error; err != nil {
		return nil, err
	}
	return services, nil
}

// GetService returns a specific service by name
func (sm *ServiceManager) GetService(name string) (*models.Service, error) {
	var service models.Service
	if err := sm.db.Where("name = ?", name).First(&service).Error; err != nil {
		return nil, err
	}
	return &service, nil
}

// UpdateService updates a service configuration
func (sm *ServiceManager) UpdateService(service *models.Service) error {
	return sm.db.Save(service).Error
}

// ToggleService enables or disables a service
func (sm *ServiceManager) ToggleService(name string, enabled bool) error {
	return sm.db.Model(&models.Service{}).Where("name = ?", name).Update("enabled", enabled).Error
}

// GenerateDockerCompose generates a docker-compose.yml configuration
func (sm *ServiceManager) GenerateDockerCompose() (string, error) {
	services, err := sm.GetAllServices()
	if err != nil {
		return "", err
	}

	compose := map[string]interface{}{
		"version": "3.8",
		"services": map[string]interface{}{},
		"networks": map[string]interface{}{
			"mediacheky": map[string]interface{}{
				"driver": "bridge",
			},
		},
		"volumes": map[string]interface{}{},
	}

	servicesMap := compose["services"].(map[string]interface{})
	volumesMap := compose["volumes"].(map[string]interface{})

	for _, service := range services {
		if !service.Enabled {
			continue
		}

		serviceConfig := map[string]interface{}{
			"image":          service.Image,
			"container_name": service.Name,
			"restart":        "unless-stopped",
			"networks":       []string{"mediacheky"},
		}

		// Add ports if service has web UI
		if service.WebPort != "" {
			serviceConfig["ports"] = []string{
				fmt.Sprintf("%s:%s", service.WebPort, service.Port),
			}
		}

		// Add environment variables
		if service.Environment != "" {
			var env map[string]string
			if err := json.Unmarshal([]byte(service.Environment), &env); err == nil && len(env) > 0 {
				serviceConfig["environment"] = env
			}
		}

		// Add volumes
		volumes := []string{
			fmt.Sprintf("./%s/config:/config", service.Name),
		}

		// Add common volumes based on service type
		switch service.Name {
		case "jellyfin":
			volumes = append(volumes,
				"./media:/media:ro",
				"./jellyfin/cache:/cache",
			)
		case "sonarr", "radarr":
			volumes = append(volumes,
				"./media:/media",
				"./downloads:/downloads",
			)
		case "qbittorrent":
			volumes = append(volumes,
				"./downloads:/downloads",
			)
		case "bazarr":
			volumes = append(volumes,
				"./media:/media",
			)
		}

		if service.Volumes != "" {
			var customVolumes []string
			if err := json.Unmarshal([]byte(service.Volumes), &customVolumes); err == nil {
				volumes = append(volumes, customVolumes...)
			}
		}

		serviceConfig["volumes"] = volumes

		// Add service-specific volume definitions
		volumeName := fmt.Sprintf("%s_config", service.Name)
		volumesMap[volumeName] = map[string]interface{}{}

		servicesMap[service.Name] = serviceConfig
	}

	// Convert to YAML-like string representation
	yamlBytes, err := json.MarshalIndent(compose, "", "  ")
	if err != nil {
		return "", err
	}

	return string(yamlBytes), nil
}
