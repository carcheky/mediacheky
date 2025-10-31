package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/carcheky/keepercheky/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// JellyfinClient implements the StreamingClient interface for Jellyfin.
type JellyfinClient struct {
	client  *resty.Client
	baseURL string
	apiKey  string
	logger  *zap.Logger
}

// NewJellyfinClient creates a new Jellyfin client.
func NewJellyfinClient(config ClientConfig, logger *zap.Logger) *JellyfinClient {
	client := resty.New()
	client.SetBaseURL(config.BaseURL)
	client.SetHeader("X-Emby-Token", config.APIKey) // Jellyfin uses X-Emby-Token header

	// Disable HTTP caching to always get fresh data
	client.SetHeader("Cache-Control", "no-cache, no-store, must-revalidate")
	client.SetHeader("Pragma", "no-cache")
	client.SetHeader("Expires", "0")

	client.SetTimeout(config.Timeout)

	if config.Timeout == 0 {
		client.SetTimeout(DefaultTimeout)
	}

	return &JellyfinClient{
		client:  client,
		baseURL: config.BaseURL,
		apiKey:  config.APIKey,
		logger:  logger,
	}
}

// jellyfinSystemInfo represents the system info response from Jellyfin API.
type jellyfinSystemInfo struct {
	Version                    string `json:"Version"`
	ID                         string `json:"Id"`
	OperatingSystem            string `json:"OperatingSystem"`
	OperatingSystemDisplayName string `json:"OperatingSystemDisplayName"`
	ServerName                 string `json:"ServerName"`
	LocalAddress               string `json:"LocalAddress"`
	ProductName                string `json:"ProductName"`
}

// JellyfinSystemInfo representa toda la información del sistema de Jellyfin
type JellyfinSystemInfo struct {
	Version       string `json:"version"`
	ServerID      string `json:"server_id"`
	ServerName    string `json:"server_name"`
	ProductName   string `json:"product_name"`
	OS            string `json:"os"`
	OSDisplayName string `json:"os_display_name"`
	LocalAddress  string `json:"local_address"`
}

// jellyfinItem represents a media item from Jellyfin API.
type jellyfinItem struct {
	ID          string    `json:"Id"`
	Name        string    `json:"Name"`
	Type        string    `json:"Type"` // "Movie", "Series", "Episode", etc.
	Path        string    `json:"Path"`
	DateCreated time.Time `json:"DateCreated"`
	Overview    string    `json:"Overview"`
	UserData    struct {
		PlayCount         int       `json:"PlayCount"`
		IsFavorite        bool      `json:"IsFavorite"`
		LastPlayedDate    time.Time `json:"LastPlayedDate"`
		PlaybackPosition  int64     `json:"PlaybackPositionTicks"`
		UnplayedItemCount int       `json:"UnplayedItemCount"` // Episodios no vistos
	} `json:"UserData"`
	MediaSources []struct {
		Size int64 `json:"Size"`
	} `json:"MediaSources"`
	ImageTags map[string]string `json:"ImageTags"`
	// Series-specific fields
	ChildCount         int `json:"ChildCount"`         // Número de temporadas (no episodios)
	RecursiveItemCount int `json:"RecursiveItemCount"` // Total de episodios incluyendo sub-items
}

// jellyfinItemsResponse represents the response from items endpoint.
type jellyfinItemsResponse struct {
	Items      []jellyfinItem `json:"Items"`
	TotalCount int            `json:"TotalRecordCount"`
}

// TestConnection verifies the connection to Jellyfin.
func (c *JellyfinClient) TestConnection(ctx context.Context) error {
	return c.callWithRetry(ctx, func() error {
		var info jellyfinSystemInfo
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&info).
			Get("/System/Info/Public")

		if err != nil {
			return fmt.Errorf("connection failed: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Jellyfin connection successful",
			zap.String("version", info.Version),
			zap.String("server_id", info.ID),
			zap.String("url", c.baseURL),
		)

		return nil
	})
}

