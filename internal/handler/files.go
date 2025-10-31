package handler

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/carcheky/keepercheky/internal/config"
	"github.com/carcheky/keepercheky/internal/repository"
	"github.com/carcheky/keepercheky/internal/service"
	"github.com/carcheky/keepercheky/internal/service/clients"
	"github.com/carcheky/keepercheky/pkg/cache"
	"github.com/carcheky/keepercheky/pkg/filesystem"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FilesHandler handles file listing and management
type FilesHandler struct {
	mediaRepo      *repository.MediaRepository
	config         *config.Config
	syncService    *service.SyncService
	healthAnalyzer *service.HealthAnalyzer
	logger         *zap.Logger
	countsCache    *cache.CountsCache
}

// NewFilesHandler creates a new files handler
func NewFilesHandler(
	mediaRepo *repository.MediaRepository,
	cfg *config.Config,
	syncService *service.SyncService,
	healthAnalyzer *service.HealthAnalyzer,
	logger *zap.Logger,
	countsCache *cache.CountsCache,
) *FilesHandler {
	return &FilesHandler{
		mediaRepo:      mediaRepo,
		config:         cfg,
		syncService:    syncService,
		healthAnalyzer: healthAnalyzer,
		logger:         logger,
		countsCache:    countsCache,
	}
}

// PathInfo represents monitored path information from external services
type PathInfo struct {
	Service string `json:"service"` // radarr, sonarr, jellyfin, qbittorrent
	Type    string `json:"type"`    // library, download
	Path    string `json:"path"`
	Label   string `json:"label"` // Human-readable label
}

// UserViewInfo represents user viewing information
type UserViewInfo struct {
	UserID     string `json:"user_id"`
	Username   string `json:"username"`
	PlayCount  int    `json:"play_count"`
	LastPlayed string `json:"last_played"`
}

// MediaFileInfo represents media file information for display
type MediaFileInfo struct {
	ID            uint   `json:"id" gorm:"column:id"`
	Title         string `json:"title" gorm:"column:title"`
	Type          string `json:"type" gorm:"column:type"`
	FilePath      string `json:"file_path" gorm:"column:file_path"`
	Size          int64  `json:"size" gorm:"column:size"`
	PosterURL     string `json:"poster_url" gorm:"column:poster_url"`
	Quality       string `json:"quality" gorm:"column:quality"`
	IsHardlink    bool   `json:"is_hardlink" gorm:"column:is_hardlink"`
	HardlinkPaths string `json:"hardlink_paths" gorm:"column:hardlink_paths"`
	PrimaryPath   string `json:"primary_path" gorm:"column:primary_path"`

	// Service flags
	InRadarr      bool `json:"in_radarr" gorm:"column:in_radarr"`
	InSonarr      bool `json:"in_sonarr" gorm:"column:in_sonarr"`
	InJellyfin    bool `json:"in_jellyfin" gorm:"column:in_jellyfin"`
	InJellyseerr  bool `json:"in_jellyseerr" gorm:"column:in_jellyseerr"`
	InJellystat   bool `json:"in_jellystat" gorm:"column:in_jellystat"`
	InQBittorrent bool `json:"in_qbittorrent" gorm:"column:in_qbittorrent"`

	// Service IDs
	RadarrID     *int    `json:"radarr_id" gorm:"column:radarr_id"`
	SonarrID     *int    `json:"sonarr_id" gorm:"column:sonarr_id"`
	JellyfinID   *string `json:"jellyfin_id" gorm:"column:jellyfin_id"`
	JellyseerrID *int    `json:"jellyseerr_id" gorm:"column:jellyseerr_id"`
	JellystatID  *string `json:"jellystat_id" gorm:"column:jellystat_id"`

	// Torrent info
	TorrentHash     string  `json:"torrent_hash" gorm:"column:torrent_hash"`
	TorrentCategory string  `json:"torrent_category" gorm:"column:torrent_category"`
	TorrentState    string  `json:"torrent_state" gorm:"column:torrent_state"`
	TorrentTags     string  `json:"torrent_tags" gorm:"column:torrent_tags"`
	IsSeeding       bool    `json:"is_seeding" gorm:"column:is_seeding"`
	SeedRatio       float64 `json:"seed_ratio" gorm:"column:seed_ratio"`

	// Viewing information (not stored in DB, calculated on-the-fly)
	HasBeenWatched  bool           `json:"has_been_watched" gorm:"-"`
	WatchedByUsers  []UserViewInfo `json:"watched_by_users" gorm:"-"`
	TotalPlayCount  int            `json:"total_play_count" gorm:"-"`
	LastWatchedDate string         `json:"last_watched_date" gorm:"-"`

	// Metadata
	Tags     []string `json:"tags" gorm:"-"`
	Excluded bool     `json:"excluded" gorm:"column:excluded"`
}

