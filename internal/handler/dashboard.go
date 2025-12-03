package handler

import (
	"context"
	"time"

	"github.com/carcheky/mediacheky/internal/config"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/sys/unix"
)

type DashboardHandler struct {
	repos        *repository.Repositories
	logger       *logger.Logger
	config       *config.Config
	dockerClient *service.DockerClient
}

func NewDashboardHandler(repos *repository.Repositories, logger *logger.Logger, cfg *config.Config, dockerClient *service.DockerClient) *DashboardHandler {
	return &DashboardHandler{
		repos:        repos,
		logger:       logger,
		config:       cfg,
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

// SystemInfo returns system information including disk space for the media library
func (h *DashboardHandler) SystemInfo(c *fiber.Ctx) error {
	// Media library is always mounted at /MEDIACHEKY_LIBRARY in container
	mediaPath := "/MEDIACHEKY_LIBRARY"

	// Get filesystem statistics
	var stat unix.Statfs_t
	err := unix.Statfs(mediaPath, &stat)
	if err != nil {
		h.logger.Error("Failed to get filesystem stats",
			"path", mediaPath,
			"error", err,
		)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve disk space information",
		})
	}

	// Calculate disk space metrics
	// Note: Statfs_t fields might be uint64 or int64 depending on platform
	totalBytes := uint64(stat.Blocks) * uint64(stat.Bsize)
	availableBytes := uint64(stat.Bavail) * uint64(stat.Bsize)
	usedBytes := totalBytes - (uint64(stat.Bfree) * uint64(stat.Bsize))
	usedPercent := 0.0
	if totalBytes > 0 {
		usedPercent = float64(usedBytes) / float64(totalBytes) * 100
	}

	return c.JSON(fiber.Map{
		"path":            mediaPath,
		"total_bytes":     totalBytes,
		"used_bytes":      usedBytes,
		"available_bytes": availableBytes,
		"used_percent":    usedPercent,
	})
}
