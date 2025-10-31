package handler

import (
	"time"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service"
	"github.com/carcheky/keepercheky/internal/service/cleanup"
	"github.com/carcheky/keepercheky/pkg/cache"
	"github.com/carcheky/keepercheky/pkg/logger"
	"gorm.io/gorm"
)

type Handlers struct {
	Health      *HealthHandler
	Dashboard   *DashboardHandler
	Media       *MediaHandler
	Schedule    *ScheduleHandler
	Settings    *SettingsHandler
	Logs        *LogsHandler
	Sync        *SyncHandler
	Files       *FilesHandler
	FileActions *FileActionsHandler
	Radarr      *RadarrHandler
	Sonarr      *SonarrHandler
	QBittorrent *QBittorrentHandler
	Bazarr      *BazarrHandler
}

func NewHandlers(db *gorm.DB, repos *repository.Repositories, logger *logger.Logger, cfg *config.Config) *Handlers {
	// Initialize category counts cache with 30 second TTL
	countsCache := cache.NewCountsCache(30 * time.Second)

	// Initialize OLD SyncService (for client access and backward compatibility)
	oldSyncService := service.NewSyncService(repos.Media, logger, cfg)

	// Initialize NEW FilesystemSyncService (filesystem-first approach)
	filesystemSyncService := service.NewFilesystemSyncService(
		repos.Media,
		repos.Settings,
		oldSyncService.GetRadarrClient(),
		oldSyncService.GetSonarrClient(),
		oldSyncService.GetJellyfinClient(),
		nil, // Jellyseerr not needed for filesystem sync
		oldSyncService.GetQBittorrentClient(),
		logger.Desugar(),
		cfg,
		countsCache,
	)

	// Initialize CleanupService
	cleanupService := cleanup.NewCleanupService(
		repos.Media,
		repos.History,
		oldSyncService.GetRadarrClient(),
		oldSyncService.GetSonarrClient(),
		oldSyncService.GetJellyfinClient(),
		oldSyncService.GetQBittorrentClient(),
		logger.Desugar(),
	)

	// Initialize HealthAnalyzer
	healthAnalyzer := service.NewHealthAnalyzer(logger.Desugar())

	return &Handlers{
		Health:    NewHealthHandler(db, logger),
		Dashboard: NewDashboardHandler(repos, logger, oldSyncService),
		Media:     NewMediaHandler(repos, cleanupService, logger),
		Schedule:  NewScheduleHandler(repos, logger),
		Settings:  NewSettingsHandler(repos, logger, cfg, oldSyncService),
		Logs:      NewLogsHandler(repos, logger),
		Sync:      NewSyncHandler(filesystemSyncService, logger), // Use NEW filesystem-first sync
		Files:     NewFilesHandler(repos.Media, cfg, oldSyncService, healthAnalyzer, logger.Desugar(), countsCache),
		FileActions: NewFileActionsHandler(
			repos.Media,
			repos.History,
			oldSyncService.GetRadarrClient(),
			oldSyncService.GetSonarrClient(),
			oldSyncService.GetQBittorrentClient(),
			oldSyncService.GetJellyfinClient(),
			logger.Desugar(),
		),
		Radarr:      NewRadarrHandler(cfg, logger),
		Sonarr:      NewSonarrHandler(cfg, logger),
		QBittorrent: NewQBittorrentHandler(cfg, logger),
		Bazarr:      NewBazarrHandler(cfg, logger),
	}
}
