package service

import (
	"context"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"go.uber.org/zap"
	"golang.org/x/crypto/pbkdf2"
)

const (
	// MediaChekyNetwork is the hardcoded network name that ALL services MUST use
	// This is the Docker Compose prefixed name (project_networkname)
	MediaChekyNetwork = "mediacheky_mediacheky-net"
)

// getHostDataPath returns the host path that is mounted to /app/data
// This allows us to build service config paths dynamically
func getHostDataPath() string {
	// First try environment variable (set by docker compose.yml)
	if hostPath := os.Getenv("MEDIACHEKY_HOST_PATH"); hostPath != "" {
		return filepath.Join(hostPath, "volumes", "mediacheky-data")
	}

	// Fallback: relative path (works in most cases)
	return "./volumes/mediacheky-data"
}

// buildServiceConfigPath constructs the full host path for a service's config directory
func buildServiceConfigPath(serviceName string) string {
	return filepath.Join(getHostDataPath(), "services-volumes", serviceName)
}

// getMediaLibraryPath returns the absolute host path for the media library
// Uses the same logic as getHostDataPath: reads env var and constructs path
func getMediaLibraryPath() string {
	// Read from environment variable (set by docker compose.yml)
	mediaPath := os.Getenv("MEDIACHEKY_MEDIA_PATH")
	if mediaPath == "" {
		mediaPath = "./volumes/library" // Default
	}

	// If path is relative, make it absolute using MEDIACHEKY_HOST_PATH
	if !filepath.IsAbs(mediaPath) {
		if hostPath := os.Getenv("MEDIACHEKY_HOST_PATH"); hostPath != "" {
			return filepath.Join(hostPath, mediaPath)
		}
	}

	// Already absolute or no MEDIACHEKY_HOST_PATH set
	return mediaPath
}

