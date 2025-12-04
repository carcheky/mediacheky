package handler

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

const (
	// defaultStopTimeout is the default timeout in seconds for stopping containers
	defaultStopTimeout = 10
	// defaultRestartTimeout is the default timeout in seconds for restarting containers
	defaultRestartTimeout = 10
)

// ServiceHandler handles service-related HTTP requests
type ServiceHandler struct {
	repos          *repository.Repositories
	logger         *logger.Logger
	dockerClient   *service.DockerClient
	serviceManager *service.ServiceManager
	proxyService   *service.ProxyService
}

// NewServiceHandler creates a new ServiceHandler instance
// NewServiceHandler creates a new ServiceHandler instance
func NewServiceHandler(repos *repository.Repositories, logger *logger.Logger, dockerClient *service.DockerClient, serviceManager *service.ServiceManager, proxyService *service.ProxyService) *ServiceHandler {
	return &ServiceHandler{
		repos:          repos,
		logger:         logger,
		dockerClient:   dockerClient,
		serviceManager: serviceManager,
		proxyService:   proxyService,
	}
}

// Index renders the top-level Services page
func (h *ServiceHandler) Index(c *fiber.Ctx) error {
	return c.Render("pages/services", fiber.Map{
		"Title": "Services - MediaCheky",
	}, "layouts/main")
}

// ListServices handles GET /api/services
// ListServices handles GET /api/services
func (h *ServiceHandler) ListServices(c *fiber.Ctx) error {
	services, err := h.repos.Service.GetAll()
	if err != nil {
		h.logger.Error("Failed to list services", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve services",
		})
	}

	// Get proxy config for SSL check (if proxyService is available)
	scheme := "http"
	if h.proxyService != nil {
		proxyCfg, _ := h.proxyService.GetProxyConfig()
		if proxyCfg != nil && proxyCfg.SSLEnabled {
			scheme = "https"
		}
	}

	// Create enriched response with calculated URLs
	type ServiceResponse struct {
		models.Service
		Endpoint string `json:"endpoint"`
		URL      string `json:"url"`
	}

	response := make([]ServiceResponse, 0, len(services))
	for _, svc := range services {
		endpoint := ""
		if h.proxyService != nil {
			endpoint, _ = h.proxyService.GetServiceEndpoint(svc.Name)
		}
		url := ""
		if endpoint != "" {
			url = fmt.Sprintf("%s://%s", scheme, endpoint)
		}

		response = append(response, ServiceResponse{
			Service:  svc,
			Endpoint: endpoint,
			URL:      url,
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    response,
	})
}

