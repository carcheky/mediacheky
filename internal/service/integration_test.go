package service_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carcheky/mediacheky/internal/database"
	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/internal/service"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRadarrConfigurationIntegration tests the full flow of configuring and starting Radarr
func TestRadarrConfigurationIntegration(t *testing.T) {
	// Setup test database
	cfg := database.Config{Path: ":memory:"}
	testLogger := logger.New("debug")
	db, err := database.Initialize(cfg, testLogger)
	require.NoError(t, err)

	// Run migrations and seed
	require.NoError(t, database.RunMigrations(db))
	require.NoError(t, database.SeedData(db))

	// Create repositories
	repos := repository.NewRepositories(db)

	// Setup temp directories
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	servicesDir := filepath.Join(tmpDir, "services")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.MkdirAll(servicesDir, 0755))

	// Create Radarr template file
	radarrTemplate := `version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
      {{- if .Umask }}
      - UMASK={{ .Umask }}
      {{- end }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
      {{- if .Paths.Downloads }}
      - {{ .Paths.Downloads }}:/downloads
      {{- end }}
    ports:
      - "{{ .Port }}:7878"
    {{- if .Network }}
    networks:
      - {{ .Network }}
    {{- end }}
    restart: {{ .RestartPolicy }}
{{- if .Network }}

networks:
  {{ .Network }}:
    external: true
{{- end }}
`
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "radarr.yml"), []byte(radarrTemplate), 0644))

	// Initialize services
	templateEngine := service.NewTemplateEngine(
		testLogger.Desugar(),
		templatesDir,
		servicesDir,
		repos.Config,
		repos.Template,
	)

	dockerCompose := service.NewDockerComposeClient(testLogger.Desugar())

	// Note: We can't test actual Docker operations in unit tests,
	// but we can test the configuration and template generation

	// Test 1: Get the Radarr service
	radarrService, err := repos.Service.GetByName("radarr")
	require.NoError(t, err)
	assert.Equal(t, "radarr", radarrService.Name)
	assert.False(t, radarrService.Enabled)

	// Test 2: Update service configuration
	newConfig := models.ServiceConfig{
		"Port":          7878,
		"Image":         "linuxserver/radarr:latest",
		"ContainerName": "radarr",
		"Paths": map[string]interface{}{
			"Config":    "/data/config/radarr",
			"Movies":    "/data/media/movies",
			"Downloads": "/data/downloads",
		},
		"RestartPolicy": "unless-stopped",
	}

	radarrService.Config = newConfig
	require.NoError(t, repos.Service.Update(radarrService))

	// Test 3: Generate docker-compose file
	composePath, err := templateEngine.GenerateCompose("radarr", newConfig)
	require.NoError(t, err)
	assert.FileExists(t, composePath)

	// Test 4: Verify compose file content
	composeContent, err := os.ReadFile(composePath)
	require.NoError(t, err)
	composeStr := string(composeContent)

	// Verify key elements are present
	assert.Contains(t, composeStr, "linuxserver/radarr:latest")
	assert.Contains(t, composeStr, "container_name: radarr")
	assert.Contains(t, composeStr, "PUID=1000")
	assert.Contains(t, composeStr, "PGID=1000")
	assert.Contains(t, composeStr, "TZ=UTC")
	assert.Contains(t, composeStr, "/data/config/radarr:/config")
	assert.Contains(t, composeStr, "/data/media/movies:/movies")
	assert.Contains(t, composeStr, "/data/downloads:/downloads")
	assert.Contains(t, composeStr, "7878:7878")
	assert.Contains(t, composeStr, "restart: unless-stopped")

	// Test 5: Validate compose file
	require.NoError(t, dockerCompose.ValidateComposeFile(composePath))

	// Test 6: Enable service
	require.NoError(t, repos.Service.SetEnabled(radarrService.ID, true))

	// Verify service is enabled
	radarrService, err = repos.Service.GetByName("radarr")
	require.NoError(t, err)
	assert.True(t, radarrService.Enabled)
}

// TestServiceManagerUpdateConfig tests the service manager's config update flow
func TestServiceManagerUpdateConfig(t *testing.T) {
	// Setup test database
	cfg := database.Config{Path: ":memory:"}
	testLogger := logger.New("debug")
	db, err := database.Initialize(cfg, testLogger)
	require.NoError(t, err)

	// Run migrations and seed
	require.NoError(t, database.RunMigrations(db))
	require.NoError(t, database.SeedData(db))

	// Create repositories
	repos := repository.NewRepositories(db)

	// Setup temp directories
	tmpDir := t.TempDir()
	templatesDir := filepath.Join(tmpDir, "templates")
	servicesDir := filepath.Join(tmpDir, "services")
	require.NoError(t, os.MkdirAll(templatesDir, 0755))
	require.NoError(t, os.MkdirAll(servicesDir, 0755))

	// Create Radarr template file
	radarrTemplate := `version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
      - {{ .Paths.Downloads }}:/downloads
    ports:
      - "{{ .Port }}:7878"
    restart: {{ .RestartPolicy }}
`
	require.NoError(t, os.WriteFile(filepath.Join(templatesDir, "radarr.yml"), []byte(radarrTemplate), 0644))

	// Initialize services
	templateEngine := service.NewTemplateEngine(
		testLogger.Desugar(),
		templatesDir,
		servicesDir,
		repos.Config,
		repos.Template,
	)

	dockerCompose := service.NewDockerComposeClient(testLogger.Desugar())

	// Note: We skip Docker client initialization for this test
	// as we're only testing config updates

	serviceManager := service.NewServiceManager(
		testLogger.Desugar(),
		templateEngine,
		dockerCompose,
		nil, // Docker client - not needed for config updates
		repos.Service,
		repos.ServiceLog,
	)

	// Test: Update service configuration
	newConfig := models.ServiceConfig{
		"Port":          7878,
		"Image":         "linuxserver/radarr:latest",
		"ContainerName": "radarr",
		"Paths": map[string]interface{}{
			"Config":    "/data/config/radarr",
			"Movies":    "/data/media/movies",
			"Downloads": "/data/downloads",
		},
		"RestartPolicy": "unless-stopped",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = serviceManager.UpdateServiceConfig(ctx, "radarr", newConfig)
	require.NoError(t, err)

	// Verify service was updated
	radarrService, err := repos.Service.GetByName("radarr")
	require.NoError(t, err)
	assert.Equal(t, 7878, radarrService.Port)
	assert.Equal(t, "linuxserver/radarr:latest", radarrService.Image)

	// Verify config content (Port might be float64 or int depending on JSON unmarshaling)
	assert.Equal(t, "radarr", radarrService.Config["ContainerName"])
	assert.Equal(t, "linuxserver/radarr:latest", radarrService.Config["Image"])
	assert.Equal(t, "unless-stopped", radarrService.Config["RestartPolicy"])

	// Check Port value (handle both int and float64)
	portVal := radarrService.Config["Port"]
	if portFloat, ok := portVal.(float64); ok {
		assert.Equal(t, float64(7878), portFloat)
	} else if portInt, ok := portVal.(int); ok {
		assert.Equal(t, 7878, portInt)
	} else {
		t.Fatalf("Port has unexpected type: %T", portVal)
	}

	// Check Paths
	paths, ok := radarrService.Config["Paths"].(map[string]interface{})
	require.True(t, ok, "Paths should be a map")
	assert.Equal(t, "/data/config/radarr", paths["Config"])
	assert.Equal(t, "/data/media/movies", paths["Movies"])
	assert.Equal(t, "/data/downloads", paths["Downloads"])
}
