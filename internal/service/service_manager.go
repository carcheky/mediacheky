package service

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"go.uber.org/zap"
)

const (
	// MediaChekyNetwork is the hardcoded network name that ALL services MUST use
	// This is the Docker Compose prefixed name (project_networkname)
	MediaChekyNetwork = "mediacheky_mediacheky-net"
)

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	GetByName(name string) (*models.Service, error)
	Create(service *models.Service) error
	Update(service *models.Service) error
	UpdateStatus(id uint, status string, containerID string) error
	SetEnabled(id uint, enabled bool) error
}

// ServiceLogRepository defines the interface for service log data access
type ServiceLogRepository interface {
	Create(log *models.ServiceLog) error
}

// ServiceManager orchestrates service lifecycle operations
type ServiceManager struct {
	logger         *zap.Logger
	templateEngine *TemplateEngine
	dockerCompose  *DockerComposeClient
	dockerClient   *DockerClient
	serviceRepo    ServiceRepository
	logRepo        ServiceLogRepository
}

// NewServiceManager creates a new ServiceManager instance
func NewServiceManager(
	logger *zap.Logger,
	templateEngine *TemplateEngine,
	dockerCompose *DockerComposeClient,
	dockerClient *DockerClient,
	serviceRepo ServiceRepository,
	logRepo ServiceLogRepository,
) *ServiceManager {
	return &ServiceManager{
		logger:         logger,
		templateEngine: templateEngine,
		dockerCompose:  dockerCompose,
		dockerClient:   dockerClient,
		serviceRepo:    serviceRepo,
		logRepo:        logRepo,
	}
}

// EnableService enables a service and generates its docker-compose file
func (sm *ServiceManager) EnableService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Enabling service", zap.String("service", serviceName))

	// Get service from database, create if doesn't exist
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		if err.Error() == "record not found" || err.Error() == "failed to get service: record not found" {
			// Create service with default configuration
			sm.logger.Info("Service not found, creating with defaults", zap.String("service", serviceName))
			defaultConfig := map[string]interface{}{
				"Image":         fmt.Sprintf("linuxserver/%s:latest", serviceName),
				"ContainerName": serviceName,
				"Paths": map[string]string{
					"Config": fmt.Sprintf("/app/data/services/%s/config", serviceName),
				},
				"RestartPolicy": "unless-stopped",
			}

			svc = &models.Service{
				Name:    serviceName,
				Enabled: false,
				Status:  "stopped",
				Config:  defaultConfig,
			}

			if err := sm.serviceRepo.Create(svc); err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			sm.logger.Info("Service created successfully", zap.String("service", serviceName))
		} else {
			return fmt.Errorf("failed to get service: %w", err)
		}
	}

	if svc.Enabled {
		sm.logger.Info("Service already enabled", zap.String("service", serviceName))
		return nil
	}

	// Generate docker-compose file
	composePath, err := sm.templateEngine.GenerateCompose(serviceName, svc.Config)
	if err != nil {
		sm.logAction(svc.ID, "enable", "error", fmt.Sprintf("Failed to generate compose: %v", err))
		return fmt.Errorf("failed to generate compose file: %w", err)
	}

	// Update service state
	if err := sm.serviceRepo.SetEnabled(svc.ID, true); err != nil {
		sm.logAction(svc.ID, "enable", "error", fmt.Sprintf("Failed to update state: %v", err))
		return fmt.Errorf("failed to enable service: %w", err)
	}

	sm.logAction(svc.ID, "enable", "success", fmt.Sprintf("Service enabled, compose generated at %s", composePath))
	sm.logger.Info("Service enabled successfully", zap.String("service", serviceName))

	// Start the service automatically with default values
	sm.logger.Info("Starting service automatically after enable", zap.String("service", serviceName))
	if err := sm.StartService(ctx, serviceName); err != nil {
		sm.logger.Error("Failed to auto-start service after enable",
			zap.String("service", serviceName),
			zap.Error(err))
		// Don't fail the enable operation, just log the error
		sm.logAction(svc.ID, "enable", "warning", fmt.Sprintf("Service enabled but failed to start: %v", err))
	} else {
		sm.logger.Info("Service auto-started successfully", zap.String("service", serviceName))
	}

	return nil
}

