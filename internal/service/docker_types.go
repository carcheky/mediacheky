package service

import (
	"time"
)

// ContainerStatus represents the status of a Docker container
type ContainerStatus string

const (
	// ContainerStatusRunning indicates the container is running
	ContainerStatusRunning ContainerStatus = "running"
	// ContainerStatusStopped indicates the container is stopped
	ContainerStatusStopped ContainerStatus = "stopped"
	// ContainerStatusError indicates the container is in an error state
	ContainerStatusError ContainerStatus = "error"
	// ContainerStatusStarting indicates the container is starting
	ContainerStatusStarting ContainerStatus = "starting"
	// ContainerStatusRestarting indicates the container is restarting
	ContainerStatusRestarting ContainerStatus = "restarting"
	// ContainerStatusPaused indicates the container is paused
	ContainerStatusPaused ContainerStatus = "paused"
	// ContainerStatusUnknown indicates the container status is unknown
	ContainerStatusUnknown ContainerStatus = "unknown"
)

// ContainerInfo represents information about a Docker container
type ContainerInfo struct {
	ID        string
	Name      string
	Image     string
	Status    ContainerStatus
	State     string
	CreatedAt time.Time
	StartedAt time.Time
	Labels    map[string]string
	Ports     []PortBinding
}

// PortBinding represents a port mapping for a container
type PortBinding struct {
	ContainerPort int
	HostPort      int
	Protocol      string
}

// ContainerStats represents resource usage statistics for a container
type ContainerStats struct {
	CPUPercent    float64
	MemoryUsage   uint64
	MemoryLimit   uint64
	MemoryPercent float64
	NetworkRx     uint64
	NetworkTx     uint64
}

// ComposeResult represents the result of a docker-compose operation
type ComposeResult struct {
	Success bool
	Output  string
	Error   string
}

// AllowedImages is a whitelist of allowed Docker images for security
var AllowedImages = map[string]bool{
	"linuxserver/radarr":          true,
	"linuxserver/sonarr":          true,
	"linuxserver/jellyfin":        true,
	"linuxserver/prowlarr":        true,
	"linuxserver/bazarr":          true,
	"linuxserver/qbittorrent":     true,
	"fallenbagel/jellyseerr":      true,
	"ghcr.io/cyb3rjak3/jellystat": true,
}

// MediaChekyLabels are the labels used to identify MediaCheky-managed containers
const (
	LabelManagedBy      = "mediacheky.managed"
	LabelServiceName    = "mediacheky.service.name"
	LabelServiceID      = "mediacheky.service.id"
	LabelManagedByValue = "true"
)