// GetService handles GET /api/services/:name
func (h *ServiceHandler) GetService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	svc, err := h.repos.Service.GetByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Service '%s' not found", name),
			})
		}
		h.logger.Error("Failed to get service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve service",
		})
	}

	// CRITICAL FIX: Update status from REAL container state before returning
	// This ensures buttons show correct state after start/stop/restart operations
	if svc.Enabled {
		// Extract container name from config
		var containerName string
		if nameVal, ok := svc.Config["ContainerName"]; ok {
			if nameStr, ok := nameVal.(string); ok && nameStr != "" {
				containerName = nameStr
			}
		}

		// If no container name in config, use service name
		if containerName == "" {
			containerName = name
		}

		h.logger.Info("🔍 GetService: Checking container status",
			"service", name,
			"container_name", containerName,
			"db_status", svc.Status)

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// Try to get container directly by name (doesn't require labels)
		realStatus := "stopped" // Default if not found
		containerFound := false

		// First, try to inspect the container directly by name
		containerInfo, err := h.dockerClient.GetContainer(ctx, containerName)
		if err == nil && containerInfo != nil {
			// Container found via direct inspect
			realStatus = string(containerInfo.Status)
			containerFound = true
			h.logger.Info("✅ Container found (direct inspect)",
				"service", name,
				"container", containerName,
				"status", realStatus)
		} else {
			// Fallback: list all containers and search
			h.logger.Debug("🔍 Container not found by name, searching all containers",
				"service", name,
				"container_name", containerName)

			containers, listErr := h.dockerClient.ListContainers(ctx)
			if listErr == nil {
				h.logger.Debug("🐳 Docker containers listed",
					"service", name,
					"total_containers", len(containers))

				for _, container := range containers {
					// Docker prefixes names with /, so strip it
					cName := container.Name
					if len(cName) > 0 && cName[0] == '/' {
						cName = cName[1:]
					}

					h.logger.Debug("Checking container",
						"service", name,
						"container_name", cName,
						"looking_for", containerName,
						"status", string(container.Status))

					if cName == containerName {
						// Container found in list
						realStatus = string(container.Status)
						containerFound = true
						h.logger.Info("✅ Container found (in list)",
							"service", name,
							"container", containerName,
							"status", realStatus)
						break
					}
				}
			} else {
				h.logger.Error("❌ Failed to list Docker containers",
					"service", name,
					"error", listErr)
			}
		}

		if !containerFound {
			h.logger.Warn("⚠️ Container NOT found in Docker",
				"service", name,
				"container_name", containerName,
				"setting_status_to", "stopped")
		}

		// Update service with real status
		if svc.Status != realStatus {
			h.logger.Info("🔄 Status changed",
				"service", name,
				"old_status", svc.Status,
				"new_status", realStatus)
			svc.Status = realStatus
			// Update in database asynchronously (don't block response)
			// Note: Added context with timeout to prevent goroutine leaks
			go func(id uint, status string) {
				if updateErr := h.repos.Service.UpdateStatus(id, status, ""); updateErr != nil {
					h.logger.Warn("Failed to update service status in DB", "name", name, "error", updateErr)
				}
			}(svc.ID, realStatus)
		} else {
			h.logger.Debug("ℹ️ Status unchanged",
				"service", name,
				"status", svc.Status)
		}

		// Add volume mount information from container inspect if available
		if containerInfo != nil && len(containerInfo.Mounts) > 0 {
			mounts := make(map[string]string)
			for _, mount := range containerInfo.Mounts {
				mounts[mount.Destination] = mount.Source
			}
			// Add mounts to response data
			return c.JSON(APIResponse{
				Success: true,
				Data: fiber.Map{
					"id":           svc.ID,
					"name":         svc.Name,
					"display_name": svc.DisplayName,
					"enabled":      svc.Enabled,
					"status":       svc.Status,
					"image":        svc.Image,
					"port":         svc.Port,
					"config":       svc.Config,
					"mounts":       mounts,
					"created_at":   svc.CreatedAt,
					"updated_at":   svc.UpdatedAt,
				},
			})
		}
	} else {
		h.logger.Debug("⏸️ Service not enabled, skipping status check",
			"service", name)
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    svc,
	})
}

// EnableService handles POST /api/services/:name/enable
func (h *ServiceHandler) EnableService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to enable service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.serviceManager.EnableService(ctx, name); err != nil {
		h.logger.Error("Failed to enable service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to enable service: %v", err),
		})
	}

	h.logger.Info("Service enabled", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service enabled successfully"},
	})
}

// DisableService handles POST /api/services/:name/disable
func (h *ServiceHandler) DisableService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to disable service
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.serviceManager.DisableService(ctx, name); err != nil {
		h.logger.Error("Failed to disable service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to disable service: %v", err),
		})
	}

	h.logger.Info("Service disabled", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service disabled successfully"},
	})
}

// StartContainer handles POST /api/services/:name/start
func (h *ServiceHandler) StartContainer(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to start service
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := h.serviceManager.StartService(ctx, name); err != nil {
		h.logger.Error("Failed to start service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to start service: %v", err),
		})
	}

	h.logger.Info("Service started", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service started successfully"},
	})
}

// StopContainer handles POST /api/services/:name/stop
func (h *ServiceHandler) StopContainer(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to stop service
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	if err := h.serviceManager.StopService(ctx, name); err != nil {
		h.logger.Error("Failed to stop service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to stop service: %v", err),
		})
	}

	h.logger.Info("Service stopped", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service stopped successfully"},
	})
}