// GetSystemInfo retrieves complete system information from Jellyfin.
func (c *JellyfinClient) GetSystemInfo(ctx context.Context) (*JellyfinSystemInfo, error) {
	var info jellyfinSystemInfo

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&info).
			Get("/System/Info/Public")

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
	systemInfo := &JellyfinSystemInfo{
		Version:       info.Version,
		ServerID:      info.ID,
		ServerName:    info.ServerName,
		ProductName:   info.ProductName,
		OS:            info.OperatingSystem,
		OSDisplayName: info.OperatingSystemDisplayName,
		LocalAddress:  info.LocalAddress,
	}

	c.logger.Info("Retrieved Jellyfin system info",
		zap.String("version", systemInfo.Version),
		zap.String("server_name", systemInfo.ServerName),
		zap.String("os", systemInfo.OS),
	)

	return systemInfo, nil
}

// GetLibrary retrieves all media items from Jellyfin with pagination for better performance.
func (c *JellyfinClient) GetLibrary(ctx context.Context) ([]*models.Media, error) {
	const pageSize = 500 // Procesar en lotes de 500 items

	var allMedia []*models.Media
	startIndex := 0
	totalCount := 0

	c.logger.Info("Starting Jellyfin library retrieval with pagination",
		zap.Int("page_size", pageSize),
	)

	for {
		var response jellyfinItemsResponse

		err := c.callWithRetry(ctx, func() error {
			resp, err := c.client.R().
				SetContext(ctx).
				SetResult(&response).
				SetQueryParams(map[string]string{
					"IncludeItemTypes": "Movie,Series,Season,Episode", // Incluir Episodes y Seasons
					"Recursive":        "true",
					"Fields":           "Path,DateCreated,MediaSources,UserData,ChildCount,RecursiveItemCount,ParentId,SeriesId,SeriesName,SeasonId,SeasonName,IndexNumber,ParentIndexNumber", // Más campos para episodes
					"StartIndex":       fmt.Sprintf("%d", startIndex),
					"Limit":            fmt.Sprintf("%d", pageSize),
					"EnableImages":     "false", // Acelerar respuesta
				}).
				Get("/Items")

			if err != nil {
				return fmt.Errorf("failed to get library page: %w", err)
			}

			if resp.StatusCode() != 200 {
				return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
			}

			return nil
		})

		if err != nil {
			return nil, err
		}

		// Guardar total en primera iteración
		if startIndex == 0 {
			totalCount = response.TotalCount
			allMedia = make([]*models.Media, 0, totalCount)
			c.logger.Info("Jellyfin library total items",
				zap.Int("total", totalCount),
			)
		}

		// Convertir items de esta página
		for i := range response.Items {
			item := &response.Items[i]

			// Log raw data for series to debug episode count
			if item.Type == "Series" {
				c.logger.Info("Jellyfin raw series data",
					zap.String("name", item.Name),
					zap.String("type", item.Type),
					zap.Int("child_count", item.ChildCount),
					zap.Int("recursive_item_count", item.RecursiveItemCount),
				)
			}

			media := c.convertToMedia(item)
			allMedia = append(allMedia, media)
		}

		c.logger.Info("Retrieved Jellyfin page",
			zap.Int("start_index", startIndex),
			zap.Int("items_in_page", len(response.Items)),
			zap.Int("total_retrieved", len(allMedia)),
			zap.Int("total_count", totalCount),
		)

		// Salir si no hay más items
		if len(response.Items) < pageSize || len(allMedia) >= totalCount {
			break
		}

		startIndex += pageSize
	}

	c.logger.Info("Jellyfin library retrieval complete",
		zap.Int("total_items", len(allMedia)),
	)

	return allMedia, nil
}

// GetPlaybackInfo retrieves playback information for media.
func (c *JellyfinClient) GetPlaybackInfo(ctx context.Context, mediaID string) (*models.PlaybackInfo, error) {
	var item jellyfinItem

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&item).
			SetPathParam("id", mediaID).
			SetQueryParam("Fields", "UserData").
			Get("/Items/{id}")

		if err != nil {
			return fmt.Errorf("failed to get playback info: %w", err)
		}

		if resp.StatusCode() == 404 {
			return fmt.Errorf("item not found: %s", mediaID)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	playbackInfo := &models.PlaybackInfo{
		MediaID:     item.ID,
		LastPlayed:  item.UserData.LastPlayedDate,
		PlayCount:   item.UserData.PlayCount,
		IsFavorite:  item.UserData.IsFavorite,
		PlaybackPos: item.UserData.PlaybackPosition,
	}

	return playbackInfo, nil
}

