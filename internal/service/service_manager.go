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

	sm.logAction(svc.ID, "reset", "success", "Service pruned: down -v, deleted directories, disabled service")
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

	// Load credentials from database or generate new ones
	dbPath := filepath.Join("/app/data/services-volumes", serviceName, "radarr.db")
	username, exists, err := sm.getRadarrCredentialsFromDB(dbPath)
	
	if err != nil {
		sm.logger.Warn("Failed to read credentials from database", zap.Error(err))
	} else if exists {
		// Credentials exist in database - load them
		config.Username = username
		config.Password = "********" // Don't expose actual password
		sm.logger.Info("Loaded existing credentials from Radarr database", zap.String("username", username))
	} else {
		// No credentials in database - generate random ones
		generatedUser, generatedPass := generateRandomCredentials()
		config.Username = generatedUser
		config.Password = generatedPass
		sm.logger.Info("Generated random credentials for Radarr",
			zap.String("username", generatedUser),
			zap.String("password_length", fmt.Sprintf("%d", len(generatedPass))))
		
		// Automatically save the generated credentials
		if err := sm.updateRadarrCredentials(dbPath, generatedUser, generatedPass); err != nil {
			sm.logger.Warn("Failed to save generated credentials", zap.Error(err))
		} else {
			sm.logger.Info("Generated credentials saved to database")
		}
	}

	return config, nil
}

// UpdateRadarrConfig updates Radarr's config.xml file and restarts the service
func (sm *ServiceManager) UpdateRadarrConfig(ctx context.Context, serviceName string, config models.RadarrConfig) error {
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
		startIdx := indexOf(content, start)
		if startIdx == -1 {
			return ""
		}
		startIdx += len(start)
		endIdx := indexOf(content[startIdx:], end)
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
			if indexOf(content, "<ApiKey>") != -1 && indexOf(content, "</ApiKey>") != -1 {
				// Extract ApiKey value
				start := indexOf(content, "<ApiKey>") + 8
				end := indexOf(content[start:], "</ApiKey>")
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
		LogLevel:               "debug",
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

// Helper functions
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// generateRandomCredentials generates random username and password
func generateRandomCredentials() (username, password string) {
	// Generate random username: "admin" + 4 random chars
	usernameBytes := make([]byte, 4)
	rand.Read(usernameBytes)
	username = fmt.Sprintf("admin%x", usernameBytes)[:10] // max 10 chars

	// Generate random password: 16 characters
	passwordBytes := make([]byte, 12)
	rand.Read(passwordBytes)
	password = base64.URLEncoding.EncodeToString(passwordBytes)[:16]

	return username, password
}

// getRadarrCredentialsFromDB reads existing credentials from Radarr's database
func (sm *ServiceManager) getRadarrCredentialsFromDB(dbPath string) (username string, exists bool, err error) {
	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return "", false, nil
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Query for existing user
	var savedUsername string
	err = db.QueryRow("SELECT Username FROM Users LIMIT 1").Scan(&savedUsername)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil // No user exists
		}
		return "", false, fmt.Errorf("failed to query users: %w", err)
	}

	return savedUsername, true, nil
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

	// PBKDF2-HMAC-SHA512 with 10000 iterations (same as Radarr)
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
