package service

import (
	"testing"

	"go.uber.org/zap"
)

// TestNewHealthAnalyzer verifies that the health analyzer can be created successfully
func TestNewHealthAnalyzer(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	analyzer := NewHealthAnalyzer(logger)

	if analyzer == nil {
		t.Fatal("Expected analyzer to be created, got nil")
	}

	if analyzer.logger == nil {
		t.Error("Expected logger to be set")
	}

	if analyzer.neverWatchedThreshold != 180 {
		t.Errorf("Expected default threshold to be 180, got %d", analyzer.neverWatchedThreshold)
	}
}

// TestSetNeverWatchedThreshold verifies that the threshold can be set
func TestSetNeverWatchedThreshold(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	analyzer.SetNeverWatchedThreshold(90)

	if analyzer.neverWatchedThreshold != 90 {
		t.Errorf("Expected threshold to be 90, got %d", analyzer.neverWatchedThreshold)
	}
}

// TestHealthAnalyzer_DetectOrphanDownloads tests orphan download detection
func TestHealthAnalyzer_DetectOrphanDownloads(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	tests := []struct {
		name           string
		file           *MediaFileInfo
		expectedStatus HealthStatus
		expectedSev    Severity
		checkIssues    bool
		checkActions   bool
	}{
		{
			name: "archivo en qBT pero no en Jellyfin - orphan_download",
			file: &MediaFileInfo{
				InQBittorrent: true,
				InJellyfin:    false,
				InRadarr:      false,
				Type:          "movie",
			},
			expectedStatus: HealthStatusOrphanDownload,
			expectedSev:    SeverityWarning,
			checkIssues:    true,
			checkActions:   true,
		},
		{
			name: "archivo en qBT y Radarr pero no en Jellyfin - orphan_download",
			file: &MediaFileInfo{
				InQBittorrent: true,
				InJellyfin:    false,
				InRadarr:      true,
				Type:          "movie",
			},
			expectedStatus: HealthStatusOrphanDownload,
			expectedSev:    SeverityWarning,
			checkIssues:    true,
			checkActions:   true,
		},
		{
			name: "archivo en todos los servicios - ok",
			file: &MediaFileInfo{
				InQBittorrent: true,
				InJellyfin:    true,
				InRadarr:      true,
				Type:          "movie",
			},
			expectedStatus: HealthStatusOK,
			expectedSev:    SeverityOK,
			checkIssues:    false,
			checkActions:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.AnalyzeFile(tt.file)

			if report.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, report.Status)
			}

			if report.Severity != tt.expectedSev {
				t.Errorf("Expected severity %s, got %s", tt.expectedSev, report.Severity)
			}

			if tt.checkIssues && len(report.Issues) == 0 {
				t.Error("Expected issues to be populated for orphan download")
			}

			if tt.checkActions && len(report.Actions) == 0 {
				t.Error("Expected actions to be populated for orphan download")
			}

			// Verify specific actions for orphan downloads
			if tt.expectedStatus == HealthStatusOrphanDownload {
				hasImportAction := false
				for _, action := range report.Actions {
					if action == "import_radarr" || action == "import_sonarr" {
						hasImportAction = true
						break
					}
				}
				if !hasImportAction {
					t.Error("Expected import action for orphan download")
				}
			}
		})
	}
}

// TestHealthAnalyzer_DetectOnlyHardlinks tests hardlink-only detection
func TestHealthAnalyzer_DetectOnlyHardlinks(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	tests := []struct {
		name           string
		file           *MediaFileInfo
		expectedStatus HealthStatus
		expectedSev    Severity
	}{
		{
			name: "hardlink con torrent activo - ok",
			file: &MediaFileInfo{
				IsHardlink:     true,
				HardlinkPaths:  "/jellyfin/movies/file.mkv|/downloads/file.mkv",
				InQBittorrent:  true,
				InJellyfin:     true,
				InRadarr:       true,
				HasBeenWatched: false,
				TotalPlayCount: 0,
			},
			expectedStatus: HealthStatusOK,
			expectedSev:    SeverityOK,
		},
		{
			name: "hardlink sin torrent - only_hardlink",
			file: &MediaFileInfo{
				IsHardlink:     true,
				HardlinkPaths:  "/jellyfin/movies/file.mkv|/downloads/file.mkv",
				InQBittorrent:  false,
				InJellyfin:     true,
				InRadarr:       true,
				HasBeenWatched: false,
				TotalPlayCount: 0,
			},
			expectedStatus: HealthStatusOnlyHardlink,
			expectedSev:    SeverityWarning,
		},
		{
			name: "archivo único (no hardlink) - ok",
			file: &MediaFileInfo{
				IsHardlink:     false,
				HardlinkPaths:  "",
				InQBittorrent:  false,
				InJellyfin:     true,
				InRadarr:       true,
				HasBeenWatched: false,
				TotalPlayCount: 0,
			},
			expectedStatus: HealthStatusOK,
			expectedSev:    SeverityOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.AnalyzeFile(tt.file)

			if report.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, report.Status)
			}

			if report.Severity != tt.expectedSev {
				t.Errorf("Expected severity %s, got %s", tt.expectedSev, report.Severity)
			}

			// Verify cleanup action for only_hardlink
			if tt.expectedStatus == HealthStatusOnlyHardlink {
				hasCleanupAction := false
				for _, action := range report.Actions {
					if action == "cleanup_hardlink" {
						hasCleanupAction = true
						break
					}
				}
				if !hasCleanupAction {
					t.Error("Expected cleanup_hardlink action for only_hardlink status")
				}
			}
		})
	}
}

