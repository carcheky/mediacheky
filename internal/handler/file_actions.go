package handler

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// FileActionsHandler handles file management actions
type FileActionsHandler struct {
	mediaRepo      *repository.MediaRepository
	historyRepo    *repository.HistoryRepository
	radarrClient   clients.MediaClient
	sonarrClient   clients.MediaClient
	qbitClient     *clients.QBittorrentClient
	jellyfinClient clients.StreamingClient
	logger         *zap.Logger
}

// NewFileActionsHandler creates a new file actions handler
func NewFileActionsHandler(
	mediaRepo *repository.MediaRepository,
	historyRepo *repository.HistoryRepository,
	radarrClient clients.MediaClient,
	sonarrClient clients.MediaClient,
	qbitClient *clients.QBittorrentClient,
	jellyfinClient clients.StreamingClient,
	logger *zap.Logger,
) *FileActionsHandler {
	return &FileActionsHandler{
		mediaRepo:      mediaRepo,
		historyRepo:    historyRepo,
		radarrClient:   radarrClient,
		sonarrClient:   sonarrClient,
		qbitClient:     qbitClient,
		jellyfinClient: jellyfinClient,
		logger:         logger,
	}
}

// IgnoreFileRequest represents the request body for ignoring a file
type IgnoreFileRequest struct {
	Reason    string `json:"reason"`
	Permanent bool   `json:"permanent"`
}

// DeleteFileRequest represents the request body for deleting a file
type DeleteFileRequest struct {
	DeleteFromServices bool `json:"delete_from_services"`
	DeleteTorrent      bool `json:"delete_torrent"`
	Confirm            bool `json:"confirm"`
}

// CleanupHardlinkRequest represents the request body for cleaning up a hardlink
type CleanupHardlinkRequest struct {
	KeepPath   string `json:"keep_path"`
	RemovePath string `json:"remove_path"`
}

// ImportToRadarrRequest represents the request body for importing to Radarr
type ImportToRadarrRequest struct {
	FilePath         string `json:"file_path"`
	QualityProfileID int    `json:"quality_profile_id"`
	RootFolderPath   string `json:"root_folder_path"`
}

// ImportToSonarrRequest represents the request body for importing to Sonarr
type ImportToSonarrRequest struct {
	FilePath         string `json:"file_path"`
	QualityProfileID int    `json:"quality_profile_id"`
	RootFolderPath   string `json:"root_folder_path"`
}

// BulkActionRequest represents the request body for bulk actions
type BulkActionRequest struct {
	FileIDs []uint                 `json:"file_ids"`
	Action  string                 `json:"action"`
	Params  map[string]interface{} `json:"params"`
}

// IgnoreFile marks a file as ignored for future analysis
// POST /api/files/:id/ignore
func (h *FileActionsHandler) IgnoreFile(c *fiber.Ctx) error {
	// Parse ID
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Error("Invalid file ID", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid file ID",
		})
	}

	// Parse request body
	var req IgnoreFileRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Get media from database
	media, err := h.mediaRepo.GetByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get media", zap.Int("id", id), zap.Error(err))
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "File not found",
		})
	}

	h.logger.Info("Ignoring file",
		zap.Int("id", id),
		zap.String("title", media.Title),
		zap.String("reason", req.Reason),
		zap.Bool("permanent", req.Permanent),
	)

	// Update media record
	media.Excluded = true
	// Note: We'd need to add ExclusionReason and ExcludedAt fields to the Media model
	// For now, we'll just set Excluded to true

	if err := h.mediaRepo.Update(media); err != nil {
		h.logger.Error("Failed to update media", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to ignore file",
		})
	}

	// Log to history
	history := &models.History{
		MediaID:    media.ID,
		MediaTitle: media.Title,
		Action:     "ignored",
		Status:     "success",
		Message:    fmt.Sprintf("File ignored: %s", req.Reason),
	}
	_ = h.historyRepo.Create(history)

	h.logger.Info("File ignored successfully",
		zap.Int("id", id),
		zap.String("title", media.Title),
	)

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Archivo marcado como ignorado",
	})
}

