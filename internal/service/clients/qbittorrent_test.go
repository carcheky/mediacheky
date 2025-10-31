package clients

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestQBittorrentClient_GetTransferInfo(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			// Set SID cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/transfer/info":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"dl_info_speed": 1024000,
				"dl_info_data": 104857600,
				"up_info_speed": 512000,
				"up_info_data": 52428800,
				"dht_nodes": 150,
				"connection_status": "connected"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "admin", logger)

	ctx := context.Background()
	transferInfo, err := client.GetTransferInfo(ctx)

	require.NoError(t, err)
	assert.NotNil(t, transferInfo)
	assert.Equal(t, int64(1024000), transferInfo.DLInfoSpeed)
	assert.Equal(t, int64(512000), transferInfo.UPInfoSpeed)
	assert.Equal(t, 150, transferInfo.DHTNodes)
	assert.Equal(t, "connected", transferInfo.ConnectionStatus)
}

func TestQBittorrentClient_GetServerState(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/sync/maindata":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"server_state": {
					"dl_info_speed": 2048000,
					"up_info_speed": 1024000,
					"dl_info_data": 209715200,
					"up_info_data": 104857600,
					"dht_nodes": 200,
					"connection_status": "connected",
					"free_space_on_disk": 107374182400
				}
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "admin", logger)

	ctx := context.Background()
	serverState, err := client.GetServerState(ctx)

	require.NoError(t, err)
	assert.NotNil(t, serverState)
	assert.Equal(t, int64(2048000), serverState.DLInfoSpeed)
	assert.Equal(t, int64(1024000), serverState.UPInfoSpeed)
	assert.Equal(t, int64(107374182400), serverState.FreeSpaceOnDisk)
}

func TestQBittorrentClient_GetTorrentProperties(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/torrents/properties":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"addition_date": 1609459200,
				"completion_date": 1609545600,
				"total_uploaded": 10737418240,
				"total_downloaded": 5368709120,
				"nb_connections": 25,
				"seeds": 10,
				"seeds_total": 50,
				"peers": 5,
				"peers_total": 20,
				"eta": 3600,
				"save_path": "/downloads/movies"
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "admin", logger)

	ctx := context.Background()
	hash := "test-hash-123"
	props, err := client.GetTorrentProperties(ctx, hash)

	require.NoError(t, err)
	assert.NotNil(t, props)
	assert.Equal(t, int64(1609459200), props.AdditionDate)
	assert.Equal(t, int64(3600), props.ETA)
	assert.Equal(t, 10, props.Seeds)
	assert.Equal(t, 5, props.Peers)
}

func TestQBittorrentClient_GetTorrentTrackers(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/torrents/trackers":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[
				{
					"url": "http://tracker1.example.com:6969/announce",
					"status": 2,
					"num_peers": 15,
					"num_seeds": 30,
					"msg": "Working"
				},
				{
					"url": "http://tracker2.example.com:6969/announce",
					"status": 4,
					"num_peers": 0,
					"num_seeds": 0,
					"msg": "Timed out"
				}
			]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "admin", logger)

	ctx := context.Background()
	hash := "test-hash-123"
	trackers, err := client.GetTorrentTrackers(ctx, hash)

	require.NoError(t, err)
	assert.NotNil(t, trackers)
	assert.Len(t, trackers, 2)
	if len(trackers) > 0 {
		assert.Equal(t, 2, trackers[0].Status)
		assert.Equal(t, 30, trackers[0].NumSeeds)
	}
	if len(trackers) > 1 {
		assert.Equal(t, 4, trackers[1].Status)
	}
}

func TestQBittorrentClient_GetEnhancedTorrentInfo(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			http.SetCookie(w, &http.Cookie{
				Name:  "SID",
				Value: "test-session-id",
			})
			w.WriteHeader(http.StatusOK)
		case "/api/v2/torrents/info":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[{
				"hash": "test-hash-123",
				"name": "Test Movie",
				"state": "uploading",
				"progress": 1.0,
				"ratio": 2.5,
				"size": 5368709120,
				"upspeed": 1024000,
				"dlspeed": 0,
				"seeding_time": 7200,
				"category": "movies",
				"tags": "action,hd",
				"save_path": "/downloads/movies",
				"added_on": 1609459200,
				"completed_on": 1609545600,
				"eta": 8640000,
				"num_seeds": 10,
				"num_leechs": 5
			}]`))
		case "/api/v2/torrents/properties":
			// Return JSON matching qBittorrent API format
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"addition_date": 1609459200,
				"completion_date": 1609545600,
				"total_uploaded": 10737418240,
				"total_downloaded": 5368709120,
				"seeds": 10,
				"peers": 5,
				"eta": 8640000
			}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "admin", logger)

	ctx := context.Background()
	hash := "test-hash-123"
	info, err := client.GetEnhancedTorrentInfo(ctx, hash)

	require.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, "test-hash-123", info.Hash)
	assert.Equal(t, "Test Movie", info.Name)
	assert.Equal(t, float64(1.0), info.Progress)
	assert.True(t, info.IsSeeding)
	assert.True(t, info.IsComplete)
	assert.Equal(t, int64(1609459200), info.AddedOn)
	assert.Equal(t, int64(1609545600), info.CompletedOn)
	assert.Equal(t, 10, info.NumSeeds)
	assert.Equal(t, 5, info.NumPeers)
}

func TestQBittorrentClient_ErrorHandling(t *testing.T) {
	// Create test server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v2/auth/login":
			w.WriteHeader(http.StatusUnauthorized)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	logger, _ := zap.NewDevelopment()
	client := NewQBittorrentClient(server.URL, "admin", "wrongpass", logger)

	ctx := context.Background()

	t.Run("GetTransferInfo with auth error", func(t *testing.T) {
		_, err := client.GetTransferInfo(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("GetServerState with auth error", func(t *testing.T) {
		_, err := client.GetServerState(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})

	t.Run("GetTorrentProperties with auth error", func(t *testing.T) {
		_, err := client.GetTorrentProperties(ctx, "test-hash")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})
}
