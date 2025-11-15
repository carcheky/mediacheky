package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// DockerHandler handles Docker-related HTTP requests
type DockerHandler struct {
	logger       *logger.Logger
	dockerClient *service.DockerClient
	rawClient    *client.Client
	db           *gorm.DB
	repos        *repository.Repositories
}

// NewDockerHandler creates a new DockerHandler instance
func NewDockerHandler(logger *logger.Logger, dockerClient *service.DockerClient, db *gorm.DB, repos *repository.Repositories) *DockerHandler {
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
		db:           db,
		repos:        repos,
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

// DockerHubTagsResponse represents the response from Docker Hub tags API
type DockerHubTagsResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

// isStableTag checks if a tag is a clean version number (X.X.X)
func isStableTag(tag string) bool {
	// Exclude "latest" as it's always available by default
	if tag == "latest" {
		return false
	}

	// Only allow tags that match version pattern: X.X.X or X.X.X.X (numbers and dots only)
	// Must start with a digit and contain only digits, dots, and optionally "v" prefix
	tagClean := strings.TrimPrefix(tag, "v")

	// Check if it contains ONLY digits and dots
	for _, char := range tagClean {
		if char != '.' && (char < '0' || char > '9') {
			return false
		}
	}

	// Must contain at least one dot (to be a version)
	if !strings.Contains(tagClean, ".") {
		return false
	}

	// Must not start or end with a dot
	if strings.HasPrefix(tagClean, ".") || strings.HasSuffix(tagClean, ".") {
		return false
	}

	return true
} // fetchDockerHubTags fetches stable tags from Docker Hub API for a given image
func (h *DockerHandler) fetchDockerHubTags(imageName string) ([]string, error) {
	// Remove tag if present
	if idx := strings.Index(imageName, ":"); idx != -1 {
		imageName = imageName[:idx]
	}

	// Build Docker Hub API URL - fetch more tags to filter stable ones
	// For official images (no namespace): library/image
	// For user/org images: namespace/image
	var apiURL string
	if !strings.Contains(imageName, "/") {
		// Official image
		apiURL = fmt.Sprintf("https://hub.docker.com/v2/repositories/library/%s/tags?page_size=100&ordering=last_updated", imageName)
	} else {
		// User/org image
		apiURL = fmt.Sprintf("https://hub.docker.com/v2/repositories/%s/tags?page_size=100&ordering=last_updated", imageName)
	}

	h.logger.Info("Fetching tags from Docker Hub", "image", imageName, "url", apiURL)

	// Make HTTP request
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tags from Docker Hub: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Docker Hub API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tagsResp DockerHubTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return nil, fmt.Errorf("failed to parse Docker Hub response: %w", err)
	}

	// Extract and filter tag names - ONLY stable tags
	stableTags := make([]string, 0, 10)

	for _, result := range tagsResp.Results {
		if result.Name == "" {
			continue
		}

		// Only collect stable tags (exclude nightly, dev, architecture-specific, etc.)
		if isStableTag(result.Name) {
			stableTags = append(stableTags, result.Name)
			if len(stableTags) >= 10 {
				break
			}
		}
	}

	// Always return only stable tags, even if there are few
	return stableTags, nil
}

// GetDockerTags handles GET /api/docker/tags?image=linuxserver/radarr
// Fetches available tags from Docker Hub for a given image and stores them in database
func (h *DockerHandler) GetDockerTags(c *fiber.Ctx) error {
	imageName := c.Query("image")
	if imageName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Image name is required",
		})
	}

	// Remove tag if present for consistency
	if idx := strings.Index(imageName, ":"); idx != -1 {
		imageName = imageName[:idx]
	}

	h.logger.Info("Fetching Docker tags", "image", imageName)

	// Fetch tags from Docker Hub
	tags, err := h.fetchDockerHubTags(imageName)
	if err != nil {
		h.logger.Error("Failed to fetch Docker Hub tags", "image", imageName, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch tags: %v", err),
		})
	}

	if len(tags) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(APIResponse{
			Success: false,
			Error:   "No tags found for this image",
		})
	}

	// Delete old tags for this image
	if err := h.db.Where("image_name = ?", imageName).Delete(&models.DockerTag{}).Error; err != nil {
		h.logger.Error("Failed to delete old tags", "image", imageName, "error", err)
		// Continue anyway - not critical
	}

	// Save new tags to database
	for _, tag := range tags {
		dockerTag := models.DockerTag{
			ImageName: imageName,
			Tag:       tag,
		}
		if err := h.db.Create(&dockerTag).Error; err != nil {
			h.logger.Error("Failed to save tag", "image", imageName, "tag", tag, "error", err)
			// Continue anyway - save what we can
		}
	}

	h.logger.Info("Docker tags fetched and saved", "image", imageName, "count", len(tags))

	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"image": imageName,
			"tags":  tags,
			"count": len(tags),
		},
	})
}

// GetSavedDockerTags handles GET /api/docker/tags/saved?image=linuxserver/radarr
// Returns tags saved in database for a given image
func (h *DockerHandler) GetSavedDockerTags(c *fiber.Ctx) error {
	imageName := c.Query("image")
	if imageName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(APIResponse{
			Success: false,
			Error:   "Image name is required",
		})
	}

	// Remove tag if present for consistency
	if idx := strings.Index(imageName, ":"); idx != -1 {
		imageName = imageName[:idx]
	}

	h.logger.Info("Fetching saved Docker tags", "image", imageName)

	// Query database for saved tags
	var dockerTags []models.DockerTag
	if err := h.db.Where("image_name = ?", imageName).Order("updated_at DESC").Find(&dockerTags).Error; err != nil {
		h.logger.Error("Failed to fetch saved tags", "image", imageName, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(APIResponse{
			Success: false,
			Error:   "Failed to fetch saved tags",
		})
	}

	// Extract tag names
	tags := make([]string, 0, len(dockerTags))
	for _, dt := range dockerTags {
		tags = append(tags, dt.Tag)
	}

	h.logger.Info("Saved Docker tags fetched", "image", imageName, "count", len(tags))

	return c.JSON(APIResponse{
		Success: true,
		Data: fiber.Map{
			"image": imageName,
			"tags":  tags,
			"count": len(tags),
		},
	})
}