// CountResult holds the result of category count aggregation query
type CountResult struct {
	Healthy   int64 `gorm:"column:healthy"`
	Attention int64 `gorm:"column:attention"`
	Critical  int64 `gorm:"column:critical"`
	Hardlinks int64 `gorm:"column:hardlinks"`
	Unwatched int64 `gorm:"column:unwatched"`
	Total     int64 `gorm:"column:total"`
}

// RenderFilesPage renders the files listing page
// This is the initial HTML render - actual data is loaded via JavaScript calling /api/files
func (h *FilesHandler) RenderFilesPage(c *fiber.Ctx) error {
	// Just render the empty page template - Alpine.js will load data from /api/files
	// This prevents slow filesystem scanning on every page load
	return c.Render("pages/files", fiber.Map{
		"Title": "Archivos del Sistema",
	}, "layouts/main")
}

// getServicePaths retrieves configured paths from external services
func (h *FilesHandler) getServicePaths(ctx context.Context) []PathInfo {
	var paths []PathInfo

	// Get qBittorrent save paths FIRST (origin/source)
	if h.config.Clients.QBittorrent.Enabled && h.syncService.GetQBittorrentClient() != nil {
		qbittorrentPaths := h.getQBittorrentPaths(ctx)
		paths = append(paths, qbittorrentPaths...)
	}

	// Get Jellyfin library paths (destination)
	if h.config.Clients.Jellyfin.Enabled && h.syncService.GetJellyfinClient() != nil {
		jellyfinPaths := h.getJellyfinPaths(ctx)
		paths = append(paths, jellyfinPaths...)
	}

	// Get Radarr root folders
	if h.config.Clients.Radarr.Enabled && h.syncService.GetRadarrClient() != nil {
		radarrPaths := h.getRadarrPaths(ctx)
		paths = append(paths, radarrPaths...)
	}

	// Get Sonarr root folders
	if h.config.Clients.Sonarr.Enabled && h.syncService.GetSonarrClient() != nil {
		sonarrPaths := h.getSonarrPaths(ctx)
		paths = append(paths, sonarrPaths...)
	}

	return paths
}

// getRadarrPaths retrieves root folder paths from Radarr
func (h *FilesHandler) getRadarrPaths(ctx context.Context) []PathInfo {
	var paths []PathInfo

	client := h.syncService.GetRadarrClient()
	if client == nil {
		return paths
	}

	// Access internal resty client via reflection/type assertion
	// We'll use a simpler approach: create a method in the client or access via config
	// For now, let's use the config-based approach

	// Since we can't easily access the internal HTTP client,
	// we'll rely on media files to discover paths for now
	// TODO: Add GetRootFolders() method to RadarrClient interface

	return paths
}

// getSonarrPaths retrieves root folder paths from Sonarr
func (h *FilesHandler) getSonarrPaths(ctx context.Context) []PathInfo {
	var paths []PathInfo

	client := h.syncService.GetSonarrClient()
	if client == nil {
		return paths
	}

	// TODO: Add GetRootFolders() method to SonarrClient interface

	return paths
}

