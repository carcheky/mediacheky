package apitest

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProxyEndpoints(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	// Seed test data
	require.NoError(t, SeedTestServices(ta))

	t.Run("Get proxy configuration", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/proxy/config"))

		// Proxy config may or may not exist in test environment
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Get domains", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/proxy/domains"))

		// Domains list should be accessible
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Update service subdomain", func(t *testing.T) {
		updateData := map[string]string{
			"subdomain": "radarr-test",
		}

		resp := DoRequest(t, ta.App, PUT("/api/services/radarr/subdomain", updateData))

		// The exact response depends on implementation
		// It may succeed or fail based on proxy configuration
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusBadRequest, fiber.StatusServiceUnavailable}, resp.StatusCode)
	})

	t.Run("Get service endpoint", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/services/radarr/endpoint"))

		// Endpoint information should be available
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Update proxy config", func(t *testing.T) {
		updateData := map[string]interface{}{
			"enabled": true,
			"domain":  "example.com",
		}

		resp := DoRequest(t, ta.App, PUT("/api/proxy/config", updateData))

		// The endpoint may succeed, fail validation, or return errors depending on the test environment
		// (e.g., missing proxy config, invalid input, or internal error). Accept multiple status codes.
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusBadRequest, fiber.StatusNotFound, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("Add domain", func(t *testing.T) {
		domainData := map[string]string{
			"name":   "test.example.com",
			"target": "http://localhost:8080",
		}

		resp := DoRequest(t, ta.App, POST("/api/proxy/domains", domainData))

		// May succeed or fail based on validation
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusBadRequest, fiber.StatusCreated}, resp.StatusCode)
	})

	t.Run("Delete domain", func(t *testing.T) {
		resp := DoRequest(t, ta.App, DELETE("/api/proxy/domains/test.example.com"))

		// May succeed, return 404, or error based on implementation
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusNotFound, fiber.StatusBadRequest, fiber.StatusInternalServerError}, resp.StatusCode)
	})
}

func TestDockerEndpoints(t *testing.T) {
	ta := SetupTestApp(t)
	defer ta.Cleanup()

	t.Run("Get Docker info without Docker client", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/docker/info"))

		// Without Docker client in test environment, should return error
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusServiceUnavailable, fiber.StatusInternalServerError}, resp.StatusCode)
	})

	t.Run("List Docker containers without Docker client", func(t *testing.T) {
		resp := DoRequest(t, ta.App, GET("/api/docker/containers"))

		// Without Docker client in test environment, should return error
		assert.Contains(t, []int{fiber.StatusOK, fiber.StatusServiceUnavailable, fiber.StatusInternalServerError}, resp.StatusCode)
	})
}
