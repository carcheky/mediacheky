package service

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockConfigRepository is a mock implementation of ConfigRepository
type MockConfigRepository struct {
	mock.Mock
}

func (m *MockConfigRepository) GetAsMap() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

// MockTemplateRepository is a mock implementation of TemplateRepository
type MockTemplateRepository struct {
	mock.Mock
}

func (m *MockTemplateRepository) GetByName(name string) (*models.Template, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Template), args.Error(1)
}

func TestNewTemplateEngine(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	te := NewTemplateEngine(logger, "/tmp/templates", "/tmp/services", configRepo, templateRepo)

	assert.NotNil(t, te)
	assert.Equal(t, "/tmp/templates", te.templatesDir)
	assert.Equal(t, "/tmp/services", te.servicesDir)
}

func TestLoadGlobalConfig(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	te := NewTemplateEngine(logger, "/tmp/templates", "/tmp/services", configRepo, templateRepo)

	tests := []struct {
		name     string
		config   map[string]string
		expected GlobalConfig
	}{
		{
			name: "full config",
			config: map[string]string{
				"PUID": "1000",
				"PGID": "1000",
				"TZ":   "Europe/Madrid",
			},
			expected: GlobalConfig{
				PUID:     "1000",
				PGID:     "1000",
				Timezone: "Europe/Madrid",
			},
		},
		{
			name:   "empty config with defaults",
			config: map[string]string{},
			expected: GlobalConfig{
				PUID:     "1000",
				PGID:     "1000",
				Timezone: "UTC",
			},
		},
		{
			name: "partial config",
			config: map[string]string{
				"TZ": "America/New_York",
			},
			expected: GlobalConfig{
				PUID:     "1000",
				PGID:     "1000",
				Timezone: "America/New_York",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configRepo.On("GetAsMap").Return(tt.config, nil).Once()

			result, err := te.loadGlobalConfig()

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
			configRepo.AssertExpectations(t)
		})
	}
}