// getJellyfinPaths retrieves library folder paths from Jellyfin
func (h *FilesHandler) getJellyfinPaths(ctx context.Context) []PathInfo {
	var paths []PathInfo

	client := h.syncService.GetJellyfinClient()
	if client == nil {
		return paths
	}

	// Try to get virtual folders (library root paths) from Jellyfin
	// This requires type assertion to access the GetVirtualFolders method
	jellyfinClient, ok := client.(*clients.JellyfinClient)
	if !ok {
		h.logger.Warn("Could not type assert JellyfinClient")
		return paths
	}

	// Get library virtual folders
	folders, err := jellyfinClient.GetVirtualFolders(ctx)
	if err != nil {
		h.logger.Error("Failed to get Jellyfin virtual folders",
			zap.Error(err),
		)
		return paths
	}

	// Extract library paths from virtual folders
	for _, folder := range folders {
		for _, location := range folder.Locations {
			label := fmt.Sprintf("📚 Jellyfin: %s", folder.Name)
			paths = append(paths, PathInfo{
				Service: "jellyfin",
				Type:    "library",
				Path:    location,
				Label:   label,
			})
		}
	}

	return paths
} // getQBittorrentPaths retrieves download paths from qBittorrent
func (h *FilesHandler) getQBittorrentPaths(ctx context.Context) []PathInfo {
	var paths []PathInfo

	client := h.syncService.GetQBittorrentClient()
	if client == nil {
		h.logger.Warn("qBittorrent client is not configured")
		return paths
	}

	// Get qBittorrent categories and their save paths
	categories, err := client.GetCategories(ctx)
	if err != nil {
		h.logger.Error("Failed to get qBittorrent categories",
			zap.Error(err),
		)
	} else {
		// Add each category's save path
		for name, category := range categories {
			if category.SavePath != "" {
				paths = append(paths, PathInfo{
					Service: "qbittorrent",
					Type:    "download",
					Path:    category.SavePath,
					Label:   fmt.Sprintf("⬇️ qBittorrent: %s", name),
				})
				h.logger.Info("Added qBittorrent category path",
					zap.String("category", name),
					zap.String("path", category.SavePath),
				)
			}
		}
	}

	// Get qBittorrent preferences (configuration) for default save path (completed downloads)
	prefs, err := client.GetPreferences(ctx)
	if err != nil {
		h.logger.Error("Failed to get qBittorrent preferences",
			zap.Error(err),
		)
		return paths
	}

	// Add completed downloads path (ExportDirFin) if available, otherwise SavePath
	// Only add if it's not already in categories
	completedPath := prefs.ExportDirFin
	if completedPath == "" {
		completedPath = prefs.SavePath
	}

	if completedPath != "" {
		// Check if this path is not already in the list
		alreadyExists := false
		for _, p := range paths {
			if p.Path == completedPath {
				alreadyExists = true
				break
			}
		}

		if !alreadyExists {
			paths = append(paths, PathInfo{
				Service: "qbittorrent",
				Type:    "download",
				Path:    completedPath,
				Label:   "⬇️ qBittorrent: Descargas completadas",
			})
			h.logger.Info("Added qBittorrent default downloads path",
				zap.String("path", completedPath),
			)
		}
	}

	return paths
}

// mergePaths combines service paths and media paths, removing duplicates
// nolint:unused // Reserved for future use
func (h *FilesHandler) mergePaths(servicePaths, mediaPaths []PathInfo) []PathInfo {
	pathMap := make(map[string]PathInfo)

	// Add service paths first (higher priority)
	for _, p := range servicePaths {
		// Normalize path for comparison
		normalizedPath := filepath.Clean(p.Path)
		key := p.Service + ":" + normalizedPath
		pathMap[key] = p
	}

	// Add media paths (only if not already present)
	for _, p := range mediaPaths {
		normalizedPath := filepath.Clean(p.Path)
		key := p.Service + ":" + normalizedPath
		if _, exists := pathMap[key]; !exists {
			pathMap[key] = p
		}
	}

	// Convert map to slice
	result := make([]PathInfo, 0, len(pathMap))
	for _, path := range pathMap {
		result = append(result, path)
	}

	return result
}

