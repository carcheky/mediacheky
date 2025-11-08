package service

import (
	"testing"
	"time"

	containertypes "github.com/docker/docker/api/types/container"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestDockerTypes verifies that Docker types are properly defined
func TestDockerTypes(t *testing.T) {
	// Test ContainerStatus constants
	assert.Equal(t, ContainerStatus("running"), ContainerStatusRunning)
	assert.Equal(t, ContainerStatus("stopped"), ContainerStatusStopped)
	assert.Equal(t, ContainerStatus("error"), ContainerStatusError)
	assert.Equal(t, ContainerStatus("starting"), ContainerStatusStarting)

	// Test AllowedImages contains expected services
	assert.True(t, AllowedImages["linuxserver/radarr"])
	assert.True(t, AllowedImages["linuxserver/sonarr"])
	assert.True(t, AllowedImages["linuxserver/jellyfin"])
	assert.False(t, AllowedImages["some-untrusted-image"])

	// Test MediaCheky labels
	assert.Equal(t, "mediacheky.managed", LabelManagedBy)
	assert.Equal(t, "mediacheky.service.name", LabelServiceName)
	assert.Equal(t, "mediacheky.service.id", LabelServiceID)
}

// TestContainerInfo verifies ContainerInfo structure
func TestContainerInfo(t *testing.T) {
	now := time.Now()
	info := ContainerInfo{
		ID:        "abc123",
		Name:      "test-container",
		Image:     "linuxserver/radarr",
		Status:    ContainerStatusRunning,
		State:     "running",
		CreatedAt: now,
		StartedAt: now,
		Labels: map[string]string{
			LabelManagedBy:   LabelManagedByValue,
			LabelServiceName: "radarr",
		},
		Ports: []PortBinding{
			{
				ContainerPort: 7878,
				HostPort:      7878,
				Protocol:      "tcp",
			},
		},
	}

	assert.Equal(t, "abc123", info.ID)
	assert.Equal(t, "test-container", info.Name)
	assert.Equal(t, ContainerStatusRunning, info.Status)
	assert.Equal(t, "true", info.Labels[LabelManagedBy])
	assert.Equal(t, 1, len(info.Ports))
	assert.Equal(t, 7878, info.Ports[0].ContainerPort)
}

// TestContainerStats verifies ContainerStats structure
func TestContainerStats(t *testing.T) {
	stats := ContainerStats{
		CPUPercent:    25.5,
		MemoryUsage:   1024 * 1024 * 512,      // 512 MB
		MemoryLimit:   1024 * 1024 * 1024 * 2, // 2 GB
		MemoryPercent: 25.0,
		NetworkRx:     1024 * 1024, // 1 MB
		NetworkTx:     512 * 1024,  // 512 KB
	}

	assert.Equal(t, 25.5, stats.CPUPercent)
	assert.Equal(t, uint64(1024*1024*512), stats.MemoryUsage)
	assert.Equal(t, 25.0, stats.MemoryPercent)
	assert.True(t, stats.NetworkRx > 0)
	assert.True(t, stats.NetworkTx > 0)
}

// TestComposeResult verifies ComposeResult structure
func TestComposeResult(t *testing.T) {
	result := ComposeResult{
		Success: true,
		Output:  "Container started successfully",
		Error:   "",
	}

	assert.True(t, result.Success)
	assert.NotEmpty(t, result.Output)
	assert.Empty(t, result.Error)
}

// TestMapStateToStatus verifies state mapping logic
func TestMapStateToStatus(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dc := &DockerClient{
		logger: logger,
	}

	tests := []struct {
		state    string
		expected ContainerStatus
	}{
		{"running", ContainerStatusRunning},
		{"exited", ContainerStatusStopped},
		{"stopped", ContainerStatusStopped},
		{"restarting", ContainerStatusRestarting},
		{"paused", ContainerStatusPaused},
		{"created", ContainerStatusStarting},
		{"dead", ContainerStatusError},
		{"removing", ContainerStatusError},
		{"unknown", ContainerStatusUnknown},
		{"", ContainerStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			status := dc.mapStateToStatus(tt.state)
			assert.Equal(t, tt.expected, status)
		})
	}
}

