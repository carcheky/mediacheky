package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/carcheky/mediacheky/internal/models"
	"github.com/carcheky/mediacheky/internal/repository"
	"github.com/carcheky/mediacheky/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func setupConfigTestApp(db *gorm.DB) (*fiber.App, *ConfigHandler) {
	app := fiber.New()
	repos := repository.NewRepositories(db)
	log := logger.New("debug")
	handler := NewConfigHandler(repos, log)
	return app, handler
}

func TestGetGlobalConfig(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupConfigTestApp(db)

	// Create test config entries
	testConfigs := []models.GlobalConfig{
		{Key: "PUID", Value: "1000", Category: "system"},
		{Key: "PGID", Value: "1000", Category: "system"},
		{Key: "TZ", Value: "Europe/Madrid", Category: "system"},
	}
	for _, cfg := range testConfigs {
		db.Create(&cfg)
	}

	app.Get("/api/config/global", handler.GetGlobalConfig)

	req := httptest.NewRequest("GET", "/api/config/global", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v", response.Success)
	}

	configMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data to be a map")
	}

	// Check if system category exists
	systemConfig, ok := configMap["system"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected system category in config")
	}

	// Verify values
	if systemConfig["PUID"] != "1000" {
		t.Errorf("Expected PUID='1000', got '%v'", systemConfig["PUID"])
	}
}

func TestUpdateGlobalConfig(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupConfigTestApp(db)

	app.Put("/api/config/global", handler.UpdateGlobalConfig)

	type ConfigEntry struct {
		Key      string `json:"key"`
		Value    string `json:"value"`
		Category string `json:"category,omitempty"`
	}

	configUpdates := []ConfigEntry{
		{Key: "PUID", Value: "1001", Category: "system"},
		{Key: "PGID", Value: "1001", Category: "system"},
	}
	configJSON, _ := json.Marshal(configUpdates)

	req := httptest.NewRequest("PUT", "/api/config/global", bytes.NewReader(configJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v. Error: %s", response.Success, response.Error)
	}

	// Verify config was updated in database
	var config models.GlobalConfig
	db.Where("key = ?", "PUID").First(&config)
	if config.Value != "1001" {
		t.Errorf("Expected PUID='1001', got '%s'", config.Value)
	}
}

func TestGetConfigValue(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupConfigTestApp(db)

	// Create test config
	testConfig := models.GlobalConfig{Key: "TZ", Value: "Europe/Madrid", Category: "system"}
	db.Create(&testConfig)

	app.Get("/api/config/global/:key", handler.GetConfigValue)

	tests := []struct {
		name           string
		key            string
		expectedStatus int
		expectedError  bool
	}{
		{"Existing key", "TZ", fiber.StatusOK, false},
		{"Non-existent key", "NONEXISTENT", fiber.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/config/global/"+tt.key, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			var response APIResponse
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}

			if tt.expectedError && response.Success {
				t.Errorf("Expected error response, got success")
			}
			if !tt.expectedError && !response.Success {
				t.Errorf("Expected success response, got error: %s", response.Error)
			}
		})
	}
}

func TestUpdateConfigValue(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupConfigTestApp(db)

	// Create test config
	testConfig := models.GlobalConfig{Key: "TZ", Value: "UTC", Category: "system"}
	db.Create(&testConfig)

	app.Put("/api/config/global/:key", handler.UpdateConfigValue)

	type ValueUpdate struct {
		Value    string `json:"value"`
		Category string `json:"category,omitempty"`
	}

	update := ValueUpdate{Value: "Europe/London", Category: "system"}
	updateJSON, _ := json.Marshal(update)

	req := httptest.NewRequest("PUT", "/api/config/global/TZ", bytes.NewReader(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected success=true, got %v. Error: %s", response.Success, response.Error)
	}

	// Verify config was updated in database
	var updatedConfig models.GlobalConfig
	db.Where("key = ?", "TZ").First(&updatedConfig)
	if updatedConfig.Value != "Europe/London" {
		t.Errorf("Expected TZ='Europe/London', got '%s'", updatedConfig.Value)
	}
}

func TestUpdateGlobalConfig_EmptyList(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupConfigTestApp(db)

	app.Put("/api/config/global", handler.UpdateGlobalConfig)

	emptyList := []interface{}{}
	configJSON, _ := json.Marshal(emptyList)

	req := httptest.NewRequest("PUT", "/api/config/global", bytes.NewReader(configJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected error response, got success")
	}
}
