package service

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
	logger         *zap.Logger
	templatesDir   string
	servicesDir    string
	configRepo     ConfigRepository
	templateRepo   TemplateRepository
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
	return &TemplateEngine{
		logger:       logger,
		templatesDir: templatesDir,
		servicesDir:  servicesDir,
		configRepo:   configRepo,
		templateRepo: templateRepo,
	}
}

// GenerateCompose generates a docker-compose.yml file for a service
func (te *TemplateEngine) GenerateCompose(serviceName string, config models.ServiceConfig) (string, error) {
	te.logger.Info("Generating docker-compose for service",
		zap.String("service", serviceName))

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
	if image, ok := config["Image"].(string); ok {
		data.Image = image
	}
	if containerName, ok := config["ContainerName"].(string); ok {
		data.ContainerName = containerName
	}
	if port, ok := config["Port"].(float64); ok {
		data.Port = int(port)
	} else if port, ok := config["Port"].(int); ok {
		data.Port = port
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

	return composePath, nil
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
	return filepath.Join(te.servicesDir, serviceName, "docker-compose.yml")
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

		_, err := fs.ReadFile(filepath.Join(templatesPath, entry.Name()))
		if err != nil {
			te.logger.Warn("Failed to read template file",
				zap.String("file", entry.Name()),
				zap.Error(err))
			continue
		}

		te.logger.Info("Loaded template from filesystem",
			zap.String("service", serviceName),
			zap.String("file", entry.Name()))
	}

	return nil
}
