package repository

import (
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.Service{},
		&models.Template{},
		&models.GlobalConfig{},
		&models.ServiceLog{},
	)
	require.NoError(t, err)

	return db
}

func TestServiceRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	service := &models.Service{
		Name:        "radarr",
		DisplayName: "Radarr",
		Icon:        "film",
		Enabled:     true,
		Status:      "running",
		Image:       "linuxserver/radarr:latest",
		Port:        7878,
		Config: models.ServiceConfig{
			"quality_profile": "HD-1080p",
		},
	}

	err := repo.Create(service)
	assert.NoError(t, err)
	assert.NotZero(t, service.ID)
}

func TestServiceRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create a test service
	service := &models.Service{
		Name:        "sonarr",
		DisplayName: "Sonarr",
		Enabled:     false,
		Status:      "stopped",
	}
	err := repo.Create(service)
	require.NoError(t, err)

	// Retrieve by name
	retrieved, err := repo.GetByName("sonarr")
	assert.NoError(t, err)
	assert.Equal(t, "sonarr", retrieved.Name)
	assert.Equal(t, "Sonarr", retrieved.DisplayName)
	assert.False(t, retrieved.Enabled)
}

func TestServiceRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create multiple services
	services := []*models.Service{
		{Name: "radarr", DisplayName: "Radarr", Enabled: true},
		{Name: "sonarr", DisplayName: "Sonarr", Enabled: true},
		{Name: "jellyfin", DisplayName: "Jellyfin", Enabled: false},
	}

	for _, s := range services {
		err := repo.Create(s)
		require.NoError(t, err)
	}

	// Get all services
	allServices, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allServices, 3)
}

func TestServiceRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create a service
	service := &models.Service{
		Name:        "radarr",
		DisplayName: "Radarr",
		Enabled:     false,
		Status:      "stopped",
	}
	err := repo.Create(service)
	require.NoError(t, err)

	// Update the service
	service.Enabled = true
	service.Status = "running"
	service.ContainerID = "abc123"
	err = repo.Update(service)
	assert.NoError(t, err)

	// Verify the update
	updated, err := repo.GetByID(service.ID)
	assert.NoError(t, err)
	assert.True(t, updated.Enabled)
	assert.Equal(t, "running", updated.Status)
	assert.Equal(t, "abc123", updated.ContainerID)
}

func TestServiceRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create a service
	service := &models.Service{
		Name:   "radarr",
		Status: "stopped",
	}
	err := repo.Create(service)
	require.NoError(t, err)

	// Update status
	err = repo.UpdateStatus(service.ID, "running", "container123")
	assert.NoError(t, err)

	// Verify
	updated, err := repo.GetByID(service.ID)
	assert.NoError(t, err)
	assert.Equal(t, "running", updated.Status)
	assert.Equal(t, "container123", updated.ContainerID)
}

func TestServiceRepository_GetEnabled(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create services
	services := []*models.Service{
		{Name: "radarr", Enabled: true},
		{Name: "sonarr", Enabled: true},
		{Name: "jellyfin", Enabled: false},
	}

	for _, s := range services {
		err := repo.Create(s)
		require.NoError(t, err)
	}

	// Get enabled services
	enabled, err := repo.GetEnabled()
	assert.NoError(t, err)
	assert.Len(t, enabled, 2)
}

func TestServiceRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create a service
	service := &models.Service{
		Name: "radarr",
	}
	err := repo.Create(service)
	require.NoError(t, err)

	// Delete the service
	err = repo.Delete(service.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(service.ID)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestServiceRepository_CreateOrUpdate(t *testing.T) {
	db := setupTestDB(t)
	repo := NewServiceRepository(db)

	// Create a service
	service := &models.Service{
		Name:        "radarr",
		DisplayName: "Radarr",
		Port:        7878,
	}
	err := repo.CreateOrUpdate(service)
	assert.NoError(t, err)
	originalID := service.ID

	// Update the same service
	service.Port = 7879
	service.DisplayName = "Radarr Updated"
	err = repo.CreateOrUpdate(service)
	assert.NoError(t, err)
	assert.Equal(t, originalID, service.ID) // ID should remain the same

	// Verify update
	updated, err := repo.GetByName("radarr")
	assert.NoError(t, err)
	assert.Equal(t, 7879, updated.Port)
	assert.Equal(t, "Radarr Updated", updated.DisplayName)
}

func TestServiceRepository_WithTemplate(t *testing.T) {
	db := setupTestDB(t)
	serviceRepo := NewServiceRepository(db)
	templateRepo := NewTemplateRepository(db)

	// Create a template
	template := &models.Template{
		Name:    "radarr",
		Version: "1.0.0",
		Content: "version: '3.8'...",
	}
	err := templateRepo.Create(template)
	require.NoError(t, err)

	// Create a service with template
	service := &models.Service{
		Name:       "radarr",
		TemplateID: template.ID,
	}
	err = serviceRepo.Create(service)
	require.NoError(t, err)

	// Retrieve with preloaded template
	retrieved, err := serviceRepo.GetByID(service.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.Template)
	assert.Equal(t, "radarr", retrieved.Template.Name)
	assert.Equal(t, "1.0.0", retrieved.Template.Version)
}
