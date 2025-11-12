package handler

import (
	"context"
	"time"

	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	repos        *repository.Repositories
	logger       *logger.Logger
	config       *config.Config
	syncService  *service.SyncService
	dockerClient *service.DockerClient
}

func NewDashboardHandler(repos *repository.Repositories, logger *logger.Logger, cfg *config.Config, syncService *service.SyncService, dockerClient *service.DockerClient) *DashboardHandler {
	return &DashboardHandler{
		repos:        repos,
		logger:       logger,
		config:       cfg,
		syncService:  syncService,
		dockerClient: dockerClient,
	}
}

func (h *DashboardHandler) Index(c *fiber.Ctx) error {
	return c.Render("pages/dashboard", fiber.Map{
		"Title": "Dashboard - KeeperCheky",
	}, "layouts/main")
}

func (h *DashboardHandler) Stats(c *fiber.Ctx) error {
	stats, err := h.repos.Media.GetStats()
	if err != nil {
		h.logger.Error("Failed to get stats", "error", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to get statistics",
		})
	}

	// Add placeholder values for features not yet implemented
	stats["to_delete"] = 0
	stats["leaving_soon"] = 0

	return c.JSON(stats)
}

// GetJellyseerrStats returns detailed Jellyseerr statistics.
func (h *DashboardHandler) GetJellyseerrStats(c *fiber.Ctx) error {
	if h.config == nil || !h.config.Clients.Jellyseerr.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Jellyseerr service is disabled",
		})
	}
	if h.config.Clients.Jellyseerr.URL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Jellyseerr URL not configured",
		})
	}
	if h.syncService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Sync service unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	requestStats, err := h.syncService.GetJellyseerrRequestStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get Jellyseerr stats",
			"error", err,
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(requestStats)
}

// GetJellyseerrRequests returns recent Jellyseerr requests.
func (h *DashboardHandler) GetJellyseerrRequests(c *fiber.Ctx) error {
	if h.config == nil || !h.config.Clients.Jellyseerr.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Jellyseerr service is disabled",
		})
	}
	if h.config.Clients.Jellyseerr.URL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Jellyseerr URL not configured",
		})
	}
	if h.syncService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Sync service unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	requests, err := h.syncService.GetJellyseerrRequests(ctx)
	if err != nil {
		h.logger.Error("Failed to get Jellyseerr requests",
			"error", err,
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"requests": requests,
		"count":    len(requests),
	})
}

// GetJellystatStats returns detailed Jellystat statistics for the dashboard.
func (h *DashboardHandler) GetJellystatStats(c *fiber.Ctx) error {
	if h.config == nil || !h.config.Clients.Jellystat.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Jellystat service is disabled",
		})
	}
	if h.config.Clients.Jellystat.URL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Jellystat URL not configured",
		})
	}
	if h.syncService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Sync service unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Get days from query params, default to 7 days for dashboard
	days := c.QueryInt("days", 7)

	stats, err := h.syncService.GetJellystatStatistics(ctx, days)
	if err != nil {
		h.logger.Error("Failed to get Jellystat stats",
			"error", err,
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(stats)
}

// GetJellystatViewsByType returns views by library type for the dashboard.
func (h *DashboardHandler) GetJellystatViewsByType(c *fiber.Ctx) error {
	if h.config == nil || !h.config.Clients.Jellystat.Enabled {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Jellystat service is disabled",
		})
	}
	if h.config.Clients.Jellystat.URL == "" {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Jellystat URL not configured",
		})
	}
	if h.syncService == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Sync service unavailable",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get days from query params, default to 7 days for dashboard
	days := c.QueryInt("days", 7)

	views, err := h.syncService.GetJellystatViewsByLibraryType(ctx, days)
	if err != nil {
		h.logger.Error("Failed to get Jellystat views by type",
			"error", err,
		)
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(views)
}

// addServiceStats adds service statistics to the stats map
func (h *DashboardHandler) addServiceStats(stats fiber.Map) {
	services, err := h.repos.Service.GetAll()
	if err != nil {
		h.logger.Error("Failed to get services", "error", err)
		return
	}

	activeServices := 0
	stoppedServices := 0
	for _, svc := range services {
		if svc.Enabled && svc.Status == "running" {
			activeServices++
		} else {
			stoppedServices++
		}
	}
	stats["services_active"] = activeServices
	stats["services_stopped"] = stoppedServices
	stats["services_total"] = len(services)
}

// addContainerStats adds Docker container statistics to the stats map
func (h *DashboardHandler) addContainerStats(stats fiber.Map) {
	if h.dockerClient == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	containers, err := h.dockerClient.ListContainers(ctx)
	if err != nil {
		h.logger.Error("Failed to list containers", "error", err)
		return
	}

	runningContainers := 0
	for _, container := range containers {
		if container.Status == service.ContainerStatusRunning {
			runningContainers++
		}
	}
	stats["containers_running"] = runningContainers
	stats["containers_total"] = len(containers)
}

// GetDashboardStats returns general statistics for the dashboard
// Handles GET /api/dashboard/stats
func (h *DashboardHandler) GetDashboardStats(c *fiber.Ctx) error {
	stats := fiber.Map{}
	h.addServiceStats(stats)
	h.addContainerStats(stats)

	return c.JSON(APIResponse{
		Success: true,
		Data:    stats,
	})
}

// HealthCheck returns health status of all enabled services
// Handles GET /api/dashboard/health
func (h *DashboardHandler) HealthCheck(c *fiber.Ctx) error {
	services, err := h.repos.Service.GetEnabled()
	if err != nil {
		h.logger.Error("Failed to get enabled services", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve services",
		})
	}

	healthStatus := make([]fiber.Map, 0, len(services))

	for _, svc := range services {
		status := fiber.Map{
			"name":         svc.Name,
			"display_name": svc.DisplayName,
			"enabled":      svc.Enabled,
			"status":       svc.Status,
			"healthy":      false,
		}

		// Check container health if Docker client is available and container exists
		if h.dockerClient != nil && svc.ContainerID != "" {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			containerStatus, err := h.dockerClient.GetContainerStatus(ctx, svc.ContainerID)
			cancel()

			if err != nil {
				h.logger.Error("Failed to get container status",
					"service", svc.Name,
					"container_id", svc.ContainerID,
					"error", err)
				status["error"] = err.Error()
			} else {
				status["container_status"] = string(containerStatus)
				status["healthy"] = containerStatus == service.ContainerStatusRunning
			}
		}

		healthStatus = append(healthStatus, status)
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    healthStatus,
	})
}