// DeleteFile deletes a file from the system safely
// POST /api/files/:id/delete
func (h *FileActionsHandler) DeleteFile(c *fiber.Ctx) error {
	ctx := context.Background()

	// Parse ID
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Error("Invalid file ID", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid file ID",
		})
	}

	// Parse request body
	var req DeleteFileRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Require confirmation
	if !req.Confirm {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Confirmation required (confirm: true)",
		})
	}

	// Get media from database
	media, err := h.mediaRepo.GetByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get media", zap.Int("id", id), zap.Error(err))
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "File not found",
		})
	}

	h.logger.Info("Deleting file",
		zap.Int("id", id),
		zap.String("title", media.Title),
		zap.String("path", media.FilePath),
		zap.Bool("delete_from_services", req.DeleteFromServices),
		zap.Bool("delete_torrent", req.DeleteTorrent),
	)

	deletedFrom := []string{}

	// Delete from services if requested
	if req.DeleteFromServices {
		// Delete from Radarr
		if media.InRadarr && media.RadarrID != nil && h.radarrClient != nil {
			h.logger.Info("Deleting from Radarr", zap.Int("radarr_id", *media.RadarrID))
			if err := h.radarrClient.DeleteItem(ctx, *media.RadarrID, true); err != nil {
				h.logger.Error("Failed to delete from Radarr", zap.Error(err))
				// Continue with other deletions
			} else {
				deletedFrom = append(deletedFrom, "radarr")
			}
		}

		// Delete from Sonarr
		if media.InSonarr && media.SonarrID != nil && h.sonarrClient != nil {
			h.logger.Info("Deleting from Sonarr", zap.Int("sonarr_id", *media.SonarrID))
			if err := h.sonarrClient.DeleteItem(ctx, *media.SonarrID, true); err != nil {
				h.logger.Error("Failed to delete from Sonarr", zap.Error(err))
				// Continue with other deletions
			} else {
				deletedFrom = append(deletedFrom, "sonarr")
			}
		}

		// Delete from Jellyfin
		if media.InJellyfin && media.JellyfinID != nil && h.jellyfinClient != nil {
			h.logger.Info("Deleting from Jellyfin", zap.String("jellyfin_id", *media.JellyfinID))
			if err := h.jellyfinClient.DeleteItem(ctx, *media.JellyfinID); err != nil {
				h.logger.Error("Failed to delete from Jellyfin", zap.Error(err))
				// Continue with other deletions
			} else {
				deletedFrom = append(deletedFrom, "jellyfin")
			}
		}
	}

	// Delete torrent if requested
	if req.DeleteTorrent && media.InQBittorrent && media.TorrentHash != "" && h.qbitClient != nil {
		h.logger.Info("Deleting torrent", zap.String("hash", media.TorrentHash))
		if err := h.qbitClient.DeleteTorrent(ctx, media.TorrentHash, true); err != nil {
			h.logger.Error("Failed to delete torrent", zap.Error(err))
			// Continue with file deletion
		} else {
			deletedFrom = append(deletedFrom, "qbittorrent")
		}
	}

	// Delete physical file
	// If it's a hardlink, only delete the specific path
	if media.IsHardlink {
		h.logger.Info("Deleting hardlink", zap.String("path", media.FilePath))
		if err := os.Remove(media.FilePath); err != nil {
			h.logger.Error("Failed to delete hardlink", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   fmt.Sprintf("Failed to delete file: %v", err),
			})
		}
		deletedFrom = append(deletedFrom, "filesystem")
	} else {
		// Regular file - delete it
		h.logger.Info("Deleting file", zap.String("path", media.FilePath))
		if err := os.Remove(media.FilePath); err != nil {
			h.logger.Error("Failed to delete file", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"error":   fmt.Sprintf("Failed to delete file: %v", err),
			})
		}
		deletedFrom = append(deletedFrom, "filesystem")
	}

	// Delete from database
	if err := h.mediaRepo.Delete(uint(id)); err != nil {
		h.logger.Error("Failed to delete from database", zap.Error(err))
		// File already deleted, so we return success but note the DB error
	}

	// Log to history
	history := &models.History{
		MediaID:    media.ID,
		MediaTitle: media.Title,
		Action:     "deleted",
		Status:     "success",
		Message:    fmt.Sprintf("File deleted from: %s", strings.Join(deletedFrom, ", ")),
	}
	_ = h.historyRepo.Create(history)

	h.logger.Info("File deleted successfully",
		zap.Int("id", id),
		zap.String("title", media.Title),
		zap.Strings("deleted_from", deletedFrom),
	)

	return c.JSON(fiber.Map{
		"success":      true,
		"deleted_from": deletedFrom,
		"message":      "Archivo eliminado exitosamente de todos los servicios",
	})
}

