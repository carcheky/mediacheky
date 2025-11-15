package models

import (
	"time"

	"gorm.io/gorm"
)

// Tag represents a tag from external services (Radarr/Sonarr).
type Tag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// External service information
	ServiceType string `json:"service_type" gorm:"not null;index"` // "radarr" or "sonarr"
	ServiceID   int    `json:"service_id" gorm:"not null;index"`   // ID in the external service

	// Tag details
	Label string `json:"label" gorm:"not null"`
}

// TableName specifies the table name for Tag model.
func (Tag) TableName() string {
	return "tags"
}

// DockerTag represents a Docker image tag from Docker Hub
type DockerTag struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Docker image information
	ImageName string `json:"image_name" gorm:"not null;index"` // Full image name (e.g., "linuxserver/radarr")
	Tag       string `json:"tag" gorm:"not null"`              // Tag name (e.g., "latest", "5.14.0")

	// Composite unique index to prevent duplicate tags for same image
	// gorm:"uniqueIndex:idx_image_tag"
}

// TableName specifies the table name for DockerTag model.
func (DockerTag) TableName() string {
	return "docker_tags"
}