// TestHealthAnalyzer_DetectDeadTorrents tests dead torrent detection
func TestHealthAnalyzer_DetectDeadTorrents(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	tests := []struct {
		name           string
		file           *MediaFileInfo
		expectedStatus HealthStatus
		expectedSev    Severity
	}{
		{
			name: "estado error - dead_torrent",
			file: &MediaFileInfo{
				InQBittorrent: true,
				TorrentState:  "error",
				InJellyfin:    true,
			},
			expectedStatus: HealthStatusDeadTorrent,
			expectedSev:    SeverityCritical,
		},
		{
			name: "estado missingFiles - dead_torrent",
			file: &MediaFileInfo{
				InQBittorrent: true,
				TorrentState:  "missingFiles",
				InJellyfin:    true,
			},
			expectedStatus: HealthStatusDeadTorrent,
			expectedSev:    SeverityCritical,
		},
		{
			name: "estado pausedUP sin seeders - dead_torrent",
			file: &MediaFileInfo{
				InQBittorrent: true,
				TorrentState:  "pausedUP",
				IsSeeding:     false,
				SeedRatio:     0.05,
				InJellyfin:    true,
			},
			expectedStatus: HealthStatusDeadTorrent,
			expectedSev:    SeverityCritical,
		},
		{
			name: "estado uploading - ok",
			file: &MediaFileInfo{
				InQBittorrent:  true,
				TorrentState:   "uploading",
				IsSeeding:      true,
				SeedRatio:      1.5,
				InJellyfin:     true,
				InRadarr:       true,
				HasBeenWatched: false,
				TotalPlayCount: 0,
			},
			expectedStatus: HealthStatusOK,
			expectedSev:    SeverityOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.AnalyzeFile(tt.file)

			if report.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, report.Status)
			}

			if report.Severity != tt.expectedSev {
				t.Errorf("Expected severity %s, got %s", tt.expectedSev, report.Severity)
			}

			// Verify critical severity for dead torrents
			if tt.expectedStatus == HealthStatusDeadTorrent && report.Severity != SeverityCritical {
				t.Error("Expected critical severity for dead torrent")
			}
		})
	}
}

// TestHealthAnalyzer_DetectNeverWatched tests never watched detection
func TestHealthAnalyzer_DetectNeverWatched(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	tests := []struct {
		name           string
		file           *MediaFileInfo
		expectedStatus HealthStatus
	}{
		{
			name: "nunca visto pero play count > 0 - ok",
			file: &MediaFileInfo{
				InJellyfin:     true,
				HasBeenWatched: false,
				TotalPlayCount: 1,
				InRadarr:       true,
				InQBittorrent:  true,
			},
			expectedStatus: HealthStatusOK,
		},
		{
			name: "visto recientemente - ok",
			file: &MediaFileInfo{
				InJellyfin:     true,
				HasBeenWatched: true,
				TotalPlayCount: 5,
				InRadarr:       true,
				InQBittorrent:  true,
			},
			expectedStatus: HealthStatusOK,
		},
		{
			name: "visto hace mucho - ok",
			file: &MediaFileInfo{
				InJellyfin:     true,
				HasBeenWatched: true,
				TotalPlayCount: 2,
				InRadarr:       true,
				InQBittorrent:  true,
			},
			expectedStatus: HealthStatusOK,
		},
		{
			name: "no está en Jellyfin - no puede ser never_watched",
			file: &MediaFileInfo{
				InJellyfin:     false,
				HasBeenWatched: false,
				TotalPlayCount: 0,
				InQBittorrent:  true,
			},
			expectedStatus: HealthStatusOrphanDownload, // Will be classified as orphan instead
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.AnalyzeFile(tt.file)

			if report.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, report.Status)
			}
		})
	}
}

