package apitest

import (
	"testing"

	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/database"
	"github.com/carcheky/mediacheky/internal/handler"
	"github.com/carcheky/mediacheky/internal/middleware"
	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestApp represents a test application instance with all dependencies
type TestApp struct {
	App      *fiber.App
	DB       *gorm.DB
	Repos    *repository.Repositories
	Handlers *handler.Handlers
	Logger   *logger.Logger
	Config   *config.Config
}

// SetupTestApp creates a new test application with in-memory database
func SetupTestApp(t *testing.T) *TestApp {
	// Initialize in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(
		&models.Service{},
		&models.Template{},
		&models.GlobalConfig{},
		&models.ServiceLog{},
		&models.ServiceCredentials{},
		&models.ProxyConfig{},
		&models.Domain{},
		&models.DockerTag{},
		&models.Media{},
		&models.Schedule{},
		&models.History{},
		&models.Settings{},
	); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Seed initial data
	if err := database.SeedData(db); err != nil {
		t.Fatalf("Failed to seed test database: %v", err)
	}

	// Initialize logger
	log := logger.New("debug")

	// Load test configuration
	cfg := &config.Config{
		App: config.AppConfig{
			Environment:      "test",
			LogLevel:         "debug",
			SchedulerEnabled: false,
		},
		Server: config.ServerConfig{
			Port: "8000",
		},
		Database: config.DatabaseConfig{
			Path: ":memory:",
		},
	}

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize handlers
	handlers := handler.NewHandlers(db, repos, log, cfg)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:      "MediaCheky Test",
		ErrorHandler: middleware.ErrorHandler(log),
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.RequestID())

	// Setup routes
	setupTestRoutes(app, handlers)

	return &TestApp{
		App:      app,
		DB:       db,
		Repos:    repos,
		Handlers: handlers,
		Logger:   log,
		Config:   cfg,
	}
}

// setupTestRoutes configures all API routes for testing
func setupTestRoutes(app *fiber.App, h *handler.Handlers) {
	// Health check
	app.Get("/health", h.Health.Check)

	// API routes
	api := app.Group("/api")
	{
		// Dashboard endpoints
		api.Get("/dashboard/stats", h.Dashboard.GetDashboardStats)
		api.Get("/dashboard/health", h.Dashboard.HealthCheck)

		// Stats (legacy endpoint)
		api.Get("/stats", h.Dashboard.Stats)

		// Service endpoints with validation
		services := api.Group("/services")
		services.Get("/", h.Service.ListServices)
		services.Get("/:name", middleware.ValidateServiceName(), h.Service.GetService)
		services.Post("/:name/enable", middleware.ValidateServiceName(), h.Service.EnableService)
		services.Post("/:name/disable", middleware.ValidateServiceName(), h.Service.DisableService)
		services.Post("/:name/start", middleware.ValidateServiceName(), h.Service.StartContainer)
		services.Post("/:name/stop", middleware.ValidateServiceName(), h.Service.StopContainer)
		services.Post("/:name/restart", middleware.ValidateServiceName(), h.Service.RestartContainer)
		services.Post("/:name/update", middleware.ValidateServiceName(), h.Service.UpdateService)
		services.Put("/:name/config", middleware.ValidateServiceName(), h.Service.UpdateServiceConfig)
		services.Get("/:name/logs", middleware.ValidateServiceName(), h.Service.GetContainerLogs)
		services.Put("/:name/subdomain", middleware.ValidateServiceName(), h.Proxy.UpdateServiceSubdomain)
		services.Get("/:name/endpoint", middleware.ValidateServiceName(), h.Proxy.GetServiceEndpoint)

		// Global configuration endpoints with validation
		globalConfig := api.Group("/config/global")
		globalConfig.Get("/", h.Config.GetGlobalConfig)
		globalConfig.Put("/", h.Config.UpdateGlobalConfig)
		globalConfig.Get("/:key", middleware.ValidateConfigKey(), h.Config.GetConfigValue)
		globalConfig.Put("/:key", middleware.ValidateConfigKey(), h.Config.UpdateConfigValue)

		// Proxy endpoints
		proxy := api.Group("/proxy")
		proxy.Get("/config", h.Proxy.GetProxyConfig)
		proxy.Put("/config", h.Proxy.UpdateProxyConfig)
		proxy.Get("/domains", h.Proxy.GetDomains)
		proxy.Post("/domains", h.Proxy.AddDomain)
		proxy.Delete("/domains/:name", h.Proxy.DeleteDomain)

		// Docker endpoints
		api.Get("/docker/info", h.Docker.GetDockerInfo)
		api.Get("/docker/containers", h.Docker.ListContainers)

		// Configuration (Settings) - legacy endpoints
		api.Get("/config", h.Settings.Get)
		api.Post("/config", h.Settings.Update)
		api.Post("/config/test/:service", h.Settings.TestConnection)

		// Jellyfin endpoints (for dashboard stats)
		api.Get("/jellyfin/stats", h.Settings.GetJellyfinStats)
		api.Get("/jellyfin/sessions", h.Settings.GetJellyfinSessions)
		api.Get("/jellyfin/recently-added", h.Settings.GetJellyfinRecentlyAdded)
		api.Get("/jellyfin/activity", h.Settings.GetJellyfinActivity)

		// Radarr/Sonarr queue endpoints
		api.Get("/radarr/queue", h.Settings.GetRadarrQueue)
		api.Get("/sonarr/queue", h.Settings.GetSonarrQueue)

		// Jellystat endpoints (for dashboard stats)
		api.Get("/jellystat/stats", h.Settings.GetJellystatStats)
		api.Get("/jellystat/views-by-type", h.Settings.GetJellystatViewsByType)
		api.Get("/jellystat/user-activity", h.Settings.GetJellystatUserActivity)
		api.Get("/jellystat/library-stats", h.Settings.GetJellystatLibraryStats)
	}
}

// Cleanup releases resources used by the test app
func (ta *TestApp) Cleanup() {
	if ta.DB != nil {
		sqlDB, err := ta.DB.DB()
		if err == nil {
			if cerr := sqlDB.Close(); cerr != nil {
				ta.Logger.Warn("Error closing test database", "error", cerr)
			}
		}
	}
}
