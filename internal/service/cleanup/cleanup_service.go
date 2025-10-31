package cleanup

import (
	"context"
	"fmt"
	"os"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"go.uber.org/zap"
)

// DeleteOptions represents the options for deleting media.
type DeleteOptions struct {
	Radarr      bool `json:"radarr"`
	Sonarr      bool `json:"sonarr"`
	Jellyfin    bool `json:"jellyfin"`
	DeleteFiles bool `json:"deleteFiles"`
	QBittorrent bool `json:"qbittorrent"`
}

// DeleteResult contains the result of a deletion operation.
type DeleteResult struct {
	Success      bool              `json:"success"`
	DeletedFrom  []string          `json:"deleted_from"`
	Errors       map[string]string `json:"errors"`
	FilesDeleted bool              `json:"files_deleted"`
}

// CleanupService handles media cleanup operations.
type CleanupService struct {
	mediaRepo         *repository.MediaRepository
	historyRepo       *repository.HistoryRepository
	radarrClient      clients.MediaClient
	sonarrClient      clients.MediaClient
	jellyfinClient    clients.StreamingClient
	qbittorrentClient *clients.QBittorrentClient
	logger            *zap.Logger
}

// NewCleanupService creates a new cleanup service.
func NewCleanupService(
	mediaRepo *repository.MediaRepository,
	historyRepo *repository.HistoryRepository,
	radarrClient clients.MediaClient,
	sonarrClient clients.MediaClient,
	jellyfinClient clients.StreamingClient,
	qbittorrentClient *clients.QBittorrentClient,
	logger *zap.Logger,
) *CleanupService {
	return &CleanupService{
		mediaRepo:         mediaRepo,
		historyRepo:       historyRepo,
		radarrClient:      radarrClient,
		sonarrClient:      sonarrClient,
		jellyfinClient:    jellyfinClient,
		qbittorrentClient: qbittorrentClient,
		logger:            logger,
	}
}

// DeleteMedia deletes media from selected services.
func (s *CleanupService) DeleteMedia(ctx context.Context, media *models.Media, options DeleteOptions) (*DeleteResult, error) {
	s.logger.Info("Starting media deletion",
		zap.Uint("id", media.ID),
		zap.String("title", media.Title),
		zap.String("type", media.Type),
		zap.Bool("radarr", options.Radarr),
		zap.Bool("sonarr", options.Sonarr),
		zap.Bool("jellyfin", options.Jellyfin),
		zap.Bool("delete_files", options.DeleteFiles),
		zap.Bool("qbittorrent", options.QBittorrent),
	)

	result := &DeleteResult{
		Success:     true,
		DeletedFrom: []string{},
		Errors:      make(map[string]string),
	}

	// Delete from Radarr
	if options.Radarr && media.RadarrID != nil && s.radarrClient != nil {
		s.logger.Info("Deleting from Radarr",
			zap.Int("radarr_id", *media.RadarrID),
		)

		err := s.radarrClient.DeleteItem(ctx, *media.RadarrID, options.DeleteFiles)
		if err != nil {
			s.logger.Error("Failed to delete from Radarr",
				zap.Int("radarr_id", *media.RadarrID),
				zap.Error(err),
			)
			result.Errors["radarr"] = err.Error()
			result.Success = false
		} else {
			result.DeletedFrom = append(result.DeletedFrom, "radarr")
			s.logger.Info("Successfully deleted from Radarr")
		}
	}

	// Delete from Sonarr
	if options.Sonarr && media.SonarrID != nil && s.sonarrClient != nil {
		s.logger.Info("Deleting from Sonarr",
			zap.Int("sonarr_id", *media.SonarrID),
		)

		err := s.sonarrClient.DeleteItem(ctx, *media.SonarrID, options.DeleteFiles)
		if err != nil {
			s.logger.Error("Failed to delete from Sonarr",
				zap.Int("sonarr_id", *media.SonarrID),
				zap.Error(err),
			)
			result.Errors["sonarr"] = err.Error()
			result.Success = false
		} else {
			result.DeletedFrom = append(result.DeletedFrom, "sonarr")
			s.logger.Info("Successfully deleted from Sonarr")
		}
	}

	// Delete from Jellyfin
	if options.Jellyfin && media.JellyfinID != nil && s.jellyfinClient != nil {
		s.logger.Info("Deleting from Jellyfin",
			zap.String("jellyfin_id", *media.JellyfinID),
		)

		err := s.jellyfinClient.DeleteItem(ctx, *media.JellyfinID)
		if err != nil {
			s.logger.Error("Failed to delete from Jellyfin",
				zap.String("jellyfin_id", *media.JellyfinID),
				zap.Error(err),
			)
			result.Errors["jellyfin"] = err.Error()
			result.Success = false
		} else {
			result.DeletedFrom = append(result.DeletedFrom, "jellyfin")
			s.logger.Info("Successfully deleted from Jellyfin")
		}
	}

	// Delete files directly if requested and not already deleted by Radarr/Sonarr
	if options.DeleteFiles && !options.Radarr && !options.Sonarr && media.FilePath != "" {
		s.logger.Info("Deleting files from disk",
			zap.String("path", media.FilePath),
		)

		err := os.RemoveAll(media.FilePath)
		if err != nil {
			s.logger.Error("Failed to delete files",
				zap.String("path", media.FilePath),
				zap.Error(err),
			)
			result.Errors["files"] = err.Error()
			result.Success = false
		} else {
			result.FilesDeleted = true
			s.logger.Info("Successfully deleted files from disk")
		}
	} else if options.DeleteFiles && (options.Radarr || options.Sonarr) {
		// Files were deleted by Radarr/Sonarr
		result.FilesDeleted = true
	}

	// Create history entry
	historyEntry := &models.History{
		MediaID:    media.ID,
		MediaTitle: media.Title,
		Action:     "deleted",
		Status:     "success",
		Message:    fmt.Sprintf("Deleted from: %v. Errors: %v", result.DeletedFrom, result.Errors),
		DryRun:     false,
	}

	if err := s.historyRepo.Create(historyEntry); err != nil {
		s.logger.Error("Failed to create history entry",
			zap.Error(err),
		)
	}

	// Delete from database if successfully deleted from at least one service
	if len(result.DeletedFrom) > 0 || result.FilesDeleted {
		if err := s.mediaRepo.Delete(media.ID); err != nil {
			s.logger.Error("Failed to delete from database",
				zap.Uint("id", media.ID),
				zap.Error(err),
			)
			result.Errors["database"] = err.Error()
			result.Success = false
			return result, fmt.Errorf("failed to delete from database: %w", err)
		}
		s.logger.Info("Successfully deleted from database")
	} // If we have any errors, return them
	if len(result.Errors) > 0 {
		return result, fmt.Errorf("deletion completed with errors")
	}

	s.logger.Info("Media deletion completed successfully",
		zap.Uint("id", media.ID),
		zap.Strings("deleted_from", result.DeletedFrom),
	)

	return result, nil
}
