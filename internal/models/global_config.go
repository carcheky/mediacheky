package models

import (
	"time"

	"gorm.io/gorm"
)

// GlobalConfig represents global configuration variables shared across services
type GlobalConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Configuration key-value pair
	Key   string `json:"key" gorm:"uniqueIndex;not null"` // e.g., "PUID", "PGID", "TZ"
	Value string `json:"value" gorm:"type:text"`          // Configuration value
	
	// Optional categorization
	Category string `json:"category" gorm:"index"` // e.g., "system", "paths", "network"
}

func (GlobalConfig) TableName() string {
	return "global_config"
}
