package handler

import (
	"fmt"

	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ConfigHandler handles global configuration HTTP requests
type ConfigHandler struct {
	repos  *repository.Repositories
	logger *logger.Logger
}

// NewConfigHandler creates a new ConfigHandler instance
func NewConfigHandler(repos *repository.Repositories, logger *logger.Logger) *ConfigHandler {
	return &ConfigHandler{
		repos:  repos,
		logger: logger,
	}
}

// GetGlobalConfig handles GET /api/config/global
func (h *ConfigHandler) GetGlobalConfig(c *fiber.Ctx) error {
	configs, err := h.repos.Config.GetAll()
	if err != nil {
		h.logger.Error("Failed to get global config", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve global configuration",
		})
	}

	// Convert to map for easier consumption
	configMap := make(map[string]interface{})
	for _, cfg := range configs {
		if cfg.Category != "" {
			if configMap[cfg.Category] == nil {
				configMap[cfg.Category] = make(map[string]string)
			}
			configMap[cfg.Category].(map[string]string)[cfg.Key] = cfg.Value
		} else {
			configMap[cfg.Key] = cfg.Value
		}
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    configMap,
	})
}

// UpdateGlobalConfig handles PUT /api/config/global
func (h *ConfigHandler) UpdateGlobalConfig(c *fiber.Ctx) error {
	type ConfigEntry struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Category string `json:"category,omitempty"`
	}

	var configUpdates []ConfigEntry
	if err := c.BodyParser(&configUpdates); err != nil {
		h.logger.Error("Failed to parse config updates", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	// Validate input
	if len(configUpdates) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "No configuration entries provided",
		})
	}

	// Update each configuration entry
	for _, entry := range configUpdates {
		if entry.Key == "" {
			continue // Skip empty keys
		}

		if err := h.repos.Config.Set(entry.Key, entry.Value, entry.Category); err != nil {
			h.logger.Error("Failed to update config entry", "key", entry.Key, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Failed to update configuration key '%s'", entry.Key),
			})
		}
	}

	h.logger.Info("Global config updated", "count", len(configUpdates))
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Global configuration updated successfully"},
	})
}

// GetConfigValue handles GET /api/config/global/:key
func (h *ConfigHandler) GetConfigValue(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Configuration key is required",
		})
	}

	config, err := h.repos.Config.GetByKey(key)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusNotFound).JSON(APIResponse{
				Success: false,
				Error:   fmt.Sprintf("Configuration key '%s' not found", key),
			})
		}
		h.logger.Error("Failed to get config value", "key", key, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve configuration value",
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    config,
	})
}

// UpdateConfigValue handles PUT /api/config/global/:key
func (h *ConfigHandler) UpdateConfigValue(c *fiber.Ctx) error {
	key := c.Params("key")
	if key == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Configuration key is required",
		})
	}

	type ValueUpdate struct {
		Value    string `json:"value"`
		Category string `json:"category,omitempty"`
	}

	var update ValueUpdate
	if err := c.BodyParser(&update); err != nil {
		h.logger.Error("Failed to parse value update", "key", key, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Invalid request body",
		})
	}

	if err := h.repos.Config.Set(key, update.Value, update.Category); err != nil {
		h.logger.Error("Failed to update config value", "key", key, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to update configuration value",
		})
	}

	h.logger.Info("Config value updated", "key", key)
	return c.JSON(APIResponse{
		Success: true,
		Data:    fiber.Map{"message": "Configuration value updated successfully"},
	})
}

// Index renders the global configuration page
func (h *ConfigHandler) Index(c *fiber.Ctx) error {
	return c.Render("pages/global", fiber.Map{
		"Title":   "Global Variables",
		"Version": "dev",
	}, "layouts/main")
}
