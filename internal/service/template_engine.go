package service

// TemplateEngine generates docker-compose files from templates

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TemplateEngine handles template parsing and docker-compose generation
type TemplateEngine struct {
	logger       *zap.Logger
	templatesDir string
	servicesDir  string
	configRepo   ConfigRepository
	templateRepo TemplateRepository
}

// ConfigRepository defines the interface for accessing global configuration
type ConfigRepository interface {
	GetAsMap() (map[string]string, error)
}

// TemplateRepository defines the interface for accessing templates
type TemplateRepository interface {
	GetByName(name string) (*models.Template, error)
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	// Service-specific configuration
	Image         string
	ContainerName string
	Port          int
	ExposePort    bool // Whether to expose port to host
	HostPort      int  // Port to expose on host (if different from Port)
	Paths         map[string]string
	Umask         string
	Network       string
	RestartPolicy string

	// Global configuration
	Global GlobalConfig

	// Additional custom fields
	Custom map[string]interface{}
}

// GlobalConfig represents global configuration variables
type GlobalConfig struct {
	PUID     string
	PGID     string
	Timezone string
}

// NewTemplateEngine creates a new TemplateEngine instance
func NewTemplateEngine(logger *zap.Logger, templatesDir, servicesDir string, configRepo ConfigRepository, templateRepo TemplateRepository) *TemplateEngine {
	// Ensure servicesDir is absolute path
	absServicesDir, err := filepath.Abs(servicesDir)
	if err != nil {
		logger.Warn("Failed to get absolute path for servicesDir, using as-is",
			zap.String("servicesDir", servicesDir),
			zap.Error(err))
		absServicesDir = servicesDir
	}

	return &TemplateEngine{
		logger:       logger,
		templatesDir: templatesDir,
		servicesDir:  absServicesDir,
		configRepo:   configRepo,
		templateRepo: templateRepo,
	}
}

// GenerateCompose validates that the service compose file exists
// Since we now use static compose files in services/ directory,
// this function only verifies the file exists and returns its path
func (te *TemplateEngine) GenerateCompose(serviceName string, config models.ServiceConfig) (string, error) {
	te.logger.Info("Validating compose file for service",
		zap.String("service", serviceName))

	// Get path to static compose file
	composePath := te.GetComposePath(serviceName)

	// Verify file exists
	if _, err := os.Stat(composePath); err != nil {
		return "", fmt.Errorf("compose file not found for service %s: %w", serviceName, err)
	}

	te.logger.Info("Compose file found",
		zap.String("service", serviceName),
		zap.String("path", composePath))

	return composePath, nil
}

// Legacy: Keep unused code for reference (can be removed later)
func (te *TemplateEngine) generateComposeOld(serviceName string, config models.ServiceConfig) (string, error) {
	// Load global configuration
	globalConfig, err := te.loadGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load global config: %w", err)
	}

	// Load template content
	templateContent, err := te.loadTemplate(serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Build template data
	templateData, err := te.buildTemplateData(config, globalConfig)
	if err != nil {
		return "", fmt.Errorf("failed to build template data: %w", err)
	}

	// Validate required fields
	if templateData.Image == "" {
		return "", fmt.Errorf("image is required but not specified in configuration")
	}

	// Parse and execute template
	tmpl, err := template.New(serviceName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	composeContent := buf.String()

	// Validate generated YAML
	if err := te.validateYAML(composeContent); err != nil {
		return "", fmt.Errorf("generated invalid YAML: %w", err)
	}

	// Write compose file
	composePath, err := te.writeComposeFile(serviceName, composeContent)
	if err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	te.logger.Info("Successfully generated docker-compose",
		zap.String("service", serviceName),
		zap.String("path", composePath))

	return composePath, nil
}

// loadGlobalConfig loads global configuration from the repository
func (te *TemplateEngine) loadGlobalConfig() (GlobalConfig, error) {
	configMap, err := te.configRepo.GetAsMap()
	if err != nil {
		return GlobalConfig{}, err
	}

	// Set defaults if not found
	puid := configMap["PUID"]
	if puid == "" {
		puid = "1000"
	}

	pgid := configMap["PGID"]
	if pgid == "" {
		pgid = "1000"
	}

	timezone := configMap["TZ"]
	if timezone == "" {
		timezone = "UTC"
	}

	return GlobalConfig{
		PUID:     puid,
		PGID:     pgid,
		Timezone: timezone,
	}, nil
}

// loadTemplate loads template content from database or filesystem
func (te *TemplateEngine) loadTemplate(serviceName string) (string, error) {
	// First try to load from database
	if te.templateRepo != nil {
		tmpl, err := te.templateRepo.GetByName(serviceName)
		if err == nil && tmpl != nil && tmpl.Content != "" {
			te.logger.Debug("Loaded template from database",
				zap.String("service", serviceName))
			return tmpl.Content, nil
		}
	}

	// Fall back to filesystem
	templatePath := filepath.Join(te.templatesDir, serviceName+".yml")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("template not found: %s: %w", templatePath, err)
	}

	te.logger.Debug("Loaded template from filesystem",
		zap.String("service", serviceName),
		zap.String("path", templatePath))

	return string(content), nil
}

