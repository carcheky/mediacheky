package service

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/carcheky/keepercheky/pkg/cache"
	"github.com/carcheky/keepercheky/pkg/filesystem"
	"go.uber.org/zap"
)

// FilesystemSyncService implements filesystem-first sync approach
type FilesystemSyncService struct {
	mediaRepo         *repository.MediaRepository
	settingsRepo      *repository.SettingsRepository
	radarrClient      clients.MediaClient
	sonarrClient      clients.MediaClient
	jellyfinClient    clients.StreamingClient
	jellyseerrClient  clients.RequestClient
	qbittorrentClient *clients.QBittorrentClient
	logger            *zap.Logger
	config            *config.Config
	countsCache       *cache.CountsCache
}

// NewFilesystemSyncService creates a new filesystem-first sync service
func NewFilesystemSyncService(
	mediaRepo *repository.MediaRepository,
	settingsRepo *repository.SettingsRepository,
	radarrClient clients.MediaClient,
	sonarrClient clients.MediaClient,
	jellyfinClient clients.StreamingClient,
	jellyseerrClient clients.RequestClient,
	qbittorrentClient *clients.QBittorrentClient,
	logger *zap.Logger,
	config *config.Config,
	countsCache *cache.CountsCache,
) *FilesystemSyncService {
	return &FilesystemSyncService{
		mediaRepo:         mediaRepo,
		settingsRepo:      settingsRepo,
		radarrClient:      radarrClient,
		sonarrClient:      sonarrClient,
		jellyfinClient:    jellyfinClient,
		jellyseerrClient:  jellyseerrClient,
		qbittorrentClient: qbittorrentClient,
		logger:            logger,
		config:            config,
		countsCache:       countsCache,
	}
}

