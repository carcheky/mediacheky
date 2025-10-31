package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// RadarrClient implements the MediaClient interface for Radarr.
type RadarrClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	logger  *zap.Logger
}

// NewRadarrClient creates a new Radarr client.
func NewRadarrClient(config ClientConfig, logger *zap.Logger) *RadarrClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetHeader("X-Api-Key", config.APIKey)

	// Disable HTTP caching to always get fresh data
	client.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	client.SetHeader("Pragma", "no-cache")
	client.SetHeader("Expires", "0")

	client.SetTimeout(config.Timeout)

	if config.Timeout == 0 {
		client.SetTimeout(DefaultTimeout)
	}

	return &RadarrClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// radarrSystemStatus represents the system status response from Radarr API.
type radarrSystemStatus struct {
	Version           string `json:"version"`
	BuildTime         string `json:"buildTime"`
	IsDebug           bool   `json:"isDebug"`
	IsProduction      bool   `json:"isProduction"`
	IsAdmin           bool   `json:"isAdmin"`
	IsUserInteractive bool   `json:"isUserInteractive"`
	StartupPath       string `json:"startupPath"`
	AppData           string `json:"appData"`
	OsName            string `json:"osName"`
	OsVersion         string `json:"osVersion"`
	IsMonoRuntime     bool   `json:"isMonoRuntime"`
	IsMono            bool   `json:"isMono"`
	IsLinux           bool   `json:"isLinux"`
	IsOsx             bool   `json:"isOsx"`
	IsWindows         bool   `json:"isWindows"`
	Mode              string `json:"mode"`
	Branch            string `json:"branch"`
	Authentication    string `json:"authentication"`
	SqliteVersion     string `json:"sqliteVersion"`
	UrlBase           string `json:"urlBase"`
	RuntimeVersion    string `json:"runtimeVersion"`
	RuntimeName       string `json:"runtimeName"`
}

// RadarrSystemInfo representa toda la información del sistema de Radarr
type RadarrSystemInfo struct {
	Version        string `json:"version"`
	BuildTime      string `json:"build_time"`
	Branch         string `json:"branch"`
	OS             string `json:"os"`
	OSVersion      string `json:"os_version"`
	Runtime        string `json:"runtime"`
	RuntimeVersion string `json:"runtime_version"`
	IsDebug        bool   `json:"is_debug"`
	IsProduction   bool   `json:"is_production"`
	Authentication string `json:"authentication"`
	URLBase        string `json:"url_base"`
	StartupPath    string `json:"startup_path"`
	AppData        string `json:"app_data"`
	SqliteVersion  string `json:"sqlite_version"`
}

// radarrMovie represents a movie from Radarr API.
type radarrMovie struct {
	ID         int       `json:"id"`
	Title      string    `json:"title"`
	Path       string    `json:"path"`
	SizeOnDisk int64     `json:"sizeOnDisk"`
	Added      time.Time `json:"added"`
	Quality    struct {
		Quality struct {
			Name string `json:"name"`
		} `json:"quality"`
	} `json:"quality"`
	Tags    []int `json:"tags"`
	HasFile bool  `json:"hasFile"`
	Images  []struct {
		CoverType string `json:"coverType"`
		URL       string `json:"url"`
	} `json:"images"`
}

// radarrTag represents a tag from Radarr API.
type radarrTag struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
}

// radarrQueueItem represents an item in the download queue.
type radarrQueueItem struct {
	ID                      int       `json:"id"`
	MovieID                 int       `json:"movieId"`
	Title                   string    `json:"title"`
	Size                    int64     `json:"size"`
	Sizeleft                int64     `json:"sizeleft"`
	Status                  string    `json:"status"`
	TrackedDownloadStatus   string    `json:"trackedDownloadStatus"`
	TrackedDownloadState    string    `json:"trackedDownloadState"`
	StatusMessages          []string  `json:"statusMessages"`
	DownloadID              string    `json:"downloadId"`
	Protocol                string    `json:"protocol"`
	DownloadClient          string    `json:"downloadClient"`
	Indexer                 string    `json:"indexer"`
	OutputPath              string    `json:"outputPath"`
	TimedOut                bool      `json:"timedOut"`
	EstimatedCompletionTime time.Time `json:"estimatedCompletionTime"`
}

// radarrQueueResponse represents the queue response.
type radarrQueueResponse struct {
	Page         int               `json:"page"`
	PageSize     int               `json:"pageSize"`
	TotalRecords int               `json:"totalRecords"`
	Records      []radarrQueueItem `json:"records"`
}

