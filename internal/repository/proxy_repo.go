package repository

import (
	"fmt"

	"github.com/carcheky/mediacheky/internal/models"
	"gorm.io/gorm"
)

// ProxyRepository handles database operations for proxy configuration
type ProxyRepository struct {
	db *gorm.DB
}

// NewProxyRepository creates a new ProxyRepository
func NewProxyRepository(db *gorm.DB) *ProxyRepository {
	return &ProxyRepository{db: db}
}

// GetProxyConfig retrieves the current proxy configuration
func (r *ProxyRepository) GetProxyConfig() (*models.ProxyConfig, error) {
	var config models.ProxyConfig
	if err := r.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config if none exists
			return &models.ProxyConfig{
				ProxyType:  "traefik",
				Enabled:    false,
				SSLEnabled: false,
			}, nil
		}
		return nil, fmt.Errorf("failed to get proxy config: %w", err)
	}
	return &config, nil
}

// UpdateProxyConfig updates or creates the proxy configuration
func (r *ProxyRepository) UpdateProxyConfig(config *models.ProxyConfig) error {
	var existing models.ProxyConfig
	err := r.db.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new config
		if err := r.db.Create(config).Error; err != nil {
			return fmt.Errorf("failed to create proxy config: %w", err)
		}
		return nil
	}

	if err != nil {
		return fmt.Errorf("failed to check existing proxy config: %w", err)
	}

	// Update existing config
	config.ID = existing.ID
	if err := r.db.Save(config).Error; err != nil {
		return fmt.Errorf("failed to update proxy config: %w", err)
	}
	return nil
}

// AddDomain adds a new domain to the configuration
func (r *ProxyRepository) AddDomain(domain *models.Domain) error {
	// Check if domain already exists
	var existing models.Domain
	err := r.db.Where("name = ?", domain.Name).First(&existing).Error
	if err == nil {
		return fmt.Errorf("domain %s already exists", domain.Name)
	}

	if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing domain: %w", err)
	}

	// If this is marked as primary, remove primary flag from other domains
	if domain.IsPrimary {
		if err := r.db.Model(&models.Domain{}).Where("is_primary = ?", true).Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("failed to update existing primary domains: %w", err)
		}
	}

	if err := r.db.Create(domain).Error; err != nil {
		return fmt.Errorf("failed to add domain: %w", err)
	}
	return nil
}

// GetDomains retrieves all configured domains
func (r *ProxyRepository) GetDomains() ([]models.Domain, error) {
	var domains []models.Domain
	if err := r.db.Order("is_primary DESC, name ASC").Find(&domains).Error; err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}
	return domains, nil
}

// GetPrimaryDomain retrieves the primary domain
func (r *ProxyRepository) GetPrimaryDomain() (*models.Domain, error) {
	var domain models.Domain
	if err := r.db.Where("is_primary = ?", true).First(&domain).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No primary domain set
		}
		return nil, fmt.Errorf("failed to get primary domain: %w", err)
	}
	return &domain, nil
}

// DeleteDomain removes a domain from the configuration
func (r *ProxyRepository) DeleteDomain(name string) error {
	result := r.db.Where("name = ?", name).Delete(&models.Domain{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete domain: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("domain %s not found", name)
	}
	return nil
}

// UpdateDomain updates an existing domain
func (r *ProxyRepository) UpdateDomain(domain *models.Domain) error {
	// If setting as primary, remove primary flag from others
	if domain.IsPrimary {
		if err := r.db.Model(&models.Domain{}).Where("is_primary = ? AND id != ?", true, domain.ID).Update("is_primary", false).Error; err != nil {
			return fmt.Errorf("failed to update existing primary domains: %w", err)
		}
	}

	if err := r.db.Save(domain).Error; err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	return nil
}