// extractPathsFromMedia extracts unique directory paths from media files
// nolint:unused // Reserved for future use
func (h *FilesHandler) extractPathsFromMedia(media []MediaFileInfo) []PathInfo {
	pathMap := make(map[string]PathInfo)

	for _, m := range media {
		if m.FilePath == "" {
			continue
		}

		// Get parent directory
		dir := m.FilePath
		// Remove filename, keep directory path
		for i := len(dir) - 1; i >= 0; i-- {
			if dir[i] == '/' {
				dir = dir[:i]
				break
			}
		}

		// Determine service and type based on flags
		var service, pathType, label string

		// Check for qBittorrent downloads
		if m.InQBittorrent && m.TorrentCategory != "" {
			service = "qbittorrent"
			pathType = "download"
			label = "⬇️ qBittorrent: " + m.TorrentCategory
			key := service + ":" + dir
			if _, exists := pathMap[key]; !exists {
				pathMap[key] = PathInfo{
					Service: service,
					Type:    pathType,
					Path:    dir,
					Label:   label,
				}
			}
		}

		// Check for Radarr movies
		if m.InRadarr && m.Type == "movie" {
			service = "radarr"
			pathType = "library"
			label = "🎬 Radarr: Películas"
			key := service + ":" + dir
			if _, exists := pathMap[key]; !exists {
				pathMap[key] = PathInfo{
					Service: service,
					Type:    pathType,
					Path:    dir,
					Label:   label,
				}
			}
		}

		// Check for Sonarr series
		if m.InSonarr && (m.Type == "series" || m.Type == "episode") {
			service = "sonarr"
			pathType = "library"
			label = "📺 Sonarr: Series"
			key := service + ":" + dir
			if _, exists := pathMap[key]; !exists {
				pathMap[key] = PathInfo{
					Service: service,
					Type:    pathType,
					Path:    dir,
					Label:   label,
				}
			}
		}

		// Check for Jellyfin library
		if m.InJellyfin {
			service = "jellyfin"
			pathType = "library"
			if m.Type == "movie" {
				label = "📚 Jellyfin: Biblioteca de Películas"
			} else if m.Type == "series" || m.Type == "episode" {
				label = "📚 Jellyfin: Biblioteca de Series"
			} else {
				label = "📚 Jellyfin: Biblioteca"
			}
			key := service + ":" + dir
			if _, exists := pathMap[key]; !exists {
				pathMap[key] = PathInfo{
					Service: service,
					Type:    pathType,
					Path:    dir,
					Label:   label,
				}
			}
		}
	}

	// Convert map to slice
	paths := make([]PathInfo, 0, len(pathMap))
	for _, path := range pathMap {
		paths = append(paths, path)
	}

	return paths
}

