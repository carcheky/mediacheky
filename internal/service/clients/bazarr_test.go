package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestBazarrClient_TestConnection(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "successful connection",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"data": bazarrSystemStatus{
					BazarrVersion: "1.2.0",
					OSName:        "Linux",
					PythonVersion: "3.10.0",
				},
			},
			wantErr: false,
		},
		{
			name:       "connection refused",
			statusCode: http.StatusServiceUnavailable,
			response:   nil,
			wantErr:    true,
		},
		{
			name:       "unauthorized",
			statusCode: http.StatusUnauthorized,
			response:   nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Check API key in query parameter
				if r.URL.Query().Get("apikey") != "test-api-key" {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			config := ClientConfig{
				BaseURL: server.URL,
				APIKey:  "test-api-key",
				Timeout: 5 * time.Second,
			}

			client := NewBazarrClient(config, logger)
			err := client.TestConnection(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("TestConnection() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBazarrClient_GetSystemInfo(t *testing.T) {
	logger := zap.NewNop()

	mockStatus := bazarrSystemStatus{
		BazarrVersion:   "1.2.0",
		SonarrVersion:   "4.0.0",
		RadarrVersion:   "5.0.0",
		OSName:          "Linux",
		OSVersion:       "5.15.0",
		PythonVersion:   "3.10.0",
		BazarrDirectory: "/app/bazarr",
		StartTime:       "2024-01-01T00:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/system/status" {
			t.Errorf("Expected path /api/system/status, got %s", r.URL.Path)
		}

		// Check API key
		if r.URL.Query().Get("apikey") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": mockStatus,
		})
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}

	client := NewBazarrClient(config, logger)
	info, err := client.GetSystemInfo(context.Background())

	if err != nil {
		t.Fatalf("GetSystemInfo() error = %v", err)
	}

	if info.BazarrVersion != mockStatus.BazarrVersion {
		t.Errorf("Expected version %s, got %s", mockStatus.BazarrVersion, info.BazarrVersion)
	}

	if info.OS != mockStatus.OSName {
		t.Errorf("Expected OS %s, got %s", mockStatus.OSName, info.OS)
	}
}

func TestBazarrClient_GetHistory(t *testing.T) {
	logger := zap.NewNop()

	mockHistory := struct {
		Data []bazarrHistoryItem `json:"data"`
	}{
		Data: []bazarrHistoryItem{
			{
				ID:          1,
				Action:      1, // download
				Title:       "Test Movie",
				Timestamp:   time.Now(),
				Description: "Downloaded subtitle",
				RadarrID:    100,
				Provider:    "opensubtitles",
				Language:    "en",
				Score:       95,
			},
			{
				ID:          2,
				Action:      1,
				Title:       "Test Series",
				Timestamp:   time.Now(),
				Description: "Downloaded subtitle",
				SonarrID:    200,
				Provider:    "subscene",
				Language:    "es",
				Score:       90,
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/history" {
			t.Errorf("Expected path /api/history, got %s", r.URL.Path)
		}

		// Check API key
		if r.URL.Query().Get("apikey") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockHistory)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}

	client := NewBazarrClient(config, logger)
	history, err := client.GetHistory(context.Background(), 50)

	if err != nil {
		t.Fatalf("GetHistory() error = %v", err)
	}

	if len(history) != 2 {
		t.Errorf("Expected 2 history items, got %d", len(history))
	}

	if len(history) > 0 {
		if history[0].Action != "download" {
			t.Errorf("Expected action 'download', got %s", history[0].Action)
		}

		if history[0].MediaType != "movie" {
			t.Errorf("Expected media type 'movie', got %s", history[0].MediaType)
		}
	}

	if len(history) > 1 {
		if history[1].MediaType != "series" {
			t.Errorf("Expected media type 'series', got %s", history[1].MediaType)
		}
	}
}

func TestBazarrClient_GetWantedMovies(t *testing.T) {
	logger := zap.NewNop()

	mockWanted := struct {
		Data []bazarrWantedItem `json:"data"`
	}{
		Data: []bazarrWantedItem{
			{
				RadarrID:         100,
				Title:            "Test Movie",
				MissingSubtitles: []string{"en", "es"},
			},
			{
				RadarrID:         101,
				Title:            "Another Movie",
				MissingSubtitles: []string{"fr"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/movies/wanted" {
			t.Errorf("Expected path /api/movies/wanted, got %s", r.URL.Path)
		}

		// Check API key
		if r.URL.Query().Get("apikey") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockWanted)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}

	client := NewBazarrClient(config, logger)
	wanted, err := client.GetWantedMovies(context.Background())

	if err != nil {
		t.Fatalf("GetWantedMovies() error = %v", err)
	}

	if len(wanted) != 2 {
		t.Errorf("Expected 2 wanted items, got %d", len(wanted))
	}

	if len(wanted) > 0 {
		if wanted[0].MediaType != "movie" {
			t.Errorf("Expected media type 'movie', got %s", wanted[0].MediaType)
		}

		if len(wanted[0].MissingSubtitles) != 2 {
			t.Errorf("Expected 2 missing subtitles, got %d", len(wanted[0].MissingSubtitles))
		}
	}
}

func TestBazarrClient_GetWantedSeries(t *testing.T) {
	logger := zap.NewNop()

	mockWanted := struct {
		Data []bazarrWantedItem `json:"data"`
	}{
		Data: []bazarrWantedItem{
			{
				SonarrID:         200,
				Title:            "Test Series",
				MissingSubtitles: []string{"en"},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/series/wanted" {
			t.Errorf("Expected path /api/series/wanted, got %s", r.URL.Path)
		}

		// Check API key
		if r.URL.Query().Get("apikey") != "test-api-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockWanted)
	}))
	defer server.Close()

	config := ClientConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
		Timeout: 5 * time.Second,
	}

	client := NewBazarrClient(config, logger)
	wanted, err := client.GetWantedSeries(context.Background())

	if err != nil {
		t.Fatalf("GetWantedSeries() error = %v", err)
	}

	if len(wanted) != 1 {
		t.Errorf("Expected 1 wanted item, got %d", len(wanted))
	}

	if len(wanted) > 0 {
		if wanted[0].MediaType != "series" {
			t.Errorf("Expected media type 'series', got %s", wanted[0].MediaType)
		}
	}
}
