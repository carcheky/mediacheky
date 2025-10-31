package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// JellyseerrClient implements the RequestClient interface for Jellyseerr/Overseerr.
type JellyseerrClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	logger  *zap.Logger
}

// NewJellyseerrClient creates a new Jellyseerr client.
func NewJellyseerrClient(config ClientConfig, logger *zap.Logger) *JellyseerrClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetHeader("X-Api-Key", config.APIKey)
	client.SetTimeout(config.Timeout)

	if config.Timeout == 0 {
		client.SetTimeout(DefaultTimeout)
	}

	return &JellyseerrClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// jellyseerrStatus represents the status response from Jellyseerr API.
type jellyseerrStatus struct {
	Version       string `json:"version"`
	CommitTag     string `json:"commitTag"`
	UpdateAvail   bool   `json:"updateAvailable"`
	CommitsBehind int    `json:"commitsBehind"`
}

// JellyseerrSystemInfo representa toda la información del sistema de Jellyseerr
type JellyseerrSystemInfo struct {
	Version         string `json:"version"`
	CommitTag       string `json:"commit_tag"`
	UpdateAvailable bool   `json:"update_available"`
	CommitsBehind   int    `json:"commits_behind"`
}

// jellyseerrRequest represents a request from Jellyseerr API.
type jellyseerrRequest struct {
	ID          int       `json:"id"`
	Status      int       `json:"status"` // 1=pending, 2=approved, 3=available, 4=denied
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Type        string    `json:"type"` // "movie" or "tv"
	RequestedBy struct {
		DisplayName string `json:"displayName"`
	} `json:"requestedBy"`
	Media struct {
		TMDBID       int    `json:"tmdbId"`
		Status       int    `json:"status"`
		ExternalID   int    `json:"serviceId"`
		ExternalType string `json:"serviceId4k"` // Radarr/Sonarr ID
	} `json:"media"`
	Seasons []struct {
		ID     int `json:"id"`
		Status int `json:"status"`
	} `json:"seasons,omitempty"`
}

// jellyseerrRequestsResponse represents the response from requests endpoint.
type jellyseerrRequestsResponse struct {
	PageInfo struct {
		Pages   int `json:"pages"`
		Results int `json:"results"`
	} `json:"pageInfo"`
	Results []jellyseerrRequest `json:"results"`
}

// TestConnection verifies the connection to Jellyseerr.
func (c *JellyseerrClient) TestConnection(ctx context.Context) error {
	return c.callWithRetry(ctx, func() error {
		var status jellyseerrStatus
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&status).
			Get("/api/v1/status")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Jellyseerr connection successful",
			zap.String("version", status.Version),
			zap.String("url", c.baseURL),
		)

		return nil
	})
}

// GetSystemInfo retrieves complete system information from Jellyseerr.
func (c *JellyseerrClient) GetSystemInfo(ctx context.Context) (*JellyseerrSystemInfo, error) {
	var status jellyseerrStatus

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&status).
			Get("/api/v1/status")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convertir a nuestro modelo
	systemInfo := &JellyseerrSystemInfo{
		Version:         status.Version,
		CommitTag:       status.CommitTag,
		UpdateAvailable: status.UpdateAvail,
		CommitsBehind:   status.CommitsBehind,
	}

	c.logger.Info("Retrieved Jellyseerr system info",
		zap.String("version", systemInfo.Version),
		zap.String("commit_tag", systemInfo.CommitTag),
		zap.Bool("update_available", systemInfo.UpdateAvailable),
	)

	return systemInfo, nil
}

