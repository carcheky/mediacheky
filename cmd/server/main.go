package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/handler"
	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/template/html/v2"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	lgr, err := logger.New(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer lgr.Sync()

	// Initialize database
	db, err := initDatabase(cfg.Database.Path, lgr)
	if err != nil {
		lgr.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize services
	serviceManager := service.NewServiceManager(db, lgr)
	if err := serviceManager.InitializeDefaultServices(); err != nil {
		lgr.Fatal("Failed to initialize default services", zap.Error(err))
	}

	// Initialize template engine
	engine := html.New("./web/templates", ".html")
	
	// Add custom template functions
	engine.AddFunc("add", func(a, b int) int {
		return a + b
	})

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		Views:             engine,
		ViewsLayout:       "layout",
		PassLocalsToViews: true,
		ErrorHandler:      errorHandler(lgr),
	})

	// Middleware
	app.Use(recover.New())
	app.Use(cors.New())

	// Initialize handlers
	h := handler.New(serviceManager, lgr)

	// Routes
	setupRoutes(app, h)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	lgr.Info("Starting MediaCheky server", zap.String("address", addr))
	
	if err := app.Listen(addr); err != nil {
		lgr.Fatal("Failed to start server", zap.Error(err))
	}
}

// initDatabase initializes the database connection and runs migrations
func initDatabase(path string, lgr *zap.Logger) (*gorm.DB, error) {
	// Ensure data directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.Service{}, &models.ComposeConfig{}); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	lgr.Info("Database initialized", zap.String("path", path))
	return db, nil
}

// setupRoutes configures application routes
func setupRoutes(app *fiber.App, h *handler.Handler) {
	// Web routes
	app.Get("/", h.Home)
	app.Get("/services", h.Services)

	// API routes
	api := app.Group("/api")
	api.Post("/services/:name/toggle", h.ToggleService)
	api.Get("/docker-compose", h.GetDockerCompose)

	// Download routes
	app.Get("/docker-compose/download", h.DownloadDockerCompose)

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
		})
	})
}

// errorHandler handles application errors
func errorHandler(lgr *zap.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		code := fiber.StatusInternalServerError
		if e, ok := err.(*fiber.Error); ok {
			code = e.Code
		}

		lgr.Error("Request error", 
			zap.Int("status", code), 
			zap.String("path", c.Path()),
			zap.Error(err))

		return c.Status(code).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
}