// TestHealthAnalyzer_GetHealthSummary tests aggregated statistics
func TestHealthAnalyzer_GetHealthSummary(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	files := []MediaFileInfo{
		// Orphan download 1
		{
			InQBittorrent: true,
			InJellyfin:    false,
			Type:          "movie",
		},
		// Orphan download 2
		{
			InQBittorrent: true,
			InJellyfin:    false,
			Type:          "series",
		},
		// Healthy file
		{
			InQBittorrent:  true,
			InJellyfin:     true,
			InRadarr:       true,
			HasBeenWatched: true,
			TotalPlayCount: 5,
		},
		// Dead torrent
		{
			InQBittorrent: true,
			TorrentState:  "error",
			InJellyfin:    true,
		},
		// Only hardlink
		{
			IsHardlink:    true,
			HardlinkPaths: "/path1|/path2",
			InQBittorrent: false,
			InJellyfin:    true,
		},
	}

	summary := analyzer.GetHealthSummary(files)

	if summary.TotalFiles != 5 {
		t.Errorf("Expected total files to be 5, got %d", summary.TotalFiles)
	}

	if summary.OrphanDownloads != 2 {
		t.Errorf("Expected orphan downloads to be 2, got %d", summary.OrphanDownloads)
	}

	if summary.Healthy != 1 {
		t.Errorf("Expected healthy files to be 1, got %d", summary.Healthy)
	}

	if summary.DeadTorrents != 1 {
		t.Errorf("Expected dead torrents to be 1, got %d", summary.DeadTorrents)
	}

	if summary.OnlyHardlinks != 1 {
		t.Errorf("Expected only hardlinks to be 1, got %d", summary.OnlyHardlinks)
	}

	// Verify needs_attention count (warnings + critical)
	// Orphan downloads (2) + only hardlinks (1) = 3 warnings
	// Dead torrents (1) = 1 critical
	// Total needs attention = 4
	if summary.NeedsAttention != 4 {
		t.Errorf("Expected needs attention to be 4, got %d", summary.NeedsAttention)
	}

	// Verify critical problems count (only dead torrents)
	if summary.CriticalProblems != 1 {
		t.Errorf("Expected critical problems to be 1, got %d", summary.CriticalProblems)
	}
}

// TestHealthAnalyzer_AnalyzeFiles tests batch analysis
func TestHealthAnalyzer_AnalyzeFiles(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	files := []MediaFileInfo{
		{
			InQBittorrent: true,
			InJellyfin:    false,
			Type:          "movie",
		},
		{
			InQBittorrent:  true,
			InJellyfin:     true,
			InRadarr:       true,
			HasBeenWatched: true,
			TotalPlayCount: 3,
		},
	}

	reports := analyzer.AnalyzeFiles(files)

	if len(reports) != 2 {
		t.Errorf("Expected 2 reports, got %d", len(reports))
	}

	if reports[0].Status != HealthStatusOrphanDownload {
		t.Errorf("Expected first report to be orphan download, got %s", reports[0].Status)
	}

	if reports[1].Status != HealthStatusOK {
		t.Errorf("Expected second report to be ok, got %s", reports[1].Status)
	}
}

// TestHealthAnalyzer_ExcludedFiles tests that excluded files are properly handled
func TestHealthAnalyzer_ExcludedFiles(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	file := &MediaFileInfo{
		InQBittorrent: true,
		InJellyfin:    false,
		Excluded:      true, // This file should be skipped
	}

	report := analyzer.AnalyzeFile(file)

	// Excluded files should always return OK status
	if report.Status != HealthStatusOK {
		t.Errorf("Expected excluded file to have OK status, got %s", report.Status)
	}

	if report.Severity != SeverityOK {
		t.Errorf("Expected excluded file to have OK severity, got %s", report.Severity)
	}
}

// TestHealthAnalyzer_UnclassifiedFiles tests unclassified file detection
func TestHealthAnalyzer_UnclassifiedFiles(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	file := &MediaFileInfo{
		InRadarr:      false,
		InSonarr:      false,
		InJellyfin:    false,
		InJellyseerr:  false,
		InQBittorrent: false,
	}

	report := analyzer.AnalyzeFile(file)

	if report.Status != HealthStatusUnclassified {
		t.Errorf("Expected unclassified status, got %s", report.Status)
	}

	if report.Severity != SeverityWarning {
		t.Errorf("Expected warning severity for unclassified file, got %s", report.Severity)
	}
}

// TestHealthAnalyzer_PriorityOrdering tests that issues are detected in the correct priority order
func TestHealthAnalyzer_PriorityOrdering(t *testing.T) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	analyzer := NewHealthAnalyzer(logger)

	// File with multiple potential issues - orphan should take priority
	file := &MediaFileInfo{
		InQBittorrent:  true,
		InJellyfin:     false, // This makes it an orphan
		IsHardlink:     true,
		HardlinkPaths:  "/path1|/path2",
		HasBeenWatched: false,
		TotalPlayCount: 0,
	}

	report := analyzer.AnalyzeFile(file)

	// Orphan download should be detected first (highest priority)
	if report.Status != HealthStatusOrphanDownload {
		t.Errorf("Expected orphan download (highest priority), got %s", report.Status)
	}
}