// ServiceRepository defines the interface for service data access
type ServiceRepository interface {
	GetByName(name string) (*models.Service, error)
	Create(service *models.Service) error
	Update(service *models.Service) error
	UpdateStatus(id uint, status string, containerID string) error
	SetEnabled(id uint, enabled bool) error
	GetServiceCredentials(serviceName string) (*models.ServiceCredentials, error)
	SaveServiceCredentials(serviceName, username, password string) error
	DeleteServiceCredentials(serviceName string) error
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

// EnableService enables a service and generates its docker compose file
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
					"Config": buildServiceConfigPath(serviceName),
					"Media":  getMediaLibraryPath(),
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

	// Generate docker compose file
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

	// Initialize Radarr config.xml with defaults if needed
	if serviceName == "radarr" {
		if err := sm.initializeRadarrConfig(ctx, serviceName); err != nil {
			sm.logger.Warn("Failed to initialize Radarr config", zap.Error(err))
			// Don't fail the start, just log the warning
		}
	}

	// Initialize Sonarr config.xml with defaults if needed
	if serviceName == "sonarr" {
		if err := sm.initializeSonarrConfig(ctx, serviceName); err != nil {
			sm.logger.Warn("Failed to initialize Sonarr config", zap.Error(err))
			// Don't fail the start, just log the warning
		}
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

	// Get compose file path
	composePath := sm.templateEngine.GetComposePath(serviceName)

	// Load global config for docker compose execution
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		sm.logAction(svc.ID, "restart", "error", fmt.Sprintf("Failed to load global config: %v", err))
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Stop the service
	if err := sm.StopService(ctx, serviceName); err != nil {
		sm.logAction(svc.ID, "restart", "error", fmt.Sprintf("Failed to stop during restart: %v", err))
		return fmt.Errorf("failed to stop service during restart: %w", err)
	}

	// Use docker compose up --force-recreate to recreate the container
	sm.logger.Info("Recreating container", zap.String("service", serviceName))
	result, err := sm.dockerCompose.ComposeUpForceRecreate(ctx, composePath, globalConfig)
	if err != nil {
		sm.logAction(svc.ID, "restart", "error", fmt.Sprintf("Failed to recreate container: %v", err))
		return fmt.Errorf("failed to recreate container: %w", err)
	}

	if !result.Success {
		sm.logAction(svc.ID, "restart", "error", result.Error)
		return fmt.Errorf("container recreation failed: %s", result.Error)
	}

	sm.logAction(svc.ID, "restart", "success", "Service restarted and container recreated")
	sm.logger.Info("Service restarted successfully", zap.String("service", serviceName))

	return nil
}

// ResetService deletes the service's config directory and recreates the container
// This will wipe all service configuration data (NOT media files)
func (sm *ServiceManager) ResetService(ctx context.Context, serviceName string) error {
	sm.logger.Info("Resetting service (deleting config and recreating container)", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Get compose file path
	composePath := sm.templateEngine.GetComposePath(serviceName)

	// Load global config
	globalConfig, err := sm.templateEngine.LoadGlobalConfigPublic()
	if err != nil {
		sm.logAction(svc.ID, "reset", "error", fmt.Sprintf("Failed to load global config: %v", err))
		return fmt.Errorf("failed to load global config: %w", err)
	}

	// Stop the service and remove volumes with docker compose down -v
	sm.logger.Info("Stopping service and removing volumes (docker compose down -v)", zap.String("service", serviceName))
	downResult, downErr := sm.dockerCompose.ComposeDownWithVolumes(ctx, composePath, globalConfig)
	if downErr != nil {
		sm.logger.Warn("Failed to stop service with docker compose down -v, continuing anyway",
			zap.String("service", serviceName),
			zap.Error(downErr))
	} else if !downResult.Success {
		sm.logger.Warn("docker compose down -v reported error, continuing anyway",
			zap.String("service", serviceName),
			zap.String("error", downResult.Error))
	}

	// Delete both directories: services/radarr and services-volumes/radarr
	serviceDirs := []string{
		filepath.Join("/app/data/services", serviceName),         // docker compose files
		filepath.Join("/app/data/services-volumes", serviceName), // config data
	}

	for _, dirPath := range serviceDirs {
		sm.logger.Info("Deleting service directory",
			zap.String("service", serviceName),
			zap.String("path", dirPath))

		if err := sm.templateEngine.RemoveDirectory(dirPath); err != nil {
			sm.logger.Warn("Failed to delete directory, continuing",
				zap.String("path", dirPath),
				zap.Error(err))
		}
	}

	sm.logger.Info("Service directories deleted",
		zap.String("service", serviceName))

	// Delete service credentials from MediaCheky database (for Radarr, etc.)
	if err := sm.serviceRepo.DeleteServiceCredentials(serviceName); err != nil {
		sm.logger.Warn("Failed to delete service credentials",
			zap.String("service", serviceName),
			zap.Error(err))
	} else {
		sm.logger.Info("Service credentials deleted from MediaCheky database",
			zap.String("service", serviceName))
	}

	// Reset service configuration to defaults so it can be enabled again
	defaultConfig := map[string]interface{}{
		"Image":         fmt.Sprintf("linuxserver/%s:latest", serviceName),
		"ContainerName": serviceName,
		"Paths": map[string]string{
			"Config": buildServiceConfigPath(serviceName),
			"Media":  getMediaLibraryPath(),
		},
		"RestartPolicy": "unless-stopped",
	}

	svc.Config = defaultConfig
	if err := sm.serviceRepo.Update(svc); err != nil {
		sm.logger.Warn("Failed to reset service config to defaults",
			zap.String("service", serviceName),
			zap.Error(err))
	} else {
		sm.logger.Info("Service config reset to defaults", zap.String("service", serviceName))
	}

	// Disable service after prune and set status to stopped
	if err := sm.serviceRepo.SetEnabled(svc.ID, false); err != nil {
		sm.logger.Warn("Failed to disable service after prune",
			zap.String("service", serviceName),
			zap.Error(err))
	} else {
		sm.logger.Info("Service disabled after prune", zap.String("service", serviceName))
	}
	if err := sm.serviceRepo.UpdateStatus(svc.ID, "stopped", ""); err != nil {
		sm.logger.Warn("Failed to update service status after prune",
			zap.String("service", serviceName),
			zap.Error(err))
	}

	sm.logAction(svc.ID, "reset", "success", "Service pruned: down -v, deleted directories, deleted credentials, disabled service")
	sm.logger.Info("Service prune completed successfully", zap.String("service", serviceName))

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

	// Get service from database, create if doesn't exist
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		if err.Error() == "record not found" || err.Error() == "failed to get service: record not found" {
			// Create service with provided configuration
			sm.logger.Info("Service not found, creating with provided config", zap.String("service", serviceName))
			svc = &models.Service{
				Name:    serviceName,
				Enabled: false,
				Status:  "stopped",
				Config:  config,
			}

			if err := sm.serviceRepo.Create(svc); err != nil {
				return fmt.Errorf("failed to create service: %w", err)
			}
			sm.logger.Info("Service created successfully", zap.String("service", serviceName))
		} else {
			return fmt.Errorf("failed to get service: %w", err)
		}
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

// GetRadarrConfig reads and parses Radarr's config.xml file
func (sm *ServiceManager) GetRadarrConfig(ctx context.Context, serviceName string) (models.RadarrConfig, error) {
	var config models.RadarrConfig

	// Build path to config.xml inside the container
	configPath := filepath.Join("/app/data/services-volumes", serviceName, "config.xml")

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			sm.logger.Info("Radarr config.xml not found, returning defaults", zap.String("path", configPath))
			// Return default config
			config = models.RadarrConfig{
				BindAddress:            "*",
				Port:                   7878,
				SslPort:                9898,
				EnableSsl:              false,
				LaunchBrowser:          true,
				ApiKey:                 "",
				AuthenticationMethod:   "Forms",
				AuthenticationRequired: "DisabledForLocalAddresses",
				Username:               "",
				Password:               "",
				PasswordConfirmation:   "",
				Branch:                 "master",
				LogLevel:               "debug",
				SslCertPath:            "",
				SslCertPassword:        "",
				UrlBase:                "",
				InstanceName:           "Radarr",
				UpdateMechanism:        "Docker",
				UseProxy:               false,
				SendAnonymousUsageData: true,
			}
		} else {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Parse XML
		if err := parseRadarrConfig(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse config XML: %w", err)
		}
	}

	// CRITICAL: Check if Radarr container is fully initialized
	// We must wait for [ls.io-init] done. in logs before accessing/modifying anything
	containerInitialized, err := sm.isRadarrContainerInitialized(ctx, serviceName)
	if err != nil {
		sm.logger.Warn("Failed to check Radarr container initialization status", zap.Error(err))
	}

	if !containerInitialized {
		sm.logger.Info("Radarr container not fully initialized yet, waiting for [ls.io-init] done.",
			zap.String("service", serviceName))
		// Return empty credentials - don't try to access DB or generate anything
		config.Username = ""
		config.Password = ""
		return config, nil
	}

	sm.logger.Info("Radarr container is fully initialized, proceeding with configuration",
		zap.String("service", serviceName))

	// Load credentials from our MediaCheky database first
	credentials, err := sm.getServiceCredentials(serviceName)
	if err == nil && credentials != nil {
		// Credentials exist in MediaCheky database - use them
		config.Username = credentials.Username
		config.Password = credentials.Password // Return real password for copy/autologin
		sm.logger.Info("Loaded existing credentials from MediaCheky database", zap.String("username", credentials.Username))
	} else {
		// Try loading from Radarr's database as fallback
		dbPath := filepath.Join("/app/data/services-volumes", serviceName, "radarr.db")
		dbStatus, err := sm.getRadarrCredentialsFromDB(dbPath)

		if err != nil {
			sm.logger.Warn("Failed to read credentials from Radarr database", zap.Error(err))
		} else if dbStatus != nil {
			// Check database status and act accordingly
			if !dbStatus.DBExists {
				// Database doesn't exist yet - Radarr hasn't been started
				// DO NOT generate credentials yet - wait for Radarr to create its DB
				sm.logger.Info("Radarr database does not exist yet, waiting for first Radarr startup")
				config.Username = ""
				config.Password = ""
			} else if !dbStatus.TableExists {
				// Database exists but Users table doesn't - Radarr is initializing
				// DO NOT generate credentials yet - wait for Radarr to create the table
				sm.logger.Info("Radarr Users table does not exist yet, waiting for Radarr initialization")
				config.Username = ""
				config.Password = ""
			} else if !dbStatus.UserExists {
				// Database exists, table exists, but NO user - safe to generate
				sm.logger.Info("Radarr database ready but no user exists, generating credentials")
				generatedUser, generatedPass, err := generateRandomCredentials()
				if err != nil {
					sm.logger.Error("Failed to generate credentials", zap.Error(err))
				} else {
					config.Username = generatedUser
					config.Password = generatedPass
					sm.logger.Info("Generated random credentials for Radarr",
						zap.String("username", generatedUser),
						zap.String("password_length", fmt.Sprintf("%d", len(generatedPass))))

					// Save to both databases
					if err := sm.updateRadarrCredentials(dbPath, generatedUser, generatedPass); err != nil {
						sm.logger.Warn("Failed to save credentials to Radarr database", zap.Error(err))
					}
					if err := sm.saveServiceCredentials(serviceName, generatedUser, generatedPass); err != nil {
						sm.logger.Warn("Failed to save credentials to MediaCheky database", zap.Error(err))
					} else {
						sm.logger.Info("Generated credentials saved to both databases")
					}

					// Ensure default root folder exists
					if err := sm.ensureRootFoldersInDB(ctx, serviceName); err != nil {
						sm.logger.Warn("Failed to ensure root folders in Radarr DB", zap.String("service", serviceName), zap.Error(err))
					}
				}
			} else {
				// User exists in Radarr database - migrate to MediaCheky DB
				password, err := sm.getRadarrPasswordFromDB(dbPath, dbStatus.Username)
				if err == nil && password != "" {
					config.Username = dbStatus.Username
					config.Password = password
					// Save to MediaCheky database for future use
					if err := sm.saveServiceCredentials(serviceName, dbStatus.Username, password); err != nil {
						sm.logger.Warn("Failed to save credentials to MediaCheky database", zap.Error(err))
					} else {
						sm.logger.Info("Migrated credentials from Radarr DB to MediaCheky DB", zap.String("username", dbStatus.Username))
					}
				} else {
					// Fallback: show masked password
					config.Username = dbStatus.Username
					config.Password = "********"
					sm.logger.Info("Loaded username from Radarr database (password masked)", zap.String("username", dbStatus.Username))
				}
			}
		}
	}

	return config, nil
}

// UpdateRadarrConfig updates Radarr's config.xml file and restarts the service
func (sm *ServiceManager) UpdateRadarrConfig(ctx context.Context, serviceName string, config models.RadarrConfig) error {
	// CRITICAL: Verify container is fully initialized before saving ANY configuration
	containerInitialized, err := sm.isRadarrContainerInitialized(ctx, serviceName)
	if err != nil {
		sm.logger.Warn("Failed to check Radarr container initialization status",
			zap.String("service", serviceName),
			zap.Error(err))
		return fmt.Errorf("cannot update Radarr config: failed to verify container initialization: %w", err)
	}
	if !containerInitialized {
		sm.logger.Warn("Radarr container not fully initialized yet, refusing to save credentials",
			zap.String("service", serviceName))
		return fmt.Errorf("Radarr container is not fully initialized yet. Please wait for the container to finish starting (look for '[ls.io-init] done.' in logs)")
	}

	// Build path to config.xml inside the container
	configPath := filepath.Join("/app/data/services-volumes", serviceName, "config.xml")

	// Check if directory exists
	configDir := filepath.Dir(configPath)
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("Radarr config directory does not exist: %s. Make sure the service is enabled and started at least once", configDir)
	}

	// Generate XML content
	xmlContent, err := generateRadarrConfigXML(config)
	if err != nil {
		return fmt.Errorf("failed to generate config XML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, []byte(xmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	sm.logger.Info("Radarr config.xml updated successfully", zap.String("path", configPath))

	// Update credentials in Radarr's database if provided
	if config.Username != "" && config.Password != "" {
		dbPath := filepath.Join("/app/data/services-volumes", serviceName, "radarr.db")
		sm.logger.Info("Updating Radarr credentials in database",
			zap.String("dbPath", dbPath),
			zap.String("username", config.Username))
		if err := sm.updateRadarrCredentials(dbPath, config.Username, config.Password); err != nil {
			sm.logger.Warn("Failed to update Radarr credentials in database", zap.Error(err))
			// Don't fail the entire operation - just log the warning
		} else {
			sm.logger.Info("Radarr credentials updated successfully in database")
		}
	}

	// Restart Radarr container to apply changes
	// Radarr only reads config.xml on startup, not in hot-reload
	sm.logger.Info("Restarting Radarr to apply configuration changes", zap.String("service", serviceName))
	if err := sm.RestartService(ctx, serviceName); err != nil {
		sm.logger.Warn("Failed to restart Radarr after config update", zap.Error(err))
		// Don't return error - config was saved successfully, restart is a bonus
		return nil
	}

	sm.logger.Info("Radarr restarted successfully", zap.String("service", serviceName))
	return nil
}

// parseRadarrConfig parses the XML config into RadarrConfig struct
func parseRadarrConfig(data []byte, config *models.RadarrConfig) error {
	// Simple XML parsing - extract values between tags
	content := string(data)

	// Helper function to extract value between tags
	extractValue := func(tag string) string {
		start := fmt.Sprintf("<%s>", tag)
		end := fmt.Sprintf("</%s>", tag)
		startIdx := strings.Index(content, start)
		if startIdx == -1 {
			return ""
		}
		startIdx += len(start)
		endIdx := strings.Index(content[startIdx:], end)
		if endIdx == -1 {
			return ""
		}
		return content[startIdx : startIdx+endIdx]
	}

	config.BindAddress = extractValue("BindAddress")
	config.Port = parseInt(extractValue("Port"), 7878)
	config.SslPort = parseInt(extractValue("SslPort"), 9898)
	config.EnableSsl = extractValue("EnableSsl") == "True"
	config.LaunchBrowser = extractValue("LaunchBrowser") == "True"
	config.ApiKey = extractValue("ApiKey")
	config.AuthenticationMethod = extractValue("AuthenticationMethod")
	config.AuthenticationRequired = extractValue("AuthenticationRequired")
	config.Username = extractValue("Username")
	config.Password = extractValue("Password")
	config.PasswordConfirmation = extractValue("PasswordConfirmation")
	config.Branch = extractValue("Branch")
	config.LogLevel = extractValue("LogLevel")
	config.SslCertPath = extractValue("SslCertPath")
	config.SslCertPassword = extractValue("SslCertPassword")
	config.UrlBase = extractValue("UrlBase")
	config.InstanceName = extractValue("InstanceName")
	config.UpdateMechanism = extractValue("UpdateMechanism")
	config.UseProxy = extractValue("UseProxy") == "True"
	config.SendAnonymousUsageData = extractValue("SendAnonymousUsageData") == "True"

	return nil
}

// generateRadarrConfigXML generates XML content from RadarrConfig
func generateRadarrConfigXML(config models.RadarrConfig) (string, error) {
	boolToStr := func(b bool) string {
		if b {
			return "True"
		}
		return "False"
	}

	xml := fmt.Sprintf(`<Config>
  <BindAddress>%s</BindAddress>
  <Port>%d</Port>
  <SslPort>%d</SslPort>
  <EnableSsl>%s</EnableSsl>
  <LaunchBrowser>%s</LaunchBrowser>
  <ApiKey>%s</ApiKey>
  <AuthenticationMethod>%s</AuthenticationMethod>
  <AuthenticationRequired>%s</AuthenticationRequired>
  <Branch>%s</Branch>
  <LogLevel>%s</LogLevel>
  <SslCertPath>%s</SslCertPath>
  <SslCertPassword>%s</SslCertPassword>
  <UrlBase>%s</UrlBase>
  <InstanceName>%s</InstanceName>
  <UpdateMechanism>%s</UpdateMechanism>
</Config>`,
		config.BindAddress,
		config.Port,
		config.SslPort,
		boolToStr(config.EnableSsl),
		boolToStr(config.LaunchBrowser),
		config.ApiKey,
		config.AuthenticationMethod,
		config.AuthenticationRequired,
		config.Branch,
		config.LogLevel,
		config.SslCertPath,
		config.SslCertPassword,
		config.UrlBase,
		config.InstanceName,
		config.UpdateMechanism,
	)

	return xml, nil
}

// initializeRadarrConfig creates config.xml with default values if it doesn't exist
// or if it has default/empty values
func (sm *ServiceManager) initializeRadarrConfig(ctx context.Context, serviceName string) error {
	configPath := filepath.Join("/app/data/services-volumes", serviceName, "config.xml")

	// Check if config file exists
	if _, err := os.Stat(configPath); err == nil {
		// File exists, check if it has valid content
		data, readErr := os.ReadFile(configPath)
		if readErr == nil && len(data) > 0 {
			// File exists and has content, check if ApiKey is set
			content := string(data)
			if strings.Index(content, "<ApiKey>") != -1 && strings.Index(content, "</ApiKey>") != -1 {
				// Extract ApiKey value
				start := strings.Index(content, "<ApiKey>") + 8
				end := strings.Index(content[start:], "</ApiKey>")
				if end > 0 {
					apiKey := content[start : start+end]
					// If ApiKey has value (not empty), assume config is valid
					if len(apiKey) > 0 {
						sm.logger.Debug("Radarr config.xml already exists with valid ApiKey", zap.String("path", configPath))
						return nil
					}
				}
			}
		}
		// If we reach here, file exists but is invalid/empty, will overwrite
		sm.logger.Info("Radarr config.xml exists but is invalid, reinitializing", zap.String("path", configPath))
	}

	// Create default config
	defaultConfig := models.RadarrConfig{
		BindAddress:            "*",
		Port:                   7878,
		SslPort:                9898,
		EnableSsl:              false,
		LaunchBrowser:          true,
		ApiKey:                 "", // Radarr will generate this on first start
		AuthenticationMethod:   "Forms",
		AuthenticationRequired: "DisabledForLocalAddresses",
		Username:               "",
		Password:               "",
		PasswordConfirmation:   "",
		Branch:                 "master",
		LogLevel:               "info",
		SslCertPath:            "",
		SslCertPassword:        "",
		UrlBase:                "",
		InstanceName:           "Radarr",
		UpdateMechanism:        "Docker",
		UseProxy:               false,
		SendAnonymousUsageData: true,
	}

	// Generate XML content
	xmlContent, err := generateRadarrConfigXML(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to generate default config XML: %w", err)
	}

	// Ensure config directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(xmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write default config file: %w", err)
	}

	sm.logger.Info("Radarr config.xml initialized with default values", zap.String("path", configPath))
	return nil
}

// ensureRootFoldersInDB ensures default root folders exist in Radarr/Sonarr database
func (sm *ServiceManager) ensureRootFoldersInDB(ctx context.Context, serviceName string) error {
	var desiredPath string
	switch serviceName {
	case "radarr":
		desiredPath = "/MEDIACHEKY_LIBRARY/library/movies/"
	case "sonarr":
		desiredPath = "/MEDIACHEKY_LIBRARY/library/tv/"
	default:
		return nil // Not applicable for this service
	}

	dbPath := filepath.Join(getHostDataPath(), "services-volumes", serviceName, serviceName+".db")

	// Wait for DB file to exist (with timeout)
	retries := 30 // 30 seconds max wait
	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if _, err := os.Stat(dbPath); err != nil {
			if os.IsNotExist(err) {
				time.Sleep(time.Second)
				continue
			}
			sm.logger.Warn("Error checking DB file for root folder",
				zap.String("dbPath", dbPath),
				zap.Error(err))
			return nil // Non-fatal
		}

		// DB file exists, try to insert root folder
		db, err := sql.Open("sqlite3", dbPath)
		if err != nil {
			sm.logger.Warn("Failed to open service DB for root folder",
				zap.String("dbPath", dbPath),
				zap.Error(err))
			return nil // Non-fatal
		}

		// Check if RootFolders table exists and if our path is already there
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM RootFolders WHERE Path = ?", desiredPath).Scan(&count)
		if err != nil {
			_ = db.Close()
			// Table might not exist yet, wait a bit
			time.Sleep(500 * time.Millisecond)
			continue
		}

		if count == 0 {
			// Insert the default root folder
			_, err = db.Exec("INSERT OR IGNORE INTO RootFolders (Path) VALUES (?)", desiredPath)
			if err != nil {
				_ = db.Close()
				if strings.Contains(err.Error(), "database is locked") || strings.Contains(err.Error(), "busy") {
					// Retry on lock
					time.Sleep(250 * time.Millisecond)
					continue
				}
				sm.logger.Warn("Failed to insert root folder",
					zap.String("dbPath", dbPath),
					zap.Error(err))
				return nil // Non-fatal
			}
			sm.logger.Info("Inserted default root folder for service",
				zap.String("service", serviceName),
				zap.String("path", desiredPath))
		} else {
			sm.logger.Debug("Root folder already exists",
				zap.String("service", serviceName),
				zap.String("path", desiredPath))
		}

		_ = db.Close()
		return nil
	}

	sm.logger.Warn("Database not ready to ensure root folder",
		zap.String("dbPath", dbPath))
	return nil
}

// Helper functions

// generateRandomCredentials generates random username and password
func generateRandomCredentials() (username, password string, err error) {
	// Generate random username: "admin" + 4 random chars
	usernameBytes := make([]byte, 4)
	if _, err := rand.Read(usernameBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random username: %w", err)
	}
	username = fmt.Sprintf("admin%x", usernameBytes)[:10] // max 10 chars

	// Generate random password: 16 characters
	passwordBytes := make([]byte, 12)
	if _, err := rand.Read(passwordBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random password: %w", err)
	}
	password = base64.URLEncoding.EncodeToString(passwordBytes)[:16]

	return username, password, nil
}

// isRadarrContainerInitialized checks if Radarr container has completed initialization
// by looking for "[ls.io-init] done." in the container logs
func (sm *ServiceManager) isRadarrContainerInitialized(ctx context.Context, serviceName string) (bool, error) {
	// Get service from database to find container name
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		sm.logger.Debug("Service not found in database, assuming not initialized",
			zap.String("service", serviceName))
		return false, nil
	}

	// Check if service is enabled and has a container
	if !svc.Enabled {
		sm.logger.Debug("Service not enabled",
			zap.String("service", serviceName))
		return false, nil
	}

	// Extract container name from config
	var containerName string
	if nameVal, ok := svc.Config["ContainerName"]; ok {
		if nameStr, ok := nameVal.(string); ok && nameStr != "" {
			containerName = nameStr
		}
	}

	// If no container name in config, use service name
	if containerName == "" {
		containerName = serviceName
	}

	sm.logger.Debug("Checking Radarr container initialization",
		zap.String("service", serviceName),
		zap.String("container", containerName))

	// Try to get container info
	containerInfo, err := sm.dockerClient.GetContainer(ctx, containerName)
	if err != nil || containerInfo == nil {
		sm.logger.Debug("Container not found or not accessible",
			zap.String("container", containerName),
			zap.Error(err))
		return false, nil
	}

	// Check if container is running
	if containerInfo.Status != "running" {
		sm.logger.Debug("Container is not running",
			zap.String("container", containerName),
			zap.String("status", string(containerInfo.Status)))
		return false, nil
	}

	// Get container logs (last 200 lines should be enough to catch init message)
	logs, err := sm.dockerClient.GetContainerLogs(ctx, containerInfo.ID, "200")
	if err != nil {
		sm.logger.Warn("Failed to get container logs for initialization check",
			zap.String("container", containerName),
			zap.Error(err))
		return false, err
	}

	// Check if logs contain the initialization complete marker
	initComplete := strings.Contains(logs, "[ls.io-init] done.")

	if initComplete {
		sm.logger.Info("Radarr container initialization complete",
			zap.String("service", serviceName),
			zap.String("container", containerName))
	} else {
		sm.logger.Debug("Radarr container still initializing, waiting for [ls.io-init] done.",
			zap.String("service", serviceName),
			zap.String("container", containerName))
	}

	return initComplete, nil
}

