package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service/clients/mocks"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a test database with migrations
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Run migrations
	if err := db.AutoMigrate(&models.Media{}, &models.History{}); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

// setupTestHandler creates a test handler with database
func setupTestHandler(t *testing.T) (*FileActionsHandler, *gorm.DB) {
	db := setupTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	historyRepo := repository.NewHistoryRepository(db)

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	handler := NewFileActionsHandler(
		mediaRepo,
		historyRepo,
		nil, // radarrClient
		nil, // sonarrClient
		nil, // qbitClient
		nil, // jellyfinClient
		logger,
	)

	return handler, db
}

// TestNewFileActionsHandler verifies that the handler can be created successfully
func TestNewFileActionsHandler(t *testing.T) {
	// Setup in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open in-memory database: %v", err)
	}

	// Create repositories
	mediaRepo := repository.NewMediaRepository(db)
	historyRepo := repository.NewHistoryRepository(db)

	// Create logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Create handler (with nil clients for this basic test)
	handler := NewFileActionsHandler(
		mediaRepo,
		historyRepo,
		nil, // radarrClient
		nil, // sonarrClient
		nil, // qbitClient
		nil, // jellyfinClient
		logger,
	)

	if handler == nil {
		t.Fatal("Expected handler to be created, got nil")
	}

	if handler.mediaRepo == nil {
		t.Error("Expected mediaRepo to be set")
	}

	if handler.historyRepo == nil {
		t.Error("Expected historyRepo to be set")
	}

	if handler.logger == nil {
		t.Error("Expected logger to be set")
	}
}

