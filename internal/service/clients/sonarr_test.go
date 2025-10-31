package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestSonarrClient_GetSystemInfo(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/system/status", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-Api-Key"))

		response := sonarrSystemStatus{
			Version:        "4.0.0.0",
			BuildTime:      "2024-01-01",
			Branch:         "master",
			OsName:         "linux",
			OsVersion:      "5.15.0",
			RuntimeName:    ".NET Core",
			RuntimeVersion: "6.0.0",
			IsDebug:        false,
			IsProduction:   true,
			IsLinux:        true,
			Authentication: "forms",
			UrlBase:        "",
			StartupPath:    "/app",
			AppData:        "/config",
			SqliteVersion:  "3.36.0",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	info, err := client.GetSystemInfo(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "4.0.0.0", info.Version)
	assert.Equal(t, "Linux", info.OS)
	assert.Equal(t, ".NET Core", info.Runtime)
	assert.True(t, info.IsProduction)
}

func TestSonarrClient_GetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/queue", r.URL.Path)
		assert.Equal(t, "test-api-key", r.Header.Get("X-Api-Key"))

		response := sonarrQueueResponse{
			Page:         1,
			PageSize:     100,
			TotalRecords: 2,
			Records: []sonarrQueueItem{
				{
					ID:                      1,
					SeriesID:                123,
					EpisodeID:               456,
					Title:                   "Test Series S01E01",
					Size:                    1000000000,
					Sizeleft:                500000000,
					Status:                  "downloading",
					TrackedDownloadStatus:   "ok",
					TrackedDownloadState:    "downloading",
					Protocol:                "torrent",
					DownloadClient:          "qBittorrent",
					Indexer:                 "Test Indexer",
					TimedOut:                false,
					EstimatedCompletionTime: time.Now().Add(1 * time.Hour),
				},
				{
					ID:                      2,
					SeriesID:                123,
					EpisodeID:               457,
					Title:                   "Test Series S01E02",
					Size:                    2000000000,
					Sizeleft:                0,
					Status:                  "completed",
					TrackedDownloadStatus:   "ok",
					TrackedDownloadState:    "importPending",
					Protocol:                "torrent",
					DownloadClient:          "qBittorrent",
					Indexer:                 "Test Indexer",
					TimedOut:                false,
					EstimatedCompletionTime: time.Now(),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	queue, err := client.GetQueue(context.Background())
	require.NoError(t, err)
	assert.Len(t, queue, 2)

	// Check first item
	assert.Equal(t, 1, queue[0].ID)
	assert.Equal(t, 123, queue[0].SeriesID)
	assert.Equal(t, 456, queue[0].EpisodeID)
	assert.Equal(t, "Test Series S01E01", queue[0].Title)
	assert.Equal(t, int64(1000000000), queue[0].Size)
	assert.Equal(t, int64(500000000), queue[0].SizeLeft)
	assert.Equal(t, 50.0, queue[0].Progress)
	assert.Equal(t, "downloading", queue[0].Status)

	// Check second item (completed)
	assert.Equal(t, 2, queue[1].ID)
	assert.Equal(t, int64(0), queue[1].SizeLeft)
	assert.Equal(t, 100.0, queue[1].Progress)
}

func TestSonarrClient_GetHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/history", r.URL.Path)
		assert.Equal(t, "50", r.URL.Query().Get("pageSize"))

		response := sonarrHistoryResponse{
			Page:         1,
			PageSize:     50,
			TotalRecords: 2,
			Records: []sonarrHistoryItem{
				{
					ID:          1,
					EpisodeID:   456,
					SeriesID:    123,
					SourceTitle: "Test.Series.S01E01.1080p.BluRay.x264",
					Quality: struct {
						Quality struct {
							Name string `json:"name"`
						} `json:"quality"`
					}{
						Quality: struct {
							Name string `json:"name"`
						}{Name: "Bluray-1080p"},
					},
					Date:       time.Now().Add(-1 * time.Hour),
					EventType:  "grabbed",
					DownloadID: "test-download-1",
				},
				{
					ID:          2,
					EpisodeID:   456,
					SeriesID:    123,
					SourceTitle: "Test.Series.S01E01.1080p.BluRay.x264",
					Quality: struct {
						Quality struct {
							Name string `json:"name"`
						} `json:"quality"`
					}{
						Quality: struct {
							Name string `json:"name"`
						}{Name: "Bluray-1080p"},
					},
					Date:       time.Now(),
					EventType:  "downloadFolderImported",
					DownloadID: "test-download-1",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	history, err := client.GetHistory(context.Background(), 50)
	require.NoError(t, err)
	assert.Len(t, history, 2)

	assert.Equal(t, 1, history[0].ID)
	assert.Equal(t, 456, history[0].EpisodeID)
	assert.Equal(t, 123, history[0].SeriesID)
	assert.Equal(t, "grabbed", history[0].EventType)
	assert.Equal(t, "Bluray-1080p", history[0].Quality)

	assert.Equal(t, 2, history[1].ID)
	assert.Equal(t, "downloadFolderImported", history[1].EventType)
}

func TestSonarrClient_GetCalendar(t *testing.T) {
	startDate := time.Now()
	endDate := time.Now().AddDate(0, 0, 30)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/calendar", r.URL.Path)
		assert.Equal(t, startDate.Format("2006-01-02"), r.URL.Query().Get("start"))
		assert.Equal(t, endDate.Format("2006-01-02"), r.URL.Query().Get("end"))

		response := []sonarrCalendarItem{
			{
				ID:            1,
				SeriesID:      123,
				SeasonNumber:  1,
				EpisodeNumber: 5,
				Title:         "Episode 5",
				AirDate:       "2024-01-05",
				AirDateUtc:    time.Now().AddDate(0, 0, 5),
				HasFile:       false,
				Monitored:     true,
				Series: struct {
					Title string `json:"title"`
				}{Title: "Test Series"},
			},
			{
				ID:            2,
				SeriesID:      123,
				SeasonNumber:  1,
				EpisodeNumber: 6,
				Title:         "Episode 6",
				AirDate:       "2024-01-12",
				AirDateUtc:    time.Now().AddDate(0, 0, 12),
				HasFile:       false,
				Monitored:     true,
				Series: struct {
					Title string `json:"title"`
				}{Title: "Test Series"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	calendar, err := client.GetCalendar(context.Background(), startDate, endDate)
	require.NoError(t, err)
	assert.Len(t, calendar, 2)

	assert.Equal(t, 1, calendar[0].ID)
	assert.Equal(t, 123, calendar[0].SeriesID)
	assert.Equal(t, "Test Series", calendar[0].SeriesTitle)
	assert.Equal(t, 1, calendar[0].SeasonNumber)
	assert.Equal(t, 5, calendar[0].EpisodeNumber)
	assert.Equal(t, "Episode 5", calendar[0].Title)
	assert.False(t, calendar[0].HasFile)
	assert.True(t, calendar[0].Monitored)
}

func TestSonarrClient_GetQualityProfiles(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v3/qualityprofile", r.URL.Path)

		response := []sonarrQualityProfile{
			{ID: 1, Name: "Any"},
			{ID: 2, Name: "SD"},
			{ID: 3, Name: "HD-720p"},
			{ID: 4, Name: "HD-1080p"},
			{ID: 5, Name: "Ultra-HD"},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	profiles, err := client.GetQualityProfiles(context.Background())
	require.NoError(t, err)
	assert.Len(t, profiles, 5)

	assert.Equal(t, 1, profiles[0].ID)
	assert.Equal(t, "Any", profiles[0].Name)

	assert.Equal(t, 5, profiles[4].ID)
	assert.Equal(t, "Ultra-HD", profiles[4].Name)
}

func TestSonarrClient_GetQueue_EmptyQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := sonarrQueueResponse{
			Page:         1,
			PageSize:     100,
			TotalRecords: 0,
			Records:      []sonarrQueueItem{},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	queue, err := client.GetQueue(context.Background())
	require.NoError(t, err)
	assert.Empty(t, queue)
}

func TestSonarrClient_ErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "invalid-key",
		Timeout: 5 * time.Second,
	}, logger)

	_, err := client.GetSystemInfo(context.Background())
	assert.Error(t, err)

	_, err = client.GetQueue(context.Background())
	assert.Error(t, err)

	_, err = client.GetHistory(context.Background(), 50)
	assert.Error(t, err)

	_, err = client.GetQualityProfiles(context.Background())
	assert.Error(t, err)
}

func TestSonarrClient_GetQueue_ProgressValidation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test cases with invalid data that could cause negative or >100% progress
		response := sonarrQueueResponse{
			Page:         1,
			PageSize:     100,
			TotalRecords: 3,
			Records: []sonarrQueueItem{
				{
					ID:       1,
					SeriesID: 123,
					Title:    "Episode with inconsistent size (negative progress)",
					Size:     1000000000,
					Sizeleft: 1500000000, // More left than total - invalid!
					Status:   "downloading",
				},
				{
					ID:       2,
					SeriesID: 456,
					Title:    "Episode with zero size",
					Size:     0,
					Sizeleft: 0,
					Status:   "downloading",
				},
				{
					ID:       3,
					SeriesID: 789,
					Title:    "Normal episode",
					Size:     1000000000,
					Sizeleft: 250000000,
					Status:   "downloading",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	logger := zap.NewNop()
	client := NewSonarrClient(ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}, logger)

	queue, err := client.GetQueue(context.Background())
	require.NoError(t, err)
	assert.Len(t, queue, 3)

	// First item: negative progress should be clamped to 0
	assert.Equal(t, 1, queue[0].ID)
	assert.Equal(t, 0.0, queue[0].Progress, "Progress should be 0 when SizeLeft > Size")

	// Second item: zero size should result in 0% progress
	assert.Equal(t, 2, queue[1].ID)
	assert.Equal(t, 0.0, queue[1].Progress, "Progress should be 0 when Size is 0")

	// Third item: normal progress calculation
	assert.Equal(t, 3, queue[2].ID)
	assert.Equal(t, 75.0, queue[2].Progress, "Progress should be 75%")
}