// RadarrQueueItem represents a processed queue item for public API.
type RadarrQueueItem struct {
	ID                  int       `json:"id"`
	MovieID             int       `json:"movie_id"`
	Title               string    `json:"title"`
	Size                int64     `json:"size"`
	SizeLeft            int64     `json:"size_left"`
	Progress            float64   `json:"progress"`
	Status              string    `json:"status"`
	DownloadStatus      string    `json:"download_status"`
	DownloadState       string    `json:"download_state"`
	Protocol            string    `json:"protocol"`
	DownloadClient      string    `json:"download_client"`
	Indexer             string    `json:"indexer"`
	TimedOut            bool      `json:"timed_out"`
	EstimatedCompletion time.Time `json:"estimated_completion"`
}

// radarrHistoryItem represents a history entry.
type radarrHistoryItem struct {
	ID          int    `json:"id"`
	MovieID     int    `json:"movieId"`
	SourceTitle string `json:"sourceTitle"`
	Quality     struct {
		Quality struct {
			Name string `json:"name"`
		} `json:"quality"`
	} `json:"quality"`
	Date       time.Time `json:"date"`
	EventType  string    `json:"eventType"`
	DownloadID string    `json:"downloadId"`
}

// radarrHistoryResponse represents the history response.
type radarrHistoryResponse struct {
	Page         int                 `json:"page"`
	PageSize     int                 `json:"pageSize"`
	TotalRecords int                 `json:"totalRecords"`
	Records      []radarrHistoryItem `json:"records"`
}

// RadarrHistoryItem represents a processed history item for public API.
type RadarrHistoryItem struct {
	ID          int       `json:"id"`
	MovieID     int       `json:"movie_id"`
	SourceTitle string    `json:"source_title"`
	Quality     string    `json:"quality"`
	Date        time.Time `json:"date"`
	EventType   string    `json:"event_type"`
	DownloadID  string    `json:"download_id"`
}

// radarrCalendarItem represents a calendar entry.
type radarrCalendarItem struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	InCinemas       time.Time `json:"inCinemas"`
	PhysicalRelease time.Time `json:"physicalRelease"`
	DigitalRelease  time.Time `json:"digitalRelease"`
	Year            int       `json:"year"`
	HasFile         bool      `json:"hasFile"`
	Monitored       bool      `json:"monitored"`
}

// RadarrCalendarItem represents a processed calendar item for public API.
type RadarrCalendarItem struct {
	ID              int       `json:"id"`
	Title           string    `json:"title"`
	InCinemas       time.Time `json:"in_cinemas"`
	PhysicalRelease time.Time `json:"physical_release"`
	DigitalRelease  time.Time `json:"digital_release"`
	Year            int       `json:"year"`
	HasFile         bool      `json:"has_file"`
	Monitored       bool      `json:"monitored"`
}

// radarrQualityProfile represents a quality profile.
type radarrQualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RadarrQualityProfile represents a quality profile for public API.
type RadarrQualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TestConnection verifies the connection to Radarr.
func (c *RadarrClient) TestConnection(ctx context.Context) error {
	return c.callWithRetry(ctx, func() error {
		var status radarrSystemStatus
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&status).
			Get("/api/v3/system/status")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Radarr connection successful",
			zap.String("version", status.Version),
			zap.String("url", c.baseURL),
		)

		return nil
	})
}

// GetSystemInfo retrieves complete system information from Radarr.
func (c *RadarrClient) GetSystemInfo(ctx context.Context) (*RadarrSystemInfo, error) {
	var status radarrSystemStatus

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&status).
			Get("/api/v3/system/status")

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

	// Construir el nombre del OS
	osName := status.OsName
	if status.IsLinux {
		osName = "Linux"
	} else if status.IsWindows {
		osName = "Windows"
	} else if status.IsOsx {
		osName = "macOS"
	}

	// Convertir a nuestro modelo
	info := &RadarrSystemInfo{
		Version:        status.Version,
		BuildTime:      status.BuildTime,
		Branch:         status.Branch,
		OS:             osName,
		OSVersion:      status.OsVersion,
		Runtime:        status.RuntimeName,
		RuntimeVersion: status.RuntimeVersion,
		IsDebug:        status.IsDebug,
		IsProduction:   status.IsProduction,
		Authentication: status.Authentication,
		URLBase:        status.UrlBase,
		StartupPath:    status.StartupPath,
		AppData:        status.AppData,
		SqliteVersion:  status.SqliteVersion,
	}

	c.logger.Info("Retrieved Radarr system info",
		zap.String("version", info.Version),
		zap.String("os", info.OS),
		zap.String("runtime", info.Runtime),
	)

	return info, nil
}

