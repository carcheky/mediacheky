package scheduler

import (
	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/pkg/logger"
	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron   *cron.Cron
	repos  *repository.Repositories
	logger *logger.Logger
	config *config.Config
}

func New(repos *repository.Repositories, logger *logger.Logger, cfg *config.Config) *Scheduler {
	return &Scheduler{
		cron:   cron.New(),
		repos:  repos,
		logger: logger,
		config: cfg,
	}
}

func (s *Scheduler) Start() {
	s.logger.Info("Starting scheduler")

	// Load enabled schedules from database
	schedules, err := s.repos.Schedule.GetEnabled()
	if err != nil {
		s.logger.Error("Failed to load schedules", "error", err)
		return
	}

	// Register each schedule
	for _, schedule := range schedules {
		s.logger.Info("Registering schedule",
			"name", schedule.Name,
			"cron", schedule.CronExpr,
		)

		// TODO: Implement actual cleanup logic
		_, err := s.cron.AddFunc(schedule.CronExpr, func() {
			s.logger.Info("Running scheduled cleanup", "schedule", schedule.Name)
			// Cleanup logic will go here
		})

		if err != nil {
			s.logger.Error("Failed to register schedule",
				"name", schedule.Name,
				"error", err,
			)
		}
	}

	s.cron.Start()
	s.logger.Info("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.logger.Info("Stopping scheduler")
	s.cron.Stop()
}
