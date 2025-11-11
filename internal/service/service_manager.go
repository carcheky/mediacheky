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

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
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

	// Stop container if running
	if svc.ContainerID != "" && svc.Status == "running" {
		sm.logger.Info("Stopping container before disabling",
			zap.String("service", serviceName),
			zap.String("container_id", svc.ContainerID))

		if err := sm.StopService(ctx, serviceName); err != nil {
			sm.logger.Warn("Failed to stop container during disable",
				zap.String("service", serviceName),
				zap.Error(err))
			// Continue with disable even if stop fails
		}
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

	// Execute docker compose up
	result, err := sm.dockerCompose.ComposeUp(ctx, composePath)
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

	// Execute docker compose down
	result, err := sm.dockerCompose.ComposeDown(ctx, composePath)
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

// UpdateServiceConfig updates a service's configuration and regenerates compose file if enabled
func (sm *ServiceManager) UpdateServiceConfig(ctx context.Context, serviceName string, config models.ServiceConfig) error {
	sm.logger.Info("Updating service configuration", zap.String("service", serviceName))

	// Get service from database
	svc, err := sm.serviceRepo.GetByName(serviceName)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Update configuration
	svc.Config = config

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
		sm.logAction(svc.ID, "config_update", "success", "Configuration updated and compose file regenerated")
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