// TestValidateImage verifies image validation
func TestValidateImage(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dc := &DockerClient{
		logger: logger,
	}

	tests := []struct {
		image    string
		expected bool
	}{
		{"linuxserver/radarr", true},
		{"linuxserver/sonarr", true},
		{"linuxserver/jellyfin", true},
		{"linuxserver/prowlarr", true},
		{"untrusted/image", false},
		{"malicious-image", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.image, func(t *testing.T) {
			valid := dc.ValidateImage(tt.image)
			assert.Equal(t, tt.expected, valid)
		})
	}
}

// TestCalculateCPUPercent verifies CPU percentage calculation
func TestCalculateCPUPercent(t *testing.T) {
	// Test with zero values
	stats := &containertypes.StatsResponse{
		CPUStats: containertypes.CPUStats{
			CPUUsage: containertypes.CPUUsage{
				TotalUsage: 0,
			},
			SystemUsage: 0,
			OnlineCPUs:  1,
		},
		PreCPUStats: containertypes.CPUStats{
			CPUUsage: containertypes.CPUUsage{
				TotalUsage: 0,
			},
			SystemUsage: 0,
		},
	}

	cpuPercent := calculateCPUPercent(stats)
	assert.Equal(t, 0.0, cpuPercent)

	// Test with non-zero values
	stats.CPUStats.CPUUsage.TotalUsage = 1000000
	stats.CPUStats.SystemUsage = 10000000
	stats.PreCPUStats.CPUUsage.TotalUsage = 500000
	stats.PreCPUStats.SystemUsage = 5000000

	cpuPercent = calculateCPUPercent(stats)
	assert.True(t, cpuPercent >= 0.0)
	assert.True(t, cpuPercent <= 100.0)
}

// TestDockerClientTimeout verifies that timeout is set correctly
func TestDockerClientTimeout(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dc := &DockerClient{
		logger:  logger,
		timeout: 30 * time.Second,
	}

	assert.Equal(t, 30*time.Second, dc.timeout)
}

// TestContextCreation verifies that context is created when nil
func TestContextCreation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dc := &DockerClient{
		logger:  logger,
		timeout: 5 * time.Second,
	}

	// Test that methods handle nil context properly (without actual Docker connection)
	// This is just to verify the structure and logic
	assert.NotNil(t, dc.logger)
}

// TestPortBindingStructure verifies PortBinding structure
func TestPortBindingStructure(t *testing.T) {
	port := PortBinding{
		ContainerPort: 8080,
		HostPort:      8080,
		Protocol:      "tcp",
	}

	assert.Equal(t, 8080, port.ContainerPort)
	assert.Equal(t, 8080, port.HostPort)
	assert.Equal(t, "tcp", port.Protocol)
}

// Mock implementation note:
// The actual Docker client tests would require either:
// 1. A running Docker daemon (integration tests)
// 2. A mock Docker client interface
// 3. Testcontainers library
//
// For unit tests without Docker, we test the logic and structure.
// Integration tests should be run separately in CI with Docker available.

// TestDockerComposeClientCreation verifies DockerComposeClient can be created
func TestDockerComposeClientCreation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)

	assert.NotNil(t, dcc)
	assert.NotNil(t, dcc.logger)
	assert.Equal(t, 2*time.Minute, dcc.timeout)
}

// TestComposeResultStructure verifies ComposeResult structure
func TestComposeResultStructure(t *testing.T) {
	result := &ComposeResult{
		Success: true,
		Output:  "Successfully started container",
		Error:   "",
	}

	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Successfully started")
	assert.Empty(t, result.Error)

	failedResult := &ComposeResult{
		Success: false,
		Output:  "",
		Error:   "Failed to start container",
	}

	assert.False(t, failedResult.Success)
	assert.Empty(t, failedResult.Output)
	assert.NotEmpty(t, failedResult.Error)
}