// scanFilesystemFromPaths scans the filesystem from service paths and returns unified files
func (h *FilesHandler) scanFilesystemFromPaths(servicePaths []PathInfo) ([]MediaFileInfo, error) {
	ctx := context.Background()

	// Extract unique root paths and categorize them
	var rootPaths []string
	var libraryPaths []string
	var downloadPaths []string

	pathSet := make(map[string]bool)

	for _, pathInfo := range servicePaths {
		if pathSet[pathInfo.Path] {
			continue
		}
		pathSet[pathInfo.Path] = true
		rootPaths = append(rootPaths, pathInfo.Path)

		if pathInfo.Type == "library" {
			libraryPaths = append(libraryPaths, pathInfo.Path)
		} else if pathInfo.Type == "download" {
			downloadPaths = append(downloadPaths, pathInfo.Path)
		}
	}

	if len(rootPaths) == 0 {
		h.logger.Warn("No paths to scan")
		return []MediaFileInfo{}, nil
	}

	// Create scanner with options
	scanOptions := filesystem.ScanOptions{
		RootPaths:       rootPaths,
		LibraryPaths:    libraryPaths,
		DownloadPaths:   downloadPaths,
		VideoExtensions: []string{".mkv", ".mp4", ".avi", ".m4v", ".ts", ".m2ts", ".wmv", ".flv", ".webm"},
		MinSize:         50 * 1024 * 1024, // 50MB minimum
		SkipHidden:      true,
	}

	scanner := filesystem.NewScanner(scanOptions, h.logger)

	// Scan filesystem
	fileEntries, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("filesystem scan failed: %w", err)
	}

	// Convert FileEntry map to EnrichedFile map for enrichment
	enrichedFiles := make(map[string]*filesystem.EnrichedFile)
	for path, entry := range fileEntries {
		enrichedFiles[path] = &filesystem.EnrichedFile{
			FileEntry: entry,
			ModTime:   entry.ModTime,
		}
	}

	// Enrich with service data
	h.enrichWithServices(ctx, enrichedFiles)

	// Convert to MediaFileInfo, grouping hardlinks
	mediaFiles := make([]MediaFileInfo, 0, len(enrichedFiles))
	processedInodes := make(map[int64]bool)

	for _, enrichedFile := range enrichedFiles {
		// Skip if we already processed this inode (hardlink group)
		if enrichedFile.IsHardlink && processedInodes[enrichedFile.Inode] {
			continue
		}

		// Mark inode as processed
		if enrichedFile.IsHardlink {
			processedInodes[enrichedFile.Inode] = true
		}

		// Use enriched title if available, otherwise infer from filename
		title := enrichedFile.Title
		if title == "" {
			title = h.inferTitleFromPath(enrichedFile.PrimaryPath)
		}

		// Build hardlink paths string
		hardlinkPathsStr := ""
		if enrichedFile.IsHardlink && len(enrichedFile.HardlinkPaths) > 0 {
			hardlinkPathsStr = strings.Join(enrichedFile.HardlinkPaths, "|")
		}

		// Convert tags
		tags := enrichedFile.Tags
		if tags == nil {
			tags = []string{}
		}

		mediaFile := MediaFileInfo{
			ID:              0, // No database ID since this is direct filesystem scan
			Title:           title,
			Type:            enrichedFile.MediaType,
			FilePath:        enrichedFile.PrimaryPath,
			Size:            enrichedFile.Size,
			PosterURL:       enrichedFile.PosterURL,
			Quality:         enrichedFile.Quality,
			IsHardlink:      enrichedFile.IsHardlink,
			HardlinkPaths:   hardlinkPathsStr,
			PrimaryPath:     enrichedFile.PrimaryPath,
			InRadarr:        enrichedFile.InRadarr,
			InSonarr:        enrichedFile.InSonarr,
			InJellyfin:      enrichedFile.InJellyfin,
			InJellyseerr:    enrichedFile.InJellyseerr,
			InQBittorrent:   enrichedFile.InQBittorrent,
			RadarrID:        enrichedFile.RadarrID,
			SonarrID:        enrichedFile.SonarrID,
			JellyfinID:      enrichedFile.JellyfinID,
			JellyseerrID:    enrichedFile.JellyseerrID,
			TorrentHash:     enrichedFile.TorrentHash,
			TorrentCategory: enrichedFile.TorrentCategory,
			TorrentState:    enrichedFile.TorrentState,
			TorrentTags:     enrichedFile.TorrentTags,
			IsSeeding:       enrichedFile.IsSeeding,
			SeedRatio:       enrichedFile.SeedRatio,
			Tags:            tags,
			Excluded:        enrichedFile.Excluded,
			HasBeenWatched:  false,
			WatchedByUsers:  []UserViewInfo{},
			TotalPlayCount:  0,
			LastWatchedDate: "",
		}

		// Add watching info if available
		if enrichedFile.LastWatched != nil {
			mediaFile.HasBeenWatched = true
			// TODO: Get detailed user watching info from Jellyfin/Jellystat
		}

		mediaFiles = append(mediaFiles, mediaFile)
	}

	h.logger.Info("Filesystem scan complete",
		zap.Int("total_entries", len(fileEntries)),
		zap.Int("unique_files", len(mediaFiles)),
		zap.Int("hardlink_groups", len(processedInodes)),
	)

	return mediaFiles, nil
}

