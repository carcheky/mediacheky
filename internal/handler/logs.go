package handler

import (
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type LogsHandler struct {
	repos  *repository.Repositories
	logger *logger.Logger
}

func NewLogsHandler(repos *repository.Repositories, logger *logger.Logger) *LogsHandler {
	return &LogsHandler{
		repos:  repos,
		logger: logger,
	}
}

func (h *LogsHandler) Index(c *fiber.Ctx) error {
	return c.Render("pages/logs", fiber.Map{
		"Title": "Logs - KeeperCheky",
	}, "layouts/main")
}
