package apitest

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalConfigEndpoints(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	// Seed test global config
	require.NoError(t, SeedTestGlobalConfig(ta))

	t.Run("Get all global config", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/config/global"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		// Config is grouped by category, so check for category presence
		assert.NotEmpty(t, data, "should have global config data")
	})

	t.Run("Get specific config value - TZ", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/config/global/TZ"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.Equal(t, "TZ", data["key"])
		assert.Equal(t, "Europe/Madrid", data["value"])
	})

	t.Run("Get specific config value - PUID", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/config/global/PUID"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.Equal(t, "PUID", data["key"])
		assert.Equal(t, "1000", data["value"])
	})

	t.Run("Get specific config value - PGID", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/config/global/PGID"))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.Equal(t, "PGID", data["key"])
		assert.Equal(t, "1000", data["value"])
	})

	t.Run("Get non-existent config returns 404", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/config/global/NONEXISTENT"))

		AssertStatusCode(t, resp, fiber.StatusNotFound)
		AssertAPIError(t, resp)
	})

	t.Run("Update config value", func(t *testing.T) {
		updateData := map[string]string{
			"value": "Europe/London",
		}

		resp := DoRequest(t, ta.App, PUT("/api/config/global/TZ", updateData))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.NotNil(t, data["message"])

		// Verify the change persisted
		cfg, err := GetTestGlobalConfig(ta, "TZ")
		require.NoError(t, err)
		assert.Equal(t, "Europe/London", cfg.Value)
	})

	t.Run("Update all global config", func(t *testing.T) {
		updateData := []map[string]string{
			{"key": "PUID", "value": "1001"},
			{"key": "PGID", "value": "1001"},
			{"key": "TZ", "value": "America/New_York"},
		}

		resp := DoRequest(t, ta.App, PUT("/api/config/global", updateData))

		AssertStatusCode(t, resp, fiber.StatusOK)
		result := AssertAPISuccess(t, resp)

		// Verify response
		data, ok := result["data"].(map[string]interface{})
		require.True(t, ok, "data should be an object")
		assert.NotNil(t, data["message"])

		// Verify the changes persisted
		puid, err := GetTestGlobalConfig(ta, "PUID")
		require.NoError(t, err)
		assert.Equal(t, "1001", puid.Value)

		tz, err := GetTestGlobalConfig(ta, "TZ")
		require.NoError(t, err)
		assert.Equal(t, "America/New_York", tz.Value)
	})

	t.Run("Update config with invalid JSON returns 400", func(t *testing.T) {
		resp := DoRequest(t, ta.App, PUT("/api/config/global/TZ", "invalid-json").WithContentType("application/json"))

		AssertStatusCode(t, resp, fiber.StatusBadRequest)
	})

	t.Run("Update config with invalid key returns 400", func(t *testing.T) {
		updateData := map[string]string{
			"value": "test",
		}

		resp := DoRequest(t, ta.App, PUT("/api/config/global/INVALID@KEY", updateData))

		AssertStatusCode(t, resp, fiber.StatusBadRequest)
		AssertAPIError(t, resp)
	})
}
