package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// DockerClient wraps the Docker SDK for managing containers
type DockerClient struct {
	client  *client.Client
	logger  *zap.Logger
	timeout time.Duration
}

// NewDockerClient creates a new Docker client wrapper
func NewDockerClient(logger *zap.Logger) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Validate socket connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := cli.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to connect to Docker socket: %w", err)
	}

	dc := &DockerClient{
		client:  cli,
		logger:  logger,
		timeout: 30 * time.Second,
	}

	logger.Info("Docker client initialized successfully")
	return dc, nil
}

// Close closes the Docker client connection
func (dc *DockerClient) Close() error {
	if dc.client != nil {
		return dc.client.Close()
	}
	return nil
}

// ListContainers lists all containers managed by MediaCheky.
// If ctx is nil, a default timeout context (30s) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dc *DockerClient) ListContainers(ctx context.Context) ([]ContainerInfo, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	// Filter for MediaCheky-managed containers
	filterArgs := filters.NewArgs()
	filterArgs.Add("label", fmt.Sprintf("%s=%s", LabelManagedBy, LabelManagedByValue))

	containers, err := dc.client.ContainerList(ctx, containertypes.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		dc.logger.Error("Failed to list containers", zap.Error(err))
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	result := make([]ContainerInfo, 0, len(containers))
	for _, c := range containers {
		info := dc.containerSummaryToInfo(c)
		result = append(result, info)
	}

	dc.logger.Debug("Listed containers", zap.Int("count", len(result)))
	return result, nil
}

// GetContainer retrieves details about a specific container
func (dc *DockerClient) GetContainer(ctx context.Context, containerID string) (*ContainerInfo, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	inspect, err := dc.client.ContainerInspect(ctx, containerID)
	if err != nil {
		dc.logger.Error("Failed to inspect container",
			zap.String("container_id", containerID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to inspect container %s: %w", containerID, err)
	}

	info := dc.inspectToInfo(inspect)
	dc.logger.Debug("Retrieved container details",
		zap.String("container_id", containerID),
		zap.String("status", string(info.Status)))

	return &info, nil
}

// StartContainer starts a stopped container
func (dc *DockerClient) StartContainer(ctx context.Context, containerID string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	dc.logger.Info("Starting container", zap.String("container_id", containerID))

	if err := dc.client.ContainerStart(ctx, containerID, containertypes.StartOptions{}); err != nil {
		dc.logger.Error("Failed to start container",
			zap.String("container_id", containerID),
			zap.Error(err))
		return fmt.Errorf("failed to start container %s: %w", containerID, err)
	}

	dc.logger.Info("Container started successfully", zap.String("container_id", containerID))
	return nil
}

// StopContainer stops a running container with a timeout
func (dc *DockerClient) StopContainer(ctx context.Context, containerID string, stopTimeout int) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	dc.logger.Info("Stopping container",
		zap.String("container_id", containerID),
		zap.Int("timeout", stopTimeout))

	timeout := stopTimeout
	if err := dc.client.ContainerStop(ctx, containerID, containertypes.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		dc.logger.Error("Failed to stop container",
			zap.String("container_id", containerID),
			zap.Error(err))
		return fmt.Errorf("failed to stop container %s: %w", containerID, err)
	}

	dc.logger.Info("Container stopped successfully", zap.String("container_id", containerID))
	return nil
}

// RestartContainer restarts a container
func (dc *DockerClient) RestartContainer(ctx context.Context, containerID string, restartTimeout int) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	dc.logger.Info("Restarting container",
		zap.String("container_id", containerID),
		zap.Int("timeout", restartTimeout))

	timeout := restartTimeout
	if err := dc.client.ContainerRestart(ctx, containerID, containertypes.StopOptions{
		Timeout: &timeout,
	}); err != nil {
		dc.logger.Error("Failed to restart container",
			zap.String("container_id", containerID),
			zap.Error(err))
		return fmt.Errorf("failed to restart container %s: %w", containerID, err)
	}

	dc.logger.Info("Container restarted successfully", zap.String("container_id", containerID))
	return nil
}