// RadarrDBStatus representa el estado de la base de datos de Radarr
type RadarrDBStatus struct {
	DBExists    bool
	TableExists bool
	UserExists  bool
	Username    string
}

// getRadarrCredentialsFromDB reads existing credentials from Radarr's database
// Returns detailed status about database, table, and user existence
func (sm *ServiceManager) getRadarrCredentialsFromDB(dbPath string) (*RadarrDBStatus, error) {
	status := &RadarrDBStatus{
		DBExists:    false,
		TableExists: false,
		UserExists:  false,
		Username:    "",
	}

	// Check if database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		sm.logger.Debug("Radarr database does not exist yet", zap.String("path", dbPath))
		return status, nil // DB doesn't exist - this is normal on first start
	}

	status.DBExists = true

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return status, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Check if Users table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='Users'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			sm.logger.Debug("Users table does not exist yet in Radarr database")
			return status, nil // Table doesn't exist - Radarr hasn't created it yet
		}
		return status, fmt.Errorf("failed to check Users table: %w", err)
	}

	status.TableExists = true

	// Query for existing user
	var savedUsername string
	err = db.QueryRow("SELECT Username FROM Users LIMIT 1").Scan(&savedUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			sm.logger.Debug("Users table is empty")
			return status, nil // Table exists but no user - we can generate credentials
		}
		return status, fmt.Errorf("failed to query users: %w", err)
	}

	status.UserExists = true
	status.Username = savedUsername
	sm.logger.Debug("Found existing user in Radarr database", zap.String("username", savedUsername))

	return status, nil
}

