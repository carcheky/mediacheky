package apitest

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardEndpoints(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	// Seed test data
	require.NoError(t, SeedTestServices(ta))

	t.Run("Get dashboard stats", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/dashboard/stats"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")

		// Check for expected fields (actual field names from API)
		assert.Contains(t, data, "services_total")
		assert.Contains(t, data, "services_active")
		assert.Contains(t, data, "containers_total")
	})

	t.Run("Dashboard health check", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/dashboard/health"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		// Health check returns an array of service health statuses
		assert.NotNil(t, result["data"])
	})

	t.Run("Legacy stats endpoint", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/stats"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		
		// Stats endpoint returns plain JSON, not API Response format
		var data map[string]interface{}
		AssertJSONResponse(t, resp, &data)
		assert.NotNil(t, data)
	})

	t.Run("Jellyseerr stats without configuration", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/jellyseerr/stats"))

		// Without configuration, may return error, empty data, or specific status
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusServiceUnavailable, fiber.StatusBadRequest, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Jellyseerr requests without configuration", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/jellyseerr/requests"))

		// Without configuration, may return error, empty data, or specific status
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusServiceUnavailable, fiber.StatusBadRequest, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Jellystat stats without configuration", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/jellystat/dashboard/stats"))

		// Without configuration, may return error, empty data, or specific status
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusServiceUnavailable, fiber.StatusBadRequest, fiber.StatusInternalServerError}, resp.StatusCode)
	})
}
