package filesystem

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/carcheky/keepercheky/internal/models"
	"go.uber.org/zap"
)

// EnrichedFile represents a file with service metadata
type EnrichedFile struct {
	*FileEntry

	// Service flags
	InRadarr      bool
	InSonarr      bool
	InJellyfin    bool
	InJellyseerr  bool
	InJellystat   bool
	InQBittorrent bool

	// Service IDs
	RadarrID     *int
	SonarrID     *int
	JellyfinID   *string
	JellyseerrID *int

	// Torrent info
	TorrentHash     string
	TorrentCategory string
	TorrentTags     string
	TorrentState    string
	IsSeeding       bool
	SeedRatio       float64

	// Additional metadata
	Title       string
	Quality     string
	PosterURL   string
	Tags        []string
	Excluded    bool
	LastWatched *int64
	ModTime     int64 // Unix timestamp from FileEntry for easy access
}

// Enricher enriches file entries with service metadata
type Enricher struct {
	logger *zap.Logger
}

// NewEnricher creates a new enricher
func NewEnricher(logger *zap.Logger) *Enricher {
	return &Enricher{
		logger: logger,
	}
}

// EnrichWithRadarr enriches files with Radarr metadata
func (e *Enricher) EnrichWithRadarr(
	ctx context.Context,
	files map[string]*EnrichedFile,
	radarrMedia []*models.Media,
) int {
	e.logger.Info("Enriching with Radarr data",
		zap.Int("radarr_items", len(radarrMedia)),
	)

	enriched := 0

	// Create path map for fast lookup
	radarrByPath := make(map[string]*models.Media)
	for _, media := range radarrMedia {
		radarrByPath[media.FilePath] = media
	}

	// Match files with Radarr items
	for path, file := range files {
		// Try exact match first
		if radarrItem, found := radarrByPath[path]; found {
			e.applyRadarrData(file, radarrItem)
			enriched++
			continue
		}

		// Try matching with primary path (for hardlinks)
		if file.IsHardlink && file.PrimaryPath != path {
			if radarrItem, found := radarrByPath[file.PrimaryPath]; found {
				e.applyRadarrData(file, radarrItem)
				enriched++
				continue
			}
		}

		// Try fuzzy matching by directory name
		for radarrPath, radarrItem := range radarrByPath {
			if e.pathsMatch(path, radarrPath) {
				e.applyRadarrData(file, radarrItem)
				enriched++
				break
			}
		}
	}

	e.logger.Info("Radarr enrichment complete",
		zap.Int("enriched", enriched),
	)

	return enriched
}

// EnrichWithSonarr enriches files with Sonarr metadata
func (e *Enricher) EnrichWithSonarr(
	ctx context.Context,
	files map[string]*EnrichedFile,
	sonarrMedia []*models.Media,
) int {
	e.logger.Info("Enriching with Sonarr data",
		zap.Int("sonarr_items", len(sonarrMedia)),
	)

	enriched := 0

	// Create path map for fast lookup
	sonarrByPath := make(map[string]*models.Media)
	for _, media := range sonarrMedia {
		sonarrByPath[media.FilePath] = media
	}

	// Match files with Sonarr items
	for path, file := range files {
		// Try exact match
		if sonarrItem, found := sonarrByPath[path]; found {
			e.applySonarrData(file, sonarrItem)
			enriched++
			continue
		}

		// Try matching with primary path
		if file.IsHardlink && file.PrimaryPath != path {
			if sonarrItem, found := sonarrByPath[file.PrimaryPath]; found {
				e.applySonarrData(file, sonarrItem)
				enriched++
				continue
			}
		}

		// Try fuzzy matching
		for sonarrPath, sonarrItem := range sonarrByPath {
			if e.pathsMatch(path, sonarrPath) {
				e.applySonarrData(file, sonarrItem)
				enriched++
				break
			}
		}
	}

	e.logger.Info("Sonarr enrichment complete",
		zap.Int("enriched", enriched),
	)

	return enriched
}

// EnrichWithJellyfin enriches files with Jellyfin metadata
func (e *Enricher) EnrichWithJellyfin(
	ctx context.Context,
	files map[string]*EnrichedFile,
	jellyfinMedia []*models.Media,
) int {
	e.logger.Info("Enriching with Jellyfin data",
		zap.Int("jellyfin_items", len(jellyfinMedia)),
	)

	enriched := 0

	// Create path map
	jellyfinByPath := make(map[string]*models.Media)
	for _, media := range jellyfinMedia {
		jellyfinByPath[media.FilePath] = media
	}

	// Match files
	for path, file := range files {
		if jfItem, found := jellyfinByPath[path]; found {
			e.applyJellyfinData(file, jfItem)
			enriched++
			continue
		}

		if file.IsHardlink && file.PrimaryPath != path {
			if jfItem, found := jellyfinByPath[file.PrimaryPath]; found {
				e.applyJellyfinData(file, jfItem)
				enriched++
				continue
			}
		}

		for jfPath, jfItem := range jellyfinByPath {
			if e.pathsMatch(path, jfPath) {
				e.applyJellyfinData(file, jfItem)
				enriched++
				break
			}
		}
	}

	e.logger.Info("Jellyfin enrichment complete",
		zap.Int("enriched", enriched),
	)

	return enriched
}

