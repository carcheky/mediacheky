package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/xeipuuv/gojsonschema"
	"go.uber.org/zap"
)

// ConfigValidator validates service configurations against JSON schemas
type ConfigValidator struct {
	logger     *zap.Logger
	schemasDir string
}

// NewConfigValidator creates a new ConfigValidator instance
func NewConfigValidator(logger *zap.Logger, schemasDir string) *ConfigValidator {
	return &ConfigValidator{
		logger:     logger,
		schemasDir: schemasDir,
	}
}

// Validate validates a service configuration against its JSON schema
func (v *ConfigValidator) Validate(serviceName string, config models.ServiceConfig) error {
	v.logger.Debug("Validating service configuration",
		zap.String("service", serviceName))

	// Load schema
	schemaContent, err := v.loadSchema(serviceName)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	// Parse schema
	schemaLoader := gojsonschema.NewStringLoader(schemaContent)

	// Convert config to JSON for validation
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	documentLoader := gojsonschema.NewBytesLoader(configJSON)

	// Validate
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if !result.Valid() {
		return v.formatValidationErrors(result.Errors())
	}

	v.logger.Debug("Service configuration is valid",
		zap.String("service", serviceName))

	return nil
}

// loadSchema loads the JSON schema for a service
func (v *ConfigValidator) loadSchema(serviceName string) (string, error) {
	schemaPath := filepath.Join(v.schemasDir, serviceName+".schema.json")
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		return "", fmt.Errorf("schema not found: %s: %w", schemaPath, err)
	}

	v.logger.Debug("Loaded schema from filesystem",
		zap.String("service", serviceName),
		zap.String("path", schemaPath))

	return string(content), nil
}

// formatValidationErrors formats validation errors into a readable error message
func (v *ConfigValidator) formatValidationErrors(errors []gojsonschema.ResultError) error {
	var msgs []string
	for _, err := range errors {
		msg := fmt.Sprintf("- %s: %s", err.Field(), err.Description())
		msgs = append(msgs, msg)
	}

	return fmt.Errorf("configuration validation failed:\n%s", strings.Join(msgs, "\n"))
}

// ValidateRequired checks that required fields are present in the configuration
func (v *ConfigValidator) ValidateRequired(config models.ServiceConfig, required []string) error {
	var missing []string
	for _, field := range required {
		if _, ok := config[field]; !ok {
			missing = append(missing, field)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missing, ", "))
	}

	return nil
}

// ValidatePort validates that a port is within the valid range
func (v *ConfigValidator) ValidatePort(port int) error {
	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535, got %d", port)
	}
	return nil
}

// ValidatePath validates that a path is not empty and is absolute
func (v *ConfigValidator) ValidatePath(path string) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	if !filepath.IsAbs(path) {
		return fmt.Errorf("path must be absolute: %s", path)
	}

	// Prevent path traversal
	cleanPath := filepath.Clean(path)
	if cleanPath != path {
		return fmt.Errorf("path contains invalid sequences: %s", path)
	}

	return nil
}

// ValidateImage validates that a Docker image is in the allowed list
func (v *ConfigValidator) ValidateImage(image string, allowedImages []string) error {
	if image == "" {
		return fmt.Errorf("image cannot be empty")
	}

	// If no whitelist provided, accept any image
	if len(allowedImages) == 0 {
		return nil
	}

	// Check if image is in whitelist
	for _, allowed := range allowedImages {
		if image == allowed || strings.HasPrefix(image, allowed+":") {
			return nil
		}
	}

	return fmt.Errorf("image not in allowed list: %s", image)
}

// ValidateContainerName validates that a container name is valid
func (v *ConfigValidator) ValidateContainerName(name string) error {
	if name == "" {
		return fmt.Errorf("container name cannot be empty")
	}

	// Container names must match [a-zA-Z0-9_-]+
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-') {
			return fmt.Errorf("container name contains invalid character: %s", name)
		}
	}

	if len(name) > 100 {
		return fmt.Errorf("container name too long (max 100 characters): %s", name)
	}

	return nil
}
