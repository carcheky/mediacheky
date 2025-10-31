package models

import (
	"time"

	"gorm.io/gorm"
)

// Request represents a media request from Jellyseerr/Overseerr.
type Request struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// External service information
	ServiceID int `json:"service_id" gorm:"not null;index"` // ID in Jellyseerr

	// Request details
	MediaType   string    `json:"media_type" gorm:"not null"` // "movie" or "tv"
	MediaTitle  string    `json:"media_title" gorm:"not null"`
	Status      string    `json:"status" gorm:"not null;index"` // "pending", "approved", "available", "denied"
	RequestedBy string    `json:"requested_by"`
	RequestedAt time.Time `json:"requested_at" gorm:"index"`

	// Associated media
	RadarrID *int `json:"radarr_id" gorm:"index"`
	SonarrID *int `json:"sonarr_id" gorm:"index"`
}

// TableName specifies the table name for Request model.
func (Request) TableName() string {
	return "requests"
}
