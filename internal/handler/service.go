package handler

import (
	"context"
	"fmt"
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
}

// NewServiceHandler creates a new ServiceHandler instance
func NewServiceHandler(repos *repository.Repositories, logger *logger.Logger, dockerClient *service.DockerClient, serviceManager *service.ServiceManager) *ServiceHandler {
	return &ServiceHandler{
		repos:          repos,
		logger:         logger,
		dockerClient:   dockerClient,
		serviceManager: serviceManager,
	}
}

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

	return c.JSON(APIResponse{
		Success: true,
		Data:    services,
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

	// Verify service exists
	_, err := h.repos.Service.GetByName(name)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).Render("pages/error", fiber.Map{
				"Title":   "Service Not Found",
				"Message": fmt.Sprintf("Service '%s' not found", name),
				"Version": "dev",
			}, "layouts/main")
		}
		h.logger.Error("Failed to get service", "name", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).Render("pages/error", fiber.Map{
			"Title":   "Server Error",
			"Message": "Failed to retrieve service information",
			"Version": "dev",
		}, "layouts/main")
	}

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
		Data:    fiber.Map{"message": "Radarr configuration updated successfully"},
	})
}

// CheckServiceReady handles GET /api/services/:name/ready
// Performs a lightweight readiness check for a service. Strategy:
// 1) If proxy endpoint is configured, attempt HTTP GET to that URL (http/https per proxy config)
// 2) For specific services (radarr), fallback to internal container address on known port
// Returns { ready: bool, via: "proxy"|"internal"|"none", status_code: int, url: string }
func (h *ServiceHandler) CheckServiceReady(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Prepare HTTP client with short timeout
	client := &http.Client{Timeout: 2 * time.Second}

	// Helper to try a URL
	tryURL := func(url string) (bool, int, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return false, 0, err
		}
		resp, err := client.Do(req)
		if err != nil {
			return false, 0, err
		}
		defer resp.Body.Close()
		// Consider 2xx and 3xx as ready (some services redirect)
		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			return true, resp.StatusCode, nil
		}
		return false, resp.StatusCode, nil
	}

	// First, attempt via proxy endpoint if available
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

	// Fallbacks for specific services using internal container address
	// Only implement for radarr for now (internal default port 7878, or from config.xml if available)
	if name == "radarr" {
		port := 7878
		// Try to read configured port from Radarr config if available
		if h.serviceManager != nil {
			if cfg, err := h.serviceManager.GetRadarrConfig(c.Context(), name); err == nil && cfg.Port > 0 {
				port = cfg.Port
			}
		}
		internalURL := fmt.Sprintf("http://%s:%d", name, port)
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