// buildTemplateData builds the template data from service config and global config
func (te *TemplateEngine) buildTemplateData(config models.ServiceConfig, globalConfig GlobalConfig) (TemplateData, error) {
	data := TemplateData{
		Global: globalConfig,
		Custom: make(map[string]interface{}),
	}

	// Extract standard fields from config
	if image, ok := config["Image"].(string); ok && image != "" {
		data.Image = image
	} else if image, ok := config["image"].(string); ok && image != "" {
		// Try lowercase variant
		data.Image = image
	} else {
		// Set empty to detect missing image later
		data.Image = ""
	}
	if containerName, ok := config["ContainerName"].(string); ok {
		data.ContainerName = containerName
	}
	if port, ok := config["Port"].(float64); ok {
		data.Port = int(port)
	} else if port, ok := config["Port"].(int); ok {
		data.Port = port
	}

	// Extract port exposure configuration
	if exposePort, ok := config["ExposePort"].(bool); ok {
		data.ExposePort = exposePort
	} else {
		data.ExposePort = false // Default: do not expose
	}
	if hostPort, ok := config["HostPort"].(float64); ok {
		data.HostPort = int(hostPort)
	} else if hostPort, ok := config["HostPort"].(int); ok {
		data.HostPort = hostPort
	} else {
		data.HostPort = 0 // 0 means use service Port
	}

	// Extract paths
	if paths, ok := config["Paths"].(map[string]interface{}); ok {
		data.Paths = make(map[string]string)
		for k, v := range paths {
			if strVal, ok := v.(string); ok {
				data.Paths[k] = strVal
			}
		}
	}

	// Extract optional fields
	if umask, ok := config["Umask"].(string); ok {
		data.Umask = umask
	}
	if network, ok := config["Network"].(string); ok {
		data.Network = network
	}
	if restartPolicy, ok := config["RestartPolicy"].(string); ok {
		data.RestartPolicy = restartPolicy
	} else {
		data.RestartPolicy = "unless-stopped" // Default
	}

	// Store any additional custom fields
	standardFields := map[string]bool{
		"Image": true, "ContainerName": true, "Port": true, "Paths": true,
		"Umask": true, "Network": true, "RestartPolicy": true,
		"ExposePort": true, "HostPort": true,
	}
	for k, v := range config {
		if !standardFields[k] {
			data.Custom[k] = v
		}
	}

	return data, nil
}

// validateYAML validates that the generated content is valid YAML
func (te *TemplateEngine) validateYAML(content string) error {
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Check for required services definition
	if _, hasServices := result["services"]; !hasServices {
		return fmt.Errorf("generated YAML does not contain services definition")
	}

	return nil
}

// writeComposeFile writes the compose content to the appropriate file location
func (te *TemplateEngine) writeComposeFile(serviceName, content string) (string, error) {
	// Create service directory
	serviceDir := filepath.Join(te.servicesDir, serviceName)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create service directory: %w", err)
	}

	composePath := filepath.Join(serviceDir, "docker-compose.yml")

	// Backup existing file if it exists
	if _, err := os.Stat(composePath); err == nil {
		if err := te.backupComposeFile(composePath); err != nil {
			te.logger.Warn("Failed to backup existing compose file",
				zap.String("path", composePath),
				zap.Error(err))
		}
	}

	// Write new compose file
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	// Get absolute path before returning (required for docker compose)
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// backupComposeFile creates a backup of an existing compose file
func (te *TemplateEngine) backupComposeFile(composePath string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := composePath + ".backup." + timestamp

	content, err := os.ReadFile(composePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return err
	}

	te.logger.Info("Created backup of compose file",
		zap.String("original", composePath),
		zap.String("backup", backupPath))

	return nil
}

// GetComposePath returns the path to a service's docker-compose.yml file
func (te *TemplateEngine) GetComposePath(serviceName string) string {
	// Return absolute path to static compose files in services/ directory
	// These are version-controlled and maintained separately
	absPath, err := filepath.Abs(filepath.Join("services", serviceName+".yml"))
	if err != nil {
		te.logger.Warn("Failed to get absolute path, returning relative",
			zap.String("service", serviceName),
			zap.Error(err))
		return filepath.Join("services", serviceName+".yml")
	}
	return absPath
}

// LoadTemplatesFromFS loads templates from filesystem into the database
// This is useful for initializing the database with default templates
func (te *TemplateEngine) LoadTemplatesFromFS(fs embed.FS, templatesPath string) error {
	entries, err := fs.ReadDir(templatesPath)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		serviceName := entry.Name()[:len(entry.Name())-4] // Remove .yml extension

		te.logger.Info("Found template file",
			zap.String("service", serviceName),
			zap.String("file", entry.Name()))
	}

	return nil
}
