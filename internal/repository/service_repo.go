package repository

import (
	"fmt"

	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// ServiceRepository handles service data access
type ServiceRepository struct {
	db *gorm.DB
}

// NewServiceRepository creates a new ServiceRepository instance
func NewServiceRepository(db *gorm.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// GetAll retrieves all services
func (r *ServiceRepository) GetAll() ([]models.Service, error) {
	var services []models.Service
	result := r.db.Preload("Template").Find(&services)
	return services, result.Error
}

// GetByName retrieves a service by its name
func (r *ServiceRepository) GetByName(name string) (*models.Service, error) {
	var service models.Service
	result := r.db.Preload("Template").Where("name = ?", name).First(&service)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// GetByID retrieves a service by its ID
func (r *ServiceRepository) GetByID(id uint) (*models.Service, error) {
	var service models.Service
	result := r.db.Preload("Template").First(&service, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &service, nil
}

// Create creates a new service
func (r *ServiceRepository) Create(service *models.Service) error {
	return r.db.Create(service).Error
}

// Update updates an existing service
func (r *ServiceRepository) Update(service *models.Service) error {
	return r.db.Save(service).Error
}

// Delete deletes a service
func (r *ServiceRepository) Delete(id uint) error {
	return r.db.Delete(&models.Service{}, id).Error
}

// UpdateStatus updates the status of a service
func (r *ServiceRepository) UpdateStatus(id uint, status string, containerID string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if containerID != "" {
		updates["container_id"] = containerID
	}
	return r.db.Model(&models.Service{}).Where("id = ?", id).Updates(updates).Error
}

// GetEnabled retrieves all enabled services
func (r *ServiceRepository) GetEnabled() ([]models.Service, error) {
	var services []models.Service
	result := r.db.Preload("Template").Where("enabled = ?", true).Find(&services)
	return services, result.Error
}

// SetEnabled sets the enabled status of a service
func (r *ServiceRepository) SetEnabled(id uint, enabled bool) error {
	return r.db.Model(&models.Service{}).Where("id = ?", id).Update("enabled", enabled).Error
}

// CreateOrUpdate creates a new service or updates if it already exists by name
func (r *ServiceRepository) CreateOrUpdate(service *models.Service) error {
	var existing models.Service
	result := r.db.Where("name = ?", service.Name).First(&existing)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new
		return r.db.Create(service).Error
	}
	
	if result.Error != nil {
		return fmt.Errorf("error checking existing service: %w", result.Error)
	}
	
	// Update existing
	service.ID = existing.ID
	service.CreatedAt = existing.CreatedAt
	return r.db.Save(service).Error
}