// updateRadarrCredentials updates username and password in Radarr's database
func (sm *ServiceManager) updateRadarrCredentials(dbPath, username, password string) error {
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Generate salt and hash password (mimicking Radarr's UserService)
	salt := make([]byte, 16) // 128 bits / 8
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// PBKDF2-HMAC-SHA512 with 10000 iterations
	// NOTE: 10,000 iterations is intentionally low to match Radarr's implementation for compatibility.
	// While OWASP recommends 120,000+ iterations for PBKDF2-SHA512, we must use Radarr's exact settings
	// to ensure generated passwords work with Radarr's authentication system.
	hashedPassword := pbkdf2.Key([]byte(password), salt, 10000, 32, sha512.New)

	saltB64 := base64.StdEncoding.EncodeToString(salt)
	passwordB64 := base64.StdEncoding.EncodeToString(hashedPassword)

	// Check if user exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM Users").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	if count == 0 {
		// Insert new user
		identifier := uuid.New().String()
		_, err = db.Exec(
			"INSERT INTO Users (Identifier, Username, Password, Salt, Iterations) VALUES (?, ?, ?, ?, ?)",
			identifier,
			strings.ToLower(username),
			passwordB64,
			saltB64,
			10000,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		sm.logger.Info("Radarr user created in database", zap.String("username", username))
	} else {
		// Update existing user
		_, err = db.Exec(
			"UPDATE Users SET Username = ?, Password = ?, Salt = ?, Iterations = ? WHERE Id = (SELECT Id FROM Users LIMIT 1)",
			strings.ToLower(username),
			passwordB64,
			saltB64,
			10000,
		)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		sm.logger.Info("Radarr user updated in database", zap.String("username", username))
	}

	return nil
}

func parseInt(s string, defaultVal int) int {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultVal
	}
	return result
}