// enrichWithServices enriches files with data from all configured services
func (h *FilesHandler) enrichWithServices(ctx context.Context, files map[string]*filesystem.EnrichedFile) {
	enricher := filesystem.NewEnricher(h.logger)

	// Enrich with Radarr
	if h.config.Clients.Radarr.Enabled && h.syncService.GetRadarrClient() != nil {
		h.logger.Info("Enriching with Radarr data")
		radarrMedia, err := h.syncService.GetRadarrClient().GetLibrary(ctx)
		if err != nil {
			h.logger.Error("Failed to get Radarr library", zap.Error(err))
		} else {
			count := enricher.EnrichWithRadarr(ctx, files, radarrMedia)
			h.logger.Info("Radarr enrichment complete", zap.Int("enriched", count))
		}
	}

	// Enrich with Sonarr
	if h.config.Clients.Sonarr.Enabled && h.syncService.GetSonarrClient() != nil {
		h.logger.Info("Enriching with Sonarr data")
		sonarrMedia, err := h.syncService.GetSonarrClient().GetLibrary(ctx)
		if err != nil {
			h.logger.Error("Failed to get Sonarr library", zap.Error(err))
		} else {
			count := enricher.EnrichWithSonarr(ctx, files, sonarrMedia)
			h.logger.Info("Sonarr enrichment complete", zap.Int("enriched", count))
		}
	}

	// Enrich with Jellyfin
	if h.config.Clients.Jellyfin.Enabled && h.syncService.GetJellyfinClient() != nil {
		h.logger.Info("Enriching with Jellyfin data")
		jellyfinMedia, err := h.syncService.GetJellyfinClient().GetLibrary(ctx)
		if err != nil {
			h.logger.Error("Failed to get Jellyfin library", zap.Error(err))
		} else {
			count := enricher.EnrichWithJellyfin(ctx, files, jellyfinMedia)
			h.logger.Info("Jellyfin enrichment complete", zap.Int("enriched", count))
		}
	}

	// Enrich with qBittorrent
	if h.config.Clients.QBittorrent.Enabled && h.syncService.GetQBittorrentClient() != nil {
		h.logger.Info("Enriching with qBittorrent data")
		torrentMap, err := h.syncService.GetQBittorrentClient().GetAllTorrentsMap(ctx)
		if err != nil {
			h.logger.Error("Failed to get qBittorrent torrents", zap.Error(err))
		} else {
			count := enricher.EnrichWithQBittorrent(ctx, files, torrentMap)
			h.logger.Info("qBittorrent enrichment complete", zap.Int("enriched", count))
		}
	}
}

// inferTitleFromPath extracts a title from the file path
func (h *FilesHandler) inferTitleFromPath(path string) string {
	// Get filename without extension
	base := filepath.Base(path)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)

	// Clean up common patterns
	// Remove quality tags like [1080p], (BluRay), etc.
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, ".", " ")
	nameWithoutExt = strings.ReplaceAll(nameWithoutExt, "_", " ")

	// Simple cleanup
	title := strings.TrimSpace(nameWithoutExt)

	// Limit length
	if len(title) > 80 {
		title = title[:80] + "..."
	}

	return title
}

