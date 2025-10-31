package handler

import (
	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
	"gorm.io/gorm"
)

type Handlers struct {
	Health    *HealthHandler
	Dashboard *DashboardHandler
	Settings  *SettingsHandler
	Logs      *LogsHandler
}

func NewHandlers(db *gorm.DB, repos *repository.Repositories, logger *logger.Logger, cfg *config.Config) *Handlers {
	return &Handlers{
		Health:    NewHealthHandler(db, logger),
		Dashboard: NewDashboardHandler(repos, logger, nil),
		Settings:  NewSettingsHandler(repos, logger, cfg, nil),
		Logs:      NewLogsHandler(repos, logger),
	}
}