// getServiceCredentials retrieves credentials from MediaCheky's database
func (sm *ServiceManager) getServiceCredentials(serviceName string) (*models.ServiceCredentials, error) {
	return sm.serviceRepo.GetServiceCredentials(serviceName)
}

// saveServiceCredentials saves or updates credentials in MediaCheky's database
func (sm *ServiceManager) saveServiceCredentials(serviceName, username, password string) error {
	return sm.serviceRepo.SaveServiceCredentials(serviceName, username, password)
}

// getRadarrPasswordFromDB retrieves the password from Radarr's database
// This is a helper method to migrate credentials from Radarr DB to MediaCheky DB
func (sm *ServiceManager) getRadarrPasswordFromDB(dbPath, username string) (string, error) {
	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database not found")
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Query for password hash and salt
	var passwordB64, saltB64 string
	var iterations int
	err = db.QueryRow("SELECT Password, Salt, Iterations FROM Users WHERE Username = ? LIMIT 1", strings.ToLower(username)).Scan(&passwordB64, &saltB64, &iterations)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// Note: We cannot reverse the hash, so we return an empty string
	// This means we need to keep the password in MediaCheky DB from the start
	return "", nil
}

// ============================================================================
// SONARR CONFIGURATION MANAGEMENT
// ============================================================================

