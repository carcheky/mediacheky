package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewConfigValidator(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	assert.NotNil(t, validator)
	assert.Equal(t, "/tmp/schemas", validator.schemasDir)
}

func TestValidatePort(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	tests := []struct {
		name        string
		port        int
		expectError bool
	}{
		{"valid port 1024", 1024, false},
		{"valid port 8080", 8080, false},
		{"valid port 65535", 65535, false},
		{"invalid port too low", 1023, true},
		{"invalid port too high", 65536, true},
		{"invalid port zero", 0, true},
		{"invalid port negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePort(tt.port)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	tests := []struct {
		name        string
		path        string
		expectError bool
	}{
		{"valid absolute path", "/data/config", false},
		{"valid absolute path with subdirs", "/data/media/movies", false},
		{"invalid empty path", "", true},
		{"invalid relative path", "data/config", true},
		{"invalid path traversal", "/data/../etc/passwd", true},
		{"invalid path with traversal", "/data/config/../../../etc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePath(tt.path)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatorValidateImage(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	allowedImages := []string{
		"linuxserver/radarr",
		"linuxserver/sonarr",
		"linuxserver/jellyfin",
	}

	tests := []struct {
		name        string
		image       string
		allowed     []string
		expectError bool
	}{
		{"valid exact match", "linuxserver/radarr", allowedImages, false},
		{"valid with tag", "linuxserver/radarr:latest", allowedImages, false},
		{"valid different tag", "linuxserver/sonarr:develop", allowedImages, false},
		{"invalid not in whitelist", "untrusted/image", allowedImages, true},
		{"invalid empty image", "", allowedImages, true},
		{"valid no whitelist", "any/image", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateImage(tt.image, tt.allowed)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateContainerName(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	tests := []struct {
		name        string
		container   string
		expectError bool
	}{
		{"valid alphanumeric", "radarr", false},
		{"valid with dash", "my-radarr", false},
		{"valid with underscore", "my_radarr", false},
		{"valid mixed", "my-radarr_2", false},
		{"invalid empty", "", true},
		{"invalid with space", "my radarr", true},
		{"invalid with special char", "radarr!", true},
		{"invalid with dot", "radarr.service", true},
		{"invalid too long", strings.Repeat("a", 101), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateContainerName(tt.container)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRequired(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/tmp/schemas")

	tests := []struct {
		name        string
		config      models.ServiceConfig
		required    []string
		expectError bool
	}{
		{
			name: "all required present",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr",
				"Port":  7878,
				"Name":  "radarr",
			},
			required:    []string{"Image", "Port", "Name"},
			expectError: false,
		},
		{
			name: "missing one required",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr",
				"Port":  7878,
			},
			required:    []string{"Image", "Port", "Name"},
			expectError: true,
		},
		{
			name:        "empty config",
			config:      models.ServiceConfig{},
			required:    []string{"Image"},
			expectError: true,
		},
		{
			name: "no required fields",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr",
			},
			required:    []string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequired(tt.config, tt.required)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_WithSchema(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tmpDir := t.TempDir()
	schemasDir := filepath.Join(tmpDir, "schemas")

	// Create schemas directory
	err := os.MkdirAll(schemasDir, 0755)
	assert.NoError(t, err)

	// Create a test schema
	schema := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["Image", "Port"],
  "properties": {
    "Image": {
      "type": "string",
      "enum": ["linuxserver/radarr:latest", "linuxserver/radarr:develop"]
    },
    "Port": {
      "type": "integer",
      "minimum": 1024,
      "maximum": 65535
    },
    "ContainerName": {
      "type": "string"
    }
  }
}`
	schemaPath := filepath.Join(schemasDir, "radarr.schema.json")
	err = os.WriteFile(schemaPath, []byte(schema), 0644)
	assert.NoError(t, err)

	validator := NewConfigValidator(logger, schemasDir)

	tests := []struct {
		name        string
		config      models.ServiceConfig
		expectError bool
	}{
		{
			name: "valid config",
			config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:latest",
				"Port":          7878,
				"ContainerName": "radarr",
			},
			expectError: false,
		},
		{
			name: "missing required field",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr:latest",
			},
			expectError: true,
		},
		{
			name: "invalid port range",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr:latest",
				"Port":  100,
			},
			expectError: true,
		},
		{
			name: "invalid image enum",
			config: models.ServiceConfig{
				"Image": "linuxserver/radarr:nightly",
				"Port":  7878,
			},
			expectError: true,
		},
		{
			name: "valid with optional field",
			config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:develop",
				"Port":          8080,
				"ContainerName": "my-radarr",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.Validate("radarr", tt.config)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidate_SchemaNotFound(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	validator := NewConfigValidator(logger, "/nonexistent/schemas")

	config := models.ServiceConfig{
		"Image": "linuxserver/radarr",
		"Port":  7878,
	}

	err := validator.Validate("radarr", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load schema")
}
