package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
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

	// Extract and handle Radarr auth if provided
	var radarrAuth map[string]interface{}
	if auth, exists := configUpdate["radarrAuth"]; exists {
		if authMap, ok := auth.(map[string]interface{}); ok {
			radarrAuth = authMap
			// Remove from config before saving to database
			delete(configUpdate, "radarrAuth")
		}
	}

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

	configPathStr, _ := paths["Config"].(string)

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

	// If Radarr auth was provided, update it asynchronously
	if name == "radarr" && radarrAuth != nil {
		username, _ := radarrAuth["username"].(string)
		password, _ := radarrAuth["password"].(string)

		if username != "" && password != "" {
			updateRadarrAuthInBackground(username, password, configPathStr, h.logger)
		}
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

	// Prepare HTTP client with shorter timeout (1 second)
	client := &http.Client{Timeout: 1 * time.Second}

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

// RadarrAuthConfig handles GET /api/services/:name/radarr/auth
// Retrieves authentication configuration from Radarr
func (h *ServiceHandler) RadarrAuthConfig(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Only Radarr has this endpoint
	if name != "radarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Radarr",
		})
	}

	// Get service config
	svc, err := h.repos.Service.GetByName(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "Service not found",
		})
	}

	if !svc.Enabled {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service is not enabled",
		})
	}

	// Get API key from config.xml
	configPath := getServiceConfigPath(svc)
	apiKey, err := extractRadarrAPIKey(configPath)
	if err != nil {
		h.logger.Error("Failed to get Radarr API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve API key from Radarr",
		})
	}

	// Call Radarr API
	url := fmt.Sprintf("http://radarr:7878/api/v3/config/auth")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		h.logger.Error("Failed to create request", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to create API request",
		})
	}

	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call Radarr API", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to connect to Radarr",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Radarr API returned error", "status", resp.StatusCode)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Radarr returned status %d", resp.StatusCode),
		})
	}

	// Parse response
	var authConfig map[string]interface{}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("Failed to read response body", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to read Radarr response",
		})
	}

	if err := json.Unmarshal(body, &authConfig); err != nil {
		h.logger.Error("Failed to parse Radarr response", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to parse Radarr response",
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    authConfig,
	})
}

// UpdateRadarrAuthConfig handles PUT /api/services/:name/radarr/auth
// Updates authentication configuration in Radarr
func (h *ServiceHandler) UpdateRadarrAuthConfig(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	// Only Radarr has this endpoint
	if name != "radarr" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "This endpoint is only available for Radarr",
		})
	}

	// Parse request body
	var authConfig map[string]interface{}
	if err := c.BodyParser(&authConfig); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Get service config
	svc, err := h.repos.Service.GetByName(name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "Service not found",
		})
	}

	if !svc.Enabled {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service is not enabled",
		})
	}

	// Get API key from config.xml
	configPath := getServiceConfigPath(svc)
	apiKey, err := extractRadarrAPIKey(configPath)
	if err != nil {
		h.logger.Error("Failed to get Radarr API key", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve API key from Radarr",
		})
	}

	// Convert to JSON
	bodyBytes, _ := json.Marshal(authConfig)

	// Call Radarr API
	url := fmt.Sprintf("http://radarr:7878/api/v3/config/auth")
	req, err := http.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
	if err != nil {
		h.logger.Error("Failed to create request", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to create API request",
		})
	}

	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		h.logger.Error("Failed to call Radarr API", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to connect to Radarr",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		h.logger.Error("Radarr API returned error", "status", resp.StatusCode)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Radarr returned status %d", resp.StatusCode),
		})
	}

	// Log success
	h.logger.Info("Radarr auth config updated successfully", "service", name)

	return c.JSON(APIResponse{
		Success: true,
		Data:    authConfig,
	})
}

// updateRadarrAuthInBackground updates Radarr authentication asynchronously
func updateRadarrAuthInBackground(username, password, configPath string, logger *logger.Logger) {
	// Use a goroutine to avoid blocking the response
	go func() {
		apiKey, err := extractRadarrAPIKey(configPath)
		if err != nil || apiKey == "" {
			logger.Warn("Failed to extract Radarr API key for auth update", "error", err)
			return
		}

		const url = "http://radarr:7878/api/v3/config/auth"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			logger.Warn("Failed to create Radarr auth request", "error", err)
			return
		}

		req.Header.Set("X-Api-Key", apiKey)
		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			logger.Warn("Failed to get Radarr auth config", "error", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("Radarr auth config returned non-OK status", "status", resp.StatusCode)
			return
		}

		var currentConfig map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &currentConfig); err != nil {
			logger.Warn("Failed to parse Radarr auth config", "error", err)
			return
		}

		// Update with new credentials
		currentConfig["username"] = username
		currentConfig["password"] = password

		// Send PUT request
		bodyBytes, _ := json.Marshal(currentConfig)
		putReq, _ := http.NewRequest("PUT", url, bytes.NewReader(bodyBytes))
		putReq.Header.Set("X-Api-Key", apiKey)
		putReq.Header.Set("Content-Type", "application/json")

		putResp, err := client.Do(putReq)
		if err != nil {
			logger.Warn("Failed to update Radarr auth", "error", err)
			return
		}
		putResp.Body.Close()

		if putResp.StatusCode == http.StatusOK {
			logger.Info("Radarr auth credentials updated successfully")
		} else {
			logger.Warn("Failed to update Radarr auth", "status", putResp.StatusCode)
		}
	}()
}

// extractRadarrAPIKey extracts the API key from Radarr's config.xml
func extractRadarrAPIKey(configPath string) (string, error) {
	// First try environment variable
	apiKey := os.Getenv("RADARR_API_KEY")
	if apiKey != "" {
		return apiKey, nil
	}

	var possiblePaths []string

	if configPath != "" {
		cleanPath := filepath.Clean(configPath)
		if filepath.Ext(cleanPath) == ".xml" {
			possiblePaths = append(possiblePaths, cleanPath)
		} else {
			possiblePaths = append(possiblePaths, filepath.Join(cleanPath, "config.xml"))
		}
	}

	// Possible paths where config.xml might be located
	possiblePaths = append(possiblePaths,
		// Docker volumes path (most common in docker-compose setup)
		"./volumes/services-volumes/radarr/config.xml",
		"/root/volumes/services-volumes/radarr/config.xml",
		"/config/config.xml", // Inside radarr container
	)

	if configPath != "" {
		possiblePaths = append(possiblePaths, filepath.Join(configPath, "config.xml"))
	}

	possiblePaths = append(possiblePaths, filepath.Join(os.Getenv("HOME"), ".config/Radarr/config.xml"))

	for _, path := range possiblePaths {
		if fileInfo, err := os.Stat(path); err == nil && !fileInfo.IsDir() {
			// File exists, try to read it
			content, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			// Parse XML to find ApiKey
			type Config struct {
				ApiKey string `xml:"ApiKey"`
			}
			var cfg Config
			if err := xml.Unmarshal(content, &cfg); err != nil {
				continue
			}

			if cfg.ApiKey != "" {
				return cfg.ApiKey, nil
			}
		}
	}

	return "", fmt.Errorf("RADARR_API_KEY not found in environment or config.xml")
}
