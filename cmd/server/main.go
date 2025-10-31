package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/handler"
	"github.com/carcheky/keepercheky/internal/middleware"
	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/scheduler"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Build-time variables injected via ldflags
var (
	Version   = "dev"
	CommitSHA = "unknown"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger with file output
	logFilePath := "./logs/keepercheky-dev.log"
	if cfg.App.Environment == "production" {
		logFilePath = "./logs/keepercheky.log"
	}
	appLogger := logger.NewWithFile(cfg.App.LogLevel, logFilePath)
	defer appLogger.Sync()

	appLogger.Info("Starting KeeperCheky",
		"version", getVersion(),
		"commit", CommitSHA,
		"env", cfg.App.Environment,
	)

	// Initialize database
	db, err := initDatabase(cfg, appLogger)
	if err != nil {
		appLogger.Fatal("Failed to initialize database", "error", err)
	}

	// Run migrations
	if err := models.RunMigrations(db); err != nil {
		appLogger.Fatal("Failed to run migrations", "error", err)
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize template engine
	engine := html.New("./web/templates", ".html")
	engine.Reload(cfg.App.Environment == "development")

	// Add custom template functions
	engine.AddFunc("toJSON", func(v interface{}) string {
		bytes, err := json.Marshal(v)
		if err != nil {
			return "[]"
		}
		return string(bytes)
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "KeeperCheky",
		Views:        engine,
		ErrorHandler: middleware.ErrorHandler(appLogger),
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.Logger(appLogger))
	app.Use(middleware.RequestID())

	// Static files
	app.Static("/static", "./web/static")

	// Initialize handlers
	handlers := handler.NewHandlers(db, repos, appLogger, cfg)

	// Setup routes
	setupRoutes(app, handlers)

	// Initialize scheduler (if enabled)
	if cfg.App.SchedulerEnabled {
		sched := scheduler.New(repos, appLogger, cfg)
		sched.Start()
		defer sched.Stop()
	}

	// Start server
	port := cfg.Server.Port
	if port == "" {
		port = "8000"
	}

	appLogger.Info("Server starting", "port", port)
	if err := app.Listen(":" + port); err != nil {
		appLogger.Fatal("Failed to start server", "error", err)
	}
}

func initDatabase(cfg *config.Config, logger *logger.Logger) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// For now, use SQLite
	// TODO: Add PostgreSQL support based on config
	dbPath := cfg.Database.Path
	if dbPath == "" {
		dbPath = "./data/keepercheky.db"
	}

	logger.Info("Initializing database", "path", dbPath)

	db, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func setupRoutes(app *fiber.App, h *handler.Handlers) {
	// Health check
	app.Get("/health", h.Health.Check)

	// Favicon (prevent 404 errors in logs)
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(204) // No Content
	})

	// Web UI routes
	app.Get("/", h.Dashboard.Index)
	// app.Get("/media", h.Media.List) // Removed from navigation - functionality available via API
	app.Get("/files", h.Files.RenderFilesPage)
	app.Get("/files-example", h.Files.RenderExamplePage) // Components demo page
	app.Get("/schedules", h.Schedule.List)
	app.Get("/settings", h.Settings.Index)
	app.Get("/logs", h.Logs.Index)

	// API routes
	api := app.Group("/api")
	{
		// Media
		api.Get("/media", h.Media.GetAll)
		api.Get("/media/stats", h.Media.GetStats)
		api.Get("/media/:id", h.Media.GetByID)
		api.Delete("/media/:id", h.Media.Delete)
		api.Post("/media/bulk-delete", h.Media.BulkDelete)
		api.Post("/media/:id/exclude", h.Media.Exclude)

		// Files
		api.Get("/files", h.Files.GetFilesAPI)
		api.Get("/files/health", h.Files.GetFilesHealthAPI) // Health analysis endpoint

		// File Actions
		api.Post("/files/:id/import-to-radarr", h.FileActions.ImportToRadarr)
		api.Post("/files/:id/import-to-sonarr", h.FileActions.ImportToSonarr)
		api.Post("/files/:id/delete", h.FileActions.DeleteFile)
		api.Post("/files/:id/ignore", h.FileActions.IgnoreFile)
		api.Post("/files/:id/cleanup-hardlink", h.FileActions.CleanupHardlink)
		api.Post("/files/bulk-action", h.FileActions.BulkAction)

		// Stats
		api.Get("/stats", h.Dashboard.Stats)

		// Configuration (Settings)
		api.Get("/config", h.Settings.Get)
		api.Post("/config", h.Settings.Update)
		api.Post("/config/test/:service", h.Settings.TestConnection)

		// Sync - GET for SSE (Server-Sent Events)
		api.Get("/sync", h.Sync.Sync)
		api.Get("/sync/files", h.Sync.SyncFiles) // Filesystem-first sync for Files view

		// Radarr endpoints
		api.Get("/radarr/system", h.Radarr.GetSystemInfo)
		api.Get("/radarr/queue", h.Radarr.GetQueue)
		api.Get("/radarr/history", h.Radarr.GetHistory)
		api.Get("/radarr/calendar", h.Radarr.GetCalendar)
		api.Get("/radarr/quality-profiles", h.Radarr.GetQualityProfiles)

		// Sonarr endpoints
		api.Get("/sonarr/system", h.Sonarr.GetSystemInfo)
		api.Get("/sonarr/queue", h.Sonarr.GetQueue)
		api.Get("/sonarr/history", h.Sonarr.GetHistory)
		api.Get("/sonarr/calendar", h.Sonarr.GetCalendar)
		api.Get("/sonarr/quality-profiles", h.Sonarr.GetQualityProfiles)

		// Bazarr endpoints
		api.Get("/bazarr/system", h.Bazarr.GetSystemInfo)
		api.Get("/bazarr/history", h.Bazarr.GetHistory)
		api.Get("/bazarr/wanted/movies", h.Bazarr.GetWantedMovies)
		api.Get("/bazarr/wanted/series", h.Bazarr.GetWantedSeries)
		api.Get("/bazarr/subtitles/movie/:radarr_id", h.Bazarr.GetMovieSubtitles)
		api.Get("/bazarr/subtitles/series/:sonarr_id", h.Bazarr.GetSeriesSubtitles)

		// Jellyfin endpoints
		api.Get("/jellyfin/stats", h.Settings.GetJellyfinStats)
		api.Get("/jellyfin/sessions", h.Settings.GetJellyfinSessions)
		api.Get("/jellyfin/recently-added", h.Settings.GetJellyfinRecentlyAdded)
		api.Get("/jellyfin/activity", h.Settings.GetJellyfinActivity)

		// Jellyseerr endpoints
		api.Get("/jellyseerr/stats", h.Dashboard.GetJellyseerrStats)
		api.Get("/jellyseerr/requests", h.Dashboard.GetJellyseerrRequests)

		// Jellystat endpoints
		api.Get("/jellystat/stats", h.Settings.GetJellystatStats)
		api.Get("/jellystat/views-by-type", h.Settings.GetJellystatViewsByType)
		api.Get("/jellystat/user-activity", h.Settings.GetJellystatUserActivity)
		api.Get("/jellystat/library-stats", h.Settings.GetJellystatLibraryStats)
		api.Get("/jellystat/dashboard/stats", h.Dashboard.GetJellystatStats)
		api.Get("/jellystat/dashboard/views-by-type", h.Dashboard.GetJellystatViewsByType)

		// qBittorrent endpoints
		api.Get("/qbittorrent/transfer", h.QBittorrent.GetTransferInfo)
		api.Get("/qbittorrent/state", h.QBittorrent.GetServerState)
		api.Get("/qbittorrent/active", h.QBittorrent.GetActiveTorrents)
		api.Get("/qbittorrent/torrent/:hash/properties", h.QBittorrent.GetTorrentProperties)
		api.Get("/qbittorrent/torrent/:hash/trackers", h.QBittorrent.GetTorrentTrackers)
	}
}

func getVersion() string {
	// Return build-time injected version, fallback to env var, then to "dev"
	if Version != "" && Version != "dev" {
		return Version
	}
	version := os.Getenv("VERSION")
	if version == "" {
		return "dev"
	}
	return version
}