func TestBuildTemplateData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	te := NewTemplateEngine(logger, "/tmp/templates", "/tmp/services", configRepo, templateRepo)

	globalConfig := GlobalConfig{
		PUID:     "1000",
		PGID:     "1000",
		Timezone: "UTC",
	}

	tests := []struct {
		name     string
		config   models.ServiceConfig
		expected TemplateData
	}{
		{
			name: "complete radarr config",
			config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:latest",
				"ContainerName": "radarr",
				"Port":          7878,
				"Paths": map[string]interface{}{
					"Config":    "/data/radarr",
					"Movies":    "/media/movies",
					"Downloads": "/data/downloads",
				},
				"Umask":         "022",
				"Network":       "media_network",
				"RestartPolicy": "unless-stopped",
			},
			expected: TemplateData{
				Image:         "linuxserver/radarr:latest",
				ContainerName: "radarr",
				Port:          7878,
				Paths: map[string]string{
					"Config":    "/data/radarr",
					"Movies":    "/media/movies",
					"Downloads": "/data/downloads",
				},
				Umask:         "022",
				Network:       "media_network",
				RestartPolicy: "unless-stopped",
				Global:        globalConfig,
				Custom:        map[string]interface{}{},
			},
		},
		{
			name: "minimal config with defaults",
			config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:latest",
				"ContainerName": "radarr",
				"Port":          7878,
				"Paths": map[string]interface{}{
					"Config": "/data/radarr",
					"Movies": "/media/movies",
				},
			},
			expected: TemplateData{
				Image:         "linuxserver/radarr:latest",
				ContainerName: "radarr",
				Port:          7878,
				Paths: map[string]string{
					"Config": "/data/radarr",
					"Movies": "/media/movies",
				},
				RestartPolicy: "unless-stopped",
				Global:        globalConfig,
				Custom:        map[string]interface{}{},
			},
		},
		{
			name: "config with custom fields",
			config: models.ServiceConfig{
				"Image":         "linuxserver/radarr:latest",
				"ContainerName": "radarr",
				"Port":          7878,
				"Paths": map[string]interface{}{
					"Config": "/data/radarr",
					"Movies": "/media/movies",
				},
				"CustomField":  "custom_value",
				"AnotherField": 123,
			},
			expected: TemplateData{
				Image:         "linuxserver/radarr:latest",
				ContainerName: "radarr",
				Port:          7878,
				Paths: map[string]string{
					"Config": "/data/radarr",
					"Movies": "/media/movies",
				},
				RestartPolicy: "unless-stopped",
				Global:        globalConfig,
				Custom: map[string]interface{}{
					"CustomField":  "custom_value",
					"AnotherField": 123,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := te.buildTemplateData(tt.config, globalConfig)

			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateYAML(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	te := NewTemplateEngine(logger, "/tmp/templates", "/tmp/services", configRepo, templateRepo)

	tests := []struct {
		name        string
		content     string
		expectError bool
	}{
		{
			name: "valid docker-compose",
			content: `version: '3.8'
services:
  radarr:
    image: linuxserver/radarr
    ports:
      - "7878:7878"
`,
			expectError: false,
		},
		{
			name:        "invalid YAML",
			content:     `invalid: yaml: content: [unclosed`,
			expectError: true,
		},
		{
			name: "valid YAML but no services",
			content: `version: '3.8'
volumes:
  data:
`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := te.validateYAML(tt.content)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteComposeFile(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	tmpDir := t.TempDir()
	servicesDir := filepath.Join(tmpDir, "services")

	te := NewTemplateEngine(logger, "/tmp/templates", servicesDir, configRepo, templateRepo)

	content := `version: '3.8'
services:
  radarr:
    image: linuxserver/radarr
`

	// Test writing new file
	composePath, err := te.writeComposeFile("radarr", content)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(servicesDir, "radarr", "docker-compose.yml"), composePath)

	// Verify file was created
	assert.FileExists(t, composePath)

	// Verify content
	writtenContent, err := os.ReadFile(composePath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(writtenContent))

	// Test writing to existing file (should create backup)
	newContent := `version: '3.8'
services:
  radarr:
    image: linuxserver/radarr:develop
`
	composePath2, err := te.writeComposeFile("radarr", newContent)
	assert.NoError(t, err)
	assert.Equal(t, composePath, composePath2)

	// Verify new content
	writtenContent, err = os.ReadFile(composePath)
	assert.NoError(t, err)
	assert.Equal(t, newContent, string(writtenContent))

	// Verify backup was created (check for any .backup.* files)
	serviceDir := filepath.Join(servicesDir, "radarr")
	entries, err := os.ReadDir(serviceDir)
	assert.NoError(t, err)

	backupFound := false
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".yml" && entry.Name() != "docker-compose.yml" {
			backupFound = true
			break
		}
	}
	assert.True(t, backupFound, "Backup file should have been created")
}

func TestGetComposePath(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	te := NewTemplateEngine(logger, "/tmp/templates", "/data/services", configRepo, templateRepo)

	path := te.GetComposePath("radarr")
	assert.Equal(t, "/data/services/radarr/docker-compose.yml", path)
}

func TestGenerateCompose_Integration(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	configRepo := &MockConfigRepository{}
	templateRepo := &MockTemplateRepository{}

	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	servicesDir := filepath.Join(tmpDir, "services")

	// Create templates directory and template file
	err := os.MkdirAll(templatesDir, 0755)
	assert.NoError(t, err)

	templateContent := `version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
    ports:
      - "{{ .Port }}:7878"
    restart: {{ .RestartPolicy }}
`
	templatePath := filepath.Join(templatesDir, "radarr.yml")
	err = os.WriteFile(templatePath, []byte(templateContent), 0644)
	assert.NoError(t, err)

	te := NewTemplateEngine(logger, templatesDir, servicesDir, configRepo, templateRepo)

	// Setup mock
	configRepo.On("GetAsMap").Return(map[string]string{
		"PUID": "1000",
		"PGID": "1000",
		"TZ":   "UTC",
	}, nil)

	templateRepo.On("GetByName", "radarr").Return(nil, os.ErrNotExist)

	config := models.ServiceConfig{
		"Image":         "linuxserver/radarr:latest",
		"ContainerName": "radarr",
		"Port":          7878,
		"RestartPolicy": "unless-stopped",
	}

	composePath, err := te.GenerateCompose("radarr", config)

	assert.NoError(t, err)
	assert.FileExists(t, composePath)

	// Verify generated content
	generatedContent, err := os.ReadFile(composePath)
	assert.NoError(t, err)

	expectedContent := `version: '3.8'
services:
  radarr:
    image: linuxserver/radarr:latest
    container_name: radarr
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
    ports:
      - "7878:7878"
    restart: unless-stopped
`
	assert.Equal(t, expectedContent, string(generatedContent))
}