// GetSonarrConfig reads and parses Sonarr's config.xml file
func (sm *ServiceManager) GetSonarrConfig(ctx context.Context, serviceName string) (models.SonarrConfig, error) {
	var config models.SonarrConfig

	// Build path to config.xml inside the container
	configPath := filepath.Join("/app/data/services-volumes", serviceName, "config.xml")

	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			sm.logger.Info("Sonarr config.xml not found, returning defaults", zap.String("path", configPath))
			// Return default config
			config = models.SonarrConfig{
				BindAddress:            "*",
				Port:                   8989,
				SslPort:                9898,
				EnableSsl:              false,
				LaunchBrowser:          true,
				ApiKey:                 "",
				AuthenticationMethod:   "Forms",
				AuthenticationRequired: "DisabledForLocalAddresses",
				Username:               "",
				Password:               "",
				PasswordConfirmation:   "",
				Branch:                 "master",
				LogLevel:               "debug",
				SslCertPath:            "",
				SslCertPassword:        "",
				UrlBase:                "",
				InstanceName:           "Sonarr",
				UpdateMechanism:        "Docker",
				UseProxy:               false,
				SendAnonymousUsageData: true,
			}
		} else {
			return config, fmt.Errorf("failed to read config file: %w", err)
		}
	} else {
		// Parse XML
		if err := parseSonarrConfig(data, &config); err != nil {
			return config, fmt.Errorf("failed to parse config XML: %w", err)
		}
	}

	// CRITICAL: Check if Sonarr container is fully initialized
	// We must wait for [ls.io-init] done. in logs before accessing/modifying anything
	containerInitialized, err := sm.isSonarrContainerInitialized(ctx, serviceName)
	if err != nil {
		sm.logger.Warn("Failed to check Sonarr container initialization status", zap.Error(err))
	}

	if !containerInitialized {
		sm.logger.Info("Sonarr container not fully initialized yet, waiting for [ls.io-init] done.",
			zap.String("service", serviceName))
		// Return empty credentials - don't try to access DB or generate anything
		config.Username = ""
		config.Password = ""
		return config, nil
	}

	sm.logger.Info("Sonarr container is fully initialized, proceeding with configuration",
		zap.String("service", serviceName))

	// Load credentials from our MediaCheky database first
	credentials, err := sm.getServiceCredentials(serviceName)
	if err == nil && credentials != nil {
		// Credentials exist in MediaCheky database - use them
		config.Username = credentials.Username
		config.Password = credentials.Password // Return real password for copy/autologin
		sm.logger.Info("Loaded existing credentials from MediaCheky database", zap.String("username", credentials.Username))
	} else {
		// Try loading from Sonarr's database as fallback
		dbPath := filepath.Join("/app/data/services-volumes", serviceName, "sonarr.db")
		dbStatus, err := sm.getSonarrCredentialsFromDB(dbPath)

		if err != nil {
			sm.logger.Warn("Failed to read credentials from Sonarr database", zap.Error(err))
		} else if dbStatus != nil {
			// Check database status and act accordingly
			if !dbStatus.DBExists {
				// Database doesn't exist yet - Sonarr hasn't been started
				// DO NOT generate credentials yet - wait for Sonarr to create its DB
				sm.logger.Info("Sonarr database does not exist yet, waiting for first Sonarr startup")
				config.Username = ""
				config.Password = ""
			} else if !dbStatus.TableExists {
				// Database exists but Users table doesn't - Sonarr is initializing
				// DO NOT generate credentials yet - wait for Sonarr to create the table
				sm.logger.Info("Sonarr Users table does not exist yet, waiting for Sonarr initialization")
				config.Username = ""
				config.Password = ""
			} else if !dbStatus.UserExists {
				// Database exists, table exists, but NO user - safe to generate
				sm.logger.Info("Sonarr database ready but no user exists, generating credentials")
				generatedUser, generatedPass, err := generateRandomCredentials()
				if err != nil {
					sm.logger.Error("Failed to generate credentials", zap.Error(err))
				} else {
					config.Username = generatedUser
					config.Password = generatedPass
					sm.logger.Info("Generated random credentials for Sonarr",
						zap.String("username", generatedUser),
						zap.String("password_length", fmt.Sprintf("%d", len(generatedPass))))

					// Save to both databases
					if err := sm.updateSonarrCredentials(dbPath, generatedUser, generatedPass); err != nil {
						sm.logger.Warn("Failed to save credentials to Sonarr database", zap.Error(err))
					}
					if err := sm.saveServiceCredentials(serviceName, generatedUser, generatedPass); err != nil {
						sm.logger.Warn("Failed to save credentials to MediaCheky database", zap.Error(err))
					} else {
						sm.logger.Info("Generated credentials saved to both databases")
					}

					// Ensure default root folder exists
					if err := sm.ensureRootFoldersInDB(ctx, serviceName); err != nil {
						sm.logger.Warn("Failed to ensure root folders in Sonarr DB", zap.String("service", serviceName), zap.Error(err))
					}
				}
			} else {
				// User exists in Sonarr database - migrate to MediaCheky DB
				password, err := sm.getSonarrPasswordFromDB(dbPath, dbStatus.Username)
				if err == nil && password != "" {
					config.Username = dbStatus.Username
					config.Password = password
					// Save to MediaCheky database for future use
					if err := sm.saveServiceCredentials(serviceName, dbStatus.Username, password); err != nil {
						sm.logger.Warn("Failed to save credentials to MediaCheky database", zap.Error(err))
					} else {
						sm.logger.Info("Migrated credentials from Sonarr DB to MediaCheky DB", zap.String("username", dbStatus.Username))
					}
				} else {
					// Fallback: show masked password
					config.Username = dbStatus.Username
					config.Password = "********"
					sm.logger.Info("Loaded username from Sonarr database (password masked)", zap.String("username", dbStatus.Username))
				}
			}
		}
	}

	return config, nil
}