// UpdateService handles POST /api/services/:name/update
// Pulls the latest image and restarts the service
func (h *ServiceHandler) UpdateService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to update service (pull image + restart)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute) // 5 min for image pull
	defer cancel()

	h.logger.Info("Updating service", "name", name)

	if err := h.serviceManager.UpdateService(ctx, name); err != nil {
		h.logger.Error("Failed to update service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update service: %v", err),
		})
	}

	h.logger.Info("Service updated successfully", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service updated and restarted successfully"},
	})
}

// RestartContainer handles POST /api/services/:name/restart
func (h *ServiceHandler) RestartContainer(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to restart service
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := h.serviceManager.RestartService(ctx, name); err != nil {
		h.logger.Error("Failed to restart service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to restart service: %v", err),
		})
	}

	h.logger.Info("Service restarted", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service restarted successfully"},
	})
}

// KillAndDownService handles POST /api/services/:name/kill-down
// Force kills the container and removes it
func (h *ServiceHandler) KillAndDownService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Get service from database to log the action
	svc, err := h.repos.Service.GetByName(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Service not found: %v", err),
		})
	}

	// Use context with timeout for force kill operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call StopService which uses docker compose stop
	// This is equivalent to kill & down for our purposes
	if err := h.serviceManager.StopService(ctx, name); err != nil {
		h.logger.Error("Failed to kill and down service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to kill service: %v", err),
		})
	}

	// Log action
	h.repos.ServiceLog.Create(&models.ServiceLog{
		ServiceID: svc.ID,
		Action:    "kill-down",
		Status:    "success",
		Message:   "Container force stopped and will be removed on next enable",
	})

	h.logger.Info("Service force killed and marked for removal", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Container force killed successfully"},
	})
}

// UpdateServiceConfig handles PUT /api/services/:name/config
func (h *ServiceHandler) UpdateServiceConfig(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	var configUpdate models.ServiceConfig
	if err := c.BodyParser(&configUpdate); err != nil {
		h.logger.Error("Failed to parse config update", "name", name, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// If HostPort is not provided or is 0, explicitly remove it from config
	// This allows switching from dynamic (with port) to static (without port) compose
	if hostPort, exists := configUpdate["HostPort"]; !exists || hostPort == nil || hostPort == 0 || hostPort == float64(0) {
		// Explicitly mark for deletion by setting to nil
		configUpdate["HostPort"] = nil
		h.logger.Info("HostPort not provided or is 0, will be removed from config", "name", name)
	}

	// Debug: Log received configuration
	h.logger.Info("Received config update",
		"name", name,
		"config", configUpdate,
		"paths", configUpdate["Paths"])

	// Ensure Paths exists and has Config path
	if configUpdate["Paths"] == nil {
		configUpdate["Paths"] = make(map[string]interface{})
	}
	paths, ok := configUpdate["Paths"].(map[string]interface{})
	if !ok {
		paths = make(map[string]interface{})
		configUpdate["Paths"] = paths
	}

	// Set default Config path if not provided
	if paths["Config"] == nil || paths["Config"] == "" {
		paths["Config"] = fmt.Sprintf("./volumes/service-configs/%s/", name)
		h.logger.Info("Using default config path",
			"name", name,
			"config_path", paths["Config"])
	} else {
		h.logger.Info("Using user-provided config path",
			"name", name,
			"config_path", paths["Config"])
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Use service manager to update configuration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := h.serviceManager.UpdateServiceConfig(ctx, name, configUpdate); err != nil {
		h.logger.Error("Failed to update service config", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update service configuration: %v", err),
		})
	}

	h.logger.Info("Service config updated", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service configuration updated successfully"},
	})
}

// GetContainerLogs handles GET /api/services/:name/logs
func (h *ServiceHandler) GetContainerLogs(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	svc, err := h.repos.Service.GetByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Service '%s' not found", name),
			})
		}
		h.logger.Error("Failed to get service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve service",
		})
	}

	if svc.ContainerID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service has no container associated",
		})
	}

	// Get tail parameter (default to 100 lines)
	tail := c.Query("tail", "100")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logs, err := h.dockerClient.GetContainerLogs(ctx, svc.ContainerID, tail)
	if err != nil {
		h.logger.Error("Failed to get container logs", "name", name, "container_id", svc.ContainerID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to retrieve container logs: %v", err),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"logs": logs,
		},
	})
}