// GetFilesAPI returns files as JSON for API access with pagination support
func (h *FilesHandler) GetFilesAPI(c *fiber.Ctx) error {
	requestStart := time.Now()

	// Get pagination parameters
	page := c.QueryInt("page", 1)
	perPage := c.QueryInt("perPage", 25)
	sortBy := c.Query("sortBy", "file_path")
	order := c.Query("order", "asc")
	tab := c.Query("tab", "") // Filter by tab: healthy, attention, critical, hardlinks, unwatched

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 25
	}
	if perPage > 100 {
		perPage = 100 // Max 100 items per page
	}

	// Validate sort field to prevent SQL injection
	allowedSortFields := map[string]bool{
		"file_path":       true,
		"title":           true,
		"size":            true,
		"type":            true,
		"in_q_bittorrent": true,
		"in_jellyfin":     true,
		"in_radarr":       true,
		"in_sonarr":       true,
		"torrent_state":   true,
		"is_seeding":      true,
		"seed_ratio":      true,
		"created_at":      true,
		"updated_at":      true,
	}
	if !allowedSortFields[sortBy] {
		sortBy = "file_path"
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "asc"
	}

	// Build query
	query := h.mediaRepo.GetDB().
		Table("media").
		Select(`
			id, title, type, file_path, size, poster_url, quality,
			is_hardlink, hardlink_paths, primary_path,
			in_radarr, in_sonarr, in_jellyfin, in_jellyseerr, in_jellystat, in_q_bittorrent,
			radarr_id, sonarr_id, jellyfin_id, jellyseerr_id, jellystat_id,
			torrent_hash, torrent_category, torrent_state, torrent_tags,
			is_seeding, seed_ratio, excluded
		`)

	// Apply tab filtering
	switch tab {
	case "attention":
		// Orphan downloads: in qBittorrent but not in Jellyfin, Radarr, or Sonarr
		query = query.Where("in_q_bittorrent = ? AND in_jellyfin = ? AND in_radarr = ? AND in_sonarr = ?",
			true, false, false, false)
	case "critical":
		// Dead torrents: error state or missing files
		query = query.Where("in_q_bittorrent = ? AND (torrent_state = ? OR torrent_state = ?)",
			true, "error", "missingFiles")
	case "hardlinks":
		// Files with hardlinks
		query = query.Where("is_hardlink = ?", true)
	case "unwatched":
		// In Jellyfin but never watched
		// Note: has_been_watched is computed at runtime, so we need a workaround
		// For now, just filter by in_jellyfin - the frontend will do final filtering
		query = query.Where("in_jellyfin = ?", true)
	case "healthy":
		// Healthy files: in Jellyfin and (in Radarr or Sonarr) and no problems
		query = query.Where("in_jellyfin = ? AND (in_radarr = ? OR in_sonarr = ?)",
			true, true, true).
			Where("(torrent_state IS NULL OR torrent_state NOT IN (?, ?))", "error", "missingFiles")
	}

	// Get total count for this filter
	countStart := time.Now()
	var totalCount int64
	countQuery := query.Session(&gorm.Session{})
	err := countQuery.Count(&totalCount).Error
	countElapsed := time.Since(countStart)
	if err != nil {
		h.logger.Error("Failed to count media", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count media"})
	}

	// Calculate offset
	offset := (page - 1) * perPage

	// Apply sorting - prioritize qBittorrent and Jellyfin, then custom sort
	if sortBy == "file_path" && order == "asc" {
		// Default sort: downloads first, then library items, then alphabetical
		query = query.Order("in_q_bittorrent DESC, in_jellyfin DESC, file_path ASC")
	} else {
		// Custom sort requested - use GORM clause for SQL injection safety
		query = query.Order(clause.OrderByColumn{
			Column: clause.Column{Name: sortBy},
			Desc:   order == "desc",
		})
	}

	// Apply pagination and fetch data
	queryStart := time.Now()
	var media []MediaFileInfo
	err = query.
		Limit(perPage).
		Offset(offset).
		Find(&media).Error
	queryElapsed := time.Since(queryStart)

	if err != nil {
		h.logger.Error("Failed to query media from database", zap.Error(err))
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))

	// Get category counts (for summary cards) - only if not filtering by tab
	var counts map[string]int64
	var countsElapsed time.Duration
	if tab == "" {
		countsStart := time.Now()
		counts = h.getCategoryCounts()
		countsElapsed = time.Since(countsStart)
	}

	// Get last sync time from settings
	var lastSyncSetting struct {
		Value string `gorm:"column:value"`
	}
	lastSyncTime := ""
	err = h.mediaRepo.GetDB().
		Table("settings").
		Select("value").
		Where("key = ?", "last_files_sync").
		First(&lastSyncSetting).Error

	if err == nil {
		lastSyncTime = lastSyncSetting.Value
	}

	totalElapsed := time.Since(requestStart)
	h.logger.Info("Files API request completed",
		zap.Int("page", page),
		zap.Int("perPage", perPage),
		zap.String("tab", tab),
		zap.Int64("total", totalCount),
		zap.Int("returned", len(media)),
		zap.Duration("total_time", totalElapsed),
		zap.Duration("count_time", countElapsed),
		zap.Duration("query_time", queryElapsed),
		zap.Duration("counts_time", countsElapsed),
	)

	response := fiber.Map{
		"files":      media,
		"total":      totalCount,
		"page":       page,
		"perPage":    perPage,
		"totalPages": totalPages,
		"last_sync":  lastSyncTime,
	}

	// Add counts if available
	if counts != nil {
		response["counts"] = counts
	}

	return c.JSON(response)
}

// getCategoryCounts returns counts for each file category with caching
func (h *FilesHandler) getCategoryCounts() map[string]int64 {
	// Check cache first
	if cached, ok := h.countsCache.Get(); ok {
		h.logger.Debug("Category counts served from cache")
		return cached
	}

	// Cache miss - recalculate
	startTime := time.Now()
	counts := h.calculateCategoryCounts()

	// Update cache only if calculation succeeded
	if len(counts) > 0 {
		h.countsCache.Set(counts)

		elapsed := time.Since(startTime)
		h.logger.Info("Category counts calculated and cached",
			zap.Duration("elapsed", elapsed),
			zap.Int("total", int(counts["total"])),
		)
	}

	return counts
}

