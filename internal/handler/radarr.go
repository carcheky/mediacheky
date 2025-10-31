package handler

import (
	"strconv"
	"time"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type RadarrHandler struct {
	config *config.Config
	logger *logger.Logger
	client *clients.RadarrClient
}

func NewRadarrHandler(cfg *config.Config, appLogger *logger.Logger) *RadarrHandler {
	var radarrClient *clients.RadarrClient

	if cfg.Clients.Radarr.Enabled {
		zapLogger := appLogger.Desugar()
		radarrClient = clients.NewRadarrClient(
			clients.ClientConfig{
				BaseURL: cfg.Clients.Radarr.URL,
				APIKey:  cfg.Clients.Radarr.APIKey,
				Timeout: 30 * time.Second,
			},
			zapLogger,
		)
	}

	return &RadarrHandler{
		config: cfg,
		logger: appLogger,
		client: radarrClient,
	}
}

// GetSystemInfo retrieves Radarr system information.
func (h *RadarrHandler) GetSystemInfo(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Radarr client not configured",
		})
	}

	info, err := h.client.GetSystemInfo(c.Context())
	if err != nil {
		h.logger.Error("Failed to get Radarr system info", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve system info: " + err.Error(),
		})
	}

	return c.JSON(info)
}

// GetQueue retrieves the current Radarr download queue.
func (h *RadarrHandler) GetQueue(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Radarr client not configured",
		})
	}

	queue, err := h.client.GetQueue(c.Context())
	if err != nil {
		h.logger.Error("Failed to get Radarr queue", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve queue: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total": len(queue),
		"items": queue,
	})
}

// GetHistory retrieves recent Radarr history.
func (h *RadarrHandler) GetHistory(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Radarr client not configured",
		})
	}

	// Parse page size from query parameter
	pageSize := 50
	if ps := c.Query("pageSize"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	history, err := h.client.GetHistory(c.Context(), pageSize)
	if err != nil {
		h.logger.Error("Failed to get Radarr history", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve history: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total": len(history),
		"items": history,
	})
}

// GetCalendar retrieves upcoming movies from Radarr.
func (h *RadarrHandler) GetCalendar(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Radarr client not configured",
		})
	}

	// Default to next 30 days
	startDate := time.Now()
	endDate := time.Now().AddDate(0, 0, 30)

	// Parse optional date range from query parameters
	if start := c.Query("start"); start != "" {
		if parsed, err := time.Parse("2006-01-02", start); err == nil {
			startDate = parsed
		}
	}
	if end := c.Query("end"); end != "" {
		if parsed, err := time.Parse("2006-01-02", end); err == nil {
			endDate = parsed
		}
	}

	calendar, err := h.client.GetCalendar(c.Context(), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get Radarr calendar", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve calendar: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"total":      len(calendar),
		"items":      calendar,
	})
}

// GetQualityProfiles retrieves available quality profiles from Radarr.
func (h *RadarrHandler) GetQualityProfiles(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Radarr client not configured",
		})
	}

	profiles, err := h.client.GetQualityProfiles(c.Context())
	if err != nil {
		h.logger.Error("Failed to get Radarr quality profiles", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve quality profiles: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total":    len(profiles),
		"profiles": profiles,
	})
}
