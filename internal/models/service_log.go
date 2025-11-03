package models

import (
	"time"
)

// ServiceLog represents the operation history for services
type ServiceLog struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`

	// Service reference
	ServiceID   uint   `json:"service_id" gorm:"index;not null"`
	ServiceName string `json:"service_name" gorm:"index"` // Denormalized for query performance
	
	// Operation details
	Action  string `json:"action" gorm:"not null"`  // e.g., "start", "stop", "restart", "config_update"
	Status  string `json:"status" gorm:"not null"`  // e.g., "success", "failed", "pending"
	Message string `json:"message" gorm:"type:text"` // Detailed message or error
}

func (ServiceLog) TableName() string {
	return "service_logs"
}