// ConfigPage renders the service configuration page
func (h *ServiceHandler) ConfigPage(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).Render("pages/error", fiber.Map{
			"Title":   "Invalid Request",
			"Message": "Service name is required",
			"Version": "dev",
		}, "layouts/main")
	}

	// Always render config page, regardless of service existence in database
	return c.Render("pages/service_config", fiber.Map{
		"Title":       name,
		"ServiceName": name,
		"Version":     "dev",
	}, "layouts/main")
}

// CheckConfigExists handles GET /api/services/:name/config-exists
// Checks if the service's config directory exists and has files
func (h *ServiceHandler) CheckConfigExists(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Build config path (inside container)
	configPath := filepath.Join("/app/data/services-volumes", name)

	// Check if directory exists and has files
	exists := false
	if info, err := os.Stat(configPath); err == nil && info.IsDir() {
		// Check if directory has any files (not just empty directory)
		entries, err := os.ReadDir(configPath)
		if err == nil && len(entries) > 0 {
			exists = true
		}
	}

	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"exists": exists,
			"path":   configPath,
		},
	})
}

// ResetService handles POST /api/services/:name/reset
// Prunes service: docker compose down -v, deletes config directories and disables the service
func (h *ServiceHandler) ResetService(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	h.logger.Info("Resetting service configuration", "name", name)

	// Use service manager to reset service (deletes config and recreates container)
	if err := h.serviceManager.ResetService(ctx, name); err != nil {
		h.logger.Error("Failed to reset service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to reset service: %v", err),
		})
	}

	h.logger.Info("Service prune completed", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service pruned and disabled successfully"},
	})
}

// GetRadarrConfig handles GET /api/services/:name/radarr-config
func (h *ServiceHandler) GetRadarrConfig(c *fiber.Ctx) error {
	name := c.Params("name")

	if name != "radarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Radarr service",
		})
	}

	config, err := h.serviceManager.GetRadarrConfig(c.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get Radarr config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get Radarr config: %v", err),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"config": config},
	})
}

// UpdateRadarrConfig handles PUT /api/services/:name/radarr-config
func (h *ServiceHandler) UpdateRadarrConfig(c *fiber.Ctx) error {
	name := c.Params("name")

	if name != "radarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Radarr service",
		})
	}

	var config models.RadarrConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	h.logger.Info("Received Radarr config update",
		"name", name,
		"username", config.Username,
		"has_password", config.Password != "")

	if err := h.serviceManager.UpdateRadarrConfig(c.Context(), name, config); err != nil {
		h.logger.Error("Failed to update Radarr config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update Radarr config: %v", err),
		})
	}

	h.logger.Info("Radarr config updated successfully", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Configuration saved successfully. Radarr is restarting to apply changes..."},
	})
}

// GetSonarrConfig handles GET /api/services/:name/sonarr-config
func (h *ServiceHandler) GetSonarrConfig(c *fiber.Ctx) error {
	name := c.Params("name")

	if name != "sonarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Sonarr service",
		})
	}

	config, err := h.serviceManager.GetSonarrConfig(c.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get Sonarr config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get Sonarr config: %v", err),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"config": config},
	})
}

// UpdateSonarrConfig handles PUT /api/services/:name/sonarr-config
func (h *ServiceHandler) UpdateSonarrConfig(c *fiber.Ctx) error {
	name := c.Params("name")

	if name != "sonarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Sonarr service",
		})
	}

	var config models.SonarrConfig
	if err := c.BodyParser(&config); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	h.logger.Info("Received Sonarr config update",
		"name", name,
		"username", config.Username,
		"has_password", config.Password != "")

	if err := h.serviceManager.UpdateSonarrConfig(c.Context(), name, config); err != nil {
		h.logger.Error("Failed to update Sonarr config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to update Sonarr config: %v", err),
		})
	}

	h.logger.Info("Sonarr config updated successfully", "name", name)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Configuration saved successfully. Sonarr is restarting to apply changes..."},
	})
}

