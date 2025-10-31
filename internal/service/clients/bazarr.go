package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// BazarrClient implements subtitle management client for Bazarr.
type BazarrClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	logger  *zap.Logger
}

// NewBazarrClient creates a new Bazarr client.
func NewBazarrClient(config ClientConfig, logger *zap.Logger) *BazarrClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)

	// Bazarr uses apikey parameter instead of header
	client.SetQueryParam("apikey", config.APIKey)

	// Disable HTTP caching to always get fresh data
	client.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	client.SetHeader("Pragma", "no-cache")
	client.SetHeader("Expires", "0")

	client.SetTimeout(config.Timeout)

	if config.Timeout == 0 {
		client.SetTimeout(DefaultTimeout)
	}

	return &BazarrClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// bazarrSystemStatus represents the system status response from Bazarr API.
type bazarrSystemStatus struct {
	BazarrVersion   string `json:"bazarr_version"`
	SonarrVersion   string `json:"sonarr_version"`
	RadarrVersion   string `json:"radarr_version"`
	OSName          string `json:"operating_system"`
	OSVersion       string `json:"operating_system_version"`
	PythonVersion   string `json:"python_version"`
	BazarrDirectory string `json:"bazarr_directory"`
	StartTime       string `json:"start_time"`
}

// BazarrSystemInfo represents complete system information from Bazarr.
type BazarrSystemInfo struct {
	BazarrVersion string `json:"bazarr_version"`
	SonarrVersion string `json:"sonarr_version"`
	RadarrVersion string `json:"radarr_version"`
	OS            string `json:"os"`
	OSVersion     string `json:"os_version"`
	PythonVersion string `json:"python_version"`
	Directory     string `json:"directory"`
	StartTime     string `json:"start_time"`
}

// bazarrMovieSubtitle represents subtitle information for a movie.
type bazarrMovieSubtitle struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Forced   bool   `json:"forced"`
	HI       bool   `json:"hi"`
	Code2    string `json:"code2"`
	Code3    string `json:"code3"`
}

// bazarrSeriesSubtitle represents subtitle information for a series.
type bazarrSeriesSubtitle struct {
	Path     string `json:"path"`
	Language string `json:"language"`
	Forced   bool   `json:"forced"`
	HI       bool   `json:"hi"`
	Code2    string `json:"code2"`
	Code3    string `json:"code3"`
}

// bazarrMovie represents a movie from Bazarr API.
type bazarrMovie struct {
	RadarrID         int                   `json:"radarrId"`
	Title            string                `json:"title"`
	Path             string                `json:"path"`
	Subtitles        []bazarrMovieSubtitle `json:"subtitles"`
	MissingSubtitles []string              `json:"missing_subtitles"`
}

// bazarrSeries represents a series from Bazarr API.
type bazarrSeries struct {
	SonarrID         int                    `json:"sonarrSeriesId"`
	Title            string                 `json:"title"`
	Path             string                 `json:"path"`
	Subtitles        []bazarrSeriesSubtitle `json:"subtitles"`
	MissingSubtitles []string               `json:"missing_subtitles"`
}

// bazarrHistoryItem represents a history entry from Bazarr.
type bazarrHistoryItem struct {
	ID              int       `json:"id"`
	Action          int       `json:"action"`
	Title           string    `json:"title"`
	Timestamp       time.Time `json:"timestamp"`
	Description     string    `json:"description"`
	SonarrID        int       `json:"sonarrSeriesId"`
	SonarrEpisodeID int       `json:"sonarrEpisodeId"`
	RadarrID        int       `json:"radarrId"`
	Provider        string    `json:"provider"`
	Language        string    `json:"language"`
	Score           int       `json:"score"`
}

// BazarrHistoryItem represents a processed history item for public API.
type BazarrHistoryItem struct {
	ID          int       `json:"id"`
	Action      string    `json:"action"`
	Title       string    `json:"title"`
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	MediaType   string    `json:"media_type"` // "movie" or "series"
	MediaID     int       `json:"media_id"`
	Provider    string    `json:"provider"`
	Language    string    `json:"language"`
	Score       int       `json:"score"`
}

