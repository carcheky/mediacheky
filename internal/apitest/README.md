# API Testing Framework

This package provides a comprehensive testing framework for MediaCheky's REST API endpoints.

## Overview

The API testing framework provides:
- **Automated test setup** with in-memory SQLite database
- **Request helpers** for common HTTP operations
- **Response assertions** for validating API responses
- **Test fixtures** for seeding test data
- **No user interaction required** - all tests run automatically

## Running Tests

### Run all API tests
```bash
go test ./internal/apitest/...
```

### Run with verbose output
```bash
go test -v ./internal/apitest/...
```

### Run specific test
```bash
go test -v ./internal/apitest/ -run TestHealthEndpoint
```

### Run all tests (including API tests)
```bash
make test
```

## Writing New Tests

### Basic Test Structure

```go
package apitest

import (
	"testing"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestMyEndpoint(t *testing.T) {
	// Setup test app with database
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	// Seed test data if needed
	require.NoError(t, SeedTestServices(ta))

	t.Run("My test case", func(t *testing.T) {
		// Make request
		resp := DoRequest(t, ta.App, GET("/api/my-endpoint"))

		// Assert response
		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		// Additional assertions
		assert.NotNil(t, result["data"])
	})
}
```

### Request Helpers

The framework provides convenient helpers for making HTTP requests:

```go
// GET request
resp := DoRequest(t, ta.App, GET("/api/services"))

// POST request with JSON body
body := map[string]string{"key": "value"}
resp := DoRequest(t, ta.App, POST("/api/endpoint", body))

// PUT request with JSON body
resp := DoRequest(t, ta.App, PUT("/api/endpoint", body))

// DELETE request
resp := DoRequest(t, ta.App, DELETE("/api/endpoint"))

// Request with custom headers
resp := DoRequest(t, ta.App, 
	GET("/api/endpoint").WithHeaders(map[string]string{
		"Authorization": "Bearer token",
	}))

// Request with custom content type
resp := DoRequest(t, ta.App, 
	POST("/api/endpoint", body).WithContentType("application/xml"))
```

### Response Assertions

```go
// Assert status code
AssertStatusCode(t, resp, fiber.StatusOK)

// Assert successful API response (checks success=true)
result := AssertAPISuccess(t, resp)

// Assert error API response (checks success=false)
result := AssertAPIError(t, resp)

// Assert valid JSON and unmarshal
var data MyStruct
AssertJSONResponse(t, resp, &data)
```

### Test Fixtures

Use fixtures to seed test data:

```go
// Seed test services
require.NoError(t, SeedTestServices(ta))

// Seed global configuration
require.NoError(t, SeedTestGlobalConfig(ta))

// Create individual service
service, err := CreateTestService(ta, "myservice", "My Service", true)
require.NoError(t, err)

// Get service from database
service, err := GetTestService(ta, "radarr")
require.NoError(t, err)

// Get global config
config, err := GetTestGlobalConfig(ta, "TZ")
require.NoError(t, err)

// Clear all services (for isolation between tests)
require.NoError(t, ClearTestServices(ta))
```

## Test Organization

Tests are organized by API domain:

- `health_test.go` - Health check endpoint tests
- `services_test.go` - Service management endpoint tests
- `config_test.go` - Global configuration endpoint tests
- `dashboard_test.go` - Dashboard statistics endpoint tests
- `proxy_docker_test.go` - Proxy and Docker endpoint tests

## Best Practices

### 1. Use Sub-tests
Group related tests using `t.Run()`:
```go
func TestMyFeature(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	t.Run("First scenario", func(t *testing.T) {
		// Test code
	})

	t.Run("Second scenario", func(t *testing.T) {
		// Test code
	})
}
```

### 2. Clean Up Resources
Always defer cleanup:
```go
ta := SetupTestApp(t)
defer ta.Cleanup()
```

### 3. Test Both Success and Error Cases
```go
t.Run("Success case", func(t *testing.T) {
	resp := DoRequest(t, ta.App, GET("/api/valid"))
	AssertStatusCode(t, resp, fiber.StatusOK)
	AssertAPISuccess(t, resp)
})

t.Run("Error case - not found", func(t *testing.T) {
	resp := DoRequest(t, ta.App, GET("/api/invalid"))
	AssertStatusCode(t, resp, fiber.StatusNotFound)
	AssertAPIError(t, resp)
})
```

### 4. Use Descriptive Test Names
```go
t.Run("Get service returns correct data for radarr", func(t *testing.T) {
	// Test code
})
```

### 5. Verify Data Persistence
After updates, verify the changes persisted to the database:
```go
// Update
resp := DoRequest(t, ta.App, PUT("/api/config/global/TZ", updateData))
AssertStatusCode(t, resp, fiber.StatusOK)

// Verify
cfg, err := GetTestGlobalConfig(ta, "TZ")
require.NoError(t, err)
assert.Equal(t, "Europe/London", cfg.Value)
```

## Adding New Fixtures

To add new test fixtures, edit `fixtures.go`:

```go
// Add new seeding function
func SeedTestMedia(ta *TestApp) error {
	media := []models.Media{
		{Title: "Movie 1", Type: "movie"},
		{Title: "Series 1", Type: "series"},
	}

	for _, m := range media {
		if err := ta.DB.Create(&m).Error; err != nil {
			return err
		}
	}

	return nil
}
```

## Extending Test Helpers

To add new helper functions, edit `helpers.go`:

```go
// Add custom assertion
func AssertServiceEnabled(t *testing.T, resp *HTTPResponse) {
	result := AssertAPISuccess(t, resp)
	data := result["data"].(map[string]interface{})
	assert.True(t, data["enabled"].(bool))
}
```

## CI/CD Integration

These tests are designed to run in CI/CD without any special setup:
- No external dependencies required
- No Docker daemon needed
- In-memory database (no file system writes)
- Fast execution (parallel test support)

## Troubleshooting

### Tests fail with "database connection" errors
Ensure GORM and SQLite driver are properly imported:
```go
import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)
```

### Tests fail with "handler not found"
Verify the route is properly registered in `setupTestRoutes()` in `setup.go`.

### JSON unmarshaling errors
Use `AssertJSONResponse()` which provides better error messages:
```go
var data MyStruct
AssertJSONResponse(t, resp, &data)  // Shows response body on error
```

## Future Enhancements

Potential improvements to the testing framework:
- Add support for authenticated requests
- Mock external service calls (Jellyfin, Radarr, etc.)
- Add performance/load testing utilities
- Generate test coverage reports
- Add integration tests with Docker containers