// calculateCategoryCounts performs the actual category count queries
// This is separated from getCategoryCounts to allow for caching
func (h *FilesHandler) calculateCategoryCounts() map[string]int64 {
	db := h.mediaRepo.GetDB().Table("media")

	// Use a single query with CASE statements for better performance
	var result CountResult
	err := db.Select(`
		COUNT(CASE 
			WHEN in_jellyfin = true 
			AND (in_radarr = true OR in_sonarr = true)
			AND (torrent_state IS NULL OR torrent_state NOT IN ('error', 'missingFiles'))
			THEN 1 END) as healthy,
		COUNT(CASE 
			WHEN in_q_bittorrent = true 
			AND in_jellyfin = false 
			AND in_radarr = false 
			AND in_sonarr = false
			THEN 1 END) as attention,
		COUNT(CASE 
			WHEN in_q_bittorrent = true 
			AND (torrent_state = 'error' OR torrent_state = 'missingFiles')
			THEN 1 END) as critical,
		COUNT(CASE WHEN is_hardlink = true THEN 1 END) as hardlinks,
		COUNT(CASE WHEN in_jellyfin = true THEN 1 END) as unwatched,
		COUNT(*) as total
	`).Scan(&result).Error

	if err != nil {
		h.logger.Error("Failed to calculate category counts", zap.Error(err))
		// Return nil to indicate error (caller should use cached values if available)
		return nil
	}

	counts := make(map[string]int64)
	counts["healthy"] = result.Healthy
	counts["attention"] = result.Attention
	counts["critical"] = result.Critical
	counts["hardlinks"] = result.Hardlinks
	counts["unwatched"] = result.Unwatched
	counts["total"] = result.Total

	return counts
}

// GetFilesHealthAPI returns files with health analysis as JSON
func (h *FilesHandler) GetFilesHealthAPI(c *fiber.Ctx) error {
	ctx := context.Background()

	// Get service paths
	servicePaths := h.getServicePaths(ctx)

	// Scan filesystem
	scannedFiles, err := h.scanFilesystemFromPaths(servicePaths)
	if err != nil {
		h.logger.Error("Failed to scan filesystem for health analysis",
			zap.Error(err),
		)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to scan filesystem",
		})
	}

	// Convert MediaFileInfo to service.MediaFileInfo for health analysis
	serviceFiles := make([]service.MediaFileInfo, len(scannedFiles))
	for i, file := range scannedFiles {
		serviceFiles[i] = service.MediaFileInfo{
			ID:              file.ID,
			Title:           file.Title,
			Type:            file.Type,
			FilePath:        file.FilePath,
			Size:            file.Size,
			PosterURL:       file.PosterURL,
			Quality:         file.Quality,
			IsHardlink:      file.IsHardlink,
			HardlinkPaths:   file.HardlinkPaths,
			PrimaryPath:     file.PrimaryPath,
			InRadarr:        file.InRadarr,
			InSonarr:        file.InSonarr,
			InJellyfin:      file.InJellyfin,
			InJellyseerr:    file.InJellyseerr,
			InQBittorrent:   file.InQBittorrent,
			RadarrID:        file.RadarrID,
			SonarrID:        file.SonarrID,
			JellyfinID:      file.JellyfinID,
			JellyseerrID:    file.JellyseerrID,
			TorrentHash:     file.TorrentHash,
			TorrentCategory: file.TorrentCategory,
			TorrentState:    file.TorrentState,
			TorrentTags:     file.TorrentTags,
			IsSeeding:       file.IsSeeding,
			SeedRatio:       file.SeedRatio,
			HasBeenWatched:  file.HasBeenWatched,
			TotalPlayCount:  file.TotalPlayCount,
			Tags:            file.Tags,
			Excluded:        file.Excluded,
		}
	}

	// Perform health analysis
	healthReports := h.healthAnalyzer.AnalyzeFiles(serviceFiles)
	summary := h.healthAnalyzer.GetHealthSummary(serviceFiles)

	h.logger.Info("Health analysis complete",
		zap.Int("total_files", len(serviceFiles)),
		zap.Int("healthy", summary.Healthy),
		zap.Int("needs_attention", summary.NeedsAttention),
	)

	return c.JSON(fiber.Map{
		"summary": summary,
		"files":   healthReports,
		"total":   len(healthReports),
	})
}

// RenderExamplePage renders the components example/demo page
func (h *FilesHandler) RenderExamplePage(c *fiber.Ctx) error {
	return c.Render("pages/files-example", fiber.Map{
		"Title": "Alpine.js Components - Demostración",
	})
}