// CleanupHardlink removes a hardlink while keeping the original file
// POST /api/files/:id/cleanup-hardlink
func (h *FileActionsHandler) CleanupHardlink(c *fiber.Ctx) error {
	// Parse ID
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Error("Invalid file ID", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid file ID",
		})
	}

	// Parse request body
	var req CleanupHardlinkRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	// Validate paths provided
	if req.KeepPath == "" || req.RemovePath == "" {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Both keep_path and remove_path are required",
		})
	}

	// Get media from database
	media, err := h.mediaRepo.GetByID(uint(id))
	if err != nil {
		h.logger.Error("Failed to get media", zap.Int("id", id), zap.Error(err))
		return c.Status(404).JSON(fiber.Map{
			"success": false,
			"error":   "File not found",
		})
	}

	h.logger.Info("Cleaning up hardlink",
		zap.Int("id", id),
		zap.String("keep_path", req.KeepPath),
		zap.String("remove_path", req.RemovePath),
	)

	// Verify both paths exist
	keepInfo, err := os.Stat(req.KeepPath)
	if err != nil {
		h.logger.Error("Keep path does not exist", zap.String("path", req.KeepPath), zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Keep path does not exist",
		})
	}

	removeInfo, err := os.Stat(req.RemovePath)
	if err != nil {
		h.logger.Error("Remove path does not exist", zap.String("path", req.RemovePath), zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Remove path does not exist",
		})
	}

	// Verify they are hardlinks of the same inode
	keepSys, ok := keepInfo.Sys().(*syscall.Stat_t)
	if !ok {
		h.logger.Error("Failed to get keep path inode")
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to verify hardlink status",
		})
	}

	removeSys, ok := removeInfo.Sys().(*syscall.Stat_t)
	if !ok {
		h.logger.Error("Failed to get remove path inode")
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to verify hardlink status",
		})
	}

	if keepSys.Ino != removeSys.Ino {
		h.logger.Error("Paths are not hardlinks of the same file",
			zap.Uint64("keep_inode", keepSys.Ino),
			zap.Uint64("remove_inode", removeSys.Ino),
		)
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Paths are not hardlinks of the same file",
		})
	}

	// Delete the remove path
	if err := os.Remove(req.RemovePath); err != nil {
		h.logger.Error("Failed to remove hardlink", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Failed to remove hardlink: %v", err),
		})
	}

	// Update media record to remove the path from hardlink_paths
	if media.IsHardlink {
		updatedPaths := []string{}
		for _, path := range media.HardlinkPaths {
			if path != req.RemovePath {
				updatedPaths = append(updatedPaths, path)
			}
		}
		media.HardlinkPaths = updatedPaths

		// Update primary path if it was the removed one
		if media.PrimaryPath == req.RemovePath {
			media.PrimaryPath = req.KeepPath
			media.FilePath = req.KeepPath
		}

		// If only one path remains, it's no longer a hardlink
		if len(updatedPaths) <= 1 {
			media.IsHardlink = false
		}

		if err := h.mediaRepo.Update(media); err != nil {
			h.logger.Error("Failed to update media", zap.Error(err))
			// Continue - file is already deleted
		}
	}

	// Log to history
	history := &models.History{
		MediaID:    media.ID,
		MediaTitle: media.Title,
		Action:     "cleanup_hardlink",
		Status:     "success",
		Message:    fmt.Sprintf("Hardlink removed: %s (kept: %s)", req.RemovePath, req.KeepPath),
	}
	_ = h.historyRepo.Create(history)

	h.logger.Info("Hardlink cleaned up successfully",
		zap.Int("id", id),
		zap.String("removed", req.RemovePath),
		zap.String("kept", req.KeepPath),
	)

	return c.JSON(fiber.Map{
		"success":     true,
		"space_freed": 0, // Hardlinks don't free space
		"message":     "Hardlink eliminado sin afectar el archivo original",
	})
}

