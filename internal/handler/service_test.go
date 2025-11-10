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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.Service{}, &models.Template{}, &models.GlobalConfig{}); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func setupTestApp(db *gorm.DB) (*fiber.App, *ServiceHandler) {
	app := fiber.New()
	repos := repository.NewRepositories(db)
	log := logger.New("debug")
	handler := NewServiceHandler(repos, log, nil, nil) // nil Docker client and service manager for tests
	return app, handler
}

func TestListServices(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test services
	testServices := []models.Service{
		{Name: "radarr", DisplayName: "Radarr", Enabled: true, Status: "running"},
		{Name: "sonarr", DisplayName: "Sonarr", Enabled: false, Status: "stopped"},
	}
	for _, svc := range testServices {
		db.Create(&svc)
	}

	app.Get("/api/services", handler.ListServices)

	req := httptest.NewRequest("GET", "/api/services", nil)
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

	services, ok := response.Data.([]interface{})
	if !ok || len(services) != 2 {
		t.Errorf("Expected 2 services, got %d", len(services))
	}
}

func TestGetService(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test service
	testService := models.Service{Name: "radarr", DisplayName: "Radarr", Enabled: true, Status: "running"}
	db.Create(&testService)

	app.Get("/api/services/:name", handler.GetService)

	tests := []struct {
		name           string
		serviceName    string
		expectedStatus int
		expectedError  bool
	}{
		{"Existing service", "radarr", fiber.StatusOK, false},
		{"Non-existent service", "nonexistent", fiber.StatusNotFound, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/services/"+tt.serviceName, nil)
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

func TestEnableService(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test service (disabled)
	testService := models.Service{Name: "radarr", DisplayName: "Radarr", Enabled: false, Status: "stopped"}
	db.Create(&testService)

	app.Post("/api/services/:name/enable", handler.EnableService)

	req := httptest.NewRequest("POST", "/api/services/radarr/enable", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Since we don't have a service manager in tests, expect 503
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success=false when service manager unavailable")
	}
}

func TestDisableService(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test service (enabled)
	testService := models.Service{Name: "radarr", DisplayName: "Radarr", Enabled: true, Status: "running"}
	db.Create(&testService)

	app.Post("/api/services/:name/disable", handler.DisableService)

	req := httptest.NewRequest("POST", "/api/services/radarr/disable", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Since we don't have a service manager in tests, expect 503
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success=false when service manager unavailable")
	}
}

func TestUpdateServiceConfig(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test service
	testService := models.Service{
		Name:        "radarr",
		DisplayName: "Radarr",
		Enabled:     true,
		Config:      models.ServiceConfig{"port": "7878"},
	}
	db.Create(&testService)

	app.Put("/api/services/:name/config", handler.UpdateServiceConfig)

	newConfig := models.ServiceConfig{
		"port":    "8080",
		"api_key": "test123",
	}
	configJSON, _ := json.Marshal(newConfig)

	req := httptest.NewRequest("PUT", "/api/services/radarr/config", bytes.NewReader(configJSON))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Since we don't have a service manager in tests, expect 503
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected success=false when service manager unavailable")
	}
}

func TestStartContainer_NoServiceManager(t *testing.T) {
	db := setupTestDB(t)
	app, handler := setupTestApp(db)

	// Create test service without container ID
	testService := models.Service{Name: "radarr", DisplayName: "Radarr", Enabled: true, ContainerID: ""}
	db.Create(&testService)

	app.Post("/api/services/:name/start", handler.StartContainer)

	req := httptest.NewRequest("POST", "/api/services/radarr/start", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Since we don't have a service manager in tests, expect 503
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", resp.StatusCode)
	}

	var response APIResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Success {
		t.Errorf("Expected error response, got success")
	}
}