// GetRequests retrieves all active requests from Jellyseerr.
func (c *JellyseerrClient) GetRequests(ctx context.Context) ([]*models.Request, error) {
	var response jellyseerrRequestsResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParams(map[string]string{
				"take":   "100",
				"skip":   "0",
				"filter": "all",
				"sort":   "added",
			}).
			Get("/api/v1/request")

		if err != nil {
			return fmt.Errorf("failed to get requests: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert Jellyseerr requests to internal Request model
	requests := make([]*models.Request, 0, len(response.Results))
	for _, req := range response.Results {
		request := c.convertToRequest(&req)
		requests = append(requests, request)
	}

	c.logger.Info("Retrieved Jellyseerr requests",
		zap.Int("total_requests", response.PageInfo.Results),
		zap.Int("retrieved", len(requests)),
	)

	return requests, nil
}

// GetRequest retrieves a specific request from Jellyseerr.
func (c *JellyseerrClient) GetRequest(ctx context.Context, id int) (*models.Request, error) {
	var req jellyseerrRequest

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&req).
			SetPathParam("id", fmt.Sprintf("%d", id)).
			Get("/api/v1/request/{id}")

		if err != nil {
			return fmt.Errorf("failed to get request: %w", err)
		}

		if resp.StatusCode() == 404 {
			return fmt.Errorf("request not found: %d", id)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return c.convertToRequest(&req), nil
}

// DeleteRequest removes a request from Jellyseerr.
func (c *JellyseerrClient) DeleteRequest(ctx context.Context, id int) error {
	return c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetPathParam("id", fmt.Sprintf("%d", id)).
			Delete("/api/v1/request/{id}")

		if err != nil {
			return fmt.Errorf("failed to delete request: %w", err)
		}

		if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Deleted request from Jellyseerr",
			zap.Int("request_id", id),
		)

		return nil
	})
}

// convertToRequest converts a Jellyseerr request to internal Request model.
func (c *JellyseerrClient) convertToRequest(req *jellyseerrRequest) *models.Request {
	// Convert status code to string
	statusMap := map[int]string{
		1: "pending",
		2: "approved",
		3: "available",
		4: "denied",
	}

	status := statusMap[req.Status]
	if status == "" {
		status = "unknown"
	}

	request := &models.Request{
		ServiceID:   req.ID,
		MediaType:   req.Type,
		MediaTitle:  fmt.Sprintf("TMDB-%d", req.Media.TMDBID), // We'll need to resolve this later
		Status:      status,
		RequestedBy: req.RequestedBy.DisplayName,
		RequestedAt: req.CreatedAt,
	}

	// Link to Radarr/Sonarr if available
	if req.Media.ExternalID > 0 {
		if req.Type == "movie" {
			request.RadarrID = &req.Media.ExternalID
		} else {
			request.SonarrID = &req.Media.ExternalID
		}
	}

	return request
}

// JellyseerrRequestStats represents statistics about requests in Jellyseerr
type JellyseerrRequestStats struct {
	TotalRequests     int `json:"total_requests"`
	PendingRequests   int `json:"pending_requests"`
	ApprovedRequests  int `json:"approved_requests"`
	AvailableRequests int `json:"available_requests"`
	DeniedRequests    int `json:"denied_requests"`
}

// GetRequestStats retrieves statistics about all requests
func (c *JellyseerrClient) GetRequestStats(ctx context.Context) (*JellyseerrRequestStats, error) {
	// Get all requests
	requests, err := c.GetRequests(ctx)
	if err != nil {
		return nil, err
	}

	// Calculate statistics
	stats := &JellyseerrRequestStats{}
	for _, req := range requests {
		stats.TotalRequests++
		switch req.Status {
		case "pending":
			stats.PendingRequests++
		case "approved":
			stats.ApprovedRequests++
		case "available":
			stats.AvailableRequests++
		case "denied":
			stats.DeniedRequests++
		}
	}

	c.logger.Info("Retrieved Jellyseerr request statistics",
		zap.Int("total", stats.TotalRequests),
		zap.Int("pending", stats.PendingRequests),
		zap.Int("approved", stats.ApprovedRequests),
		zap.Int("available", stats.AvailableRequests),
	)

	return stats, nil
}

// callWithRetry executes a function with retry logic.
func (c *JellyseerrClient) callWithRetry(ctx context.Context, fn func() error) error {
	var lastErr error
	backoff := RetryDelay

	for i := 0; i < MaxRetries; i++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if i < MaxRetries-1 {
			c.logger.Warn("Jellyseerr API call failed, retrying",
				zap.Int("attempt", i+1),
				zap.Int("max_retries", MaxRetries),
				zap.Error(err),
				zap.Duration("backoff", backoff),
			)

			select {
			case <-time.After(backoff):
				backoff *= 2 // Exponential backoff
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