// bazarrWantedItem represents a wanted subtitle item.
type bazarrWantedItem struct {
	RadarrID         int      `json:"radarrId"`
	SonarrID         int      `json:"sonarrSeriesId"`
	SonarrEpisodeID  int      `json:"sonarrEpisodeId"`
	Title            string   `json:"title"`
	MissingSubtitles []string `json:"missing_subtitles"`
}

// BazarrWantedItem represents a wanted subtitle for public API.
type BazarrWantedItem struct {
	MediaType        string   `json:"media_type"` // "movie" or "series"
	MediaID          int      `json:"media_id"`
	Title            string   `json:"title"`
	MissingSubtitles []string `json:"missing_subtitles"`
}

// TestConnection verifies the connection to Bazarr.
func (c *BazarrClient) TestConnection(ctx context.Context) error {
	return c.callWithRetry(ctx, func() error {
		var resp struct {
			Data bazarrSystemStatus `json:"data"`
		}
		httpResp, err := c.client.R().
			SetContext(ctx).
			SetResult(&resp).
			Get("/api/system/status")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if httpResp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode())
		}

		c.logger.Info("Bazarr connection successful",
			zap.String("version", resp.Data.BazarrVersion),
			zap.String("url", c.baseURL),
		)

		return nil
	})
}

