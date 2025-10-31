package clients

import (
	"context"
	"fmt"
	"time"

	"github.com/autobrr/go-qbittorrent"
	"github.com/rs/zerolog"
)

// QBittorrentClient handles interactions with qBittorrent API for seeding validation
type QBittorrentClient struct {
	client   *qbittorrent.Client
	baseURL  string
	username string
	password string
	logger   zerolog.Logger
}

// TorrentClient interface for torrent client operations
type TorrentClient interface {
	TestConnection(ctx context.Context) error
	GetTorrentByPath(ctx context.Context, filePath string) (*TorrentInfo, error)
	IsSeeding(ctx context.Context, filePath string) (bool, error)
}

// TorrentInfo contains torrent information
type TorrentInfo struct {
	Hash       string
	Name       string
	State      string
	Progress   float64
	Ratio      float64
	Size       int64
	UpSpeed    int64
	DlSpeed    int64
	SeedTime   int64
	Category   string
	Tags       string
	SavePath   string
	IsSeeding  bool
	IsComplete bool
}

// NewQBittorrentClient creates a new qBittorrent client
func NewQBittorrentClient(baseURL, username, password string, logger zerolog.Logger) *QBittorrentClient {
	client := qbittorrent.NewClient(qbittorrent.Config{
		Host:          baseURL,
		Username:      username,
		Password:      password,
		TLSSkipVerify: true,
		Timeout:       30,
		RetryAttempts: 3,
		RetryDelay:    1,
	})

	return &QBittorrentClient{
		client:   client,
		baseURL:  baseURL,
		username: username,
		password: password,
		logger:   logger.With().Str("client", "qbittorrent").Logger(),
	}
}

// TestConnection verifies connection to qBittorrent
func (c *QBittorrentClient) TestConnection(ctx context.Context) error {
	c.logger.Debug().Msg("testing qBittorrent connection")

	if err := c.client.LoginCtx(ctx); err != nil {
		c.logger.Error().Err(err).Msg("failed to authenticate with qBittorrent")
		return fmt.Errorf("qBittorrent authentication failed: %w", err)
	}

	buildInfo, err := c.client.GetBuildInfoCtx(ctx)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get build info")
		return fmt.Errorf("failed to get build info: %w", err)
	}

	c.logger.Info().
		Str("version", buildInfo.Qt).
		Str("libtorrent", buildInfo.Libtorrent).
		Str("platform", buildInfo.Platform).
		Msg("qBittorrent connection successful")

	return nil
}

// GetTorrentByPath finds a torrent by its file path
func (c *QBittorrentClient) GetTorrentByPath(ctx context.Context, filePath string) (*TorrentInfo, error) {
	c.logger.Debug().Str("path", filePath).Msg("searching torrent by path")

	// Login first
	if err := c.client.LoginCtx(ctx); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get all torrents
	torrents, err := c.client.GetTorrentsCtx(ctx, qbittorrent.TorrentFilterOptions{})
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to get torrents")
		return nil, fmt.Errorf("failed to get torrents: %w", err)
	}

	// Search for torrent containing the file path
	for _, t := range torrents {
		// Get torrent properties to check save path
		props, err := c.client.GetTorrentPropertiesCtx(ctx, t.Hash)
		if err != nil {
			c.logger.Warn().Err(err).Str("hash", t.Hash).Msg("failed to get torrent properties")
			continue
		}

		// Check if the file path matches this torrent's save path
		if props.SavePath == filePath || containsPath(props.SavePath, filePath) {
			info := &TorrentInfo{
				Hash:       t.Hash,
				Name:       t.Name,
				State:      t.State,
				Progress:   t.Progress,
				Ratio:      t.Ratio,
				Size:       t.Size,
				UpSpeed:    t.UpSpeed,
				DlSpeed:    t.DlSpeed,
				SeedTime:   t.SeedingTime,
				Category:   t.Category,
				Tags:       t.Tags,
				SavePath:   props.SavePath,
				IsSeeding:  isStateSeeding(t.State),
				IsComplete: t.Progress >= 1.0,
			}

			c.logger.Info().
				Str("hash", info.Hash).
				Str("name", info.Name).
				Str("state", info.State).
				Float64("progress", info.Progress*100).
				Float64("ratio", info.Ratio).
				Bool("is_seeding", info.IsSeeding).
				Msg("found torrent")

			return info, nil
		}
	}

	c.logger.Debug().Str("path", filePath).Msg("no torrent found for path")
	return nil, fmt.Errorf("no torrent found for path: %s", filePath)
}

// IsSeeding checks if a file is currently seeding
func (c *QBittorrentClient) IsSeeding(ctx context.Context, filePath string) (bool, error) {
	c.logger.Debug().Str("path", filePath).Msg("checking if file is seeding")

	torrent, err := c.GetTorrentByPath(ctx, filePath)
	if err != nil {
		// If torrent not found, it's not seeding
		if err.Error() == fmt.Sprintf("no torrent found for path: %s", filePath) {
			c.logger.Debug().Str("path", filePath).Msg("file not found in torrents, not seeding")
			return false, nil
		}
		return false, err
	}

	isSeeding := torrent.IsSeeding
	c.logger.Info().
		Str("path", filePath).
		Str("hash", torrent.Hash).
		Bool("is_seeding", isSeeding).
		Float64("ratio", torrent.Ratio).
		Int64("seed_time", torrent.SeedTime).
		Msg("seeding status checked")

	return isSeeding, nil
}

// Helper functions

// containsPath checks if a file path is within a directory
func containsPath(basePath, filePath string) bool {
	// Simple check - could be improved with filepath.Clean and comparison
	return len(filePath) > len(basePath) && filePath[:len(basePath)] == basePath
}

// isStateSeeding determines if a torrent state indicates seeding
func isStateSeeding(state string) bool {
	seedingStates := map[string]bool{
		"uploading":       true,
		"stalledUP":       true,
		"checkingUP":      true,
		"forcedUP":        true,
		"queuedUP":        true,
		"pausedUP":        true,
		"metaDL":          false,
		"downloading":     false,
		"stalledDL":       false,
		"checkingDL":      false,
		"forcedDL":        false,
		"queuedDL":        false,
		"pausedDL":        false,
		"allocating":      false,
		"checkingResumeData": false,
		"moving":          false,
		"error":           false,
		"missingFiles":    false,
		"unknown":         false,
	}

	return seedingStates[state]
}
