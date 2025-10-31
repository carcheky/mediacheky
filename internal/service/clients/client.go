// Package clients provides interfaces and implementations for external service clients.
// It includes clients for Radarr, Sonarr, Jellyfin, Jellyseerr, and Bazarr APIs.
package clients

import (
	"context"
	"time"

	"github.com/carcheky/keepercheky/internal/models"
)

// MediaClient defines the interface for media server clients (Radarr, Sonarr).
type MediaClient interface {
	// TestConnection verifies the connection to the service
	TestConnection(ctx context.Context) error

	// GetLibrary retrieves all media items from the service
	GetLibrary(ctx context.Context) ([]*models.Media, error)

	// GetItem retrieves a specific media item by ID
	GetItem(ctx context.Context, id int) (*models.Media, error)

	// DeleteItem removes a media item from the service
	// deleteFiles indicates whether to also delete the media files from disk
	DeleteItem(ctx context.Context, id int, deleteFiles bool) error

	// GetTags retrieves all available tags from the service
	GetTags(ctx context.Context) ([]models.Tag, error)
}

// StreamingClient defines the interface for streaming service clients (Jellyfin).
type StreamingClient interface {
	// TestConnection verifies the connection to the service
	TestConnection(ctx context.Context) error

	// GetLibrary retrieves all media items from the service
	GetLibrary(ctx context.Context) ([]*models.Media, error)

	// GetPlaybackInfo retrieves playback information for media
	GetPlaybackInfo(ctx context.Context, mediaID string) (*models.PlaybackInfo, error)

	// DeleteItem removes a media item from the service
	DeleteItem(ctx context.Context, id string) error
}

// RequestClient defines the interface for request management clients (Jellyseerr).
type RequestClient interface {
	// TestConnection verifies the connection to the service
	TestConnection(ctx context.Context) error

	// GetRequests retrieves all active requests
	GetRequests(ctx context.Context) ([]*models.Request, error)

	// GetRequest retrieves a specific request by ID
	GetRequest(ctx context.Context, id int) (*models.Request, error)

	// DeleteRequest removes a request from the service
	DeleteRequest(ctx context.Context, id int) error
}

// SubtitleClient defines the interface for subtitle management clients (Bazarr).
type SubtitleClient interface {
	// TestConnection verifies the connection to the service
	TestConnection(ctx context.Context) error

	// GetMovieSubtitles retrieves subtitle information for a specific movie
	GetMovieSubtitles(ctx context.Context, radarrID int) ([]models.Subtitle, error)

	// GetSeriesSubtitles retrieves subtitle information for a specific series
	GetSeriesSubtitles(ctx context.Context, sonarrID int) ([]models.Subtitle, error)
}

// ClientConfig holds common configuration for all clients.
type ClientConfig struct {
	BaseURL string
	APIKey  string
	Timeout time.Duration
}

// DefaultTimeout is the default timeout for HTTP requests.
const DefaultTimeout = 30 * time.Second

// MaxRetries is the maximum number of retry attempts for failed requests.
const MaxRetries = 3

// RetryDelay is the initial delay between retry attempts.
const RetryDelay = 1 * time.Second

// DefaultRecentlyAddedLimit is the default number of recently added items to retrieve.
const DefaultRecentlyAddedLimit = 20

// DefaultActivityLogLimit is the default number of activity log entries to retrieve.
const DefaultActivityLogLimit = 50