// ImportToRadarr imports an orphan file to Radarr
// POST /api/files/:id/import-to-radarr
func (h *FileActionsHandler) ImportToRadarr(c *fiber.Ctx) error {
	// Parse ID
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Error("Invalid file ID", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid file ID",
		})
	}

	// Parse request body
	var req ImportToRadarrRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.Info("Import to Radarr requested",
		zap.Int("id", id),
		zap.String("file_path", req.FilePath),
	)

	// TODO: Implement actual import logic
	// This requires:
	// 1. Parse file name to extract movie title and year
	// 2. Search TMDB/Radarr for the movie
	// 3. Add movie to Radarr with the existing file
	// 4. Update database record

	return c.Status(501).JSON(fiber.Map{
		"success": false,
		"error":   "Import to Radarr not yet implemented - requires metadata parsing",
		"message": "Esta funcionalidad está en desarrollo",
	})
}

// ImportToSonarr imports an orphan file to Sonarr
// POST /api/files/:id/import-to-sonarr
func (h *FileActionsHandler) ImportToSonarr(c *fiber.Ctx) error {
	// Parse ID
	id, err := c.ParamsInt("id")
	if err != nil {
		h.logger.Error("Invalid file ID", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid file ID",
		})
	}

	// Parse request body
	var req ImportToSonarrRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.Info("Import to Sonarr requested",
		zap.Int("id", id),
		zap.String("file_path", req.FilePath),
	)

	// TODO: Implement actual import logic
	// This requires:
	// 1. Parse file name to extract series, season, episode info
	// 2. Search TVDB/Sonarr for the series
	// 3. Add series to Sonarr with the existing file
	// 4. Update database record

	return c.Status(501).JSON(fiber.Map{
		"success": false,
		"error":   "Import to Sonarr not yet implemented - requires metadata parsing",
		"message": "Esta funcionalidad está en desarrollo",
	})
}

