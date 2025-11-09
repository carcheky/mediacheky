package handler

import (
	"context"
	"time"

	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
)

// DockerHandler handles Docker-related HTTP requests
type DockerHandler struct {
	logger       *logger.Logger
	dockerClient *service.DockerClient
	rawClient    *client.Client
}

// NewDockerHandler creates a new DockerHandler instance
func NewDockerHandler(logger *logger.Logger, dockerClient *service.DockerClient) *DockerHandler {
	// Get raw Docker client for info API
	var rawClient *client.Client
	if dockerClient != nil {
		// Create a new raw client for Docker info access
		raw, err := client.NewClientWithOpts(
			client.FromEnv,
			client.WithAPIVersionNegotiation(),
		)
		if err != nil {
			logger.Error("Failed to create raw Docker client", "error", err)
		} else {
			rawClient = raw
		}
	}

	return &DockerHandler{
		logger:       logger,
		dockerClient: dockerClient,
		rawClient:    rawClient,
	}
}

// GetDockerInfo handles GET /api/docker/info
func (h *DockerHandler) GetDockerInfo(c *fiber.Ctx) error {
	if h.rawClient == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Docker client not available",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	info, err := h.rawClient.Info(ctx)
	if err != nil {
		h.logger.Error("Failed to get Docker info", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve Docker information",
		})
	}

	// Extract relevant information
	dockerInfo := fiber.Map{
		"id":                 info.ID,
		"name":               info.Name,
		"server_version":     info.ServerVersion,
		"operating_system":   info.OperatingSystem,
		"os_type":            info.OSType,
		"architecture":       info.Architecture,
		"ncpu":               info.NCPU,
		"memory_total":       info.MemTotal,
		"docker_root_dir":    info.DockerRootDir,
		"containers":         info.Containers,
		"containers_running": info.ContainersRunning,
		"containers_paused":  info.ContainersPaused,
		"containers_stopped": info.ContainersStopped,
		"images":             info.Images,
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    dockerInfo,
	})
}

// ListContainers handles GET /api/docker/containers
func (h *DockerHandler) ListContainers(c *fiber.Ctx) error {
	if h.dockerClient == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(APIResponse{
			Success: false,
			Error:   "Docker client not available",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	containers, err := h.dockerClient.ListContainers(ctx)
	if err != nil {
		h.logger.Error("Failed to list containers", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to retrieve containers",
		})
	}

	return c.JSON(APIResponse{
		Success: true,
		Data:    containers,
	})
}
