package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestJellyfinDefaultsPath verifies that DefaultsPath and EntrypointPath are correctly set
// for Jellyfin service initialization
func TestJellyfinDefaultsPath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	// Create temporary directories
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	servicesDir := filepath.Join(tmpDir, "services")

	err := os.MkdirAll(templatesDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(servicesDir, 0755)
	assert.NoError(t, err)

	te := NewTemplateEngine(logger, templatesDir, servicesDir, configRepo, templateRepo)

	// Mock global config
	configRepo.On("GetAsMap").Return(map[string]string{
		"PUID": "1000",
		"PGID": "1000",
		"TZ":   "UTC",
	}, nil)

	// Create Jellyfin config
	jellyfinConfig := models.ServiceConfig{
		"Image":         "linuxserver/jellyfin:latest",
		"ContainerName": "jellyfin",
		"Port":          8096,
		"Paths": map[string]interface{}{
			"Config": "/data/jellyfin",
		},
		"RestartPolicy": "unless-stopped",
	}

	// Build template data
	globalConfig := GlobalConfig{
		PUID:      "1000",
		PGID:      "1000",
		Timezone:  "UTC",
		MediaPath: "/media",
	}

	data, err := te.buildTemplateData("jellyfin", jellyfinConfig, globalConfig)
	assert.NoError(t, err)

	// Verify DefaultsPath is set
	assert.NotEmpty(t, data.DefaultsPath)
	assert.Contains(t, data.DefaultsPath, "templates")
	assert.Contains(t, data.DefaultsPath, "jellyfin.defaults")
	t.Logf("DefaultsPath: %s", data.DefaultsPath)

	// Verify EntrypointPath is set
	assert.NotEmpty(t, data.EntrypointPath)
	assert.Contains(t, data.EntrypointPath, "scripts")
	assert.Contains(t, data.EntrypointPath, "jellyfin-entrypoint.sh")
	t.Logf("EntrypointPath: %s", data.EntrypointPath)

	// Verify paths have expected structure
	assert.True(t, filepath.IsAbs(data.DefaultsPath), "DefaultsPath should be absolute")
	assert.True(t, filepath.IsAbs(data.EntrypointPath), "EntrypointPath should be absolute")

	// Verify other fields are still set correctly
	assert.Equal(t, "linuxserver/jellyfin:latest", data.Image)
	assert.Equal(t, "jellyfin", data.ContainerName)
	assert.Equal(t, 8096, data.Port)
}

// TestTemplateDataDefaultsPathGeneration verifies that DefaultsPath and EntrypointPath
// are correctly included when rendering templates
func TestTemplateDataDefaultsPathGeneration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	servicesDir := filepath.Join(tmpDir, "services")

	err := os.MkdirAll(templatesDir, 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(servicesDir, 0755)
	assert.NoError(t, err)

	te := NewTemplateEngine(logger, templatesDir, servicesDir, configRepo, templateRepo)

	configRepo.On("GetAsMap").Return(map[string]string{
		"PUID": "1000",
		"PGID": "1000",
		"TZ":   "UTC",
	}, nil)

	jellyfinConfig := models.ServiceConfig{
		"Image":         "linuxserver/jellyfin:latest",
		"ContainerName": "jellyfin",
		"Port":          8096,
		"Paths": map[string]interface{}{
			"Config": "/data/jellyfin",
		},
		"RestartPolicy": "unless-stopped",
	}

	globalConfig := GlobalConfig{
		PUID:      "1000",
		PGID:      "1000",
		Timezone:  "UTC",
		MediaPath: "/media",
	}

	data, err := te.buildTemplateData("jellyfin", jellyfinConfig, globalConfig)

	// Verify structure matches expected pattern
	defaultsName := filepath.Base(data.DefaultsPath)
	assert.Equal(t, "jellyfin.defaults", defaultsName)

	entrypointName := filepath.Base(data.EntrypointPath)
	assert.Equal(t, "jellyfin-entrypoint.sh", entrypointName)
}
