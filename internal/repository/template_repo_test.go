package repository

import (
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	template := &models.Template{
		Name:    "radarr",
		Version: "1.0.0",
		Content: `version: "3.8"
services:
  radarr:
    image: linuxserver/radarr:latest
`,
		Schema: models.JSONSchema{
			"type": "object",
			"properties": map[string]interface{}{
				"port": map[string]interface{}{
					"type":    "integer",
					"default": 7878,
				},
			},
		},
	}

	err := repo.Create(template)
	assert.NoError(t, err)
	assert.NotZero(t, template.ID)
}

func TestTemplateRepository_GetByName(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	// Create a template
	template := &models.Template{
		Name:    "sonarr",
		Version: "1.0.0",
		Content: "docker-compose content",
	}
	err := repo.Create(template)
	require.NoError(t, err)

	// Retrieve by name
	retrieved, err := repo.GetByName("sonarr")
	assert.NoError(t, err)
	assert.Equal(t, "sonarr", retrieved.Name)
	assert.Equal(t, "1.0.0", retrieved.Version)
	assert.Equal(t, "docker-compose content", retrieved.Content)
}

func TestTemplateRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	// Create multiple templates
	templates := []*models.Template{
		{Name: "radarr", Version: "1.0.0", Content: "radarr template"},
		{Name: "sonarr", Version: "1.0.0", Content: "sonarr template"},
		{Name: "jellyfin", Version: "1.0.0", Content: "jellyfin template"},
	}

	for _, tmpl := range templates {
		err := repo.Create(tmpl)
		require.NoError(t, err)
	}

	// Get all templates
	allTemplates, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, allTemplates, 3)
}

func TestTemplateRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	// Create a template
	template := &models.Template{
		Name:    "radarr",
		Version: "1.0.0",
		Content: "old content",
	}
	err := repo.Create(template)
	require.NoError(t, err)

	// Update the template
	template.Version = "2.0.0"
	template.Content = "new content"
	err = repo.Update(template)
	assert.NoError(t, err)

	// Verify the update
	updated, err := repo.GetByID(template.ID)
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", updated.Version)
	assert.Equal(t, "new content", updated.Content)
}

func TestTemplateRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	// Create a template
	template := &models.Template{
		Name:    "radarr",
		Version: "1.0.0",
		Content: "content",
	}
	err := repo.Create(template)
	require.NoError(t, err)

	// Delete the template
	err = repo.Delete(template.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(template.ID)
	assert.Error(t, err)
}

func TestTemplateRepository_SchemaHandling(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTemplateRepository(db)

	// Create template with complex schema
	template := &models.Template{
		Name:    "radarr",
		Version: "1.0.0",
		Content: "content",
		Schema: models.JSONSchema{
			"type": "object",
			"properties": map[string]interface{}{
				"port": map[string]interface{}{
					"type":    "integer",
					"default": 7878,
					"minimum": 1024,
					"maximum": 65535,
				},
				"quality": map[string]interface{}{
					"type": "string",
					"enum": []interface{}{"HD-1080p", "HD-720p", "4K"},
				},
			},
			"required": []interface{}{"port"},
		},
	}

	err := repo.Create(template)
	require.NoError(t, err)

	// Retrieve and verify schema
	retrieved, err := repo.GetByID(template.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved.Schema)
	
	// Verify schema structure
	props, ok := retrieved.Schema["properties"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, props, "port")
	assert.Contains(t, props, "quality")
}
