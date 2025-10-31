package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// setupJellyfinTestServer creates a test HTTP server for Jellyfin API mocking.
func setupJellyfinTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func TestJellyfinClient_GetSystemInfo(t *testing.T) {
	expectedInfo := jellyfinSystemInfo{
		Version:                    "10.8.13",
		ID:                         "test-server-id",
		OperatingSystem:            "Linux",
		OperatingSystemDisplayName: "Ubuntu 22.04",
		ServerName:                 "TestJellyfin",
		LocalAddress:               "http://localhost:8096",
		ProductName:                "Jellyfin Server",
	}

	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/System/Info/Public", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedInfo)
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	info, err := client.GetSystemInfo(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, expectedInfo.Version, info.Version)
	assert.Equal(t, expectedInfo.ServerName, info.ServerName)
	assert.Equal(t, expectedInfo.OperatingSystem, info.OS)
}

func TestJellyfinClient_GetActiveSessions(t *testing.T) {
	expectedSessions := []SessionInfo{
		{
			ID:               "session1",
			UserID:           "user1",
			UserName:         "TestUser",
			Client:           "Web",
			DeviceName:       "Chrome",
			LastActivityDate: time.Now(),
			NowPlayingItem: &struct {
				ID        string `json:"Id"`
				Name      string `json:"Name"`
				Type      string `json:"Type"`
				MediaType string `json:"MediaType"`
			}{
				ID:        "item1",
				Name:      "Test Movie",
				Type:      "Movie",
				MediaType: "Video",
			},
		},
	}

	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/Sessions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(expectedSessions)
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	sessions, err := client.GetActiveSessions(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 1)
	assert.Equal(t, "session1", sessions[0].ID)
	assert.Equal(t, "TestUser", sessions[0].UserName)
	assert.NotNil(t, sessions[0].NowPlayingItem)
	assert.Equal(t, "Test Movie", sessions[0].NowPlayingItem.Name)
}

func TestJellyfinClient_GetLibraryStats(t *testing.T) {
	mockItems := jellyfinItemsResponse{
		Items: []jellyfinItem{
			{
				ID:          "movie1",
				Name:        "Test Movie",
				Type:        "Movie",
				DateCreated: time.Now(),
				MediaSources: []struct {
					Size int64 `json:"Size"`
				}{
					{Size: 1073741824}, // 1 GB
				},
			},
			{
				ID:          "series1",
				Name:        "Test Series",
				Type:        "Series",
				DateCreated: time.Now(),
				MediaSources: []struct {
					Size int64 `json:"Size"`
				}{
					{Size: 536870912}, // 512 MB
				},
			},
		},
		TotalCount: 2,
	}

	mockFolders := []VirtualFolder{
		{
			Name:      "Movies",
			Locations: []string{"/media/movies"},
			ItemID:    "folder1",
		},
	}

	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/Items" {
			json.NewEncoder(w).Encode(mockItems)
		} else if r.URL.Path == "/Library/VirtualFolders" {
			json.NewEncoder(w).Encode(mockFolders)
		}
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	stats, err := client.GetLibraryStats(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 2, stats.TotalItems)
	assert.Equal(t, 1, stats.MovieCount)
	assert.Equal(t, 1, stats.SeriesCount)
	assert.Equal(t, int64(1073741824+536870912), stats.TotalSize)
	assert.Len(t, stats.LibraryFolders, 1)
}

func TestJellyfinClient_GetRecentlyAdded(t *testing.T) {
	now := time.Now()
	mockResponse := jellyfinItemsResponse{
		Items: []jellyfinItem{
			{
				ID:          "recent1",
				Name:        "New Movie",
				Type:        "Movie",
				DateCreated: now,
				ImageTags: map[string]string{
					"Primary": "tag123",
				},
			},
			{
				ID:          "recent2",
				Name:        "New Series",
				Type:        "Series",
				DateCreated: now.Add(-24 * time.Hour),
			},
		},
		TotalCount: 2,
	}

	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/Items", r.URL.Path)
		assert.Equal(t, "DateCreated", r.URL.Query().Get("SortBy"))
		assert.Equal(t, "Descending", r.URL.Query().Get("SortOrder"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	items, err := client.GetRecentlyAdded(ctx, 20)

	assert.NoError(t, err)
	assert.NotNil(t, items)
	assert.Len(t, items, 2)
	assert.Equal(t, "New Movie", items[0].Name)
	assert.Contains(t, items[0].PosterURL, "recent1")
}

func TestJellyfinClient_GetActivityLog(t *testing.T) {
	now := time.Now()
	mockResponse := jellyfinActivityResponse{
		Items: []struct {
			ID            int64     `json:"Id"`
			Name          string    `json:"Name"`
			Type          string    `json:"Type"`
			UserID        string    `json:"UserId"`
			Date          time.Time `json:"Date"`
			Severity      string    `json:"Severity"`
			ShortOverview string    `json:"ShortOverview"`
		}{
			{
				ID:            1,
				Name:          "AuthenticationSucceeded",
				Type:          "AuthenticationSuccess",
				UserID:        "user1",
				Date:          now,
				Severity:      "Info",
				ShortOverview: "User logged in",
			},
		},
		TotalRecordCount: 1,
	}

	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/System/ActivityLog/Entries", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResponse)
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	entries, err := client.GetActivityLog(ctx, 50)

	assert.NoError(t, err)
	assert.NotNil(t, entries)
	assert.Len(t, entries, 1)
	assert.Equal(t, "AuthenticationSucceeded", entries[0].Name)
	assert.Equal(t, "Info", entries[0].Severity)
}

func TestJellyfinClient_ErrorHandling(t *testing.T) {
	// Test server returning errors
	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()

	// Test GetActiveSessions with error
	_, err := client.GetActiveSessions(ctx)
	assert.Error(t, err)

	// Test GetLibraryStats with error
	_, err = client.GetLibraryStats(ctx)
	assert.Error(t, err)

	// Test GetRecentlyAdded with error
	_, err = client.GetRecentlyAdded(ctx, 20)
	assert.Error(t, err)

	// Test GetActivityLog with error
	_, err = client.GetActivityLog(ctx, 50)
	assert.Error(t, err)
}

func TestJellyfinClient_GetActiveSessions_NoActiveSessions(t *testing.T) {
	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/Sessions", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]SessionInfo{})
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	sessions, err := client.GetActiveSessions(ctx)

	assert.NoError(t, err)
	assert.NotNil(t, sessions)
	assert.Len(t, sessions, 0)
}

func TestJellyfinClient_GetRecentlyAdded_DefaultLimit(t *testing.T) {
	server := setupJellyfinTestServer(func(w http.ResponseWriter, r *http.Request) {
		// Should use default limit of 20
		assert.Equal(t, "20", r.URL.Query().Get("Limit"))
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jellyfinItemsResponse{Items: []jellyfinItem{}, TotalCount: 0})
	})
	defer server.Close()

	client := &JellyfinClient{
		client:  resty.New().SetBaseURL(server.URL),
		baseURL: server.URL,
		apiKey:  "test-api-key",
		logger:  zap.NewNop(),
	}

	ctx := context.Background()
	_, err := client.GetRecentlyAdded(ctx, 0) // Pass 0 to trigger default

	assert.NoError(t, err)
}