// TestFileActionsHandler_DeleteFile_RequiresConfirmation tests that delete requires confirmation
func TestFileActionsHandler_DeleteFile_RequiresConfirmation(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test without confirmation
	reqBody := DeleteFileRequest{
		Confirm: false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/delete", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "Confirmation required")
}

// TestFileActionsHandler_DeleteFile_Success tests successful file deletion
func TestFileActionsHandler_DeleteFile_Success(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mkv")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: testFile,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test with confirmation
	reqBody := DeleteFileRequest{
		Confirm:            true,
		DeleteFromServices: false,
		DeleteTorrent:      false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/delete", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	assert.Contains(t, response["deleted_from"], "filesystem")

	// Verify file was deleted
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))

	// Verify media was deleted from database
	var count int64
	db.Model(&models.Media{}).Where("id = ?", media.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestFileActionsHandler_DeleteFile_WithServices tests deletion with service integration
func TestFileActionsHandler_DeleteFile_WithServices(t *testing.T) {
	db := setupTestDB(t)
	mediaRepo := repository.NewMediaRepository(db)
	historyRepo := repository.NewHistoryRepository(db)
	logger, _ := zap.NewDevelopment()

	// Create mocks
	mockRadarr := new(mocks.MockMediaClient)
	mockJellyfin := new(mocks.MockStreamingClient)

	// Setup mock expectations
	radarrID := 123
	jellyfinID := "jf-456"

	mockRadarr.On("DeleteItem", mock.Anything, radarrID, true).Return(nil)
	mockJellyfin.On("DeleteItem", mock.Anything, jellyfinID).Return(nil)

	handler := NewFileActionsHandler(
		mediaRepo,
		historyRepo,
		mockRadarr,
		nil, // sonarrClient
		nil, // qbitClient - skip qbit for this test
		mockJellyfin,
		logger,
	)

	// Create temporary test file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.mkv")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Create test media with service IDs (without qBittorrent)
	media := &models.Media{
		Title:      "Test Movie",
		Type:       "movie",
		FilePath:   testFile,
		InRadarr:   true,
		RadarrID:   &radarrID,
		InJellyfin: true,
		JellyfinID: &jellyfinID,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test with confirmation and service deletion
	reqBody := DeleteFileRequest{
		Confirm:            true,
		DeleteFromServices: true,
		DeleteTorrent:      false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/delete", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])
	deletedFrom := response["deleted_from"].([]interface{})
	assert.Contains(t, deletedFrom, "radarr")
	assert.Contains(t, deletedFrom, "jellyfin")
	assert.Contains(t, deletedFrom, "filesystem")

	// Verify mocks were called
	mockRadarr.AssertCalled(t, "DeleteItem", mock.Anything, radarrID, true)
	mockJellyfin.AssertCalled(t, "DeleteItem", mock.Anything, jellyfinID)
}

// TestFileActionsHandler_IgnoreFile_Success tests successful file ignoring
func TestFileActionsHandler_IgnoreFile_Success(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
		Excluded: false,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/ignore", handler.IgnoreFile)

	// Test ignore
	reqBody := IgnoreFileRequest{
		Reason:    "Not interested",
		Permanent: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/ignore", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])

	// Verify media was marked as excluded
	var updatedMedia models.Media
	db.First(&updatedMedia, media.ID)
	assert.Equal(t, true, updatedMedia.Excluded)
}

// TestFileActionsHandler_CleanupHardlink_ValidatesSameInode tests hardlink validation
func TestFileActionsHandler_CleanupHardlink_ValidatesSameInode(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary test files (not hardlinks)
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.mkv")
	file2 := filepath.Join(tmpDir, "file2.mkv")
	err := os.WriteFile(file1, []byte("test content 1"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(file2, []byte("test content 2"), 0644)
	assert.NoError(t, err)

	// Create test media
	media := &models.Media{
		Title:         "Test Movie",
		Type:          "movie",
		FilePath:      file1,
		IsHardlink:    true,
		HardlinkPaths: models.StringSlice{file1, file2},
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test cleanup with non-hardlinked files
	reqBody := CleanupHardlinkRequest{
		KeepPath:   file1,
		RemovePath: file2,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/cleanup-hardlink", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not hardlinks of the same file")
}

// TestFileActionsHandler_CleanupHardlink_Success tests successful hardlink cleanup
func TestFileActionsHandler_CleanupHardlink_Success(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary test file and hardlink
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.mkv")
	file2 := filepath.Join(tmpDir, "file2.mkv")
	err := os.WriteFile(file1, []byte("test content"), 0644)
	assert.NoError(t, err)
	err = os.Link(file1, file2) // Create hardlink
	assert.NoError(t, err)

	// Create test media
	media := &models.Media{
		Title:         "Test Movie",
		Type:          "movie",
		FilePath:      file1,
		IsHardlink:    true,
		HardlinkPaths: models.StringSlice{file1, file2},
		PrimaryPath:   file1,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test cleanup
	reqBody := CleanupHardlinkRequest{
		KeepPath:   file1,
		RemovePath: file2,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/cleanup-hardlink", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, _ = io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, true, response["success"])

	// Verify file2 was deleted
	_, err = os.Stat(file2)
	assert.True(t, os.IsNotExist(err))

	// Verify file1 still exists
	_, err = os.Stat(file1)
	assert.NoError(t, err)
}

// TestFileActionsHandler_BulkAction_PartialFailures tests bulk action with partial failures
func TestFileActionsHandler_BulkAction_PartialFailures(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary test files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.mkv")
	file2 := filepath.Join(tmpDir, "file2.mkv")
	err := os.WriteFile(file1, []byte("test content 1"), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(file2, []byte("test content 2"), 0644)
	assert.NoError(t, err)

	// Create test media records
	media1 := &models.Media{Title: "Movie 1", Type: "movie", FilePath: file1}
	media2 := &models.Media{Title: "Movie 2", Type: "movie", FilePath: file2}
	media3 := &models.Media{Title: "Movie 3", Type: "movie", FilePath: "/nonexistent/file.mkv"}

	db.Create(media1)
	db.Create(media2)
	db.Create(media3)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/bulk-action", handler.BulkAction)

	// Test bulk delete with one failure
	reqBody := BulkActionRequest{
		FileIDs: []uint{media1.ID, media2.ID, media3.ID},
		Action:  "delete",
		Params: map[string]interface{}{
			"delete_from_services": false,
			"delete_torrent":       false,
		},
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/bulk-action", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, _ = io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, true, response["success"])

	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(3), summary["total"])
	assert.Equal(t, float64(2), summary["succeeded"]) // file1 and file2 should succeed
	assert.Equal(t, float64(1), summary["failed"])    // file3 should fail

	// Verify results
	results := response["results"].([]interface{})
	assert.Len(t, results, 3)
}

// TestFileActionsHandler_BulkAction_Ignore tests bulk ignore action
func TestFileActionsHandler_BulkAction_Ignore(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media records
	media1 := &models.Media{Title: "Movie 1", Type: "movie", FilePath: "/tmp/file1.mkv", Excluded: false}
	media2 := &models.Media{Title: "Movie 2", Type: "movie", FilePath: "/tmp/file2.mkv", Excluded: false}

	db.Create(media1)
	db.Create(media2)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/bulk-action", handler.BulkAction)

	// Test bulk ignore
	reqBody := BulkActionRequest{
		FileIDs: []uint{media1.ID, media2.ID},
		Action:  "ignore",
		Params:  map[string]interface{}{},
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/bulk-action", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, true, response["success"])

	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["total"])
	assert.Equal(t, float64(2), summary["succeeded"])
	assert.Equal(t, float64(0), summary["failed"])

	// Verify both media were marked as excluded
	var updatedMedia1, updatedMedia2 models.Media
	db.First(&updatedMedia1, media1.ID)
	db.First(&updatedMedia2, media2.ID)
	assert.Equal(t, true, updatedMedia1.Excluded)
	assert.Equal(t, true, updatedMedia2.Excluded)
}

// TestFileActionsHandler_ImportToRadarr_NotImplemented tests import to radarr
func TestFileActionsHandler_ImportToRadarr_NotImplemented(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-radarr", handler.ImportToRadarr)

	// Test import
	reqBody := ImportToRadarrRequest{
		FilePath:         "/tmp/test.mkv",
		QualityProfileID: 1,
		RootFolderPath:   "/movies",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/import-to-radarr", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 501, resp.StatusCode) // Not Implemented

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not yet implemented")
}

// TestFileActionsHandler_ImportToSonarr_NotImplemented tests import to sonarr
func TestFileActionsHandler_ImportToSonarr_NotImplemented(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Series",
		Type:     "series",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-sonarr", handler.ImportToSonarr)

	// Test import
	reqBody := ImportToSonarrRequest{
		FilePath:         "/tmp/test.mkv",
		QualityProfileID: 1,
		RootFolderPath:   "/series",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/import-to-sonarr", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 501, resp.StatusCode) // Not Implemented

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not yet implemented")
}

// TestFileActionsHandler_DeleteFile_NotFound tests delete with non-existent file
func TestFileActionsHandler_DeleteFile_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test with non-existent ID
	reqBody := DeleteFileRequest{
		Confirm: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/99999/delete", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not found")
}

// TestFileActionsHandler_IgnoreFile_NotFound tests ignore with non-existent file
func TestFileActionsHandler_IgnoreFile_NotFound(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/ignore", handler.IgnoreFile)

	// Test with non-existent ID
	reqBody := IgnoreFileRequest{
		Reason:    "Test",
		Permanent: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/99999/ignore", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 404, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not found")
}

// TestFileActionsHandler_CleanupHardlink_MissingPaths tests cleanup with missing paths
func TestFileActionsHandler_CleanupHardlink_MissingPaths(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:      "Test Movie",
		Type:       "movie",
		FilePath:   "/tmp/file1.mkv",
		IsHardlink: true,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test with missing paths
	reqBody := CleanupHardlinkRequest{
		KeepPath:   "",
		RemovePath: "",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/cleanup-hardlink", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "required")
}

// TestFileActionsHandler_CleanupHardlink_KeepPathNotExist tests cleanup when keep path doesn't exist
func TestFileActionsHandler_CleanupHardlink_KeepPathNotExist(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary test file
	tmpDir := t.TempDir()
	file2 := filepath.Join(tmpDir, "file2.mkv")
	err := os.WriteFile(file2, []byte("test content"), 0644)
	assert.NoError(t, err)

	// Create test media
	media := &models.Media{
		Title:         "Test Movie",
		Type:          "movie",
		FilePath:      "/nonexistent/file1.mkv",
		IsHardlink:    true,
		HardlinkPaths: models.StringSlice{"/nonexistent/file1.mkv", file2},
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test with non-existent keep path
	reqBody := CleanupHardlinkRequest{
		KeepPath:   "/nonexistent/file1.mkv",
		RemovePath: file2,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/cleanup-hardlink", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not exist")
}

// TestFileActionsHandler_BulkAction_InvalidAction tests bulk action with invalid action
func TestFileActionsHandler_BulkAction_InvalidAction(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/bulk-action", handler.BulkAction)

	// Test with invalid action
	reqBody := BulkActionRequest{
		FileIDs: []uint{1, 2, 3},
		Action:  "invalid_action",
		Params:  map[string]interface{}{},
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/bulk-action", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, err = io.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &response)
	assert.NoError(t, err)

	assert.Equal(t, false, response["success"])
	assert.Contains(t, response["error"], "not allowed")
}

// TestFileActionsHandler_DeleteFile_HardlinkPath tests deleting a hardlink
func TestFileActionsHandler_DeleteFile_HardlinkPath(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create temporary hardlink files
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "file1.mkv")
	file2 := filepath.Join(tmpDir, "file2.mkv")
	err := os.WriteFile(file1, []byte("test content"), 0644)
	assert.NoError(t, err)
	err = os.Link(file1, file2)
	assert.NoError(t, err)

	// Create test media as hardlink
	media := &models.Media{
		Title:         "Test Movie",
		Type:          "movie",
		FilePath:      file2,
		IsHardlink:    true,
		HardlinkPaths: models.StringSlice{file1, file2},
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test delete hardlink
	reqBody := DeleteFileRequest{
		Confirm:            true,
		DeleteFromServices: false,
		DeleteTorrent:      false,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/delete", media.ID), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Parse response
	var response map[string]interface{}
	bodyBytes, _ = io.ReadAll(resp.Body)
	json.Unmarshal(bodyBytes, &response)

	assert.Equal(t, true, response["success"])

	// Verify hardlink was deleted but original remains
	_, err = os.Stat(file2)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(file1)
	assert.NoError(t, err) // Original should still exist
}

// TestFileActionsHandler_DeleteFile_InvalidID tests delete with invalid ID
func TestFileActionsHandler_DeleteFile_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test with invalid ID
	reqBody := DeleteFileRequest{
		Confirm: true,
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/invalid/delete", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_IgnoreFile_InvalidID tests ignore with invalid ID
func TestFileActionsHandler_IgnoreFile_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/ignore", handler.IgnoreFile)

	// Test with invalid ID
	reqBody := IgnoreFileRequest{
		Reason: "Test",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/invalid/ignore", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_CleanupHardlink_InvalidID tests cleanup with invalid ID
func TestFileActionsHandler_CleanupHardlink_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test with invalid ID
	reqBody := CleanupHardlinkRequest{
		KeepPath:   "/tmp/file1.mkv",
		RemovePath: "/tmp/file2.mkv",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/invalid/cleanup-hardlink", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_ImportToRadarr_InvalidID tests import with invalid ID
func TestFileActionsHandler_ImportToRadarr_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-radarr", handler.ImportToRadarr)

	// Test with invalid ID
	reqBody := ImportToRadarrRequest{
		FilePath: "/tmp/test.mkv",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/invalid/import-to-radarr", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_ImportToSonarr_InvalidID tests import with invalid ID
func TestFileActionsHandler_ImportToSonarr_InvalidID(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-sonarr", handler.ImportToSonarr)

	// Test with invalid ID
	reqBody := ImportToSonarrRequest{
		FilePath: "/tmp/test.mkv",
	}
	bodyBytes, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/files/invalid/import-to-sonarr", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_IgnoreFile_InvalidBody tests ignore with invalid JSON body
func TestFileActionsHandler_IgnoreFile_InvalidBody(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/ignore", handler.IgnoreFile)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/ignore", media.ID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_DeleteFile_InvalidBody tests delete with invalid JSON body
func TestFileActionsHandler_DeleteFile_InvalidBody(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/delete", handler.DeleteFile)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/delete", media.ID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_CleanupHardlink_InvalidBody tests cleanup with invalid JSON body
func TestFileActionsHandler_CleanupHardlink_InvalidBody(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:      "Test Movie",
		Type:       "movie",
		FilePath:   "/tmp/test.mkv",
		IsHardlink: true,
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/cleanup-hardlink", handler.CleanupHardlink)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/cleanup-hardlink", media.ID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_ImportToRadarr_InvalidBody tests import with invalid JSON body
func TestFileActionsHandler_ImportToRadarr_InvalidBody(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Movie",
		Type:     "movie",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-radarr", handler.ImportToRadarr)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/import-to-radarr", media.ID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_ImportToSonarr_InvalidBody tests import with invalid JSON body
func TestFileActionsHandler_ImportToSonarr_InvalidBody(t *testing.T) {
	handler, db := setupTestHandler(t)

	// Create test media
	media := &models.Media{
		Title:    "Test Series",
		Type:     "series",
		FilePath: "/tmp/test.mkv",
	}
	db.Create(media)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/:id/import-to-sonarr", handler.ImportToSonarr)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", fmt.Sprintf("/api/files/%d/import-to-sonarr", media.ID), bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestFileActionsHandler_BulkAction_InvalidBody tests bulk action with invalid JSON body
func TestFileActionsHandler_BulkAction_InvalidBody(t *testing.T) {
	handler, _ := setupTestHandler(t)

	// Setup Fiber app
	app := fiber.New()
	app.Post("/api/files/bulk-action", handler.BulkAction)

	// Test with invalid JSON
	req := httptest.NewRequest("POST", "/api/files/bulk-action", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)
}

// TestGetBoolParam tests the helper function
func TestGetBoolParam(t *testing.T) {
	tests := []struct {
		name         string
		params       map[string]interface{}
		key          string
		defaultValue bool
		expected     bool
	}{
		{
			name:         "returns true when param is true",
			params:       map[string]interface{}{"test": true},
			key:          "test",
			defaultValue: false,
			expected:     true,
		},
		{
			name:         "returns false when param is false",
			params:       map[string]interface{}{"test": false},
			key:          "test",
			defaultValue: true,
			expected:     false,
		},
		{
			name:         "returns default when key missing",
			params:       map[string]interface{}{},
			key:          "test",
			defaultValue: true,
			expected:     true,
		},
		{
			name:         "returns default when value is not bool",
			params:       map[string]interface{}{"test": "not a bool"},
			key:          "test",
			defaultValue: true,
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBoolParam(tt.params, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getBoolParam() = %v, want %v", result, tt.expected)
			}
		})
	}
}
