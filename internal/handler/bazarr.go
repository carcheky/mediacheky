package handler

import (
	"strconv"
	"time"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type BazarrHandler struct {
	config *config.Config
	logger *logger.Logger
	client *clients.BazarrClient
}

func NewBazarrHandler(cfg *config.Config, appLogger *logger.Logger) *BazarrHandler {
	var bazarrClient *clients.BazarrClient

	if cfg.Clients.Bazarr.Enabled {
		zapLogger := appLogger.Desugar()
		bazarrClient = clients.NewBazarrClient(
			clients.ClientConfig{
				BaseURL: cfg.Clients.Bazarr.URL,
				APIKey:  cfg.Clients.Bazarr.APIKey,
				Timeout: 30 * time.Second,
			},
			zapLogger,
		)
	}

	return &BazarrHandler{
		config: cfg,
		logger: appLogger,
		client: bazarrClient,
	}
}

// GetSystemInfo retrieves Bazarr system information.
func (h *BazarrHandler) GetSystemInfo(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
		})
	}

	info, err := h.client.GetSystemInfo(c.Context())
	if err != nil {
		h.logger.Error("Failed to get Bazarr system info", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve system info: " + err.Error(),
		})
	}

	return c.JSON(info)
}

// GetHistory retrieves recent Bazarr subtitle download history.
func (h *BazarrHandler) GetHistory(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
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
		h.logger.Error("Failed to get Bazarr history", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve history: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total": len(history),
		"items": history,
	})
}

// GetWantedMovies retrieves movies with missing subtitles.
func (h *BazarrHandler) GetWantedMovies(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
		})
	}

	wanted, err := h.client.GetWantedMovies(c.Context())
	if err != nil {
		h.logger.Error("Failed to get wanted movies", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve wanted movies: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total": len(wanted),
		"items": wanted,
	})
}

// GetWantedSeries retrieves series with missing subtitles.
func (h *BazarrHandler) GetWantedSeries(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
		})
	}

	wanted, err := h.client.GetWantedSeries(c.Context())
	if err != nil {
		h.logger.Error("Failed to get wanted series", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve wanted series: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"total": len(wanted),
		"items": wanted,
	})
}

// GetMovieSubtitles retrieves subtitle information for a specific movie.
func (h *BazarrHandler) GetMovieSubtitles(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
		})
	}

	// Parse Radarr ID from path parameter
	radarrID, err := c.ParamsInt("radarr_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Radarr ID",
		})
	}

	subtitles, err := h.client.GetMovieSubtitles(c.Context(), radarrID)
	if err != nil {
		h.logger.Error("Failed to get movie subtitles", "radarr_id", radarrID, "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve movie subtitles: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"radarr_id": radarrID,
		"total":     len(subtitles),
		"subtitles": subtitles,
	})
}

// GetSeriesSubtitles retrieves subtitle information for a specific series.
func (h *BazarrHandler) GetSeriesSubtitles(c *fiber.Ctx) error {
	if h.client == nil {
		return c.Status(503).JSON(fiber.Map{
			"error": "Bazarr client not configured",
		})
	}

	// Parse Sonarr ID from path parameter
	sonarrID, err := c.ParamsInt("sonarr_id")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Sonarr ID",
		})
	}

	subtitles, err := h.client.GetSeriesSubtitles(c.Context(), sonarrID)
	if err != nil {
		h.logger.Error("Failed to get series subtitles", "sonarr_id", sonarrID, "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve series subtitles: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"sonarr_id": sonarrID,
		"total":     len(subtitles),
		"subtitles": subtitles,
	})
}