// GetContainerStatus retrieves the current status of a container
func (dc *DockerClient) GetContainerStatus(ctx context.Context, containerID string) (ContainerStatus, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	inspect, err := dc.client.ContainerInspect(ctx, containerID)
	if err != nil {
		dc.logger.Error("Failed to get container status",
			zap.String("container_id", containerID),
			zap.Error(err))
		return ContainerStatusUnknown, fmt.Errorf("failed to get container status %s: %w", containerID, err)
	}

	status := dc.mapDockerStateToStatus(inspect.State)
	dc.logger.Debug("Retrieved container status",
		zap.String("container_id", containerID),
		zap.String("status", string(status)))

	return status, nil
}

// GetContainerStats retrieves resource usage statistics for a container
func (dc *DockerClient) GetContainerStats(ctx context.Context, containerID string) (*ContainerStats, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	stats, err := dc.client.ContainerStats(ctx, containerID, false)
	if err != nil {
		dc.logger.Error("Failed to get container stats",
			zap.String("container_id", containerID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get container stats %s: %w", containerID, err)
	}
	defer stats.Body.Close()

	// Read stats
	var v containertypes.StatsResponse
	if err := json.NewDecoder(stats.Body).Decode(&v); err != nil {
		return nil, fmt.Errorf("failed to decode container stats: %w", err)
	}

	// Calculate CPU percentage
	cpuPercent := calculateCPUPercent(&v)

	// Calculate memory percentage
	memoryPercent := 0.0
	if v.MemoryStats.Limit > 0 {
		memoryPercent = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
	}

	// Network stats
	var networkRx, networkTx uint64
	for _, netStats := range v.Networks {
		networkRx += netStats.RxBytes
		networkTx += netStats.TxBytes
	}

	containerStats := &ContainerStats{
		CPUPercent:    cpuPercent,
		MemoryUsage:   v.MemoryStats.Usage,
		MemoryLimit:   v.MemoryStats.Limit,
		MemoryPercent: memoryPercent,
		NetworkRx:     networkRx,
		NetworkTx:     networkTx,
	}

	dc.logger.Debug("Retrieved container stats",
		zap.String("container_id", containerID),
		zap.Float64("cpu_percent", cpuPercent),
		zap.Float64("memory_percent", memoryPercent))

	return containerStats, nil
}

// GetContainerLogs retrieves logs from a container
func (dc *DockerClient) GetContainerLogs(ctx context.Context, containerID string, tail string) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	options := containertypes.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
		Timestamps: true,
	}

	logs, err := dc.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		dc.logger.Error("Failed to get container logs",
			zap.String("container_id", containerID),
			zap.Error(err))
		return "", fmt.Errorf("failed to get container logs %s: %w", containerID, err)
	}
	defer logs.Close()

	// Read logs
	logBytes, err := io.ReadAll(logs)
	if err != nil {
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	return string(logBytes), nil
}

// ValidateImage checks if an image is in the whitelist
func (dc *DockerClient) ValidateImage(image string) bool {
	allowed := AllowedImages[image]
	if !allowed {
		dc.logger.Warn("Image not in whitelist", zap.String("image", image))
	}
	return allowed
}

// containerSummaryToInfo converts Docker API container summary to ContainerInfo
func (dc *DockerClient) containerSummaryToInfo(c types.Container) ContainerInfo {
	name := ""
	if len(c.Names) > 0 {
		name = c.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:] // Remove leading slash
		}
	}

	ports := make([]PortBinding, 0, len(c.Ports))
	for _, port := range c.Ports {
		if port.PublicPort > 0 {
			ports = append(ports, PortBinding{
				ContainerPort: int(port.PrivatePort),
				HostPort:      int(port.PublicPort),
				Protocol:      port.Type,
			})
		}
	}

	return ContainerInfo{
		ID:        c.ID,
		Name:      name,
		Image:     c.Image,
		Status:    dc.mapStateToStatus(c.State),
		State:     c.State,
		CreatedAt: time.Unix(c.Created, 0),
		Labels:    c.Labels,
		Ports:     ports,
	}
}