// BulkAction executes an action on multiple files
// POST /api/files/bulk-action
func (h *FileActionsHandler) BulkAction(c *fiber.Ctx) error {
	// Parse request body
	var req BulkActionRequest
	if err := c.BodyParser(&req); err != nil {
		h.logger.Error("Failed to parse request body", zap.Error(err))
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request body",
		})
	}

	h.logger.Info("Bulk action requested",
		zap.String("action", req.Action),
		zap.Int("file_count", len(req.FileIDs)),
	)

	// Validate action
	allowedActions := map[string]bool{
		"delete": true,
		"ignore": true,
	}

	if !allowedActions[req.Action] {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("Action '%s' is not allowed for bulk operations", req.Action),
		})
	}

	type BulkResult struct {
		ID      uint   `json:"id"`
		Success bool   `json:"success"`
		Error   string `json:"error,omitempty"`
	}

	results := make([]BulkResult, 0, len(req.FileIDs))
	succeeded := 0
	failed := 0

	// Process each file
	for _, fileID := range req.FileIDs {
		result := BulkResult{ID: fileID, Success: false}

		switch req.Action {
		case "delete":
			// Create a delete request from params
			deleteReq := DeleteFileRequest{
				DeleteFromServices: getBoolParam(req.Params, "delete_from_services", false),
				DeleteTorrent:      getBoolParam(req.Params, "delete_torrent", false),
				Confirm:            true, // Auto-confirm for bulk
			}

			// Get media
			media, err := h.mediaRepo.GetByID(fileID)
			if err != nil {
				result.Error = "File not found"
				failed++
				results = append(results, result)
				continue
			}

			// Execute delete logic inline
			ctx := context.Background()
			deletedFrom := []string{}

			// Delete from services
			if deleteReq.DeleteFromServices {
				if media.InRadarr && media.RadarrID != nil && h.radarrClient != nil {
					_ = h.radarrClient.DeleteItem(ctx, *media.RadarrID, true)
					deletedFrom = append(deletedFrom, "radarr")
				}
				if media.InSonarr && media.SonarrID != nil && h.sonarrClient != nil {
					_ = h.sonarrClient.DeleteItem(ctx, *media.SonarrID, true)
					deletedFrom = append(deletedFrom, "sonarr")
				}
				if media.InJellyfin && media.JellyfinID != nil && h.jellyfinClient != nil {
					_ = h.jellyfinClient.DeleteItem(ctx, *media.JellyfinID)
					deletedFrom = append(deletedFrom, "jellyfin")
				}
			}

			// Delete torrent
			if deleteReq.DeleteTorrent && media.TorrentHash != "" && h.qbitClient != nil {
				_ = h.qbitClient.DeleteTorrent(ctx, media.TorrentHash, true)
				deletedFrom = append(deletedFrom, "qbittorrent")
			}

			// Delete file
			if err := os.Remove(media.FilePath); err != nil {
				result.Error = fmt.Sprintf("Failed to delete file: %v", err)
				failed++
				results = append(results, result)
				continue
			}
			deletedFrom = append(deletedFrom, "filesystem")

			// Delete from DB
			_ = h.mediaRepo.Delete(fileID)

			// Log to history
			history := &models.History{
				MediaID:    media.ID,
				MediaTitle: media.Title,
				Action:     "deleted_bulk",
				Status:     "success",
				Message:    fmt.Sprintf("Bulk deleted from: %s", strings.Join(deletedFrom, ", ")),
			}
			_ = h.historyRepo.Create(history)

			result.Success = true
			succeeded++

		case "ignore":
			media, err := h.mediaRepo.GetByID(fileID)
			if err != nil {
				result.Error = "File not found"
				failed++
				results = append(results, result)
				continue
			}

			media.Excluded = true
			if err := h.mediaRepo.Update(media); err != nil {
				result.Error = fmt.Sprintf("Failed to update: %v", err)
				failed++
				results = append(results, result)
				continue
			}

			// Log to history
			history := &models.History{
				MediaID:    media.ID,
				MediaTitle: media.Title,
				Action:     "ignored_bulk",
				Status:     "success",
				Message:    "Bulk ignored",
			}
			_ = h.historyRepo.Create(history)

			result.Success = true
			succeeded++
		}

		results = append(results, result)
	}

	h.logger.Info("Bulk action completed",
		zap.String("action", req.Action),
		zap.Int("total", len(req.FileIDs)),
		zap.Int("succeeded", succeeded),
		zap.Int("failed", failed),
	)

	return c.JSON(fiber.Map{
		"success": true,
		"results": results,
		"summary": fiber.Map{
			"total":     len(req.FileIDs),
			"succeeded": succeeded,
			"failed":    failed,
		},
	})
}

// Helper function to get boolean parameter from map
func getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := params[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}
