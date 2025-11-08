package repository

import (
	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// ServiceLogRepository handles service log data access
type ServiceLogRepository struct {
	db *gorm.DB
}

// NewServiceLogRepository creates a new ServiceLogRepository instance
func NewServiceLogRepository(db *gorm.DB) *ServiceLogRepository {
	return &ServiceLogRepository{db: db}
}

// Create creates a new service log entry
func (r *ServiceLogRepository) Create(log *models.ServiceLog) error {
	return r.db.Create(log).Error
}

// GetByServiceID retrieves all logs for a specific service
func (r *ServiceLogRepository) GetByServiceID(serviceID uint, limit int) ([]models.ServiceLog, error) {
	var logs []models.ServiceLog
	query := r.db.Where("service_id = ?", serviceID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	result := query.Find(&logs)
	return logs, result.Error
}

// GetRecent retrieves recent service logs across all services
func (r *ServiceLogRepository) GetRecent(limit int) ([]models.ServiceLog, error) {
	var logs []models.ServiceLog
	result := r.db.Order("created_at DESC").Limit(limit).Find(&logs)
	return logs, result.Error
}

// GetByAction retrieves logs by action type
func (r *ServiceLogRepository) GetByAction(action string, limit int) ([]models.ServiceLog, error) {
	var logs []models.ServiceLog
	result := r.db.Where("action = ?", action).Order("created_at DESC").Limit(limit).Find(&logs)
	return logs, result.Error
}

// DeleteOlderThan deletes logs older than specified days
func (r *ServiceLogRepository) DeleteOlderThan(days int) (int64, error) {
	cutoffDate := r.db.NowFunc().AddDate(0, 0, -days)
	result := r.db.Where("created_at < ?", cutoffDate).Delete(&models.ServiceLog{})
	return result.RowsAffected, result.Error
}