// GetLibrary retrieves all movies from Radarr.
func (c *RadarrClient) GetLibrary(ctx context.Context) ([]*models.Media, error) {
	var radarrMovies []radarrMovie

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&radarrMovies).
			Get("/api/v3/movie")

		if err != nil {
			return fmt.Errorf("failed to get library: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert Radarr movies to internal Media model
	mediaList := make([]*models.Media, 0, len(radarrMovies))
	for _, movie := range radarrMovies {
		if !movie.HasFile {
			continue // Skip movies without files
		}

		media := c.convertToMedia(&movie)
		mediaList = append(mediaList, media)
	}

	c.logger.Info("Retrieved Radarr library",
		zap.Int("total_movies", len(radarrMovies)),
		zap.Int("with_files", len(mediaList)),
	)

	return mediaList, nil
}

// GetItem retrieves a specific movie from Radarr.
func (c *RadarrClient) GetItem(ctx context.Context, id int) (*models.Media, error) {
	var movie radarrMovie

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&movie).
			SetPathParam("id", fmt.Sprintf("%d", id)).
			Get("/api/v3/movie/{id}")

		if err != nil {
			return fmt.Errorf("failed to get movie: %w", err)
		}

		if resp.StatusCode() == 404 {
			return fmt.Errorf("movie not found: %d", id)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return c.convertToMedia(&movie), nil
}

// DeleteItem removes a movie from Radarr.
func (c *RadarrClient) DeleteItem(ctx context.Context, id int, deleteFiles bool) error {
	return c.callWithRetry(ctx, func() error {
		deleteFilesStr := "false"
		if deleteFiles {
			deleteFilesStr = "true"
		}

		resp, err := c.client.R().
			SetContext(ctx).
			SetPathParam("id", fmt.Sprintf("%d", id)).
			SetQueryParam("deleteFiles", deleteFilesStr).
			SetQueryParam("addImportExclusion", "false").
			Delete("/api/v3/movie/{id}")

		if err != nil {
			return fmt.Errorf("failed to delete movie: %w", err)
		}

		if resp.StatusCode() != 200 && resp.StatusCode() != 204 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Deleted movie from Radarr",
			zap.Int("movie_id", id),
			zap.Bool("deleted_files", deleteFiles),
		)

		return nil
	})
}

