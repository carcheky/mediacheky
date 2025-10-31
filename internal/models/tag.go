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
