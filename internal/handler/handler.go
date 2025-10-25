package handler

import (
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// Handler holds all HTTP handlers
type Handler struct {
	serviceManager *service.ServiceManager
	logger         *zap.Logger
}

// New creates a new handler instance
func New(serviceManager *service.ServiceManager, logger *zap.Logger) *Handler {
	return &Handler{
		serviceManager: serviceManager,
		logger:         logger,
	}
}

// Home renders the home page
func (h *Handler) Home(c *fiber.Ctx) error {
	services, err := h.serviceManager.GetAllServices()
	if err != nil {
		h.logger.Error("Failed to get services", zap.Error(err))
		return c.Status(500).Render("error", fiber.Map{
			"error": "Failed to load services",
		})
	}

	return c.Render("index", fiber.Map{
		"services": services,
		"title":    "MediaCheky - Multimedia Server Manager",
	})
}

// Services returns the services management page
func (h *Handler) Services(c *fiber.Ctx) error {
	services, err := h.serviceManager.GetAllServices()
	if err != nil {
		h.logger.Error("Failed to get services", zap.Error(err))
		return c.Status(500).SendString("Failed to load services")
	}

	return c.Render("services", fiber.Map{
		"services": services,
		"title":    "Services - MediaCheky",
	})
}

// ToggleService handles enabling/disabling a service
func (h *Handler) ToggleService(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	
	type ToggleRequest struct {
		Enabled bool `json:"enabled"`
	}
	
	var req ToggleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request",
		})
	}

	if err := h.serviceManager.ToggleService(serviceName, req.Enabled); err != nil {
		h.logger.Error("Failed to toggle service", 
			zap.String("service", serviceName), 
			zap.Error(err))
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to toggle service",
		})
	}

	h.logger.Info("Service toggled", 
		zap.String("service", serviceName), 
		zap.Bool("enabled", req.Enabled))

	return c.JSON(fiber.Map{
		"success": true,
		"enabled": req.Enabled,
	})
}

// GetDockerCompose generates and returns docker-compose configuration
func (h *Handler) GetDockerCompose(c *fiber.Ctx) error {
	compose, err := h.serviceManager.GenerateDockerCompose()
	if err != nil {
		h.logger.Error("Failed to generate docker-compose", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate configuration",
		})
	}

	return c.SendString(compose)
}

// DownloadDockerCompose generates and downloads docker-compose.yml
func (h *Handler) DownloadDockerCompose(c *fiber.Ctx) error {
	compose, err := h.serviceManager.GenerateDockerCompose()
	if err != nil {
		h.logger.Error("Failed to generate docker-compose", zap.Error(err))
		return c.Status(500).SendString("Failed to generate configuration")
	}

	c.Set("Content-Type", "application/x-yaml")
	c.Set("Content-Disposition", "attachment; filename=docker-compose.yml")
	
	return c.SendString(compose)
}
