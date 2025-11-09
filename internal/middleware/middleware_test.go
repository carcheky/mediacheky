package middleware

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestValidateServiceName(t *testing.T) {
	app := fiber.New()
	
	app.Get("/services/:name", ValidateServiceName(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	tests := []struct {
		name           string
		serviceName    string
		expectedStatus int
		shouldPass     bool
	}{
		{"Valid lowercase", "radarr", fiber.StatusOK, true},
		{"Valid with hyphen", "my-service", fiber.StatusOK, true},
		{"Valid with underscore", "my_service", fiber.StatusOK, true},
		{"Valid mixed case", "MyService", fiber.StatusOK, true},
		{"Invalid special chars", "service@test", fiber.StatusBadRequest, false},
		{"Invalid dots", "service.test", fiber.StatusBadRequest, false},
		{"Too long", "averylongservicenamethatexceedsthemaximumlengthof50characters", fiber.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/services/"+tt.serviceName, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			success, ok := response["success"].(bool)
			if !ok {
				t.Fatalf("Response missing 'success' field")
			}

			if tt.shouldPass && !success {
				t.Errorf("Expected validation to pass, but got error: %v", response["error"])
			}
			if !tt.shouldPass && success {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestValidateConfigKey(t *testing.T) {
	app := fiber.New()
	
	app.Get("/config/:key", ValidateConfigKey(), func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	tests := []struct {
		name           string
		key            string
		expectedStatus int
		shouldPass     bool
	}{
		{"Valid uppercase", "PUID", fiber.StatusOK, true},
		{"Valid with underscore", "MY_CONFIG", fiber.StatusOK, true},
		{"Valid with numbers", "CONFIG123", fiber.StatusOK, true},
		{"Invalid lowercase", "puid", fiber.StatusBadRequest, false},
		{"Invalid mixed case", "MyConfig", fiber.StatusBadRequest, false},
		{"Invalid special chars", "CONFIG-KEY", fiber.StatusBadRequest, false},
		{"Invalid dots", "CONFIG.KEY", fiber.StatusBadRequest, false},
		{"Too long", "AVERYLONGCONFIGKEYTHATEXCEEDSTHEMAXIMUMLENGTHOF50CHARS", fiber.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/config/"+tt.key, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var response map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			success, ok := response["success"].(bool)
			if !ok {
				t.Fatalf("Response missing 'success' field")
			}

			if tt.shouldPass && !success {
				t.Errorf("Expected validation to pass, but got error: %v", response["error"])
			}
			if !tt.shouldPass && success {
				t.Errorf("Expected validation to fail, but it passed")
			}
		})
	}
}

func TestCORS(t *testing.T) {
	app := fiber.New()
	
	app.Use(CORS())
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"success": true})
	})

	// Test regular GET request
	t.Run("GET request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Check CORS headers
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin header to be '*'")
		}
		if resp.Header.Get("Access-Control-Allow-Methods") == "" {
			t.Errorf("Expected Access-Control-Allow-Methods header to be set")
		}
	})

	// Test OPTIONS preflight request
	t.Run("OPTIONS preflight", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		if resp.StatusCode != fiber.StatusNoContent {
			t.Errorf("Expected status 204, got %d", resp.StatusCode)
		}

		// Check CORS headers
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin header to be '*'")
		}
	})
}

func TestRequestID(t *testing.T) {
	app := fiber.New()
	
	app.Use(RequestID())
	app.Get("/test", func(c *fiber.Ctx) error {
		requestID := c.Locals("requestid")
		return c.JSON(fiber.Map{"request_id": requestID})
	})

	t.Run("Generates request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		// Check header is set
		requestID := resp.Header.Get("X-Request-ID")
		if requestID == "" {
			t.Errorf("Expected X-Request-ID header to be set")
		}

		var response map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		if response["request_id"] == nil || response["request_id"] == "" {
			t.Errorf("Expected request_id in response")
		}
	})

	t.Run("Uses existing request ID", func(t *testing.T) {
		existingID := "test-request-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("Failed to execute request: %v", err)
		}

		requestID := resp.Header.Get("X-Request-ID")
		if requestID != existingID {
			t.Errorf("Expected request ID to be '%s', got '%s'", existingID, requestID)
		}
	})
}