// CheckServiceReady handles GET /api/services/:name/ready
// Performs a lightweight readiness check for a service. Strategy:
// 1) For radarr: Try Radarr API health endpoint (fastest, most reliable)
// 2) If proxy endpoint is configured, attempt HTTP HEAD to that URL (http/https per proxy config)
// 3) Fallback to internal container address on known port
// Returns { ready: bool, via: "api"|"proxy"|"internal"|"none", status_code: int, url: string }
func (h *ServiceHandler) CheckServiceReady(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Check if service exists in database first to avoid unnecessary queries
	_, err := h.repos.Service.GetByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Service not configured yet - just check container availability directly
			return h.checkServiceReadyDirect(c, name)
		}
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to check service",
		})
	}

	// Prepare HTTP client with increased timeout (3 seconds)
	// Services may be slow to start or under load, so we need a reasonable timeout
	client := &http.Client{Timeout: 3 * time.Second}

	// Helper to try a URL with HEAD request (faster than GET)
	tryURL := func(url string) (bool, int, error) {
		req, err := http.NewRequest(http.MethodHead, url, nil)
		if err != nil {
			return false, 0, err
		}
		resp, err := client.Do(req)
		if err != nil {
			// If HEAD fails, try GET as fallback
			req, _ = http.NewRequest(http.MethodGet, url, nil)
			resp, err = client.Do(req)
			if err != nil {
				return false, 0, err
			}
		}
		defer resp.Body.Close()
		// Consider 2xx and 3xx as ready (some services redirect)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return true, resp.StatusCode, nil
		}
		return false, resp.StatusCode, nil
	}

	// For Radarr: Try API health endpoint with API key from config
	if name == "radarr" {
		// Get API key from Radarr config
		var apiKey string
		if h.serviceManager != nil {
			if cfg, err := h.serviceManager.GetRadarrConfig(c.Context(), name); err == nil && cfg.ApiKey != "" {
				apiKey = cfg.ApiKey
			}
		}

		// Only try API endpoint if we have the API key
		if apiKey != "" {
			apiURL := fmt.Sprintf("http://%s:7878/api/v3/health", name)
			req, err := http.NewRequest(http.MethodGet, apiURL, nil)
			if err == nil {
				req.Header.Set("X-Api-Key", apiKey)
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						return c.JSON(APIResponse{
							Success: true,
							Data: fiber.Map{
								"ready":       true,
								"via":         "api",
								"status_code": resp.StatusCode,
								"url":         apiURL,
							},
						})
					}
				}
			}
		}

		// If API check didn't work, just check if port is responding (TCP check only)
		// This avoids authentication challenges
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:7878", name), 1*time.Second)
		if err == nil {
			conn.Close()
			return c.JSON(APIResponse{
				Success: true,
				Data: fiber.Map{
					"ready":       true,
					"via":         "tcp",
					"status_code": 0,
					"url":         fmt.Sprintf("http://%s:7878", name),
				},
			})
		}
	}

	// For Sonarr: Try API health endpoint with API key from config
	if name == "sonarr" {
		// Get API key from Sonarr config
		var apiKey string
		if h.serviceManager != nil {
			if cfg, err := h.serviceManager.GetSonarrConfig(c.Context(), name); err == nil && cfg.ApiKey != "" {
				apiKey = cfg.ApiKey
			}
		}

		// Only try API endpoint if we have the API key
		if apiKey != "" {
			apiURL := fmt.Sprintf("http://%s:8989/api/v3/health", name)
			req, err := http.NewRequest(http.MethodGet, apiURL, nil)
			if err == nil {
				req.Header.Set("X-Api-Key", apiKey)
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 400 {
						return c.JSON(APIResponse{
							Success: true,
							Data: fiber.Map{
								"ready":       true,
								"via":         "api",
								"status_code": resp.StatusCode,
								"url":         apiURL,
							},
						})
					}
				}
			}
		}

		// If API check didn't work, just check if port is responding (TCP check only)
		// This avoids authentication challenges
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:8989", name), 1*time.Second)
		if err == nil {
			conn.Close()
			return c.JSON(APIResponse{
				Success: true,
				Data: fiber.Map{
					"ready":       true,
					"via":         "tcp",
					"status_code": 0,
					"url":         fmt.Sprintf("http://%s:8989", name),
				},
			})
		}
	}

	// Then, attempt via proxy endpoint if available
	proxySvc := service.NewProxyService(h.repos, h.logger)
	endpoint, err := proxySvc.GetServiceEndpoint(name)
	if err == nil && endpoint != "" {
		// Determine scheme from proxy config
		proxyCfg, _ := proxySvc.GetProxyConfig()
		scheme := "http"
		if proxyCfg != nil && proxyCfg.SSLEnabled {
			scheme = "https"
		}
		url := fmt.Sprintf("%s://%s", scheme, endpoint)
		if ok, code, _ := tryURL(url); ok {
			return c.JSON(APIResponse{
				Success: true,
				Data: fiber.Map{
					"ready":       true,
					"via":         "proxy",
					"status_code": code,
					"url":         url,
				},
			})
		}
	}

	// Fallback for specific services using internal container address
	if name == "radarr" {
		internalURL := fmt.Sprintf("http://%s:7878", name)
		if ok, code, _ := tryURL(internalURL); ok {
			return c.JSON(APIResponse{
				Success: true,
				Data: fiber.Map{
					"ready":       true,
					"via":         "internal",
					"status_code": code,
					"url":         internalURL,
				},
			})
		}
	}

	// Not ready
	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"ready":       false,
			"via":         "none",
			"status_code": 0,
		},
	})
}

