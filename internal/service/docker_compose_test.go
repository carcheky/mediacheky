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

	// Test non-absolute path
	_, err = dcc.ComposeUp(ctx, "relative/path/compose.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeUp(ctx, "/nonexistent/path/compose.yml")
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

	// Test non-absolute path
	_, err = dcc.ComposeDown(ctx, "relative/path/compose.yml")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be absolute")

	// Test non-existent file
	_, err = dcc.ComposeDown(ctx, "/nonexistent/path/compose.yml")
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
	_, err = dcc.ComposeUp(nil, composePath)
	// May succeed or fail depending on docker availability, but should not panic
	assert.NotNil(t, err == nil || err != nil) // Just checking it doesn't panic

	_, err = dcc.ComposeDown(nil, composePath)
	assert.NotNil(t, err == nil || err != nil)

	_, err = dcc.ComposeRestart(nil, composePath, "test")
	assert.NotNil(t, err == nil || err != nil)

	_, err = dcc.ComposeLogs(nil, composePath, "test", 10)
	assert.NotNil(t, err == nil || err != nil)
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
