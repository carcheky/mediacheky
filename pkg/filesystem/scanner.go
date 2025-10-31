package filesystem

import (
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

// FileEntry represents a real file on the filesystem
type FileEntry struct {
	Path          string   // Absolute path to the file
	Size          int64    // File size in bytes
	Inode         int64    // Inode number (for hardlink detection)
	ModTime       int64    // Modification time (Unix timestamp)
	IsHardlink    bool     // Whether this file has hardlinks
	HardlinkPaths []string // All paths that are hardlinks to this file
	PrimaryPath   string   // The "canonical" path (prioritized)
	MediaType     string   // "movie" or "series" (inferred from path)
}

// ScanOptions configures the filesystem scan
type ScanOptions struct {
	RootPaths       []string // Paths to scan (e.g., ["/BibliotecaMultimedia"])
	LibraryPaths    []string // Library paths to prioritize (e.g., ["/BibliotecaMultimedia/Peliculas", "/BibliotecaMultimedia/Series"])
	DownloadPaths   []string // Download paths (lower priority, e.g., ["/BibliotecaMultimedia/Descargas"])
	VideoExtensions []string // File extensions to include (e.g., [".mkv", ".mp4", ".avi"])
	MinSize         int64    // Minimum file size in bytes (skip small files)
	SkipHidden      bool     // Skip hidden files/directories
}

// Scanner scans the filesystem and detects media files
type Scanner struct {
	options ScanOptions
	logger  *zap.Logger

	// Internal state
	inodeMap map[int64][]*FileEntry // Map of inode -> list of paths (for hardlink detection)
	mu       sync.RWMutex
}

// NewScanner creates a new filesystem scanner
func NewScanner(options ScanOptions, logger *zap.Logger) *Scanner {
	// Set defaults if not provided
	if len(options.VideoExtensions) == 0 {
		options.VideoExtensions = []string{".mkv", ".mp4", ".avi", ".m4v", ".ts", ".m2ts"}
	}
	if options.MinSize == 0 {
		options.MinSize = 100 * 1024 * 1024 // 100MB default
	}

	return &Scanner{
		options:  options,
		logger:   logger,
		inodeMap: make(map[int64][]*FileEntry),
	}
}

// Scan performs a filesystem scan and returns all media files
func (s *Scanner) Scan() (map[string]*FileEntry, error) {
	s.logger.Info("Starting filesystem scan",
		zap.Strings("root_paths", s.options.RootPaths),
		zap.Int64("min_size", s.options.MinSize),
	)

	result := make(map[string]*FileEntry)

	for _, rootPath := range s.options.RootPaths {
		if err := s.scanPath(rootPath, result); err != nil {
			s.logger.Error("Failed to scan path",
				zap.String("path", rootPath),
				zap.Error(err),
			)
			return nil, err
		}
	}

	// Second pass: detect hardlinks and set primary paths
	s.detectHardlinks(result)

	s.logger.Info("Filesystem scan complete",
		zap.Int("total_files", len(result)),
		zap.Int("total_inodes", len(s.inodeMap)),
	)

	return result, nil
}

// scanPath recursively scans a directory
func (s *Scanner) scanPath(rootPath string, result map[string]*FileEntry) error {
	return filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			s.logger.Warn("Error accessing path",
				zap.String("path", path),
				zap.Error(err),
			)
			return nil // Continue scanning
		}

		// Skip hidden files/directories if configured
		if s.options.SkipHidden && strings.HasPrefix(d.Name(), ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Check if file has video extension
		if !s.isVideoFile(path) {
			return nil
		}

		// Get file info
		info, err := d.Info()
		if err != nil {
			s.logger.Warn("Failed to get file info",
				zap.String("path", path),
				zap.Error(err),
			)
			return nil
		}

		// Skip small files
		if info.Size() < s.options.MinSize {
			return nil
		}

		// Get inode (Unix-specific)
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			s.logger.Warn("Failed to get inode",
				zap.String("path", path),
			)
			return nil
		}

		// Create file entry
		// Note: stat.Ino (uint64) is converted to int64 for SQLite compatibility.
		// While this may theoretically cause overflow on systems with very large inode numbers,
		// in practice, inode numbers exceeding int64 max are extremely rare on modern filesystems.
		// Negative inodes after conversion are also extremely rare and don't affect hardlink detection
		// since we only compare inodes for equality, not magnitude.
		entry := &FileEntry{
			Path:        path,
			Size:        info.Size(),
			Inode:       int64(stat.Ino),
			ModTime:     info.ModTime().Unix(),
			MediaType:   s.inferMediaType(path),
			PrimaryPath: path, // Will be updated in second pass
		}

		// Add to result
		result[path] = entry

		// Add to inode map for hardlink detection
		s.mu.Lock()
		s.inodeMap[entry.Inode] = append(s.inodeMap[entry.Inode], entry)
		s.mu.Unlock()

		return nil
	})
}

