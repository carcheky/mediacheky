package models

import (
	"time"

	"gorm.io/gorm"
)

// Service represents a multimedia service configuration
type Service struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	DisplayName string `gorm:"not null"`
	Enabled     bool   `gorm:"default:false"`
	Image       string `gorm:"not null"`
	Port        string
	WebPort     string
	Environment string `gorm:"type:text"` // JSON encoded map[string]string
	Volumes     string `gorm:"type:text"` // JSON encoded []string
	Status      string `gorm:"default:'stopped'"`
	LastChecked *time.Time
}

// ServiceStatus represents the current status of a service
type ServiceStatus struct {
	Name      string
	Enabled   bool
	Running   bool
	Healthy   bool
	UpdatedAt time.Time
}

// ComposeConfig represents a generated docker-compose configuration
type ComposeConfig struct {
	gorm.Model
	Version  string `gorm:"not null"`
	Services string `gorm:"type:text"` // JSON encoded
	Networks string `gorm:"type:text"` // JSON encoded
	Volumes  string `gorm:"type:text"` // JSON encoded
	Active   bool   `gorm:"default:true"`
}