// EnrichWithQBittorrent enriches files with qBittorrent metadata
func (e *Enricher) EnrichWithQBittorrent(
	ctx context.Context,
	files map[string]*EnrichedFile,
	torrentMap map[string]*models.TorrentInfo,
) int {
	e.logger.Info("Enriching with qBittorrent data",
		zap.Int("torrents", len(torrentMap)),
	)

	enriched := 0

	for path, file := range files {
		// Try exact match
		if torrent, found := torrentMap[path]; found {
			e.applyTorrentData(file, torrent)
			enriched++
			continue
		}

		// Try matching with primary path
		if file.IsHardlink && file.PrimaryPath != path {
			if torrent, found := torrentMap[file.PrimaryPath]; found {
				e.applyTorrentData(file, torrent)
				enriched++
				continue
			}
		}

		// Try matching hardlink paths
		if file.IsHardlink {
			for _, hlPath := range file.HardlinkPaths {
				if torrent, found := torrentMap[hlPath]; found {
					e.applyTorrentData(file, torrent)
					enriched++
					break
				}
			}
		}
	}

	e.logger.Info("qBittorrent enrichment complete",
		zap.Int("enriched", enriched),
	)

	return enriched
}

// Helper methods to apply service data

func (e *Enricher) applyRadarrData(file *EnrichedFile, radarr *models.Media) {
	file.InRadarr = true
	file.RadarrID = radarr.RadarrID
	if file.Title == "" {
		file.Title = radarr.Title
	}
	if file.Quality == "" {
		file.Quality = radarr.Quality
	}
	if file.PosterURL == "" {
		file.PosterURL = radarr.PosterURL
	}
	file.Excluded = radarr.Excluded
}

func (e *Enricher) applySonarrData(file *EnrichedFile, sonarr *models.Media) {
	file.InSonarr = true
	file.SonarrID = sonarr.SonarrID
	if file.Title == "" {
		file.Title = sonarr.Title
	}
	if file.Quality == "" {
		file.Quality = sonarr.Quality
	}
	if file.PosterURL == "" {
		file.PosterURL = sonarr.PosterURL
	}
	file.Excluded = sonarr.Excluded
}

func (e *Enricher) applyJellyfinData(file *EnrichedFile, jellyfin *models.Media) {
	file.InJellyfin = true
	file.JellyfinID = jellyfin.JellyfinID
	if file.Title == "" {
		file.Title = jellyfin.Title
	}
	if file.PosterURL == "" {
		file.PosterURL = jellyfin.PosterURL
	}
	if jellyfin.LastWatched != nil {
		timestamp := jellyfin.LastWatched.Unix()
		file.LastWatched = &timestamp
	}
}

// EnrichWithJellystat sets the InJellystat flag for files tracked by Jellystat
// Since Jellystat monitors all Jellyfin activity, files in Jellyfin are tracked by Jellystat
func (e *Enricher) EnrichWithJellystat(
	ctx context.Context,
	files map[string]*EnrichedFile,
	jellystatEnabled bool,
) int {
	if !jellystatEnabled {
		return 0
	}

	e.logger.Info("Enriching with Jellystat tracking")

	enriched := 0
	for _, file := range files {
		// Jellystat tracks all items in Jellyfin
		if file.InJellyfin {
			file.InJellystat = true
			enriched++
		}
	}

	e.logger.Info("Jellystat enrichment complete",
		zap.Int("files_tracked", enriched),
	)

	return enriched
}

func (e *Enricher) applyTorrentData(file *EnrichedFile, torrent *models.TorrentInfo) {
	file.InQBittorrent = true
	file.TorrentHash = torrent.Hash
	file.TorrentCategory = torrent.Category
	file.TorrentTags = torrent.Tags
	file.TorrentState = torrent.State
	file.IsSeeding = torrent.IsSeeding
	file.SeedRatio = torrent.Ratio
}

// pathsMatch performs fuzzy path matching
func (e *Enricher) pathsMatch(path1, path2 string) bool {
	// Normalize paths
	norm1 := strings.ToLower(strings.TrimSuffix(path1, "/"))
	norm2 := strings.ToLower(strings.TrimSuffix(path2, "/"))

	// Check if one contains the other
	if strings.Contains(norm1, norm2) || strings.Contains(norm2, norm1) {
		return true
	}

	// Check if base directories match
	base1 := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(filepath.Base(norm1), ".", " "), "_", " "))
	base2 := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(filepath.Base(norm2), ".", " "), "_", " "))

	return strings.Contains(base1, base2) || strings.Contains(base2, base1)
}
