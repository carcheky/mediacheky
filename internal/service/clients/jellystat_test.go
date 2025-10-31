package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewJellystatClient(t *testing.T) {
	logger := zap.NewNop()
	config := ClientConfig{
		BaseURL: "http://localhost:3000",
		APIKey:  "test-api-key",
		Timeout: 30 * time.Second,
	}

	client := NewJellystatClient(config, logger)

	assert.NotNil(t, client)
	assert.Equal(t, config.BaseURL, client.baseURL)
	assert.Equal(t, config.APIKey, client.apiKey)
}

func TestJellystatClient_TestConnection(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/getconfig", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("x-api-token"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"JF_HOST":"http://jellyfin:8096","APP_USER":"admin","REQUIRE_LOGIN":false,"IS_JELLYFIN":true}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	err := client.TestConnection(context.Background())
	require.NoError(t, err)
}

func TestJellystatClient_GetSystemInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/getconfig", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"JF_HOST":"http://jellyfin:8096","APP_USER":"admin","REQUIRE_LOGIN":false,"IS_JELLYFIN":true}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	info, err := client.GetSystemInfo(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "connected", info.Status)
}

func TestJellystatClient_GetStatistics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/statistics", r.URL.Path)
		assert.Equal(t, "30", r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Match the exact struct field names in the client code
		response := `{"days":30,"movies":150,"episodes":500,"songs":1000}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	stats, err := client.GetStatistics(context.Background(), 30)
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 30, stats.Days)
	assert.Equal(t, 150, stats.Movies)
	assert.Equal(t, 500, stats.Episodes)
	assert.Equal(t, 1000, stats.Songs)
	assert.Equal(t, 1650, stats.Total)
}

func TestJellystatClient_GetViewsByLibraryType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/stats/getViewsByLibraryType", r.URL.Path)
		assert.Equal(t, "7", r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `{"music":10,"movie":50,"episode":100,"book":5}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	views, err := client.GetViewsByLibraryType(context.Background(), 7)
	require.NoError(t, err)
	assert.NotNil(t, views)
	assert.Equal(t, 10, views.Music)
	assert.Equal(t, 50, views.Movie)
	assert.Equal(t, 100, views.Episode)
	assert.Equal(t, 5, views.Book)
}

func TestJellystatClient_GetUserActivity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/stats/getUserActivity", r.URL.Path)
		assert.Equal(t, "30", r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `[{"user_id":"user1","user_name":"John Doe","total_plays":50,"total_minutes":3000},{"user_id":"user2","user_name":"Jane Smith","total_plays":30,"total_minutes":1500}]`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	activities, err := client.GetUserActivity(context.Background(), 30)
	require.NoError(t, err)
	assert.NotNil(t, activities)
	assert.Len(t, activities, 2)
	if len(activities) >= 2 {
		assert.Equal(t, "user1", activities[0].UserID)
		assert.Equal(t, "John Doe", activities[0].UserName)
		assert.Equal(t, 50, activities[0].TotalPlays)
		assert.Equal(t, 3000, activities[0].TotalMinutes)
	}
}

func TestJellystatClient_GetLibraryStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/stats/getLibraryStats", r.URL.Path)
		assert.Equal(t, "30", r.URL.Query().Get("days"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := `[{"library_id":"lib1","library_name":"Movies","total_items":100,"total_plays":500,"total_minutes":30000},{"library_id":"lib2","library_name":"TV Shows","total_items":50,"total_plays":800,"total_minutes":45000}]`
		w.Write([]byte(response))
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	libraries, err := client.GetLibraryStats(context.Background(), 30)
	require.NoError(t, err)
	assert.NotNil(t, libraries)
	assert.Len(t, libraries, 2)
	if len(libraries) >= 2 {
		assert.Equal(t, "lib1", libraries[0].LibraryID)
		assert.Equal(t, "Movies", libraries[0].LibraryName)
		assert.Equal(t, 100, libraries[0].TotalItems)
		assert.Equal(t, 500, libraries[0].TotalPlays)
		assert.Equal(t, 30000, libraries[0].TotalMinutes)
	}
}

func TestJellystatClient_GetUserActivity_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	// Should return nil error when endpoint doesn't exist (404)
	activities, err := client.GetUserActivity(context.Background(), 30)
	require.NoError(t, err)
	assert.Nil(t, activities)
}

func TestJellystatClient_GetLibraryStats_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewJellystatClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	// Should return nil error when endpoint doesn't exist (404)
	libraries, err := client.GetLibraryStats(context.Background(), 30)
	require.NoError(t, err)
	assert.Nil(t, libraries)
}
