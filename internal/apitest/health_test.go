package apitest

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestHealthEndpoint(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	t.Run("Health check returns healthy status", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/health"))

		AssertStatusCode(t, resp, fiber.StatusOK)

		var result map[string]interface{}
		AssertJSONResponse(t, resp, &result)

		assert.Equal(t, "healthy", result["status"])
		assert.NotNil(t, result["timestamp"])
	})

	t.Run("Health check includes timestamp", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/health"))

		AssertStatusCode(t, resp, fiber.StatusOK)

		var result map[string]interface{}
		AssertJSONResponse(t, resp, &result)

		timestamp, ok := result["timestamp"].(float64)
		assert.True(t, ok, "timestamp should be a number")
		assert.Greater(t, timestamp, float64(0), "timestamp should be greater than 0")
	})
}
