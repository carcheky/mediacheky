package handler

import (
	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"gorm.io/gorm"
)

type Handlers struct {
	Health    *HealthHandler
	Dashboard *DashboardHandler
	Settings  *SettingsHandler
	Logs      *LogsHandler
	Service   *ServiceHandler
	Config    *ConfigHandler
	Docker    *DockerHandler
}

func NewHandlers(db *gorm.DB, repos *repository.Repositories, logger *logger.Logger, cfg *config.Config) *Handlers {
	// Initialize Docker client
	dockerClient, err := service.NewDockerClient(logger.Desugar())
	if err != nil {
		logger.Error("Failed to initialize Docker client", "error", err)
		// Continue without Docker client - some features will be unavailable
		dockerClient = nil
	}

	// Initialize Docker Compose client
	dockerCompose := service.NewDockerComposeClient(logger.Desugar())

	// Initialize Template Engine
	// Use /app/data/services for generated compose files (writable)
	// Static templates remain at /app/services (read-only)
	templateEngine := service.NewTemplateEngine(
		logger.Desugar(),
		"/app/templates",
		"/app/data/services", // Writable directory for generated compose files
		repos.Config,
		repos.Template,
	)

	// Initialize Service Manager
	var serviceManager *service.ServiceManager
	if dockerClient != nil {
		serviceManager = service.NewServiceManager(
			logger.Desugar(),
			templateEngine,
			dockerCompose,
			dockerClient,
			repos.Service,
			repos.ServiceLog,
		)
	}

	return &Handlers{
		Health:    NewHealthHandler(db, logger),
		Dashboard: NewDashboardHandler(repos, logger, nil, dockerClient),
		Settings:  NewSettingsHandler(repos, logger, cfg, nil),
		Logs:      NewLogsHandler(repos, logger),
		Service:   NewServiceHandler(repos, logger, dockerClient, serviceManager),
		Config:    NewConfigHandler(repos, logger),
		Docker:    NewDockerHandler(logger, dockerClient),
	}
}
