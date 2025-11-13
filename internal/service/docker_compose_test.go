package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// TestNewDockerComposeClient verifies DockerComposeClient creation
func TestNewDockerComposeClient(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)

	assert.NotNil(t, dcc)
	assert.NotNil(t, dcc.logger)
	assert.Equal(t, 2*time.Minute, dcc.timeout)
}

// TestValidateComposeFile verifies compose file validation
func TestValidateComposeFile(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)

	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		filename    string
		expectError bool
	}{
		{
			name: "valid compose file with services",
			content: `version: '3.8'
services:
  radarr:
    image: linuxserver/radarr
    ports:
      - "7878:7878"
`,
			filename:    "docker-compose.yml",
			expectError: false,
		},
		{
			name: "valid compose file with version only",
			content: `version: '3.8'
services:
  test:
    image: test
`,
			filename:    "docker-compose-version.yml",
			expectError: false,
		},
		{
			name:        "invalid compose file - no services or version",
			content:     `just some random text`,
			filename:    "invalid.yml",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test file
			filePath := filepath.Join(tmpDir, tt.filename)
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			assert.NoError(t, err)

			// Validate
			err = dcc.ValidateComposeFile(filePath)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestValidateComposeFile_PathValidation verifies path validation
func TestValidateComposeFile_PathValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)

	// Test non-absolute path
	err = dcc.ValidateComposeFile("relative/path/compose.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	err = dcc.ValidateComposeFile("/nonexistent/path/compose.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test directory instead of file
	tmpDir := t.TempDir()
	err = dcc.ValidateComposeFile(tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "directory")
}

// TestComposeUp_PathValidation verifies ComposeUp path validation
func TestComposeUp_PathValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()
	emptyConfig := map[string]string{}

	// Test non-absolute path
	_, err = dcc.ComposeUp(ctx, "relative/path/compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeUp(ctx, "/nonexistent/path/compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestComposeDown_PathValidation verifies ComposeDown path validation
func TestComposeDown_PathValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()
	emptyConfig := map[string]string{}

	// Test non-absolute path
	_, err = dcc.ComposeDown(ctx, "relative/path/compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeDown(ctx, "/nonexistent/path/compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestComposeRestart_PathValidation verifies ComposeRestart path validation
func TestComposeRestart_PathValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()

	// Test non-absolute path
	_, err = dcc.ComposeRestart(ctx, "relative/path/compose.yml", "service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeRestart(ctx, "/nonexistent/path/compose.yml", "service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestComposeLogs_PathValidation verifies ComposeLogs path validation
func TestComposeLogs_PathValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()

	// Test non-absolute path
	_, err = dcc.ComposeLogs(ctx, "relative/path/compose.yml", "service", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeLogs(ctx, "/nonexistent/path/compose.yml", "service", 100)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestContextHandling verifies that nil context is handled properly
func TestContextHandling(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)

	// Create a temporary compose file
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "docker-compose.yml")
	content := `version: '3.8'
services:
  test:
    image: hello-world
`
	err = os.WriteFile(composePath, []byte(content), 0644)
	assert.NoError(t, err)

	// These operations will execute if docker compose is available
	// We're testing that nil context doesn't cause a panic and is handled properly
	emptyConfig := map[string]string{}
	assert.NotPanics(t, func() {
		_, _ = dcc.ComposeUp(context.TODO(), composePath, emptyConfig)
	}, "ComposeUp with context.TODO should not panic")

	assert.NotPanics(t, func() {
		_, _ = dcc.ComposeDown(context.TODO(), composePath, emptyConfig)
	}, "ComposeDown with context.TODO should not panic")

	assert.NotPanics(t, func() {
		_, _ = dcc.ComposeRestart(context.TODO(), composePath, "test")
	}, "ComposeRestart with context.TODO should not panic")

	assert.NotPanics(t, func() {
		_, _ = dcc.ComposeLogs(context.TODO(), composePath, "test", 10)
	}, "ComposeLogs with nil context should not panic")
}

// TestComposeResultFields verifies ComposeResult field access
func TestComposeResultFields(t *testing.T) {
	result := &ComposeResult{
		Success: true,
		Output:  "test output",
		Error:   "test error",
	}

	assert.True(t, result.Success)
	assert.Equal(t, "test output", result.Output)
	assert.Equal(t, "test error", result.Error)

	// Test empty result
	emptyResult := &ComposeResult{}
	assert.False(t, emptyResult.Success)
	assert.Empty(t, emptyResult.Output)
	assert.Empty(t, emptyResult.Error)
}

// TestPathTraversalPrevention verifies that path traversal is prevented
func TestPathTraversalPrevention(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()
	emptyConfig := map[string]string{}

	// Test path with traversal sequences
	_, err = dcc.ComposeUp(ctx, "/tmp/../etc/docker-compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sequences")

	_, err = dcc.ComposeDown(ctx, "/tmp/./../../docker-compose.yml", emptyConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid sequences")
}

// TestServiceNameValidation verifies that service names are validated
func TestServiceNameValidation(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	ctx := context.Background()
	tmpDir := t.TempDir()
	composePath := filepath.Join(tmpDir, "docker-compose.yml")

	// Create a valid compose file
	content := `version: '3.8'
services:
  test:
    image: hello-world
`
	err = os.WriteFile(composePath, []byte(content), 0644)
	assert.NoError(t, err)

	// Test with invalid service names
	tests := []struct {
		name        string
		serviceName string
		shouldFail  bool
	}{
		{"valid service name", "test-service", false},
		{"valid with underscores", "test_service", false},
		{"valid with numbers", "service123", false},
		{"invalid with space", "test service", true},
		{"invalid with semicolon", "test;service", true},
		{"invalid with pipe", "test|service", true},
		{"invalid with ampersand", "test&service", true},
		{"command injection attempt", "test; rm -rf /", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dcc.ComposeRestart(ctx, composePath, tt.serviceName)
			if tt.shouldFail {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid service name")
			}
			// Note: if !tt.shouldFail, it may still error due to docker not running,
			// but it should NOT be a service name validation error
		})
	}
}

// TestValidateComposeFile_YAMLParsing verifies proper YAML validation
func TestValidateComposeFile_YAMLParsing(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	dcc := NewDockerComposeClient(logger)
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		content     string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid compose file",
			content: `version: '3.8'
services:
  test:
    image: hello-world
`,
			expectError: false,
		},
		{
			name:        "invalid YAML",
			content:     "not: valid: yaml: structure",
			expectError: true,
			errorMsg:    "invalid YAML format",
		},
		{
			name: "missing services",
			content: `version: '3.8'
networks:
  default:
`,
			expectError: true,
			errorMsg:    "does not contain services definition",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, "compose-"+tt.name+".yml")
			err := os.WriteFile(filePath, []byte(tt.content), 0644)
			assert.NoError(t, err)

			err = dcc.ValidateComposeFile(filePath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
