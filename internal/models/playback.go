package models

import "time"

// PlaybackInfo represents playback information from Jellyfin.
type PlaybackInfo struct {
	MediaID     string    `json:"media_id"`
	LastPlayed  time.Time `json:"last_played"`
	PlayCount   int       `json:"play_count"`
	IsFavorite  bool      `json:"is_favorite"`
	PlaybackPos int64     `json:"playback_position_ticks"` // Position in ticks (100-nanosecond units)
}