// checkServiceReadyDirect checks if a service container is ready without database config
// Used when service is not yet configured in the database
func (h *ServiceHandler) checkServiceReadyDirect(c *fiber.Ctx, name string) error {
	// Check if container port is responding (TCP check only)
	var port int
	switch name {
	case "radarr":
		port = 7878
	case "sonarr":
		port = 8989
	case "jellyfin":
		port = 8096
	case "prowlarr":
		port = 9696
	case "bazarr":
		port = 6767
	case "qbittorrent":
		port = 8080
	default:
		// Unknown service, assume not ready
		return c.JSON(APIResponse{
			Success: true,
			Data: fiber.Map{
				"ready": false,
				"via":   "unknown",
			},
		})
	}

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", name, port), 1*time.Second)
	if err == nil {
		conn.Close()
		return c.JSON(APIResponse{
			Success: true,
			Data: fiber.Map{
				"ready":       true,
				"via":         "tcp",
				"status_code": 0,
				"url":         fmt.Sprintf("http://%s:%d", name, port),
			},
		})
	}

	// Not ready
	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"ready": false,
			"via":   "tcp-failed",
		},
	})
}

// getServiceConfigPath attempts to extract the config path from the stored service configuration
func getServiceConfigPath(svc *models.Service) string {
	if svc == nil || svc.Config == nil {
		return ""
	}

	if pathsRaw, ok := svc.Config["Paths"]; ok && pathsRaw != nil {
		switch paths := pathsRaw.(type) {
		case map[string]interface{}:
			if configPath, ok := paths["Config"].(string); ok {
				return configPath
			}
		case models.ServiceConfig:
			if configPath, ok := paths["Config"].(string); ok {
				return configPath
			}
		case map[string]string:
			if configPath, ok := paths["Config"]; ok {
				return configPath
			}
		}
	}

	if configPath, ok := svc.Config["ConfigPath"].(string); ok {
		return configPath
	}

	return ""
}

// GetRootFolders handles GET /api/services/:name/rootfolders
func (h *ServiceHandler) GetRootFolders(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Only supported for Radarr and Sonarr
	if name != "radarr" && name != "sonarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Radarr and Sonarr services",
		})
	}

	// Check if service manager is available
	if h.serviceManager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Service manager is not available",
		})
	}

	// Get root folders from service via API
	ctx := c.Context()
	rootFolders, err := h.serviceManager.GetRootFolders(ctx, name)
	if err != nil {
		h.logger.Error("Failed to get root folders", "service", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get root folders: %v", err),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    rootFolders,
	})
}
