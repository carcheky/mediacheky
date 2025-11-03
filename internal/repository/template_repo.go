package repository

import (
	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// TemplateRepository handles template data access
type TemplateRepository struct {
	db *gorm.DB
}

// NewTemplateRepository creates a new TemplateRepository instance
func NewTemplateRepository(db *gorm.DB) *TemplateRepository {
	return &TemplateRepository{db: db}
}

// GetAll retrieves all templates
func (r *TemplateRepository) GetAll() ([]models.Template, error) {
	var templates []models.Template
	result := r.db.Order("name ASC").Find(&templates)
	return templates, result.Error
}

// GetByName retrieves a template by its name
func (r *TemplateRepository) GetByName(name string) (*models.Template, error) {
	var template models.Template
	result := r.db.Where("name = ?", name).First(&template)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// GetByID retrieves a template by its ID
func (r *TemplateRepository) GetByID(id uint) (*models.Template, error) {
	var template models.Template
	result := r.db.First(&template, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &template, nil
}

// Create creates a new template
func (r *TemplateRepository) Create(template *models.Template) error {
	return r.db.Create(template).Error
}

// Update updates an existing template
func (r *TemplateRepository) Update(template *models.Template) error {
	return r.db.Save(template).Error
}

// Delete deletes a template
func (r *TemplateRepository) Delete(id uint) error {
	return r.db.Delete(&models.Template{}, id).Error
}
