package handler

import (
	"strconv"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/cleanup"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type MediaHandler struct {
	repos          *repository.Repositories
	cleanupService *cleanup.CleanupService
	logger         *logger.Logger
}

func NewMediaHandler(repos *repository.Repositories, cleanupService *cleanup.CleanupService, logger *logger.Logger) *MediaHandler {
	return &MediaHandler{
		repos:          repos,
		cleanupService: cleanupService,
		logger:         logger,
	}
}

func (h *MediaHandler) List(c *fiber.Ctx) error {
	return c.Render("pages/media", fiber.Map{
		"Title": "Media Management - KeeperCheky",
	}, "layouts/main")
}

func (h *MediaHandler) GetAll(c *fiber.Ctx) error {
	// Get pagination parameters
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("pageSize", "20"))

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// Build filters from query parameters
	filters := repository.MediaFilters{
		Type:              c.Query("type", "all"),
		Status:            c.Query("status", "all"),
		Search:            c.Query("search", ""),
		Service:           c.Query("service", "all"),
		SizeRange:         c.Query("sizeRange", "all"),
		AddedDate:         c.Query("addedDate", "all"),
		SeedRatio:         c.Query("seedRatio", "all"),
		Quality:           c.Query("quality", ""),
		EpisodeCompletion: c.Query("episodeCompletion", "all"),
	}

	var media []models.Media
	var total int64
	var err error

	media, total, err = h.repos.Media.GetPaginatedWithFilters(page, pageSize, filters)

	if err != nil {
		h.logger.Error("Failed to get media", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve media",
		})
	}

	return c.JSON(fiber.Map{
		"data":     media,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"pages":    (total + int64(pageSize) - 1) / int64(pageSize),
	})
}

func (h *MediaHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid media ID",
		})
	}

	media, err := h.repos.Media.GetByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get media", "id", id, "error", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Media not found",
		})
	}

	return c.JSON(media)
}

func (h *MediaHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid media ID",
		})
	}

	// Parse delete options from request body
	var options cleanup.DeleteOptions
	if err := c.BodyParser(&options); err != nil {
		// If no body provided, use defaults (backward compatibility)
		options = cleanup.DeleteOptions{
			Radarr:      false,
			Sonarr:      false,
			Jellyfin:    false,
			DeleteFiles: false,
			QBittorrent: false,
		}
	}

	// Get media info before deletion
	media, err := h.repos.Media.GetByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get media", "id", id, "error", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Media not found",
		})
	}

	h.logger.Info("Deleting media with options",
		"id", id,
		"title", media.Title,
		"radarr", options.Radarr,
		"sonarr", options.Sonarr,
		"jellyfin", options.Jellyfin,
		"delete_files", options.DeleteFiles,
		"qbittorrent", options.QBittorrent,
	)

	// Use CleanupService to delete from services
	result, err := h.cleanupService.DeleteMedia(c.Context(), media, options)
	if err != nil {
		h.logger.Error("Failed to delete media from services", "id", id, "error", err)
		// Return the detailed result even if there were errors
		if result != nil {
			return c.Status(500).JSON(fiber.Map{
				"error":        "Failed to delete media from some services",
				"message":      err.Error(),
				"deleted_from": result.DeletedFrom,
				"errors":       result.Errors,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to delete media",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message":       "Media deleted successfully",
		"deleted_from":  result.DeletedFrom,
		"files_deleted": result.FilesDeleted,
	})
}

func (h *MediaHandler) Exclude(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid media ID",
		})
	}

	// Check if media exists
	media, err := h.repos.Media.GetByID(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Media not found",
		})
	}

	// Toggle excluded status
	newExcludedStatus := !media.Excluded
	if err := h.repos.Media.SetExcluded(uint(id), newExcludedStatus); err != nil {
		h.logger.Error("Failed to toggle exclude status", "id", id, "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update media",
		})
	}

	return c.JSON(fiber.Map{
		"message":  "Media status updated successfully",
		"excluded": newExcludedStatus,
	})
}

func (h *MediaHandler) GetStats(c *fiber.Ctx) error {
	stats, err := h.repos.Media.GetStats()
	if err != nil {
		h.logger.Error("Failed to get stats", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to retrieve statistics",
		})
	}

	return c.JSON(stats)
}

// BulkDelete deletes multiple media items at once
func (h *MediaHandler) BulkDelete(c *fiber.Ctx) error {
	var request struct {
		IDs     []uint                `json:"ids"`
		Options cleanup.DeleteOptions `json:"options"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(request.IDs) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "No media IDs provided",
		})
	}

	h.logger.Info("Bulk deleting media items",
		"count", len(request.IDs),
		"radarr", request.Options.Radarr,
		"sonarr", request.Options.Sonarr,
		"jellyfin", request.Options.Jellyfin,
		"delete_files", request.Options.DeleteFiles,
		"qbittorrent", request.Options.QBittorrent,
	)

	results := make(map[uint]interface{})
	successCount := 0
	failureCount := 0

	// Process deletions
	for _, id := range request.IDs {
		media, err := h.repos.Media.GetByID(id)
		if err != nil {
			h.logger.Error("Failed to get media for bulk delete", "id", id, "error", err)
			results[id] = fiber.Map{
				"success": false,
				"error":   "Media not found",
			}
			failureCount++
			continue
		}

		result, err := h.cleanupService.DeleteMedia(c.Context(), media, request.Options)
		if err != nil {
			h.logger.Error("Failed to delete media in bulk operation", "id", id, "title", media.Title, "error", err)
			results[id] = fiber.Map{
				"success":      false,
				"error":        err.Error(),
				"deleted_from": result.DeletedFrom,
				"errors":       result.Errors,
			}
			failureCount++
		} else {
			results[id] = fiber.Map{
				"success":       true,
				"deleted_from":  result.DeletedFrom,
				"files_deleted": result.FilesDeleted,
			}
			successCount++
		}
	}

	return c.JSON(fiber.Map{
		"message":       "Bulk delete completed",
		"total":         len(request.IDs),
		"success_count": successCount,
		"failure_count": failureCount,
		"results":       results,
	})
}