// SyncAllWithProgress performs a complete filesystem-first sync with progress reporting
func (s *FilesystemSyncService) SyncAllWithProgress(ctx context.Context, progressChan chan<- SyncProgress) error {
	defer close(progressChan)

	s.logger.Info("🗂️  Starting FILESYSTEM-FIRST sync")

	// STEP 1: Get dynamic paths from services
	rootPaths := s.getDynamicRootPaths(ctx)

	if len(rootPaths) == 0 {
		progressChan <- SyncProgress{
			Step:    "error",
			Message: "❌ No se detectaron rutas para escanear. Verifica la configuración de qBittorrent y Jellyfin.",
			Status:  "error",
		}
		return fmt.Errorf("no root paths detected")
	}

	// STEP 2: Scan filesystem (source of truth)
	progressChan <- SyncProgress{
		Step:    "scan_filesystem",
		Message: fmt.Sprintf("🗂️  Escaneando sistema de archivos (%d rutas)...", len(rootPaths)),
		Status:  "processing",
	}

	scanOptions := filesystem.ScanOptions{
		RootPaths:       rootPaths,
		VideoExtensions: s.config.Filesystem.VideoExtensions,
		MinSize:         s.config.Filesystem.MinSizeMB * 1024 * 1024, // Convert MB to bytes
		SkipHidden:      s.config.Filesystem.SkipHidden,
	}

	scanner := filesystem.NewScanner(scanOptions, s.logger)
	fileEntries, err := scanner.Scan()
	if err != nil {
		return fmt.Errorf("filesystem scan failed: %w", err)
	}

	stats := scanner.GetStats()
	progressChan <- SyncProgress{
		Step:    "scan_filesystem_complete",
		Message: fmt.Sprintf("✅ Escaneados: %d archivos únicos, %d grupos de hardlinks", stats["unique_inodes"], stats["hardlink_groups"]),
		Status:  "success",
	}

	// STEP 2: Convert FileEntry to EnrichedFile
	enrichedFiles := make(map[string]*filesystem.EnrichedFile)
	for path, entry := range fileEntries {
		enrichedFiles[path] = &filesystem.EnrichedFile{
			FileEntry: entry,
			ModTime:   entry.ModTime, // Copy for easy access
		}
	}

	// STEP 3: Enrich with Radarr
	if s.radarrClient != nil {
		progressChan <- SyncProgress{
			Step:    "enrich_radarr",
			Message: "🎬 Enriqueciendo con Radarr...",
			Status:  "processing",
		}

		radarrMedia, err := s.radarrClient.GetLibrary(ctx)
		if err != nil {
			s.logger.Error("Failed to get Radarr library", zap.Error(err))
		} else {
			enricher := filesystem.NewEnricher(s.logger)
			count := enricher.EnrichWithRadarr(ctx, enrichedFiles, radarrMedia)

			progressChan <- SyncProgress{
				Step:    "enrich_radarr_complete",
				Message: fmt.Sprintf("✅ Radarr: %d archivos enriquecidos", count),
				Status:  "success",
			}
		}
	}

	// STEP 4: Enrich with Sonarr
	if s.sonarrClient != nil {
		progressChan <- SyncProgress{
			Step:    "enrich_sonarr",
			Message: "📺 Enriqueciendo con Sonarr...",
			Status:  "processing",
		}

		sonarrMedia, err := s.sonarrClient.GetLibrary(ctx)
		if err != nil {
			s.logger.Error("Failed to get Sonarr library", zap.Error(err))
		} else {
			enricher := filesystem.NewEnricher(s.logger)
			count := enricher.EnrichWithSonarr(ctx, enrichedFiles, sonarrMedia)

			progressChan <- SyncProgress{
				Step:    "enrich_sonarr_complete",
				Message: fmt.Sprintf("✅ Sonarr: %d archivos enriquecidos", count),
				Status:  "success",
			}
		}
	}

	// STEP 5: Enrich with Jellyfin
	if s.jellyfinClient != nil {
		progressChan <- SyncProgress{
			Step:    "enrich_jellyfin",
			Message: "🎥 Enriqueciendo con Jellyfin...",
			Status:  "processing",
		}

		jellyfinMedia, err := s.jellyfinClient.GetLibrary(ctx)
		if err != nil {
			s.logger.Error("Failed to get Jellyfin library", zap.Error(err))
		} else {
			enricher := filesystem.NewEnricher(s.logger)
			count := enricher.EnrichWithJellyfin(ctx, enrichedFiles, jellyfinMedia)

			progressChan <- SyncProgress{
				Step:    "enrich_jellyfin_complete",
				Message: fmt.Sprintf("✅ Jellyfin: %d archivos enriquecidos", count),
				Status:  "success",
			}

			// Enrich with Jellystat tracking (tracks Jellyfin items)
			if s.config.Clients.Jellystat.Enabled {
				jellystatCount := enricher.EnrichWithJellystat(ctx, enrichedFiles, true)
				s.logger.Info("Jellystat tracking enabled",
					zap.Int("tracked_files", jellystatCount),
				)
			}
		}
	}

	// STEP 6: Enrich with qBittorrent
	if s.qbittorrentClient != nil {
		progressChan <- SyncProgress{
			Step:    "enrich_qbittorrent",
			Message: "🧲 Enriqueciendo con qBittorrent...",
			Status:  "processing",
		}

		torrentMap, err := s.qbittorrentClient.GetAllTorrentsMap(ctx)
		if err != nil {
			s.logger.Error("Failed to get qBittorrent torrents", zap.Error(err))
		} else {
			enricher := filesystem.NewEnricher(s.logger)
			count := enricher.EnrichWithQBittorrent(ctx, enrichedFiles, torrentMap)

			progressChan <- SyncProgress{
				Step:    "enrich_qbittorrent_complete",
				Message: fmt.Sprintf("✅ qBittorrent: %d archivos enriquecidos", count),
				Status:  "success",
			}
		}
	}

	// STEP 7: Clear database and save enriched files
	progressChan <- SyncProgress{
		Step:    "clear_database",
		Message: "🗑️  Limpiando base de datos...",
		Status:  "processing",
	}

	if err := s.mediaRepo.DeleteAll(); err != nil {
		return fmt.Errorf("failed to clear database: %w", err)
	}

	progressChan <- SyncProgress{
		Step:    "save_database",
		Message: fmt.Sprintf("💾 Guardando %d archivos en base de datos...", len(enrichedFiles)),
		Status:  "processing",
	}

	savedCount := 0
	errorCount := 0

	for _, enrichedFile := range enrichedFiles {
		media := s.convertToMedia(enrichedFile)

		if err := s.mediaRepo.Create(media); err != nil {
			s.logger.Error("Failed to save media",
				zap.String("path", media.FilePath),
				zap.Error(err),
			)
			errorCount++
			continue
		}
		savedCount++
	}

	progressChan <- SyncProgress{
		Step:    "save_database_complete",
		Message: fmt.Sprintf("✅ Guardados: %d archivos (%d errores)", savedCount, errorCount),
		Status:  "success",
	}

	// Save last sync timestamp
	if err := s.settingsRepo.Set("last_files_sync", time.Now().Format(time.RFC3339)); err != nil {
		s.logger.Error("Failed to save last sync timestamp", zap.Error(err))
	}

	// Invalidate category counts cache to force refresh with new data
	if s.countsCache != nil {
		s.countsCache.Invalidate()
		s.logger.Debug("Category counts cache invalidated after sync")
	}

	progressChan <- SyncProgress{
		Step:    "complete",
		Message: "✅ Sincronización filesystem-first completada exitosamente",
		Status:  "success",
	}

	s.logger.Info("✅ Filesystem-first sync completed",
		zap.Int("total_files", len(enrichedFiles)),
		zap.Int("saved", savedCount),
		zap.Int("errors", errorCount),
	)

	return nil
}

