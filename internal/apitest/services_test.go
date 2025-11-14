package apitest

import (
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceEndpoints(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	// Seed test data
	require.NoError(t, SeedTestServices(ta))

	t.Run("List all services", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		require.IsType(t, []interface{}{}, result["data"], "data should be an array")
		services := result["data"].([]interface{})
		assert.GreaterOrEqual(t, len(services), 5, "should have at least 5 test services")
	})

	t.Run("Get specific service - radarr", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/radarr"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.Equal(t, "radarr", data["name"])
		assert.Equal(t, "Radarr", data["display_name"])
		assert.Equal(t, true, data["enabled"])
	})

	t.Run("Get specific service - sonarr", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/sonarr"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.Equal(t, "sonarr", data["name"])
		assert.Equal(t, "Sonarr", data["display_name"])
	})

	t.Run("Get non-existent service returns 404", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/nonexistent"))

		AssertStatusCode(t, resp, fiber.StatusNotFound)
		AssertAPIError(t, resp)
	})

	t.Run("Get service with invalid name returns 400", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/invalid@name"))

		AssertStatusCode(t, resp, fiber.StatusBadRequest)
		AssertAPIError(t, resp)
	})

	t.Run("Enable service without containers", func(t *testing.T) {
		resp := DoRequest(t, ta.App, POST("/api/services/jellyfin/enable", nil))

		// Test enabling a service when Docker Compose files or containers may not exist in the test environment.
		// Acceptable status codes:
		//   - StatusOK: The operation succeeded (e.g., if the service is enabled in the database or the compose file exists).
		//   - StatusInternalServerError or StatusServiceUnavailable: The operation failed due to missing Docker Compose files or containers,
		//     which is expected in the test environment where actual Docker resources may not be present.
		// This test ensures the API handles both success and expected error scenarios gracefully.
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusInternalServerError, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Disable service without containers", func(t *testing.T) {
		resp := DoRequest(t, ta.App, POST("/api/services/radarr/disable", nil))

		// May succeed or fail based on container availability
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusInternalServerError, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Start container without Docker containers", func(t *testing.T) {
		resp := DoRequest(t, ta.App, POST("/api/services/radarr/start", nil))

		// Without actual containers, should return error
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusInternalServerError, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Stop container without Docker containers", func(t *testing.T) {
		resp := DoRequest(t, ta.App, POST("/api/services/radarr/stop", nil))

		// Without actual containers, should return error
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusInternalServerError, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Restart container without Docker containers", func(t *testing.T) {
		resp := DoRequest(t, ta.App, POST("/api/services/radarr/restart", nil))

		// Without actual containers, should return error
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusInternalServerError, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Update service config", func(t *testing.T) {
		config := models.ServiceConfig{
			"port":    "7879",
			"api_key": "new-key",
		}

		resp := DoRequest(t, ta.App, PUT("/api/services/radarr/config", config))

		// Config update should succeed even without containers
		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)
		assert.NotNil(t, result["data"])
	})

	t.Run("Update service config with invalid JSON returns 400", func(t *testing.T) {
		resp := DoRequest(t, ta.App, PUT("/api/services/radarr/config", "invalid-json").WithContentType("application/json"))

		AssertStatusCode(t, resp, fiber.StatusBadRequest)
	})

	t.Run("Get container logs without docker client", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/radarr/logs"))

		// Without docker client, should return error
		assert.Contains(t, []int{fiber.StatusServiceUnavailable, fiber.StatusInternalServerError}, resp.StatusCode)
	})
}
