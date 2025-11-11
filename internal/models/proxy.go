package models

import (
	"time"

	"gorm.io/gorm"
)

// ProxyConfig represents the proxy/reverse-proxy configuration
type ProxyConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Proxy technology selection
	ProxyType string `json:"proxy_type" gorm:"default:'traefik'"` // traefik, nginx, caddy
	Enabled   bool   `json:"enabled" gorm:"default:false"`        // Whether proxy is enabled

	// Global domains configuration
	Domains string `json:"domains" gorm:"type:text"` // Comma-separated list of domains

	// SSL/TLS configuration
	SSLEnabled bool   `json:"ssl_enabled" gorm:"default:false"`
	SSLEmail   string `json:"ssl_email"` // Email for Let's Encrypt

	// Proxy container state
	ContainerID string `json:"container_id"` // Docker container ID
	Status      string `json:"status" gorm:"default:'stopped'"`
}

func (ProxyConfig) TableName() string {
	return "proxy_config"
}

// Domain represents a configured domain
type Domain struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Domain information
	Name      string `json:"name" gorm:"uniqueIndex;not null"` // e.g., "mediacheky.local", "example.com"
	IsPrimary bool   `json:"is_primary" gorm:"default:false"`  // Primary domain for auto-assignment
}

func (Domain) TableName() string {
	return "domains"
}
