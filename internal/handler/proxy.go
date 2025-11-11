package handler

import (
	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

// ProxyHandler handles proxy configuration HTTP requests
type ProxyHandler struct {
	repos        *repository.Repositories
	logger       *logger.Logger
	proxyService *service.ProxyService
}

// NewProxyHandler creates a new ProxyHandler instance
func NewProxyHandler(repos *repository.Repositories, logger *logger.Logger) *ProxyHandler {
	return &ProxyHandler{
		repos:        repos,
		logger:       logger,
		proxyService: service.NewProxyService(repos, logger),
	}
}

// GetProxyConfig handles GET /api/proxy/config
func (h *ProxyHandler) GetProxyConfig(c *fiber.Ctx) error {
	config, err := h.proxyService.GetProxyConfig()
	if err != nil {
		h.logger.Error("Failed to get proxy config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve proxy configuration",
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    config,
	})
}

// UpdateProxyConfig handles PUT /api/proxy/config
func (h *ProxyHandler) UpdateProxyConfig(c *fiber.Ctx) error {
	var config models.ProxyConfig
	if err := c.BodyParser(&config); err != nil {
		h.logger.Error("Failed to parse proxy config", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.proxyService.UpdateProxyConfig(&config); err != nil {
		h.logger.Error("Failed to update proxy config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to update proxy configuration",
		})
	}

	h.logger.Info("Proxy config updated")
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Proxy configuration updated successfully"},
	})
}

// AddDomain handles POST /api/proxy/domains
func (h *ProxyHandler) AddDomain(c *fiber.Ctx) error {
	type DomainRequest struct {
		Name      string `json:"name"`
		IsPrimary bool   `json:"is_primary"`
	}

	var req DomainRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse domain request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Domain name is required",
		})
	}

	domain, err := h.proxyService.AddDomain(req.Name, req.IsPrimary)
	if err != nil {
		h.logger.Error("Failed to add domain", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(APIResponse{
		Success: true,
		Data:    domain,
	})
}

// GetDomains handles GET /api/proxy/domains
func (h *ProxyHandler) GetDomains(c *fiber.Ctx) error {
	domains, err := h.proxyService.GetDomains()
	if err != nil {
		h.logger.Error("Failed to get domains", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve domains",
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    domains,
	})
}

// DeleteDomain handles DELETE /api/proxy/domains/:name
func (h *ProxyHandler) DeleteDomain(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Domain name is required",
		})
	}

	if err := h.proxyService.DeleteDomain(name); err != nil {
		h.logger.Error("Failed to delete domain", "domain", name, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Domain deleted successfully"},
	})
}

// UpdateServiceSubdomain handles PUT /api/services/:name/subdomain
func (h *ProxyHandler) UpdateServiceSubdomain(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	type SubdomainRequest struct {
		Subdomain string `json:"subdomain"`
		Domain    string `json:"domain"`
	}

	var req SubdomainRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse subdomain request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Check for conflicts if subdomain is provided
	if req.Subdomain != "" {
		domain := req.Domain
		if domain == "" {
			// Use primary domain for conflict check
			primaryDomain, err := h.repos.Proxy.GetPrimaryDomain()
			if err != nil || primaryDomain == nil {
				return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
					Success: false,
					Error:   "No primary domain configured",
				})
			}
			domain = primaryDomain.Name
		}

		hasConflict, conflictService, err := h.proxyService.CheckSubdomainConflict(req.Subdomain, domain, serviceName)
		if err != nil {
			h.logger.Error("Failed to check subdomain conflict", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Error:   "Failed to validate subdomain",
			})
		}

		if hasConflict {
			return c.Status(fiber.StatusConflict).JSON(APIResponse{
				Success: false,
				Error:   "Subdomain already in use by service: " + conflictService,
			})
		}
	}

	if err := h.proxyService.UpdateServiceSubdomain(serviceName, req.Subdomain, req.Domain); err != nil {
		h.logger.Error("Failed to update service subdomain", "service", serviceName, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Service subdomain updated successfully"},
	})
}

// GetServiceEndpoint handles GET /api/services/:name/endpoint
func (h *ProxyHandler) GetServiceEndpoint(c *fiber.Ctx) error {
	serviceName := c.Params("name")
	if serviceName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Service name is required",
		})
	}

	endpoint, err := h.proxyService.GetServiceEndpoint(serviceName)
	if err != nil {
		h.logger.Error("Failed to get service endpoint", "service", serviceName, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   err.Error(),
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"endpoint": endpoint,
			"url":      "http://" + endpoint,
		},
	})
}