// inspectToInfo converts Docker API container inspect to ContainerInfo
func (dc *DockerClient) inspectToInfo(inspect types.ContainerJSON) ContainerInfo {
	name := inspect.Name
	if len(name) > 0 && name[0] == '/' {
		name = name[1:] // Remove leading slash
	}

	ports := make([]PortBinding, 0)
	if inspect.NetworkSettings != nil {
		// Pre-calculate total number of port bindings for slice capacity
		totalBindings := 0
		for _, bindings := range inspect.NetworkSettings.Ports {
			totalBindings += len(bindings)
		}
		ports = make([]PortBinding, 0, totalBindings)

		for portProto, bindings := range inspect.NetworkSettings.Ports {
			for _, binding := range bindings {
				hostPort := 0
				if binding.HostPort != "" {
					if parsed, err := strconv.Atoi(binding.HostPort); err == nil {
						hostPort = parsed
					}
				}
				ports = append(ports, PortBinding{
					ContainerPort: portProto.Int(),
					HostPort:      hostPort,
					Protocol:      portProto.Proto(),
				})
			}
		}
	}

	createdAt, err := time.Parse(time.RFC3339, inspect.Created)
	if err != nil {
		dc.logger.Debug("Failed to parse created time", zap.Error(err), zap.String("value", inspect.Created))
		createdAt = time.Time{}
	}
	startedAt, err := time.Parse(time.RFC3339, inspect.State.StartedAt)
	if err != nil {
		dc.logger.Debug("Failed to parse started time", zap.Error(err), zap.String("value", inspect.State.StartedAt))
		startedAt = time.Time{}
	}

	return ContainerInfo{
		ID:        inspect.ID,
		Name:      name,
		Image:     inspect.Config.Image,
		Status:    dc.mapDockerStateToStatus(inspect.State),
		State:     inspect.State.Status,
		CreatedAt: createdAt,
		StartedAt: startedAt,
		Labels:    inspect.Config.Labels,
		Ports:     ports,
	}
}

// mapStateToStatus maps Docker container state string to ContainerStatus
func (dc *DockerClient) mapStateToStatus(state string) ContainerStatus {
	switch state {
	case "running":
		return ContainerStatusRunning
	case "exited", "stopped":
		return ContainerStatusStopped
	case "restarting":
		return ContainerStatusRestarting
	case "paused":
		return ContainerStatusPaused
	case "created":
		return ContainerStatusStarting
	case "dead", "removing":
		return ContainerStatusError
	default:
		return ContainerStatusUnknown
	}
}

// mapDockerStateToStatus maps Docker State struct to ContainerStatus
func (dc *DockerClient) mapDockerStateToStatus(state *types.ContainerState) ContainerStatus {
	if state == nil {
		return ContainerStatusUnknown
	}

	if state.Running {
		return ContainerStatusRunning
	}
	if state.Restarting {
		return ContainerStatusRestarting
	}
	if state.Paused {
		return ContainerStatusPaused
	}
	if state.Dead || state.OOMKilled {
		return ContainerStatusError
	}
	if state.Status == "created" {
		return ContainerStatusStarting
	}

	return ContainerStatusStopped
}

// calculateCPUPercent calculates CPU usage percentage
func calculateCPUPercent(stats *containertypes.StatsResponse) float64 {
	cpuPercent := 0.0
	cpuDelta := float64(stats.CPUStats.CPUUsage.TotalUsage) - float64(stats.PreCPUStats.CPUUsage.TotalUsage)
	systemDelta := float64(stats.CPUStats.SystemUsage) - float64(stats.PreCPUStats.SystemUsage)
	onlineCPUs := float64(stats.CPUStats.OnlineCPUs)

	if onlineCPUs == 0 {
		onlineCPUs = float64(len(stats.CPUStats.CPUUsage.PercpuUsage))
	}

	// Only calculate if we have valid values to avoid division by zero
	if systemDelta > 0.0 && cpuDelta > 0.0 && onlineCPUs > 0 {
		cpuPercent = (cpuDelta / systemDelta) * onlineCPUs * 100.0
	}

	return cpuPercent
}

// ConnectContainerToNetwork connects a container to a Docker network.
// If ctx is nil, a default timeout context (30s) will be created automatically.
func (dc *DockerClient) ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dc.timeout)
		defer cancel()
	}

	// Check if container is already connected to the network
	container, err := dc.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return fmt.Errorf("failed to inspect container: %w", err)
	}

	// Check if already connected
	if container.NetworkSettings != nil {
		for netName := range container.NetworkSettings.Networks {
			if netName == networkName {
				dc.logger.Debug("Container already connected to network",
					zap.String("container_id", containerID),
					zap.String("network", networkName))
				return nil
			}
		}
	}

	// Connect to network
	err = dc.client.NetworkConnect(ctx, networkName, containerID, nil)
	if err != nil {
		return fmt.Errorf("failed to connect container to network: %w", err)
	}

	dc.logger.Info("Container connected to network",
		zap.String("container_id", containerID),
		zap.String("network", networkName))

	return nil
}