// DisableService disables a service and stops its container if running
func (sm *ServiceManager) DisableService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Disabling service", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !svc.Enabled {
		sm.logger.Info("Service already disabled", zap.String("service", serviceName))
		return nil
	}

	// Always try to stop the service when disabling
	// This ensures we stop the container even if the DB state is not accurate
	sm.logger.Info("Stopping service before disabling", zap.String("service", serviceName))
	if err := sm.StopService(ctx, serviceName); err != nil {
		sm.logger.Warn("Failed to stop container during disable",
			zap.String("service", serviceName),
			zap.Error(err))
		// Continue with disable even if stop fails
	}

	// Update service state
	if err := sm.serviceRepo.SetEnabled(svc.ID, false); err != nil {
		sm.logAction(svc.ID, "disable", "error", fmt.Sprintf("Failed to update state: %v", err))
		return fmt.Errorf("failed to disable service: %w", err)
	}

	sm.logAction(svc.ID, "disable", "success", "Service disabled")
	sm.logger.Info("Service disabled successfully", zap.String("service", serviceName))

	return nil
}

// StartService starts a service using docker compose
func (sm *ServiceManager) StartService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Starting service", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !svc.Enabled {
		return fmt.Errorf("service is not enabled")
	}

	// Get compose file path
	composePath := sm.templateEngine.GetComposePath(serviceName)

	// Load global config for volume creation and docker compose execution
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		sm.logAction(svc.ID, "start", "error", fmt.Sprintf("Failed to load global config: %v", err))
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Ensure all volume directories exist before starting
	if err := sm.ensureVolumesExist(composePath); err != nil {
		sm.logAction(svc.ID, "start", "error", fmt.Sprintf("Failed to create volumes: %v", err))
		return fmt.Errorf("failed to ensure volumes exist: %w", err)
	}

	// Execute docker compose up with config variables
	result, err := sm.dockerCompose.ComposeUp(ctx, composePath, globalConfig)
	if err != nil {
		sm.logAction(svc.ID, "start", "error", fmt.Sprintf("Failed to start: %v", err))
		return fmt.Errorf("failed to start service: %w", err)
	}

	// Extract container ID from compose result and update status
	// For now, we'll use the container name to find the container
	containerName, _ := svc.Config["ContainerName"].(string)
	if containerName == "" {
		containerName = serviceName
	}

	// Poll for container to be created, up to 10 seconds
	var containerID string
	const pollTimeout = 10 * time.Second
	const pollInterval = 200 * time.Millisecond
	start := time.Now()
	for time.Since(start) < pollTimeout {
		var pollErr error
		containerID, pollErr = sm.findContainerByName(ctx, containerName)
		if pollErr == nil && containerID != "" {
			break
		}
		time.Sleep(pollInterval)
	}
	if containerID == "" {
		sm.logger.Warn("Could not find container after start",
			zap.String("service", serviceName),
			zap.String("container_name", containerName))
	}

	// Update service status
	if err := sm.serviceRepo.UpdateStatus(svc.ID, "running", containerID); err != nil {
		sm.logger.Error("Failed to update service status",
			zap.String("service", serviceName),
			zap.Error(err))
	}

	sm.logAction(svc.ID, "start", "success", fmt.Sprintf("Service started: %s", result.Output))
	sm.logger.Info("Service started successfully", zap.String("service", serviceName))

	return nil
}