// convertToMedia converts an EnrichedFile to a Media model
func (s *FilesystemSyncService) convertToMedia(ef *filesystem.EnrichedFile) *models.Media {
	// Use title from services if available, otherwise use filename
	title := ef.Title
	if title == "" {
		title = ef.FileEntry.Path
	}

	media := &models.Media{
		// Basic info
		Title:     title,
		Type:      ef.MediaType,
		PosterURL: ef.PosterURL,
		FilePath:  ef.Path,
		Size:      ef.Size,
		AddedDate: time.Unix(ef.ModTime, 0),
		Quality:   ef.Quality,
		Excluded:  ef.Excluded,

		// Filesystem metadata
		FileInode:     ef.Inode,
		FileModTime:   ef.ModTime,
		IsHardlink:    ef.IsHardlink,
		HardlinkPaths: models.StringSlice(ef.HardlinkPaths),
		PrimaryPath:   ef.PrimaryPath,

		// Service flags
		InRadarr:      ef.InRadarr,
		InSonarr:      ef.InSonarr,
		InJellyfin:    ef.InJellyfin,
		InJellyseerr:  ef.InJellyseerr,
		InJellystat:   ef.InJellystat, // Set by enrichment process
		InQBittorrent: ef.InQBittorrent,

		// Service IDs
		RadarrID:     ef.RadarrID,
		SonarrID:     ef.SonarrID,
		JellyfinID:   ef.JellyfinID,
		JellyseerrID: ef.JellyseerrID,
		// JellystatID not needed - Jellystat uses Jellyfin IDs for tracking

		// Torrent info
		TorrentHash:     ef.TorrentHash,
		TorrentCategory: ef.TorrentCategory,
		TorrentTags:     ef.TorrentTags,
		TorrentState:    ef.TorrentState,
		IsSeeding:       ef.IsSeeding,
		SeedRatio:       ef.SeedRatio,

		// Metadata
		Tags: models.StringSlice(ef.Tags),
	}

	// Last watched
	if ef.LastWatched != nil {
		lastWatched := time.Unix(*ef.LastWatched, 0)
		media.LastWatched = &lastWatched
	}

	return media
}

// getDynamicRootPaths retrieves paths dynamically from qBittorrent and Jellyfin
func (s *FilesystemSyncService) getDynamicRootPaths(ctx context.Context) []string {
	var rootPaths []string
	pathSet := make(map[string]bool)

	// Get qBittorrent download paths
	if s.qbittorrentClient != nil {
		prefs, err := s.qbittorrentClient.GetPreferences(ctx)
		if err == nil {
			// Use ExportDirFin (completed downloads) if available, otherwise SavePath
			completedPath := prefs.ExportDirFin
			if completedPath == "" {
				completedPath = prefs.SavePath
			}

			if completedPath != "" && !pathSet[completedPath] {
				rootPaths = append(rootPaths, completedPath)
				pathSet[completedPath] = true
				s.logger.Info("Added qBittorrent completed downloads path",
					zap.String("path", completedPath))
			}
		}
	}

	// Get Jellyfin library paths
	if s.jellyfinClient != nil {
		// Type assert to access JellyfinClient-specific methods
		if jfClient, ok := s.jellyfinClient.(*clients.JellyfinClient); ok {
			folders, err := jfClient.GetVirtualFolders(ctx)
			if err == nil {
				for _, folder := range folders {
					for _, location := range folder.Locations {
						if !pathSet[location] {
							rootPaths = append(rootPaths, location)
							pathSet[location] = true
							s.logger.Info("Added Jellyfin path",
								zap.String("folder", folder.Name),
								zap.String("path", location))
						}
					}
				}
			}
		}
	}

	// Fallback to config if no dynamic paths found
	if len(rootPaths) == 0 && len(s.config.Filesystem.RootPaths) > 0 {
		s.logger.Warn("No dynamic paths found, using config root paths")
		rootPaths = s.config.Filesystem.RootPaths
	}

	return rootPaths
}
