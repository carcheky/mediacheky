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

	return &Handlers{
		Health:    NewHealthHandler(db, logger),
		Dashboard: NewDashboardHandler(repos, logger, nil, dockerClient),
		Settings:  NewSettingsHandler(repos, logger, cfg, nil),
		Logs:      NewLogsHandler(repos, logger),
		Service:   NewServiceHandler(repos, logger, dockerClient),
		Config:    NewConfigHandler(repos, logger),
		Docker:    NewDockerHandler(logger, dockerClient),
	}
}