// StopService stops a service using docker compose
func (sm *ServiceManager) StopService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Stopping service", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Get compose file path
	composePath := sm.templateEngine.GetComposePath(serviceName)

	// Load global config for docker compose execution
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		sm.logAction(svc.ID, "stop", "error", fmt.Sprintf("Failed to load global config: %v", err))
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Execute docker compose stop (preserves container for auto-restart)
	// Use stop instead of down to allow Docker restart policy to work
	result, err := sm.dockerCompose.ComposeStop(ctx, composePath, globalConfig)
	if err != nil {
		sm.logAction(svc.ID, "stop", "error", fmt.Sprintf("Failed to stop: %v", err))
		return fmt.Errorf("failed to stop service: %w", err)
	}

	// Update service status
	if err := sm.serviceRepo.UpdateStatus(svc.ID, "stopped", ""); err != nil {
		sm.logger.Error("Failed to update service status",
			zap.String("service", serviceName),
			zap.Error(err))
	}

	sm.logAction(svc.ID, "stop", "success", fmt.Sprintf("Service stopped: %s", result.Output))
	sm.logger.Info("Service stopped successfully", zap.String("service", serviceName))

	return nil
}

// RestartService restarts a service
func (sm *ServiceManager) RestartService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Restarting service", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Stop then start
	if err := sm.StopService(ctx, serviceName); err != nil {
		sm.logAction(svc.ID, "restart", "error", fmt.Sprintf("Failed to stop during restart: %v", err))
		return fmt.Errorf("failed to stop service during restart: %w", err)
	}

	if err := sm.StartService(ctx, serviceName); err != nil {
		sm.logAction(svc.ID, "restart", "error", fmt.Sprintf("Failed to start during restart: %v", err))
		return fmt.Errorf("failed to start service during restart: %w", err)
	}

	sm.logAction(svc.ID, "restart", "success", "Service restarted")
	sm.logger.Info("Service restarted successfully", zap.String("service", serviceName))

	return nil
}

// UpdateService updates a service by pulling the latest image and restarting
func (sm *ServiceManager) UpdateService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Updating service (pull + restart)", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	if !svc.Enabled {
		return fmt.Errorf("service is not enabled")
	}

	// Get compose file path
	composePath := sm.templateEngine.GetComposePath(serviceName)

	// Load global config for docker compose execution
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		sm.logAction(svc.ID, "update", "error", fmt.Sprintf("Failed to load global config: %v", err))
		return fmt.Errorf("failed to load global config: %w", err)
	}

	sm.logger.Info("Pulling latest image", zap.String("service", serviceName))
	sm.logAction(svc.ID, "update", "info", "Pulling latest image...")

	// Pull the latest image using docker compose
	result, err := sm.dockerCompose.ComposePull(ctx, composePath, globalConfig)
	if err != nil {
		sm.logAction(svc.ID, "update", "error", fmt.Sprintf("Failed to pull image: %v", err))
		return fmt.Errorf("failed to pull image: %w", err)
	}

	sm.logger.Info("Image pulled successfully",
		zap.String("service", serviceName),
		zap.String("output", result.Output))
	sm.logAction(svc.ID, "update", "info", "Image pulled successfully, restarting service...")

	// Restart the service to use the new image
	if err := sm.RestartService(ctx, serviceName); err != nil {
		sm.logAction(svc.ID, "update", "error", fmt.Sprintf("Image pulled but failed to restart: %v", err))
		return fmt.Errorf("image pulled but failed to restart service: %w", err)
	}

	sm.logAction(svc.ID, "update", "success", "Service updated and restarted with latest image")
	sm.logger.Info("Service updated successfully", zap.String("service", serviceName))

	return nil
}

