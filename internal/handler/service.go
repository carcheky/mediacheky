package handler

import (
	"context"
	"fmt"
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
