package handler

import (
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type ScheduleHandler struct {
	repos  *repository.Repositories
	logger *logger.Logger
}

func NewScheduleHandler(repos *repository.Repositories, logger *logger.Logger) *ScheduleHandler {
	return &ScheduleHandler{
		repos:  repos,
		logger: logger,
	}
}

func (h *ScheduleHandler) List(c *fiber.Ctx) error {
	return c.Render("pages/schedules", fiber.Map{
		"Title": "Cleanup Schedules - KeeperCheky",
	}, "layouts/main")
}
