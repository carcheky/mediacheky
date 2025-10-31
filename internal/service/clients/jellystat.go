package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// JellystatClient implements a client for Jellystat statistics service.
type JellystatClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	logger  *zap.Logger
}

// NewJellystatClient creates a new Jellystat client.
func NewJellystatClient(config ClientConfig, logger *zap.Logger) *JellystatClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetHeader("x-api-token", config.APIKey)
	client.SetTimeout(config.Timeout)

	if config.Timeout == 0 {
		client.SetTimeout(DefaultTimeout)
	}

	return &JellystatClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// jellystatSystemInfo represents the system info response from Jellystat API.
// nolint:unused // Reserved for future use
type jellystatSystemInfo struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// JellystatSystemInfo represents complete system information from Jellystat.
type JellystatSystemInfo struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

// JellystatStatistics represents general statistics from Jellystat.
type JellystatStatistics struct {
	Days     int `json:"days"`
	Movies   int `json:"movies"`
	Episodes int `json:"episodes"`
	Songs    int `json:"songs"`
	Total    int `json:"total"`
}

// ViewsByLibraryType represents views aggregated by library type.
type ViewsByLibraryType struct {
	Music   int `json:"music"`
	Movie   int `json:"movie"`
	Episode int `json:"episode"`
	Book    int `json:"book"`
}

// UserActivity represents user activity statistics.
type UserActivity struct {
	UserID       string `json:"user_id"`
	UserName     string `json:"user_name"`
	TotalPlays   int    `json:"total_plays"`
	TotalMinutes int    `json:"total_minutes"`
}

// JellystatLibraryStats represents statistics for a specific library in Jellystat.
type JellystatLibraryStats struct {
	LibraryID    string `json:"library_id"`
	LibraryName  string `json:"library_name"`
	TotalItems   int    `json:"total_items"`
	TotalPlays   int    `json:"total_plays"`
	TotalMinutes int    `json:"total_minutes"`
}

// TestConnection verifies the connection to Jellystat.
func (c *JellystatClient) TestConnection(ctx context.Context) error {
	return c.callWithRetry(ctx, func() error {
		// Try the getconfig endpoint which is publicly available
		resp, err := c.client.R().
			SetContext(ctx).
			Get("/api/getconfig")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d - %s", resp.StatusCode(), string(resp.Body()))
		}

		c.logger.Info("Jellystat connection successful",
			zap.String("url", c.baseURL),
		)

		return nil
	})
}

// GetSystemInfo retrieves complete system information from Jellystat.
func (c *JellystatClient) GetSystemInfo(ctx context.Context) (*JellystatSystemInfo, error) {
	var info struct {
		JFHOST       string `json:"JF_HOST"`
		APPUSER      string `json:"APP_USER"`
		RequireLogin bool   `json:"REQUIRE_LOGIN"`
		IsJellyfin   bool   `json:"IS_JELLYFIN"`
	}

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&info).
			Get("/api/getconfig")

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

	// Convert to our model
	systemInfo := &JellystatSystemInfo{
		Version: "unknown", // Jellystat doesn't expose version in getconfig
		Status:  "connected",
	}

	c.logger.Info("Retrieved Jellystat system info",
		zap.String("host", info.JFHOST),
		zap.String("status", systemInfo.Status),
	)

	return systemInfo, nil
}

// GetStatistics retrieves general statistics from Jellystat.
func (c *JellystatClient) GetStatistics(ctx context.Context, days int) (*JellystatStatistics, error) {
	var stats struct {
		Days     int `json:"days"`
		Movies   int `json:"movies"`
		Episodes int `json:"episodes"`
		Songs    int `json:"songs"`
	}

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&stats).
			SetQueryParam("days", fmt.Sprintf("%d", days)).
			Get("/api/statistics")

		if err != nil {
			return fmt.Errorf("failed to get statistics: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	result := &JellystatStatistics{
		Days:     stats.Days,
		Movies:   stats.Movies,
		Episodes: stats.Episodes,
		Songs:    stats.Songs,
		Total:    stats.Movies + stats.Episodes + stats.Songs,
	}

	c.logger.Info("Retrieved Jellystat statistics",
		zap.Int("days", days),
		zap.Int("movies", result.Movies),
		zap.Int("episodes", result.Episodes),
		zap.Int("songs", result.Songs),
		zap.Int("total", result.Total),
	)

	return result, nil
}

// GetViewsByLibraryType retrieves views aggregated by library type.
func (c *JellystatClient) GetViewsByLibraryType(ctx context.Context, days int) (*ViewsByLibraryType, error) {
	var views ViewsByLibraryType

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&views).
			SetQueryParam("days", fmt.Sprintf("%d", days)).
			Get("/api/stats/getViewsByLibraryType")

		if err != nil {
			return fmt.Errorf("failed to get views by library type: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved Jellystat views by library type",
		zap.Int("days", days),
		zap.Int("movies", views.Movie),
		zap.Int("episodes", views.Episode),
		zap.Int("music", views.Music),
		zap.Int("books", views.Book),
	)

	return &views, nil
}

// GetUserActivity retrieves user activity statistics.
func (c *JellystatClient) GetUserActivity(ctx context.Context, days int) ([]UserActivity, error) {
	var activities []UserActivity

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&activities).
			SetQueryParam("days", fmt.Sprintf("%d", days)).
			Get("/api/stats/getUserActivity")

		if err != nil {
			return fmt.Errorf("failed to get user activity: %w", err)
		}

		if resp.StatusCode() != 200 {
			// If endpoint doesn't exist, return empty array
			if resp.StatusCode() == 404 {
				c.logger.Warn("User activity endpoint not available")
				return nil
			}
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved Jellystat user activity",
		zap.Int("days", days),
		zap.Int("user_count", len(activities)),
	)

	return activities, nil
}

// GetLibraryStats retrieves statistics for all libraries.
func (c *JellystatClient) GetLibraryStats(ctx context.Context, days int) ([]JellystatLibraryStats, error) {
	var stats []JellystatLibraryStats

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&stats).
			SetQueryParam("days", fmt.Sprintf("%d", days)).
			Get("/api/stats/getLibraryStats")

		if err != nil {
			return fmt.Errorf("failed to get library stats: %w", err)
		}

		if resp.StatusCode() != 200 {
			// If endpoint doesn't exist, return empty array
			if resp.StatusCode() == 404 {
				c.logger.Warn("Library stats endpoint not available")
				return nil
			}
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved Jellystat library stats",
		zap.Int("days", days),
		zap.Int("library_count", len(stats)),
	)

	return stats, nil
}

// callWithRetry executes a function with retry logic.
func (c *JellystatClient) callWithRetry(ctx context.Context, fn func() error) error {
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
			c.logger.Warn("Jellystat API call failed, retrying",
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