// UpdateServiceConfig updates a service's configuration and regenerates compose file if enabled
func (sm *ServiceManager) UpdateServiceConfig(ctx context.Context, serviceName string, config models.ServiceConfig) error {
	sm.logger.Info("Updating service configuration", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Keep previous port exposure values to decide if restart is needed
	prevExpose, _ := svc.Config["ExposePort"].(bool)
	prevHostPortFloat, hostFloatOk := svc.Config["HostPort"].(float64)
	prevHostPortInt, hostIntOk := svc.Config["HostPort"].(int)
	prevHostPort := 0
	if hostFloatOk {
		prevHostPort = int(prevHostPortFloat)
	} else if hostIntOk {
		prevHostPort = prevHostPortInt
	}

	// If HostPort is explicitly set to nil, remove it from the saved config
	// This allows switching from dynamic (with port) to static (without port) compose
	if hostPortValue, exists := config["HostPort"]; exists && hostPortValue == nil {
		delete(config, "HostPort")
		sm.logger.Info("Removing HostPort from configuration (switching to static compose)",
			zap.String("service", serviceName))
	}

	// Update configuration in memory
	svc.Config = config

	// Debug: Log what we're saving
	sm.logger.Info("Saving configuration to database",
		zap.String("service", serviceName),
		zap.Any("config", config),
		zap.Any("config_paths", config["Paths"]))

	// Extract port and image for quick access
	if port, ok := config["Port"].(float64); ok {
		svc.Port = int(port)
	} else if port, ok := config["Port"].(int); ok {
		svc.Port = port
	}

	if image, ok := config["Image"].(string); ok {
		svc.Image = image
	}

	// Save to database
	if err := sm.serviceRepo.Update(svc); err != nil {
		sm.logAction(svc.ID, "config_update", "error", fmt.Sprintf("Failed to save config: %v", err))
		return fmt.Errorf("failed to update service: %w", err)
	}

	// Regenerate compose file if service is enabled
	if svc.Enabled {
		if _, err := sm.templateEngine.GenerateCompose(serviceName, config); err != nil {
			sm.logAction(svc.ID, "config_update", "warning", fmt.Sprintf("Config saved but compose regeneration failed: %v", err))
			return fmt.Errorf("config saved but failed to regenerate compose file: %w", err)
		}

		// Determine new exposure values
		newExpose, _ := config["ExposePort"].(bool)
		newHostPort := 0
		if hpFloat, ok := config["HostPort"].(float64); ok {
			newHostPort = int(hpFloat)
		} else if hpInt, ok := config["HostPort"].(int); ok {
			newHostPort = hpInt
		}

		// If exposure settings changed, restart to apply port mapping
		if prevExpose != newExpose || prevHostPort != newHostPort {
			sm.logger.Info("Port exposure settings changed, recreating service to apply", zap.String("service", serviceName))

			// Recreate service instead of restart (avoid docker compose down issues)
			composePath := sm.templateEngine.GetComposePath(serviceName)
			globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
			if err != nil {
				sm.logAction(svc.ID, "config_update", "warning", fmt.Sprintf("Compose regenerated but failed to load config: %v", err))
				return fmt.Errorf("compose regenerated but failed to load global config: %w", err)
			}

			// Force recreate the container with new settings
			result, err := sm.dockerCompose.ComposeUpRecreate(ctx, composePath, globalConfig, serviceName)
			if err != nil {
				errMsg := fmt.Sprintf("Compose regenerated but recreate failed: %v | Output: %s | Error: %s",
					err, result.Output, result.Error)
				sm.logger.Error("Failed to recreate service",
					zap.String("service", serviceName),
					zap.String("output", result.Output),
					zap.String("error", result.Error),
					zap.Error(err))
				sm.logAction(svc.ID, "config_update", "warning", errMsg)
				return fmt.Errorf("compose regenerated but failed to recreate service: %s", errMsg)
			}

			sm.logAction(svc.ID, "config_update", "success", fmt.Sprintf("Configuration updated and service recreated (port changes applied): %s", result.Output))
		} else {
			sm.logAction(svc.ID, "config_update", "success", "Configuration updated and compose file regenerated")
		}
	} else {
		sm.logAction(svc.ID, "config_update", "success", "Configuration updated")
	}

	sm.logger.Info("Service configuration updated", zap.String("service", serviceName))

	return nil
}

// findContainerByName finds a container ID by its name
func (sm *ServiceManager) findContainerByName(ctx context.Context, containerName string) (string, error) {
	containers, err := sm.dockerClient.ListContainers(ctx)
	if err != nil {
		return "", err
	}

	for _, container := range containers {
		// Docker prefixes names with /, so strip it
		name := container.Name
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		if name == containerName {
			return container.ID, nil
		}
	}

	return "", fmt.Errorf("container not found: %s", containerName)
}

// ensureVolumesExist creates volume directories if they don't exist
func (sm *ServiceManager) ensureVolumesExist(composePath string) error {
	// Load global config to get path variables
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Parse compose file to extract volume paths
	volumeDirs, err := sm.templateEngine.ExtractVolumePaths(composePath, globalConfig)
	if err != nil {
		return fmt.Errorf("failed to extract volume paths: %w", err)
	}

	// Create each volume directory
	for _, dir := range volumeDirs {
		if err := sm.templateEngine.EnsureDirectoryExists(dir); err != nil {
			return fmt.Errorf("failed to create volume directory %s: %w", dir, err)
		}
	}

	sm.logger.Debug("Volume directories ensured",
		zap.String("compose_path", composePath),
		zap.Int("volume_count", len(volumeDirs)))

	return nil
}

// logAction logs a service action to the service log
func (sm *ServiceManager) logAction(serviceID uint, action, status, message string) {
	log := &models.ServiceLog{
		ServiceID: serviceID,
		Action:    action,
		Status:    status,
		Message:   message,
	}

	if err := sm.logRepo.Create(log); err != nil {
		sm.logger.Error("Failed to create service log",
			zap.Uint("service_id", serviceID),
			zap.String("action", action),
			zap.Error(err))
	}
}

// AutoStartEnabledServices starts all enabled services on app startup
// and removes containers for disabled services
func (sm *ServiceManager) AutoStartEnabledServices() error {
	sm.logger.Info("Auto-starting enabled services on startup")

	// Get all services from database
	// We need to use the repository interface, which doesn't have ListAll
	// So we'll try to get known services one by one
	knownServices := []string{"radarr", "sonarr", "jellyfin", "prowlarr", "qbittorrent", "jellyseerr", "bazarr", "jellystat"}

	for _, serviceName := range knownServices {
		svc, err := sm.serviceRepo.GetByName(serviceName)
		if err != nil {
			// Service not in database yet, skip
			continue
		}

		if svc.Enabled {
			// Service is enabled → always start it
			sm.logger.Info("Auto-starting enabled service",
				zap.String("service", serviceName))

			// Generate compose file
			composePath, err := sm.templateEngine.GenerateCompose(serviceName, svc.Config)
			if err != nil {
				sm.logger.Error("Failed to generate compose for auto-start",
					zap.String("service", serviceName),
					zap.Error(err))
				continue
			}

			// Start container (docker compose up is idempotent - won't recreate if already running)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			result, err := sm.dockerCompose.ComposeUp(ctx, composePath, map[string]string{})
			cancel()

			if err != nil {
				sm.logger.Error("Failed to auto-start service",
					zap.String("service", serviceName),
					zap.Error(err))
				sm.logAction(svc.ID, "auto_start", "error", fmt.Sprintf("Failed: %v", err))
				continue
			}

			if !result.Success {
				sm.logger.Warn("Auto-start completed with errors",
					zap.String("service", serviceName),
					zap.String("error", result.Error))
			} else {
				sm.logger.Info("Service auto-started successfully",
					zap.String("service", serviceName))
				sm.logAction(svc.ID, "auto_start", "success", "Service started on app init")
			}
		} else {
			// Service is disabled → kill and remove container if exists
			sm.logger.Debug("Removing container for disabled service",
				zap.String("service", serviceName))

			// Use docker kill + docker rm for fast cleanup
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			_ = sm.dockerClient.StopContainer(ctx, serviceName, 0)
			cancel()
		}
	}

	sm.logger.Info("Auto-start process completed")
	return nil
}