// UpdateSonarrConfig updates Sonarr's config.xml file and restarts the service
func (sm *ServiceManager) UpdateSonarrConfig(ctx context.Context, serviceName string, config models.SonarrConfig) error {
	// CRITICAL: Check if Sonarr container is fully initialized before updating
	containerInitialized, err := sm.isSonarrContainerInitialized(ctx, serviceName)
	if err != nil {
		sm.logger.Warn("Failed to check Sonarr container initialization status",
			zap.String("service", serviceName),
			zap.Error(err))
		return fmt.Errorf("cannot update Sonarr config: failed to verify container initialization: %w", err)
	}
	if !containerInitialized {
		sm.logger.Warn("Sonarr container not fully initialized yet, refusing to save credentials",
			zap.String("service", serviceName))
		return fmt.Errorf("Sonarr container is not fully initialized yet. Please wait for the container to finish starting (look for '[ls.io-init] done.' in logs)")
	}

	// Build path to config directory
	configDir := filepath.Join("/app/data/services-volumes", serviceName)
	configPath := filepath.Join(configDir, "config.xml")

	// Ensure config directory exists
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		return fmt.Errorf("Sonarr config directory does not exist: %s. Make sure the service is enabled and started at least once", configDir)
	}

	// Generate XML content
	xmlContent, err := generateSonarrConfigXML(config)
	if err != nil {
		return fmt.Errorf("failed to generate config XML: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(xmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	sm.logger.Info("Sonarr config.xml updated successfully", zap.String("path", configPath))

	// Update credentials in Sonarr's database if provided
	if config.Username != "" && config.Password != "" {
		dbPath := filepath.Join(configDir, "sonarr.db")
		sm.logger.Info("Updating Sonarr credentials in database",
			zap.String("username", config.Username))
		if err := sm.updateSonarrCredentials(dbPath, config.Username, config.Password); err != nil {
			sm.logger.Warn("Failed to update Sonarr credentials in database", zap.Error(err))
			// Continue anyway - config.xml update succeeded
		} else {
			sm.logger.Info("Sonarr credentials updated successfully in database")
		}
	}

	// Restart Sonarr container to apply changes
	// Sonarr only reads config.xml on startup, not in hot-reload
	sm.logger.Info("Restarting Sonarr to apply configuration changes", zap.String("service", serviceName))
	if err := sm.RestartService(ctx, serviceName); err != nil {
		sm.logger.Warn("Failed to restart Sonarr after config update", zap.Error(err))
		// Don't return error - config was saved successfully, restart is a bonus
		return nil
	}

	sm.logger.Info("Sonarr restarted successfully", zap.String("service", serviceName))
	return nil
}

// parseSonarrConfig parses the XML config into SonarrConfig struct
func parseSonarrConfig(data []byte, config *models.SonarrConfig) error {
	// Simple XML parsing - extract values between tags
	content := string(data)

	// Helper function to extract value between tags
	extractValue := func(tag string) string {
		start := fmt.Sprintf("<%s>", tag)
		end := fmt.Sprintf("</%s>", tag)
		startIdx := strings.Index(content, start)
		if startIdx == -1 {
			return ""
		}
		startIdx += len(start)
		endIdx := strings.Index(content[startIdx:], end)
		if endIdx == -1 {
			return ""
		}
		return content[startIdx : startIdx+endIdx]
	}

	config.BindAddress = extractValue("BindAddress")
	config.Port = parseInt(extractValue("Port"), 8989)
	config.SslPort = parseInt(extractValue("SslPort"), 9898)
	config.EnableSsl = extractValue("EnableSsl") == "True"
	config.LaunchBrowser = extractValue("LaunchBrowser") == "True"
	config.ApiKey = extractValue("ApiKey")
	config.AuthenticationMethod = extractValue("AuthenticationMethod")
	config.AuthenticationRequired = extractValue("AuthenticationRequired")
	config.Username = extractValue("Username")
	config.Password = extractValue("Password")
	config.PasswordConfirmation = extractValue("PasswordConfirmation")
	config.Branch = extractValue("Branch")
	config.LogLevel = extractValue("LogLevel")
	config.SslCertPath = extractValue("SslCertPath")
	config.SslCertPassword = extractValue("SslCertPassword")
	config.UrlBase = extractValue("UrlBase")
	config.InstanceName = extractValue("InstanceName")
	config.UpdateMechanism = extractValue("UpdateMechanism")
	config.UseProxy = extractValue("UseProxy") == "True"
	config.SendAnonymousUsageData = extractValue("SendAnonymousUsageData") == "True"

	return nil
}

// generateSonarrConfigXML generates XML content from SonarrConfig
func generateSonarrConfigXML(config models.SonarrConfig) (string, error) {
	boolToStr := func(b bool) string {
		if b {
			return "True"
		}
		return "False"
	}

	xml := fmt.Sprintf(`<Config>
  <BindAddress>%s</BindAddress>
  <Port>%d</Port>
  <SslPort>%d</SslPort>
  <EnableSsl>%s</EnableSsl>
  <LaunchBrowser>%s</LaunchBrowser>
  <ApiKey>%s</ApiKey>
  <AuthenticationMethod>%s</AuthenticationMethod>
  <AuthenticationRequired>%s</AuthenticationRequired>
  <Branch>%s</Branch>
  <LogLevel>%s</LogLevel>
  <SslCertPath>%s</SslCertPath>
  <SslCertPassword>%s</SslCertPassword>
  <UrlBase>%s</UrlBase>
  <InstanceName>%s</InstanceName>
  <UpdateMechanism>%s</UpdateMechanism>
</Config>`,
		config.BindAddress,
		config.Port,
		config.SslPort,
		boolToStr(config.EnableSsl),
		boolToStr(config.LaunchBrowser),
		config.ApiKey,
		config.AuthenticationMethod,
		config.AuthenticationRequired,
		config.Branch,
		config.LogLevel,
		config.SslCertPath,
		config.SslCertPassword,
		config.UrlBase,
		config.InstanceName,
		config.UpdateMechanism,
	)

	return xml, nil
}

// initializeSonarrConfig creates config.xml with default values if it doesn't exist
// This is called during service Enable operation
func (sm *ServiceManager) initializeSonarrConfig(ctx context.Context, serviceName string) error {
	configPath := filepath.Join("/app/data/services-volumes", serviceName, "config.xml")

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		// Config exists, check if it's valid
		data, err := os.ReadFile(configPath)
		if err == nil {
			var existing models.SonarrConfig
			if err := parseSonarrConfig(data, &existing); err == nil && existing.ApiKey != "" {
				// Config is valid, don't overwrite
				sm.logger.Debug("Sonarr config.xml already exists with valid ApiKey", zap.String("path", configPath))
				return nil
			}
		}
		// Config exists but is invalid, will be overwritten below
		sm.logger.Info("Sonarr config.xml exists but is invalid, reinitializing", zap.String("path", configPath))
	}

	// Create default config
	defaultConfig := models.SonarrConfig{
		BindAddress:            "*",
		Port:                   8989,
		SslPort:                9898,
		EnableSsl:              false,
		LaunchBrowser:          true,
		ApiKey:                 "", // Sonarr will generate this on first start
		AuthenticationMethod:   "Forms",
		AuthenticationRequired: "DisabledForLocalAddresses",
		Username:               "",
		Password:               "",
		PasswordConfirmation:   "",
		Branch:                 "master",
		LogLevel:               "debug",
		SslCertPath:            "",
		SslCertPassword:        "",
		UrlBase:                "",
		InstanceName:           "Sonarr",
		UpdateMechanism:        "Docker",
		UseProxy:               false,
		SendAnonymousUsageData: true,
	}

	xmlContent, err := generateSonarrConfigXML(defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to generate default config: %w", err)
	}

	// Ensure directory exists
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file
	if err := os.WriteFile(configPath, []byte(xmlContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	sm.logger.Info("Sonarr config.xml initialized with default values", zap.String("path", configPath))
	return nil
}

// initializeJellyfinConfig creates necessary directories for Jellyfin
func (sm *ServiceManager) initializeJellyfinConfig(ctx context.Context, serviceName string) error {
	// Jellyfin doesn't use a single config file like *arr services, but we should ensure directories exist
	configDir := filepath.Join("/app/data/services-volumes", serviceName)

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	sm.logger.Info("Jellyfin config directory ensured", zap.String("path", configDir))
	return nil
}

// generateRandomPassword creates a random password of the specified length
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return string(b), nil
}

// isSonarrContainerInitialized checks if Sonarr container has completed initialization
// LinuxServer.io images output "[ls.io-init] done." when ready
func (sm *ServiceManager) isSonarrContainerInitialized(ctx context.Context, serviceName string) (bool, error) {
	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return false, fmt.Errorf("failed to get service: %w", err)
	}

	// Extract container name from config
	var containerName string
	if nameVal, ok := svc.Config["ContainerName"]; ok {
		if nameStr, ok := nameVal.(string); ok && nameStr != "" {
			containerName = nameStr
		}
	}

	// If no container name in config, use service name
	if containerName == "" {
		containerName = serviceName
	}

	sm.logger.Debug("Checking Sonarr container initialization",
		zap.String("service", serviceName),
		zap.String("container", containerName))

	// Try to get container info
	containerInfo, err := sm.dockerClient.GetContainer(ctx, containerName)
	if err != nil || containerInfo == nil {
		sm.logger.Debug("Container not found or not accessible",
			zap.String("container", containerName),
			zap.Error(err))
		return false, nil
	}

	// Check if container is running
	if containerInfo.Status != "running" {
		sm.logger.Debug("Container is not running",
			zap.String("container", containerName),
			zap.String("status", string(containerInfo.Status)))
		return false, nil
	}

	// Get container logs (last 200 lines should be enough to catch init message)
	logs, err := sm.dockerClient.GetContainerLogs(ctx, containerInfo.ID, "200")
	if err != nil {
		sm.logger.Warn("Failed to get container logs for initialization check",
			zap.String("container", containerName),
			zap.Error(err))
		return false, err
	}

	// Check if logs contain the initialization complete marker
	initComplete := strings.Contains(logs, "[ls.io-init] done.")

	if initComplete {
		sm.logger.Info("Sonarr container initialization complete",
			zap.String("container", containerName))
	} else {
		sm.logger.Debug("Sonarr container still initializing, waiting for [ls.io-init] done.",
			zap.String("container", containerName))
	}

	return initComplete, nil
}

