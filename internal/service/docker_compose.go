package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
)

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

// ComposeUp executes 'docker compose up -d' in the specified directory
func (dcc *DockerComposeClient) ComposeUp(ctx context.Context, composePath string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	composeDir := filepath.Dir(composePath)
	if _, err := os.Stat(composePath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", composePath, err)
	}

	dcc.logger.Info("Executing docker compose up",
		zap.String("path", composePath),
		zap.String("directory", composeDir))

	// Prepare command
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "up", "-d")
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
		dcc.logger.Error("Failed to execute docker compose up",
			zap.String("path", composePath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose up failed: %w", err)
	}

	dcc.logger.Info("Docker compose up executed successfully",
		zap.String("path", composePath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeDown executes 'docker compose down' in the specified directory
func (dcc *DockerComposeClient) ComposeDown(ctx context.Context, composePath string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	composeDir := filepath.Dir(composePath)
	if _, err := os.Stat(composePath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", composePath, err)
	}

	dcc.logger.Info("Executing docker compose down",
		zap.String("path", composePath),
		zap.String("directory", composeDir))

	// Prepare command
	cmd := exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "down")
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
		dcc.logger.Error("Failed to execute docker compose down",
			zap.String("path", composePath),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose down failed: %w", err)
	}

	dcc.logger.Info("Docker compose down executed successfully",
		zap.String("path", composePath),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeRestart executes 'docker compose restart' for a specific service
func (dcc *DockerComposeClient) ComposeRestart(ctx context.Context, composePath string, serviceName string) (*ComposeResult, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate path
	if !filepath.IsAbs(composePath) {
		return nil, fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	composeDir := filepath.Dir(composePath)
	if _, err := os.Stat(composePath); err != nil {
		return nil, fmt.Errorf("compose file not found: %s: %w", composePath, err)
	}

	dcc.logger.Info("Executing docker compose restart",
		zap.String("path", composePath),
		zap.String("service", serviceName),
		zap.String("directory", composeDir))

	// Prepare command
	args := []string{"compose", "-f", composePath, "restart"}
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
			zap.String("path", composePath),
			zap.String("service", serviceName),
			zap.String("stdout", result.Output),
			zap.String("stderr", result.Error),
			zap.Error(err))
		return result, fmt.Errorf("docker compose restart failed: %w", err)
	}

	dcc.logger.Info("Docker compose restart executed successfully",
		zap.String("path", composePath),
		zap.String("service", serviceName),
		zap.String("output", result.Output))

	return result, nil
}

// ComposeLogs retrieves logs from a compose service
func (dcc *DockerComposeClient) ComposeLogs(ctx context.Context, composePath string, serviceName string, tail int) (string, error) {
	if ctx == nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), dcc.timeout)
		defer cancel()
	}

	// Validate path
	if !filepath.IsAbs(composePath) {
		return "", fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	composeDir := filepath.Dir(composePath)
	if _, err := os.Stat(composePath); err != nil {
		return "", fmt.Errorf("compose file not found: %s: %w", composePath, err)
	}

	dcc.logger.Debug("Retrieving compose logs",
		zap.String("path", composePath),
		zap.String("service", serviceName),
		zap.Int("tail", tail))

	// Prepare command
	args := []string{"compose", "-f", composePath, "logs"}
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
			zap.String("path", composePath),
			zap.String("service", serviceName),
			zap.String("stderr", stderr.String()),
			zap.Error(err))
		return "", fmt.Errorf("failed to retrieve compose logs: %w", err)
	}

	return stdout.String(), nil
}

// ValidateComposeFile validates that a docker-compose.yml file is well-formed
func (dcc *DockerComposeClient) ValidateComposeFile(composePath string) error {
	// Validate path
	if !filepath.IsAbs(composePath) {
		return fmt.Errorf("compose path must be absolute: %s", composePath)
	}

	// Check if file exists
	info, err := os.Stat(composePath)
	if err != nil {
		return fmt.Errorf("compose file not found: %s: %w", composePath, err)
	}

	if info.IsDir() {
		return fmt.Errorf("compose path is a directory, not a file: %s", composePath)
	}

	// Read file content
	content, err := os.ReadFile(composePath)
	if err != nil {
		return fmt.Errorf("failed to read compose file: %w", err)
	}

	// Basic validation - check if it contains key compose keywords
	contentStr := string(content)
	if !strings.Contains(contentStr, "services:") && !strings.Contains(contentStr, "version:") {
		return fmt.Errorf("file does not appear to be a valid docker-compose file")
	}

	dcc.logger.Debug("Compose file validated", zap.String("path", composePath))
	return nil
}

// IsDockerComposeAvailable checks if docker compose command is available
func (dcc *DockerComposeClient) IsDockerComposeAvailable() bool {
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err != nil {
		dcc.logger.Warn("Docker compose command not available", zap.Error(err))
		return false
	}
	return true
}
