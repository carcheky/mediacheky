package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// StringSlice is a custom type for string slices that can be serialized to JSON
type StringSlice []string

// Scan implements the sql.Scanner interface
func (s *StringSlice) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// Value implements the driver.Valuer interface
func (s StringSlice) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

// Media represents a media item (movie or TV show)
type Media struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Title       string     `json:"title" gorm:"not null;index"`
	Type        string     `json:"type" gorm:"not null;index"` // "movie" or "series"
	PosterURL   string     `json:"poster_url"`
	FilePath    string     `json:"file_path" gorm:"not null;uniqueIndex"` // Primary path (canonical)
	Size        int64      `json:"size"`
	AddedDate   time.Time  `json:"added_date" gorm:"index"`
	LastWatched *time.Time `json:"last_watched"`

	// Filesystem metadata (source of truth)
	FileInode     int64       `json:"file_inode" gorm:"index"`                // Inode number for hardlink detection
	FileModTime   int64       `json:"file_mod_time"`                          // Last modification time (Unix timestamp)
	IsHardlink    bool        `json:"is_hardlink" gorm:"default:false;index"` // Has hardlinks
	HardlinkPaths StringSlice `json:"hardlink_paths" gorm:"type:text"`        // All hardlink paths
	PrimaryPath   string      `json:"primary_path"`                           // Canonical path (same as FilePath but explicit)

	// Series specific
	EpisodeCount     int `json:"episode_count"`
	SeasonCount      int `json:"season_count"`
	EpisodeFileCount int `json:"episode_file_count"` // Downloaded episodes

	// Torrent status
	IsSeeding       bool    `json:"is_seeding" gorm:"default:false;index"`
	TorrentHash     string  `json:"torrent_hash" gorm:"index"` // Hash from qBittorrent
	SeedRatio       float64 `json:"seed_ratio"`
	TorrentCategory string  `json:"torrent_category"`           // Category in qBittorrent
	TorrentTags     string  `json:"torrent_tags"`               // Tags in qBittorrent
	TorrentState    string  `json:"torrent_state" gorm:"index"` // State (uploading, stalledUP, etc.)

	// Service IDs
	RadarrID     *int    `json:"radarr_id" gorm:"index"`
	SonarrID     *int    `json:"sonarr_id" gorm:"index"`
	JellyfinID   *string `json:"jellyfin_id" gorm:"index"`
	JellyseerrID *int    `json:"jellyseerr_id" gorm:"index"`
	JellystatID  *string `json:"jellystat_id" gorm:"index"`

	// Service flags (filesystem-first approach)
	InRadarr      bool `json:"in_radarr" gorm:"default:false;index"`
	InSonarr      bool `json:"in_sonarr" gorm:"default:false;index"`
	InJellyfin    bool `json:"in_jellyfin" gorm:"default:false;index"`
	InJellyseerr  bool `json:"in_jellyseerr" gorm:"default:false;index"`
	InJellystat   bool `json:"in_jellystat" gorm:"default:false;index"`
	InQBittorrent bool `json:"in_qbittorrent" gorm:"column:in_q_bittorrent;default:false;index"`

	// Metadata
	Tags     StringSlice `json:"tags" gorm:"type:text"`
	Quality  string      `json:"quality"`
	Excluded bool        `json:"excluded" gorm:"default:false;index"`

	// Relationships
	History []History `json:"history,omitempty" gorm:"foreignKey:MediaID"`
}

func (Media) TableName() string {
	return "media"
}

// Schedule represents a cleanup schedule
type Schedule struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string     `json:"name" gorm:"not null"`
	Enabled     bool       `json:"enabled" gorm:"default:true"`
	CronExpr    string     `json:"cron_expr" gorm:"not null"`
	LastRun     *time.Time `json:"last_run"`
	NextRun     *time.Time `json:"next_run"`
	Description string     `json:"description"`

	// Rules (stored as JSON for now, can be normalized later)
	Rules string `json:"rules" gorm:"type:json"`
}