// SonarrDBStatus represents the status of Sonarr's database
type SonarrDBStatus struct {
	DBExists    bool
	TableExists bool
	UserExists  bool
	Username    string
}

// getSonarrCredentialsFromDB reads existing credentials from Sonarr's database
// Returns status information about the database and user
func (sm *ServiceManager) getSonarrCredentialsFromDB(dbPath string) (*SonarrDBStatus, error) {
	status := &SonarrDBStatus{
		DBExists:    false,
		TableExists: false,
		UserExists:  false,
		Username:    "",
	}

	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		sm.logger.Debug("Sonarr database does not exist yet", zap.String("path", dbPath))
		return status, fmt.Errorf("database does not exist")
	}
	status.DBExists = true

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return status, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Check if Users table exists
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='Users'").Scan(&tableName)
	if err != nil {
		if err == sql.ErrNoRows {
			sm.logger.Debug("Users table does not exist yet in Sonarr database")
			return status, nil // Table doesn't exist - Sonarr hasn't created it yet
		}
		return status, fmt.Errorf("failed to check for Users table: %w", err)
	}
	status.TableExists = true

	// Try to get username from Users table
	var savedUsername string
	err = db.QueryRow("SELECT Username FROM Users LIMIT 1").Scan(&savedUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			// No user exists yet
			return status, nil
		}
		return status, fmt.Errorf("failed to query user: %w", err)
	}

	status.UserExists = true
	status.Username = savedUsername

	sm.logger.Debug("Found existing user in Sonarr database", zap.String("username", savedUsername))
	return status, nil
}

// updateSonarrCredentials updates username and password in Sonarr's database
func (sm *ServiceManager) updateSonarrCredentials(dbPath, username, password string) error {
	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Generate salt and hash password (mimicking Sonarr's UserService)
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("failed to generate salt: %w", err)
	}

	// NOTE: 10,000 iterations is intentionally low to match Sonarr's implementation for compatibility.
	// While OWASP recommends 120,000+ iterations for PBKDF2-SHA512, we must use Sonarr's exact settings
	// to ensure generated passwords work with Sonarr's authentication system.
	// Reference: https://github.com/Sonarr/Sonarr/blob/develop/src/NzbDrone.Core/Authentication/UserService.cs
	iterations := 10000
	hash := pbkdf2.Key([]byte(password), salt, iterations, 64, sha512.New)

	// Encode to base64
	saltB64 := base64.StdEncoding.EncodeToString(salt)
	hashB64 := base64.StdEncoding.EncodeToString(hash)

	// Check if user exists
	var existingID int
	err = db.QueryRow("SELECT Id FROM Users WHERE Username = ? LIMIT 1", strings.ToLower(username)).Scan(&existingID)

	if err == sql.ErrNoRows {
		// User doesn't exist, insert
		_, err = db.Exec(`
			INSERT INTO Users (Identifier, Username, Password, Salt, Iterations) 
			VALUES (?, ?, ?, ?, ?)`,
			uuid.New().String(),
			strings.ToLower(username),
			hashB64,
			saltB64,
			iterations,
		)
		if err != nil {
			return fmt.Errorf("failed to insert user: %w", err)
		}
		sm.logger.Info("Sonarr user created in database", zap.String("username", username))
	} else if err != nil {
		return fmt.Errorf("failed to check for existing user: %w", err)
	} else {
		// User exists, update
		_, err = db.Exec(`
			UPDATE Users 
			SET Password = ?, Salt = ?, Iterations = ? 
			WHERE Username = ?`,
			hashB64,
			saltB64,
			iterations,
			strings.ToLower(username),
		)
		if err != nil {
			return fmt.Errorf("failed to update user: %w", err)
		}
		sm.logger.Info("Sonarr user updated in database", zap.String("username", username))
	}

	// Note: Credentials are saved to MediaCheky database separately by the caller

	return nil
}

// getSonarrPasswordFromDB retrieves the password from Sonarr's database
// This is a helper method to migrate credentials from Sonarr DB to MediaCheky DB
func (sm *ServiceManager) getSonarrPasswordFromDB(dbPath, username string) (string, error) {
	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", fmt.Errorf("database not found")
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Query for password hash and salt
	var passwordB64, saltB64 string
	var iterations int
	err = db.QueryRow("SELECT Password, Salt, Iterations FROM Users WHERE Username = ? LIMIT 1", strings.ToLower(username)).Scan(&passwordB64, &saltB64, &iterations)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("user not found")
		}
		return "", fmt.Errorf("failed to query user: %w", err)
	}

	// Note: We cannot reverse the hash, so we return an empty string
	// This means we need to keep the password in MediaCheky DB from the start
	return "", nil
}