// GetTags retrieves all tags from Radarr.
func (c *RadarrClient) GetTags(ctx context.Context) ([]models.Tag, error) {
	var radarrTags []radarrTag

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&radarrTags).
			Get("/api/v3/tag")

		if err != nil {
			return fmt.Errorf("failed to get tags: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to internal Tag model
	tags := make([]models.Tag, 0, len(radarrTags))
	for _, tag := range radarrTags {
		tags = append(tags, models.Tag{
			ServiceType: "radarr",
			ServiceID:   tag.ID,
			Label:       tag.Label,
		})
	}

	return tags, nil
}

// convertToMedia converts a Radarr movie to internal Media model.
func (c *RadarrClient) convertToMedia(movie *radarrMovie) *models.Media {
	media := &models.Media{
		Title:     movie.Title,
		Type:      "movie",
		FilePath:  movie.Path,
		Size:      movie.SizeOnDisk,
		AddedDate: movie.Added,
		RadarrID:  &movie.ID,
		Quality:   movie.Quality.Quality.Name,
	}

	// Extract poster URL
	for _, image := range movie.Images {
		if image.CoverType == "poster" {
			media.PosterURL = image.URL
			break
		}
	}

	// Convert tag IDs to strings (we'll need to map these to tag labels later)
	if len(movie.Tags) > 0 {
		media.Tags = make([]string, 0, len(movie.Tags))
		for _, tagID := range movie.Tags {
			media.Tags = append(media.Tags, fmt.Sprintf("radarr_tag_%d", tagID))
		}
	}

	return media
}

// GetQueue retrieves the current download queue from Radarr.
func (c *RadarrClient) GetQueue(ctx context.Context) ([]RadarrQueueItem, error) {
	var queueResp radarrQueueResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&queueResp).
			SetQueryParam("pageSize", "100").
			SetQueryParam("includeUnknownMovieItems", "false").
			Get("/api/v3/queue")

		if err != nil {
			return fmt.Errorf("failed to get queue: %w", err)
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
	items := make([]RadarrQueueItem, 0, len(queueResp.Records))
	for _, record := range queueResp.Records {
		progress := 0.0
		if record.Size > 0 {
			progress = float64(record.Size-record.Sizeleft) / float64(record.Size) * 100
			// Ensure progress is within 0-100 range (handle inconsistent data from Radarr)
			if progress < 0 {
				progress = 0
			} else if progress > 100 {
				progress = 100
			}
		}

		items = append(items, RadarrQueueItem{
			ID:                  record.ID,
			MovieID:             record.MovieID,
			Title:               record.Title,
			Size:                record.Size,
			SizeLeft:            record.Sizeleft,
			Progress:            progress,
			Status:              record.Status,
			DownloadStatus:      record.TrackedDownloadStatus,
			DownloadState:       record.TrackedDownloadState,
			Protocol:            record.Protocol,
			DownloadClient:      record.DownloadClient,
			Indexer:             record.Indexer,
			TimedOut:            record.TimedOut,
			EstimatedCompletion: record.EstimatedCompletionTime,
		})
	}

	c.logger.Info("Retrieved Radarr queue",
		zap.Int("total_items", len(items)),
	)

	return items, nil
}

// GetHistory retrieves recent history from Radarr.
func (c *RadarrClient) GetHistory(ctx context.Context, pageSize int) ([]RadarrHistoryItem, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var historyResp radarrHistoryResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&historyResp).
			SetQueryParam("pageSize", fmt.Sprintf("%d", pageSize)).
			SetQueryParam("sortKey", "date").
			SetQueryParam("sortDirection", "descending").
			Get("/api/v3/history")

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
	items := make([]RadarrHistoryItem, 0, len(historyResp.Records))
	for _, record := range historyResp.Records {
		items = append(items, RadarrHistoryItem{
			ID:          record.ID,
			MovieID:     record.MovieID,
			SourceTitle: record.SourceTitle,
			Quality:     record.Quality.Quality.Name,
			Date:        record.Date,
			EventType:   record.EventType,
			DownloadID:  record.DownloadID,
		})
	}

	c.logger.Info("Retrieved Radarr history",
		zap.Int("total_items", len(items)),
	)

	return items, nil
}

// GetCalendar retrieves upcoming movies from Radarr.
func (c *RadarrClient) GetCalendar(ctx context.Context, startDate, endDate time.Time) ([]RadarrCalendarItem, error) {
	var calendarItems []radarrCalendarItem

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&calendarItems).
			SetQueryParam("start", startDate.Format("2006-01-02")).
			SetQueryParam("end", endDate.Format("2006-01-02")).
			Get("/api/v3/calendar")

		if err != nil {
			return fmt.Errorf("failed to get calendar: %w", err)
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
	items := make([]RadarrCalendarItem, 0, len(calendarItems))
	for _, record := range calendarItems {
		items = append(items, RadarrCalendarItem{
			ID:              record.ID,
			Title:           record.Title,
			InCinemas:       record.InCinemas,
			PhysicalRelease: record.PhysicalRelease,
			DigitalRelease:  record.DigitalRelease,
			Year:            record.Year,
			HasFile:         record.HasFile,
			Monitored:       record.Monitored,
		})
	}

	c.logger.Info("Retrieved Radarr calendar",
		zap.Int("total_items", len(items)),
		zap.String("start_date", startDate.Format("2006-01-02")),
		zap.String("end_date", endDate.Format("2006-01-02")),
	)

	return items, nil
}

// GetQualityProfiles retrieves available quality profiles from Radarr.
func (c *RadarrClient) GetQualityProfiles(ctx context.Context) ([]RadarrQualityProfile, error) {
	var profiles []radarrQualityProfile

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&profiles).
			Get("/api/v3/qualityprofile")

		if err != nil {
			return fmt.Errorf("failed to get quality profiles: %w", err)
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
	items := make([]RadarrQualityProfile, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, RadarrQualityProfile{
			ID:   profile.ID,
			Name: profile.Name,
		})
	}

	c.logger.Info("Retrieved Radarr quality profiles",
		zap.Int("total_profiles", len(items)),
	)

	return items, nil
}

// callWithRetry executes a function with retry logic.
func (c *RadarrClient) callWithRetry(ctx context.Context, fn func() error) error {
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
			c.logger.Warn("Radarr API call failed, retrying",
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
