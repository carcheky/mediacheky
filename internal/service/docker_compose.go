package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// serviceNameRegex validates that service names only contain safe characters
var serviceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// DockerComposeClient handles docker-compose operations
type DockerComposeClient struct {
	logger  *zap.Logger
	timeout time.Duration
}

// NewDockerComposeClient creates a new Docker Compose client
func NewDockerComposeClient(logger *zap.Logger) *DockerComposeClient {
	return &DockerComposeClient{
		logger:  logger,
		timeout: 2 * time.Minute,
	}
}

// ComposeUp executes 'docker compose up -d' with multiple compose files.
// The first file should be the main docker-compose.yml, subsequent files override/extend it.
// If ctx is nil, a default timeout context (2 minutes) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dcc *DockerComposeClient) ComposeUp(ctx context.Context, composePath string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Sanitize path to prevent traversal
	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return nil, fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	composeDir := filepath.Dir(cleanPath)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	// Get project root (where main docker-compose.yml is)
	projectRoot := filepath.Dir(composeDir) // services/ -> project root
	mainCompose := filepath.Join(projectRoot, "docker-compose.yml")

	dcc.logger.Info("Executing docker compose up with multiple files",
		zap.String("main", mainCompose),
		zap.String("service", cleanPath),
		zap.String("directory", projectRoot))

	// Prepare command: Only use the service file, NOT the main compose
	// Each service is an independent project but shares the external network
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "up", "-d")
	cmd.Dir = projectRoot

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	result := &ComposeResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil {
		dcc.logger.Error("Failed to execute docker compose up",
			zap.String("path", cleanPath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose up failed: %w", err)
	}

	dcc.logger.Info("Docker compose up executed successfully",
		zap.String("path", cleanPath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeDown executes 'docker compose down' with multiple compose files.
// The first file should be the main docker-compose.yml, subsequent files override/extend it.
// If ctx is nil, a default timeout context (2 minutes) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dcc *DockerComposeClient) ComposeDown(ctx context.Context, composePath string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Sanitize path to prevent traversal
	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return nil, fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	composeDir := filepath.Dir(cleanPath)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	// Get project root (where main docker-compose.yml is)
	projectRoot := filepath.Dir(composeDir) // services/ -> project root
	mainCompose := filepath.Join(projectRoot, "docker-compose.yml")

	dcc.logger.Info("Executing docker compose down with multiple files",
		zap.String("main", mainCompose),
		zap.String("service", cleanPath),
		zap.String("directory", projectRoot))

	// Prepare command: Only use the service file, NOT the main compose
	// Each service is an independent project but shares the external network
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "down")
	cmd.Dir = projectRoot

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	result := &ComposeResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil {
		dcc.logger.Error("Failed to execute docker compose down",
			zap.String("path", cleanPath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose down failed: %w", err)
	}

	dcc.logger.Info("Docker compose down executed successfully",
		zap.String("path", cleanPath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeRestart executes 'docker compose restart' for a specific service.
// If ctx is nil, a default timeout context (2 minutes) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dcc *DockerComposeClient) ComposeRestart(ctx context.Context, composePath string, serviceName string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Sanitize path to prevent traversal
	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return nil, fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	// Validate serviceName to prevent command injection
	if serviceName != "" && !serviceNameRegex.MatchString(serviceName) {
		return nil, fmt.Errorf("invalid service name: %s", serviceName)
	}

	composeDir := filepath.Dir(cleanPath)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	dcc.logger.Info("Executing docker compose restart",
		zap.String("path", cleanPath),
		zap.String("service", serviceName),
		zap.String("directory", composeDir))

	// Prepare command
	args := []string{"compose", "-f", cleanPath, "restart"}
	if serviceName != "" {
		args = append(args, serviceName)
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = composeDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	err := cmd.Run()

	result := &ComposeResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil {
		dcc.logger.Error("Failed to execute docker compose restart",
			zap.String("path", cleanPath),
			zap.String("service", serviceName),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose restart failed: %w", err)
	}

	dcc.logger.Info("Docker compose restart executed successfully",
		zap.String("path", cleanPath),
		zap.String("service", serviceName),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeLogs retrieves logs from a compose service.
// If ctx is nil, a default timeout context (2 minutes) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dcc *DockerComposeClient) ComposeLogs(ctx context.Context, composePath string, serviceName string, tail int) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return "", fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Sanitize path to prevent traversal
	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return "", fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	// Validate serviceName to prevent command injection
	if serviceName != "" && !serviceNameRegex.MatchString(serviceName) {
		return "", fmt.Errorf("invalid service name: %s", serviceName)
	}

	composeDir := filepath.Dir(cleanPath)
	if _, err := os.Stat(cleanPath); err != nil {
		return "", fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	dcc.logger.Debug("Retrieving compose logs",
		zap.String("path", cleanPath),
		zap.String("service", serviceName),
		zap.Int("tail", tail))

	// Prepare command
	args := []string{"compose", "-f", cleanPath, "logs"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	if serviceName != "" {
		args = append(args, serviceName)
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = composeDir

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute command
	if err := cmd.Run(); err != nil {
		dcc.logger.Error("Failed to retrieve compose logs",
			zap.String("path", cleanPath),
			zap.String("service", serviceName),
			zap.String("stderr", stderr.String()),
			zap.Error(err))
		return "", fmt.Errorf("failed to retrieve compose logs: %w", err)
	}

	return stdout.String(), nil
}

// ValidateComposeFile validates that a docker-compose.yml file is well-formed
func (dcc *DockerComposeClient) ValidateComposeFile(composePath string) error {
	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Sanitize path to prevent traversal
	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	// Check if file exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		return fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("compose path is a directory, not a file: %s", cleanPath)
	}

	// Read file content
	content, err := os.ReadFile(cleanPath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	// Parse and validate YAML structure
	var composeFile map[string]interface{}
	if err := yaml.Unmarshal(content, &composeFile); err != nil {
		return fmt.Errorf("invalid YAML format: %w", err)
	}

	// Check for required services definition
	if _, hasServices := composeFile["services"]; !hasServices {
		return fmt.Errorf("file does not contain services definition")
	}

	dcc.logger.Debug("Compose file validated", zap.String("path", cleanPath))
	return nil
}

// IsDockerComposeAvailable checks if the docker compose command is available.
// Returns true if docker compose is installed and accessible, false otherwise.
// Logs a warning if docker compose is not available.
func (dcc *DockerComposeClient) IsDockerComposeAvailable() bool {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		dcc.logger.Warn("Docker compose command not available", zap.Error(err))
		return false
	}
	return true
}