// DeleteItem removes a media item from Jellyfin.
func (c *JellyfinClient) DeleteItem(ctx context.Context, id string) error {
	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetPathParam("id", id).
			Delete("/Items/{id}")

		if err != nil {
			return fmt.Errorf("failed to delete item: %w", err)
		}

		if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Deleted item from Jellyfin",
			zap.String("item_id", id),
		)

		return nil
	})

	if err != nil {
		return err
	}

	// Invalidate Jellyfin cache after deletion
	if err := c.InvalidateCache(ctx); err != nil {
		c.logger.Warn("Failed to invalidate Jellyfin cache after deletion",
			zap.String("item_id", id),
			zap.Error(err),
		)
		// Don't fail the deletion if cache invalidation fails
	}

	return nil
}

// InvalidateCache forces Jellyfin to refresh its library cache.
// This ensures that deleted items don't appear in subsequent library scans.
func (c *JellyfinClient) InvalidateCache(ctx context.Context) error {
	c.logger.Info("Invalidating Jellyfin cache")

	return c.callWithRetry(ctx, func() error {
		// POST to Library/Refresh to force a library scan
		resp, err := c.client.R().
			SetContext(ctx).
			Post("/Library/Refresh")

		if err != nil {
			return fmt.Errorf("failed to refresh library: %w", err)
		}

		if resp.StatusCode() != 204 && resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		c.logger.Info("Jellyfin cache invalidated successfully")
		return nil
	})
}

// convertToMedia converts a Jellyfin item to internal Media model.
func (c *JellyfinClient) convertToMedia(item *jellyfinItem) *models.Media {
	mediaType := "movie"
	if item.Type == "Series" {
		mediaType = "series"
	}

	var size int64
	if len(item.MediaSources) > 0 {
		size = item.MediaSources[0].Size
	}

	media := &models.Media{
		Title:      item.Name,
		Type:       mediaType,
		FilePath:   item.Path,
		Size:       size,
		AddedDate:  item.DateCreated,
		JellyfinID: &item.ID,
	}

	// For series, set episode count from RecursiveItemCount
	if mediaType == "series" {
		media.EpisodeFileCount = item.RecursiveItemCount
		// RecursiveItemCount includes all child items (episodes) in Jellyfin
	}

	// Set last watched if available
	if !item.UserData.LastPlayedDate.IsZero() {
		media.LastWatched = &item.UserData.LastPlayedDate
	}

	// Extract poster URL if available
	if primaryTag, ok := item.ImageTags["Primary"]; ok {
		media.PosterURL = fmt.Sprintf("%s/Items/%s/Images/Primary?tag=%s",
			c.baseURL, item.ID, primaryTag)
	}

	// Add favorite tag
	if item.UserData.IsFavorite {
		media.Tags = []string{"jellyfin_favorite"}
	}

	return media
}