func (Schedule) TableName() string {
	return "schedules"
}

// History represents an action log entry
type History struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`

	MediaID    uint   `json:"media_id" gorm:"index"`
	MediaTitle string `json:"media_title"`
	Action     string `json:"action" gorm:"not null"` // "deleted", "excluded", "unmonitored"
	Status     string `json:"status" gorm:"not null"` // "success", "failed"
	Message    string `json:"message"`
	DryRun     bool   `json:"dry_run" gorm:"default:false"`
}

func (History) TableName() string {
	return "history"
}

// Settings represents application settings
type Settings struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UpdatedAt time.Time `json:"updated_at"`

	Key   string `json:"key" gorm:"uniqueIndex;not null"`
	Value string `json:"value" gorm:"type:text"`
}

func (Settings) TableName() string {
	return "settings"
}

// GlobalStats represents global statistics for the media library
type GlobalStats struct {
	TotalMedia            int64 `json:"total_media"`
	TotalMovies           int64 `json:"total_movies"`
	TotalSeries           int64 `json:"total_series"`
	TotalSize             int64 `json:"total_size"`
	TotalEpisodes         int   `json:"total_episodes"`
	TotalEpisodesDownload int   `json:"total_episodes_download"`
}

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// Run standard AutoMigrate first
	if err := db.AutoMigrate(
		&Media{},
		&Schedule{},
		&History{},
		&Settings{},
	); err != nil {
		return err
	}

	// Add composite indices for common query patterns to improve performance
	// These indices significantly speed up the Files tab filtering
	return addPerformanceIndices(db)
}

// addPerformanceIndices creates composite indices for frequently used queries
// in the Files tab to dramatically improve query performance on large libraries
func addPerformanceIndices(db *gorm.DB) error {
	// Detect database dialect to use appropriate syntax
	dialectName := db.Dialector.Name()

	// PostgreSQL and SQLite both support partial indices with WHERE clause
	// MySQL doesn't, so we skip the WHERE clause for MySQL
	if dialectName == "sqlite" || dialectName == "postgres" {
		// Use CREATE INDEX IF NOT EXISTS for SQLite
		// PostgreSQL supports this syntax since version 9.5

		// Index for healthy files query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_healthy_files 
			ON media(in_jellyfin, in_radarr, in_sonarr, torrent_state) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create healthy files index: %w", err)
		}

		// Index for orphan downloads query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_orphan_downloads 
			ON media(in_q_bittorrent, in_jellyfin, in_radarr, in_sonarr) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create orphan downloads index: %w", err)
		}

		// Index for dead torrents query
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_dead_torrents 
			ON media(in_q_bittorrent, torrent_state) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create dead torrents index: %w", err)
		}

		// Index for default sorting
		if err := db.Exec(`
			CREATE INDEX IF NOT EXISTS idx_media_default_sort 
			ON media(in_q_bittorrent DESC, in_jellyfin DESC, file_path ASC) 
			WHERE deleted_at IS NULL
		`).Error; err != nil {
			return fmt.Errorf("failed to create default sort index: %w", err)
		}
	} else if dialectName == "mysql" {
		// MySQL doesn't support partial indices, create without WHERE clause
		// Also MySQL doesn't support CREATE INDEX IF NOT EXISTS until 8.0.27
		// We'll use a simpler approach - just try to create and ignore errors if exists

		db.Exec(`CREATE INDEX idx_media_healthy_files ON media(in_jellyfin, in_radarr, in_sonarr, torrent_state)`)
		db.Exec(`CREATE INDEX idx_media_orphan_downloads ON media(in_q_bittorrent, in_jellyfin, in_radarr, in_sonarr)`)
		db.Exec(`CREATE INDEX idx_media_dead_torrents ON media(in_q_bittorrent, torrent_state)`)
		db.Exec(`CREATE INDEX idx_media_default_sort ON media(in_q_bittorrent, in_jellyfin, file_path)`)
	}

	return nil
}
