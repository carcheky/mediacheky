package repository

import (
	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// ConfigRepository handles global configuration data access
type ConfigRepository struct {
	db *gorm.DB
}

// NewConfigRepository creates a new ConfigRepository instance
func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// GetAll retrieves all configuration entries
func (r *ConfigRepository) GetAll() ([]models.GlobalConfig, error) {
	var configs []models.GlobalConfig
	result := r.db.Order("category ASC, key ASC").Find(&configs)
	return configs, result.Error
}

// GetByKey retrieves a configuration entry by its key
func (r *ConfigRepository) GetByKey(key string) (*models.GlobalConfig, error) {
	var config models.GlobalConfig
	result := r.db.Where("key = ?", key).First(&config)
	if result.Error != nil {
		return nil, result.Error
	}
	return &config, nil
}

// GetByCategory retrieves all configuration entries in a category
func (r *ConfigRepository) GetByCategory(category string) ([]models.GlobalConfig, error) {
	var configs []models.GlobalConfig
	result := r.db.Where("category = ?", category).Order("key ASC").Find(&configs)
	return configs, result.Error
}

// Set creates or updates a configuration entry
func (r *ConfigRepository) Set(key, value, category string) error {
	config := models.GlobalConfig{
		Key:      key,
		Value:    value,
		Category: category,
	}

	// Upsert: update if exists, create if not
	return r.db.
		Where("key = ?", key).
		Assign(map[string]interface{}{
			"value":    value,
			"category": category,
		}).
		FirstOrCreate(&config).
		Error
}

// Delete deletes a configuration entry
func (r *ConfigRepository) Delete(key string) error {
	return r.db.Where("key = ?", key).Delete(&models.GlobalConfig{}).Error
}

// GetAsMap retrieves all configuration as a key-value map
func (r *ConfigRepository) GetAsMap() (map[string]string, error) {
	configs, err := r.GetAll()
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, len(configs))
	for _, config := range configs {
		result[config.Key] = config.Value
	}

	return result, nil
}