// callWithRetry executes a function with retry logic.
func (c *JellyfinClient) callWithRetry(ctx context.Context, fn func() error) error {
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
			c.logger.Warn("Jellyfin API call failed, retrying",
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

// VirtualFolder represents a Jellyfin library virtual folder.
type VirtualFolder struct {
	Name      string   `json:"Name"`
	Locations []string `json:"Locations"`
	ItemID    string   `json:"ItemId"`
}

// GetVirtualFolders retrieves library virtual folders (root paths) from Jellyfin.
func (c *JellyfinClient) GetVirtualFolders(ctx context.Context) ([]VirtualFolder, error) {
	var folders []VirtualFolder

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&folders).
			Get("/Library/VirtualFolders")

		if err != nil {
			return fmt.Errorf("failed to get virtual folders: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved Jellyfin virtual folders",
		zap.Int("folder_count", len(folders)),
	)

	return folders, nil
}

// SessionInfo represents an active Jellyfin session.
type SessionInfo struct {
	ID                    string    `json:"Id"`
	UserID                string    `json:"UserId"`
	UserName              string    `json:"UserName"`
	Client                string    `json:"Client"`
	DeviceName            string    `json:"DeviceName"`
	DeviceID              string    `json:"DeviceId"`
	ApplicationVersion    string    `json:"ApplicationVersion"`
	RemoteEndPoint        string    `json:"RemoteEndPoint"`
	LastActivityDate      time.Time `json:"LastActivityDate"`
	SupportsRemoteControl bool      `json:"SupportsRemoteControl"`
	NowPlayingItem        *struct {
		ID        string `json:"Id"`
		Name      string `json:"Name"`
		Type      string `json:"Type"`
		MediaType string `json:"MediaType"`
	} `json:"NowPlayingItem,omitempty"`
	PlayState *struct {
		PositionTicks       int64  `json:"PositionTicks"`
		CanSeek             bool   `json:"CanSeek"`
		IsPaused            bool   `json:"IsPaused"`
		IsMuted             bool   `json:"IsMuted"`
		VolumeLevel         int    `json:"VolumeLevel"`
		AudioStreamIndex    int    `json:"AudioStreamIndex"`
		SubtitleStreamIndex int    `json:"SubtitleStreamIndex"`
		MediaSourceID       string `json:"MediaSourceId"`
		PlayMethod          string `json:"PlayMethod"` // DirectPlay, DirectStream, Transcode
	} `json:"PlayState,omitempty"`
	TranscodingInfo *struct {
		VideoCodec           string   `json:"VideoCodec"`
		AudioCodec           string   `json:"AudioCodec"`
		Container            string   `json:"Container"`
		IsVideoDirect        bool     `json:"IsVideoDirect"`
		IsAudioDirect        bool     `json:"IsAudioDirect"`
		Bitrate              int      `json:"Bitrate"`
		Framerate            float64  `json:"Framerate"`
		CompletionPercentage float64  `json:"CompletionPercentage"`
		Width                int      `json:"Width"`
		Height               int      `json:"Height"`
		AudioChannels        int      `json:"AudioChannels"`
		TranscodeReasons     []string `json:"TranscodeReasons"`
	} `json:"TranscodingInfo,omitempty"`
}

// GetActiveSessions retrieves all active sessions from Jellyfin.
func (c *JellyfinClient) GetActiveSessions(ctx context.Context) ([]SessionInfo, error) {
	var sessions []SessionInfo

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&sessions).
			Get("/Sessions")

		if err != nil {
			return fmt.Errorf("failed to get active sessions: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	c.logger.Info("Retrieved Jellyfin active sessions",
		zap.Int("session_count", len(sessions)),
	)

	return sessions, nil
}

// LibraryStats represents statistics for a Jellyfin library.
type LibraryStats struct {
	TotalItems     int             `json:"total_items"`
	MovieCount     int             `json:"movie_count"`
	SeriesCount    int             `json:"series_count"`
	EpisodeCount   int             `json:"episode_count"`
	AlbumCount     int             `json:"album_count"`
	SongCount      int             `json:"song_count"`
	TotalSize      int64           `json:"total_size"`
	LibraryFolders []VirtualFolder `json:"library_folders,omitempty"`
}

// GetLibraryStats retrieves detailed statistics about the Jellyfin library.
func (c *JellyfinClient) GetLibraryStats(ctx context.Context) (*LibraryStats, error) {
	stats := &LibraryStats{}

	// Get all items to calculate stats
	var response jellyfinItemsResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParams(map[string]string{
				"Recursive": "true",
				"Fields":    "MediaSources",
			}).
			Get("/Items")

		if err != nil {
			return fmt.Errorf("failed to get library items: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	stats.TotalItems = response.TotalCount

	// Calculate type-specific counts and total size
	var totalSize int64
	for _, item := range response.Items {
		switch item.Type {
		case "Movie":
			stats.MovieCount++
		case "Series":
			stats.SeriesCount++
		case "Episode":
			stats.EpisodeCount++
		case "MusicAlbum":
			stats.AlbumCount++
		case "Audio":
			stats.SongCount++
		}

		// Sum file sizes
		if len(item.MediaSources) > 0 {
			totalSize += item.MediaSources[0].Size
		}
	}
	stats.TotalSize = totalSize

	// Get virtual folders
	folders, err := c.GetVirtualFolders(ctx)
	if err == nil {
		stats.LibraryFolders = folders
	}

	c.logger.Info("Retrieved Jellyfin library statistics",
		zap.Int("total_items", stats.TotalItems),
		zap.Int("movies", stats.MovieCount),
		zap.Int("series", stats.SeriesCount),
		zap.Int64("total_size_bytes", stats.TotalSize),
	)

	return stats, nil
}

// RecentlyAddedItem represents a recently added media item.
type RecentlyAddedItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	DateCreated time.Time `json:"date_created"`
	PosterURL   string    `json:"poster_url,omitempty"`
	Overview    string    `json:"overview,omitempty"`
}

// GetRecentlyAdded retrieves recently added items from Jellyfin.
func (c *JellyfinClient) GetRecentlyAdded(ctx context.Context, limit int) ([]RecentlyAddedItem, error) {
	if limit <= 0 {
		limit = DefaultRecentlyAddedLimit
	}

	var response jellyfinItemsResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParams(map[string]string{
				"SortBy":           "DateCreated",
				"SortOrder":        "Descending",
				"Recursive":        "true",
				"IncludeItemTypes": "Movie,Series",
				"Limit":            fmt.Sprintf("%d", limit),
				"Fields":           "Overview,DateCreated",
			}).
			Get("/Items")

		if err != nil {
			return fmt.Errorf("failed to get recently added items: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	items := make([]RecentlyAddedItem, 0, len(response.Items))
	for _, item := range response.Items {
		recentItem := RecentlyAddedItem{
			ID:          item.ID,
			Name:        item.Name,
			Type:        item.Type,
			DateCreated: item.DateCreated,
			Overview:    item.Overview,
		}

		// Extract poster URL if available
		if primaryTag, ok := item.ImageTags["Primary"]; ok {
			recentItem.PosterURL = fmt.Sprintf("%s/Items/%s/Images/Primary?tag=%s",
				c.baseURL, item.ID, primaryTag)
		}

		items = append(items, recentItem)
	}

	c.logger.Info("Retrieved recently added items",
		zap.Int("count", len(items)),
	)

	return items, nil
}

// ActivityLogEntry represents an entry in the Jellyfin activity log.
type ActivityLogEntry struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	UserID        string    `json:"user_id"`
	Date          time.Time `json:"date"`
	Severity      string    `json:"severity"`
	ShortOverview string    `json:"short_overview,omitempty"`
}

// jellyfinActivityResponse represents the response from the activity log endpoint.
type jellyfinActivityResponse struct {
	Items []struct {
		ID            int64     `json:"Id"`
		Name          string    `json:"Name"`
		Type          string    `json:"Type"`
		UserID        string    `json:"UserId"`
		Date          time.Time `json:"Date"`
		Severity      string    `json:"Severity"`
		ShortOverview string    `json:"ShortOverview"`
	} `json:"Items"`
	TotalRecordCount int `json:"TotalRecordCount"`
}

// GetActivityLog retrieves the activity log from Jellyfin.
func (c *JellyfinClient) GetActivityLog(ctx context.Context, limit int) ([]ActivityLogEntry, error) {
	if limit <= 0 {
		limit = DefaultActivityLogLimit
	}

	var response jellyfinActivityResponse

	err := c.callWithRetry(ctx, func() error {
		resp, err := c.client.R().
			SetContext(ctx).
			SetResult(&response).
			SetQueryParams(map[string]string{
				"StartIndex": "0",
				"Limit":      fmt.Sprintf("%d", limit),
			}).
			Get("/System/ActivityLog/Entries")

		if err != nil {
			return fmt.Errorf("failed to get activity log: %w", err)
		}

		if resp.StatusCode() != 200 {
			return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	entries := make([]ActivityLogEntry, 0, len(response.Items))
	for _, item := range response.Items {
		entries = append(entries, ActivityLogEntry{
			ID:            item.ID,
			Name:          item.Name,
			Type:          item.Type,
			UserID:        item.UserID,
			Date:          item.Date,
			Severity:      item.Severity,
			ShortOverview: item.ShortOverview,
		})
	}

	c.logger.Info("Retrieved activity log entries",
		zap.Int("count", len(entries)),
	)

	return entries, nil
}