// GetSystemInfo retrieves complete system information from Bazarr.
func (c *BazarrClient) GetSystemInfo(ctx context.Context) (*BazarrSystemInfo, error) {
	var resp struct {
		Data bazarrSystemStatus `json:"data"`
	}

	err := c.callWithRetry(ctx, func() error {
		httpResp, err := c.client.R().
			SetContext(ctx).
			SetResult(&resp).
			Get("/api/system/status")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if httpResp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	status := resp.Data

	// Convert to our model
	info := &BazarrSystemInfo{
		BazarrVersion: status.BazarrVersion,
		SonarrVersion: status.SonarrVersion,
		RadarrVersion: status.RadarrVersion,
		OS:            status.OSName,
		OSVersion:     status.OSVersion,
		PythonVersion: status.PythonVersion,
		Directory:     status.BazarrDirectory,
		StartTime:     status.StartTime,
	}

	c.logger.Info("Retrieved Bazarr system info",
		zap.String("version", info.BazarrVersion),
		zap.String("os", info.OS),
		zap.String("python", info.PythonVersion),
	)

	return info, nil
}

// GetMovieSubtitles retrieves subtitle information for a specific movie by Radarr ID.
func (c *BazarrClient) GetMovieSubtitles(ctx context.Context, radarrID int) ([]models.Subtitle, error) {
	var movies []bazarrMovie

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&movies).
			Get("/api/movies")

		if err != nil {
			return fmt.Errorf("failed to get movies: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Find the specific movie
	for _, movie := range movies {
		if movie.RadarrID == radarrID {
			subtitles := make([]models.Subtitle, 0, len(movie.Subtitles))
			for _, sub := range movie.Subtitles {
				subtitles = append(subtitles, models.Subtitle{
					Language: sub.Language,
					Path:     sub.Path,
					Forced:   sub.Forced,
					// Bazarr API does not provide a "default" flag for subtitles.
					// As a result, we always set Default: false here.
					// TODO: Investigate if future Bazarr versions add support for the default flag.
					// Implication: The application cannot determine which subtitle is marked as default in Bazarr,
					// so UI and selection logic should not rely on this field for Bazarr-managed subtitles.
					Default: false,
				})
			}

			c.logger.Info("Retrieved movie subtitles from Bazarr",
				zap.Int("radarr_id", radarrID),
				zap.Int("subtitle_count", len(subtitles)),
			)

			return subtitles, nil
		}
	}

	return []models.Subtitle{}, nil
}

// GetSeriesSubtitles retrieves subtitle information for a specific series by Sonarr ID.
func (c *BazarrClient) GetSeriesSubtitles(ctx context.Context, sonarrID int) ([]models.Subtitle, error) {
	var series []bazarrSeries

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&series).
			Get("/api/series")

		if err != nil {
			return fmt.Errorf("failed to get series: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Find the specific series
	for _, s := range series {
		if s.SonarrID == sonarrID {
			subtitles := make([]models.Subtitle, 0, len(s.Subtitles))
			for _, sub := range s.Subtitles {
				subtitles = append(subtitles, models.Subtitle{
					Language: sub.Language,
					Path:     sub.Path,
					Forced:   sub.Forced,
					// Bazarr API does not provide a "default" flag for subtitles.
					// As a result, we always set Default: false here.
					// TODO: Investigate if future Bazarr versions add support for the default flag.
					// Implication: The application cannot determine which subtitle is marked as default in Bazarr,
					// so UI and selection logic should not rely on this field for Bazarr-managed subtitles.
					Default: false,
				})
			}

			c.logger.Info("Retrieved series subtitles from Bazarr",
				zap.Int("sonarr_id", sonarrID),
				zap.Int("subtitle_count", len(subtitles)),
			)

			return subtitles, nil
		}
	}

	return []models.Subtitle{}, nil
}

// GetHistory retrieves recent subtitle download history from Bazarr.
func (c *BazarrClient) GetHistory(ctx context.Context, pageSize int) ([]BazarrHistoryItem, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var historyResp struct {
		Data []bazarrHistoryItem `json:"data"`
	}

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&historyResp).
			SetQueryParam("length", fmt.Sprintf("%d", pageSize)).
			Get("/api/history")

		if err != nil {
			return fmt.Errorf("failed to get history: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to public API model
	items := make([]BazarrHistoryItem, 0, len(historyResp.Data))
	for _, record := range historyResp.Data {
		action := "unknown"
		switch record.Action {
		case 1:
			action = "download"
		case 2:
			action = "manual"
		case 3:
			action = "upgrade"
		case 4:
			action = "manual_upgrade"
		case 5:
			action = "delete"
		}

		mediaType := "movie"
		mediaID := record.RadarrID
		if record.SonarrID > 0 {
			mediaType = "series"
			mediaID = record.SonarrID
		}

		items = append(items, BazarrHistoryItem{
			ID:          record.ID,
			Action:      action,
			Title:       record.Title,
			Timestamp:   record.Timestamp,
			Description: record.Description,
			MediaType:   mediaType,
			MediaID:     mediaID,
			Provider:    record.Provider,
			Language:    record.Language,
			Score:       record.Score,
		})
	}

	c.logger.Info("Retrieved Bazarr history",
		zap.Int("total_items", len(items)),
	)

	return items, nil
}

// GetWantedMovies retrieves movies with missing subtitles.
func (c *BazarrClient) GetWantedMovies(ctx context.Context) ([]BazarrWantedItem, error) {
	var wantedResp struct {
		Data []bazarrWantedItem `json:"data"`
	}

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&wantedResp).
			Get("/api/movies/wanted")

		if err != nil {
			return fmt.Errorf("failed to get wanted movies: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to public API model
	items := make([]BazarrWantedItem, 0, len(wantedResp.Data))
	for _, record := range wantedResp.Data {
		items = append(items, BazarrWantedItem{
			MediaType:        "movie",
			MediaID:          record.RadarrID,
			Title:            record.Title,
			MissingSubtitles: record.MissingSubtitles,
		})
	}

	c.logger.Info("Retrieved wanted movie subtitles from Bazarr",
		zap.Int("total_items", len(items)),
	)

	return items, nil
}

// GetWantedSeries retrieves series with missing subtitles.
func (c *BazarrClient) GetWantedSeries(ctx context.Context) ([]BazarrWantedItem, error) {
	var wantedResp struct {
		Data []bazarrWantedItem `json:"data"`
	}

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&wantedResp).
			Get("/api/series/wanted")

		if err != nil {
			return fmt.Errorf("failed to get wanted series: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to public API model
	items := make([]BazarrWantedItem, 0, len(wantedResp.Data))
	for _, record := range wantedResp.Data {
		items = append(items, BazarrWantedItem{
			MediaType:        "series",
			MediaID:          record.SonarrID,
			Title:            record.Title,
			MissingSubtitles: record.MissingSubtitles,
		})
	}

	c.logger.Info("Retrieved wanted series subtitles from Bazarr",
		zap.Int("total_items", len(items)),
	)

	return items, nil
}

// callWithRetry executes a function with retry logic.
func (c *BazarrClient) callWithRetry(ctx context.Context, fn func() error) error {
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
			c.logger.Warn("Bazarr API call failed, retrying",
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