// isVideoFile checks if a file has a video extension
func (s *Scanner) isVideoFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, videoExt := range s.options.VideoExtensions {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// inferMediaType infers whether a file is a movie or series based on path
func (s *Scanner) inferMediaType(path string) string {
	lower := strings.ToLower(path)

	// Check if path contains known series indicators
	if strings.Contains(lower, "/series/") ||
		strings.Contains(lower, "/tv/") ||
		strings.Contains(lower, "/shows/") {
		return "series"
	}

	// Check if path contains known movie indicators
	if strings.Contains(lower, "/peliculas/") ||
		strings.Contains(lower, "/movies/") ||
		strings.Contains(lower, "/films/") {
		return "movie"
	}

	// Check filename for series patterns (S01E01, 1x01, etc.)
	base := filepath.Base(lower)
	if strings.Contains(base, "s0") && strings.Contains(base, "e0") {
		return "series"
	}
	if matched, _ := filepath.Match("*[0-9]x[0-9]*", base); matched {
		return "series"
	}

	// Default to movie
	return "movie"
}

// detectHardlinks detects hardlinks and sets primary paths
func (s *Scanner) detectHardlinks(entries map[string]*FileEntry) {
	hardlinkCount := 0

	for inode, files := range s.inodeMap {
		if len(files) <= 1 {
			continue // Not a hardlink
		}

		// Multiple files with same inode = hardlinks
		hardlinkCount++

		// Determine primary path (prioritize library over downloads)
		primaryPath := s.choosePrimaryPath(files)

		// Update all entries
		var allPaths []string
		for _, entry := range files {
			allPaths = append(allPaths, entry.Path)
		}

		for _, entry := range files {
			entry.IsHardlink = true
			entry.HardlinkPaths = allPaths
			entry.PrimaryPath = primaryPath
		}

		s.logger.Debug("Detected hardlink group",
			zap.Int64("inode", inode),
			zap.Int("count", len(files)),
			zap.String("primary", primaryPath),
			zap.Strings("all_paths", allPaths),
		)
	}

	s.logger.Info("Hardlink detection complete",
		zap.Int("hardlink_groups", hardlinkCount),
	)
}

// choosePrimaryPath selects the primary path from hardlinks
// Priority: Library paths > Download paths > Alphabetical
func (s *Scanner) choosePrimaryPath(files []*FileEntry) string {
	if len(files) == 0 {
		return ""
	}

	// First priority: Library paths
	for _, libraryPath := range s.options.LibraryPaths {
		for _, entry := range files {
			if strings.HasPrefix(entry.Path, libraryPath) {
				return entry.Path
			}
		}
	}

	// Second priority: Non-download paths
	for _, entry := range files {
		isDownload := false
		for _, downloadPath := range s.options.DownloadPaths {
			if strings.HasPrefix(entry.Path, downloadPath) {
				isDownload = true
				break
			}
		}
		if !isDownload {
			return entry.Path
		}
	}

	// Fallback: First path alphabetically
	primaryPath := files[0].Path
	for _, entry := range files[1:] {
		if entry.Path < primaryPath {
			primaryPath = entry.Path
		}
	}

	return primaryPath
}

// GetStats returns statistics about the scan
func (s *Scanner) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	totalFiles := 0
	totalHardlinks := 0
	movieCount := 0
	seriesCount := 0

	for _, files := range s.inodeMap {
		totalFiles += len(files)
		if len(files) > 1 {
			totalHardlinks += len(files)
		}

		if len(files) > 0 {
			if files[0].MediaType == "movie" {
				movieCount++
			} else {
				seriesCount++
			}
		}
	}

	return map[string]interface{}{
		"total_files":     totalFiles,
		"unique_inodes":   len(s.inodeMap),
		"total_hardlinks": totalHardlinks,
		"hardlink_groups": totalHardlinks / 2, // Approximate
		"movies":          movieCount,
		"series":          seriesCount,
	}
}
