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

// DockerComposeClient handles docker compose operations
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

// ComposeUp runs 'docker compose up -d' for the specified compose file
func (dcc *DockerComposeClient) ComposeUp(ctx context.Context, composePath string, configMap map[string]string) (*ComposeResult, error) {
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

	dcc.logger.Info("Executing docker compose up",
		zap.String("path", cleanPath),
		zap.String("directory", composeDir))

	// Build command
	// Note: Using 'docker compose' (new) instead of 'docker compose' (old)
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "up", "-d")
	cmd.Dir = composeDir // Run from the directory containing the compose file

	// Set environment variables from config (will be used by docker compose)
	cmd.Env = os.Environ()
	for key, value := range configMap {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

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

// ComposeUpRecreate executes 'docker compose up -d --force-recreate' to recreate containers
func (dcc *DockerComposeClient) ComposeUpRecreate(ctx context.Context, composePath string, configMap map[string]string, serviceName string) (*ComposeResult, error) {
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

	dcc.logger.Info("Executing docker compose up --force-recreate",
		zap.String("path", cleanPath),
		zap.String("directory", composeDir),
		zap.String("service", serviceName))

	// ALWAYS force remove the container first
	// Step 1: Kill immediately (faster than graceful stop)
	killCmd := exec.CommandContext(ctx, "docker", "kill", serviceName)
	if err := killCmd.Run(); err != nil {
		dcc.logger.Debug("docker kill failed (container may not be running)",
			zap.String("container", serviceName))
	} else {
		dcc.logger.Info("Killed container", zap.String("container", serviceName))
	}

	// Step 2: Force remove
	removeCmd := exec.CommandContext(ctx, "docker", "rm", "-f", serviceName)
	var removeStdout, removeStderr bytes.Buffer
	removeCmd.Stdout = &removeStdout
	removeCmd.Stderr = &removeStderr

	if err := removeCmd.Run(); err != nil {
		dcc.logger.Warn("docker rm -f failed (container may not exist)",
			zap.String("container", serviceName),
			zap.String("stderr", removeStderr.String()))
	} else {
		dcc.logger.Info("Force removed container",
			zap.String("container", serviceName))
	}

	// Build command with --force-recreate flag
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "up", "-d", "--force-recreate")
	cmd.Dir = composeDir // Run from the directory containing the compose file	// Set environment variables from config
	cmd.Env = os.Environ()
	for key, value := range configMap {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

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
		dcc.logger.Error("Failed to execute docker compose up --force-recreate",
			zap.String("path", cleanPath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose up --force-recreate failed: %w", err)
	}

	dcc.logger.Info("Docker compose up --force-recreate executed successfully",
		zap.String("path", cleanPath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeUpForceRecreate is a convenience wrapper for ComposeUpRecreate
// It extracts the service name from the compose file and calls ComposeUpRecreate
func (dcc *DockerComposeClient) ComposeUpForceRecreate(ctx context.Context, composePath string, configMap map[string]string) (*ComposeResult, error) {
	// Extract service name from compose file
	serviceName, err := dcc.extractServiceName(composePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract service name: %w", err)
	}

	return dcc.ComposeUpRecreate(ctx, composePath, configMap, serviceName)
}

// extractServiceName reads the compose file and returns the first service name
func (dcc *DockerComposeClient) extractServiceName(composePath string) (string, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return "", fmt.Errorf("failed to read compose file: %w", err)
	}

	var compose struct {
		Services map[string]interface{} `yaml:"services"`
	}

	if err := yaml.Unmarshal(data, &compose); err != nil {
		return "", fmt.Errorf("failed to parse compose file: %w", err)
	}

	if len(compose.Services) == 0 {
		return "", fmt.Errorf("no services found in compose file")
	}

	// Return the first service name
	for serviceName := range compose.Services {
		return serviceName, nil
	}

	return "", fmt.Errorf("no services found")
}

// ComposeDown executes 'docker compose down' with multiple compose files.
// The first file should be the main docker compose.yml, subsequent files override/extend it.
// If ctx is nil, a default timeout context (2 minutes) will be created automatically.
// To maintain control over operation cancellation, pass a valid context.
func (dcc *DockerComposeClient) ComposeDown(ctx context.Context, composePath string, configMap map[string]string) (*ComposeResult, error) {
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

	dcc.logger.Info("Executing docker compose down",
		zap.String("path", cleanPath),
		zap.String("directory", composeDir))

	// Prepare command: Only use the service file, NOT the main compose
	// Each service is an independent project but shares the external network
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "down")
	cmd.Dir = composeDir // Run from the directory containing the compose file

	// Set environment variables from config (will be used by docker compose)
	cmd.Env = os.Environ()
	for key, value := range configMap {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

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

// ComposeStop executes 'docker compose stop' to stop containers without removing them.
// This preserves the container so Docker can auto-restart it based on restart policy.
func (dcc *DockerComposeClient) ComposeStop(ctx context.Context, composePath string, configMap map[string]string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate and sanitize path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	cleanPath := filepath.Clean(composePath)
	if cleanPath != composePath {
		return nil, fmt.Errorf("compose path contains invalid sequences: %s", composePath)
	}

	composeDir := filepath.Dir(cleanPath)
	if _, err := os.Stat(cleanPath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", cleanPath, err)
	}

	dcc.logger.Info("Executing docker compose stop",
		zap.String("path", cleanPath),
		zap.String("directory", composeDir))

	// Execute: docker compose -f <file> stop
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "stop")
	cmd.Dir = composeDir

	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range configMap {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := &ComposeResult{
		Success: err == nil,
		Output:  stdout.String(),
		Error:   stderr.String(),
	}

	if err != nil {
		dcc.logger.Error("Failed to execute docker compose stop",
			zap.String("path", cleanPath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose stop failed: %w", err)
	}

	dcc.logger.Info("Docker compose stop executed successfully",
		zap.String("path", cleanPath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposePull executes 'docker compose pull' to update images
func (dcc *DockerComposeClient) ComposePull(ctx context.Context, composePath string, configMap map[string]string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute) // 5 min for image pull
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

	// Get project root (where main docker compose.yml is)
	projectRoot := filepath.Dir(composeDir) // services/ -> project root

	dcc.logger.Info("Executing docker compose pull",
		zap.String("path", cleanPath),
		zap.String("directory", projectRoot))

	// Build command
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", cleanPath, "pull")
	cmd.Dir = projectRoot

	// Set environment variables from config (will be used by docker compose)
	cmd.Env = os.Environ()
	for key, value := range configMap {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

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
		dcc.logger.Error("Failed to execute docker compose pull",
			zap.String("path", cleanPath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose pull failed: %w", err)
	}

	dcc.logger.Info("Docker compose pull executed successfully",
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

// ValidateComposeFile validates that a docker compose.yml file is well-formed
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
