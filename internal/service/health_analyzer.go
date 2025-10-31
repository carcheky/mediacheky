package service

import (
	"fmt"

	"go.uber.org/zap"
)

// HealthStatus represents the health status of a file
type HealthStatus string

const (
	HealthStatusOK             HealthStatus = "ok"
	HealthStatusOrphanDownload HealthStatus = "orphan_download"
	HealthStatusOnlyHardlink   HealthStatus = "only_hardlink"
	HealthStatusDeadTorrent    HealthStatus = "dead_torrent"
	HealthStatusNeverWatched   HealthStatus = "never_watched"
	HealthStatusUnclassified   HealthStatus = "unclassified"
)

// Severity levels for health issues
type Severity string

const (
	SeverityOK       Severity = "ok"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// MediaFileInfo is duplicated here to avoid import cycle
// This should match handler.MediaFileInfo structure
type MediaFileInfo struct {
	ID              uint     `json:"id"`
	Title           string   `json:"title"`
	Type            string   `json:"type"`
	FilePath        string   `json:"file_path"`
	Size            int64    `json:"size"`
	PosterURL       string   `json:"poster_url"`
	Quality         string   `json:"quality"`
	IsHardlink      bool     `json:"is_hardlink"`
	HardlinkPaths   string   `json:"hardlink_paths"`
	PrimaryPath     string   `json:"primary_path"`
	InRadarr        bool     `json:"in_radarr"`
	InSonarr        bool     `json:"in_sonarr"`
	InJellyfin      bool     `json:"in_jellyfin"`
	InJellyseerr    bool     `json:"in_jellyseerr"`
	InQBittorrent   bool     `json:"in_qbittorrent"`
	RadarrID        *int     `json:"radarr_id"`
	SonarrID        *int     `json:"sonarr_id"`
	JellyfinID      *string  `json:"jellyfin_id"`
	JellyseerrID    *int     `json:"jellyseerr_id"`
	TorrentHash     string   `json:"torrent_hash"`
	TorrentCategory string   `json:"torrent_category"`
	TorrentState    string   `json:"torrent_state"`
	TorrentTags     string   `json:"torrent_tags"`
	IsSeeding       bool     `json:"is_seeding"`
	SeedRatio       float64  `json:"seed_ratio"`
	HasBeenWatched  bool     `json:"has_been_watched"`
	TotalPlayCount  int      `json:"total_play_count"`
	Tags            []string `json:"tags"`
	Excluded        bool     `json:"excluded"`
}

// FileHealthReport represents the health analysis of a single file
type FileHealthReport struct {
	File        *MediaFileInfo `json:"file"`
	Status      HealthStatus   `json:"status"`
	Severity    Severity       `json:"severity"`
	Issues      []string       `json:"issues"`
	Suggestions []string       `json:"suggestions"`
	Actions     []string       `json:"actions"` // Available actions: import_radarr, import_sonarr, delete, ignore, cleanup_hardlink
}

// HealthSummary represents aggregated health statistics
type HealthSummary struct {
	Healthy          int `json:"healthy"`
	OrphanDownloads  int `json:"orphan_downloads"`
	OnlyHardlinks    int `json:"only_hardlinks"`
	DeadTorrents     int `json:"dead_torrents"`
	NeverWatched     int `json:"never_watched"`
	Unclassified     int `json:"unclassified"`
	TotalFiles       int `json:"total_files"`
	NeedsAttention   int `json:"needs_attention"`   // warning + critical
	CriticalProblems int `json:"critical_problems"` // only critical
}

// HealthAnalyzer analyzes files and provides health reports
type HealthAnalyzer struct {
	logger                *zap.Logger
	neverWatchedThreshold int // days threshold for "never watched" classification
}

// NewHealthAnalyzer creates a new health analyzer
func NewHealthAnalyzer(logger *zap.Logger) *HealthAnalyzer {
	return &HealthAnalyzer{
		logger:                logger,
		neverWatchedThreshold: 180, // 6 months default
	}
}

// SetNeverWatchedThreshold sets the threshold in days for never watched classification
func (ha *HealthAnalyzer) SetNeverWatchedThreshold(days int) {
	ha.neverWatchedThreshold = days
}

// AnalyzeFile performs health analysis on a single file
func (ha *HealthAnalyzer) AnalyzeFile(file *MediaFileInfo) *FileHealthReport {
	report := &FileHealthReport{
		File:        file,
		Status:      HealthStatusOK,
		Severity:    SeverityOK,
		Issues:      []string{},
		Suggestions: []string{},
		Actions:     []string{},
	}

	// Skip excluded files
	if file.Excluded {
		return report
	}

	// Check for orphan downloads (highest priority)
	if ha.isOrphanDownload(file) {
		ha.markAsOrphanDownload(report, file)
		return report
	}

	// Check for dead torrents (critical issue)
	if ha.isDeadTorrent(file) {
		ha.markAsDeadTorrent(report, file)
		return report
	}

	// Check for only hardlinks (can free up space)
	if ha.isOnlyHardlink(file) {
		ha.markAsOnlyHardlink(report, file)
		return report
	}

	// Check for never watched (lower priority)
	if ha.isNeverWatched(file) {
		ha.markAsNeverWatched(report, file)
		return report
	}

	// Check for unclassified files (no service metadata)
	if ha.isUnclassified(file) {
		ha.markAsUnclassified(report, file)
		return report
	}

	// File is healthy
	report.Suggestions = []string{"Archivo correctamente gestionado"}
	return report
}

// AnalyzeFiles performs health analysis on multiple files
func (ha *HealthAnalyzer) AnalyzeFiles(files []MediaFileInfo) []*FileHealthReport {
	reports := make([]*FileHealthReport, 0, len(files))

	for i := range files {
		report := ha.AnalyzeFile(&files[i])
		reports = append(reports, report)
	}

	ha.logger.Info("Health analysis complete",
		zap.Int("total_files", len(files)),
		zap.Int("analyzed", len(reports)),
	)

	return reports
}

// GetHealthSummary generates aggregated health statistics
func (ha *HealthAnalyzer) GetHealthSummary(files []MediaFileInfo) *HealthSummary {
	summary := &HealthSummary{
		TotalFiles: len(files),
	}

	for i := range files {
		report := ha.AnalyzeFile(&files[i])

		switch report.Status {
		case HealthStatusOK:
			summary.Healthy++
		case HealthStatusOrphanDownload:
			summary.OrphanDownloads++
		case HealthStatusOnlyHardlink:
			summary.OnlyHardlinks++
		case HealthStatusDeadTorrent:
			summary.DeadTorrents++
		case HealthStatusNeverWatched:
			summary.NeverWatched++
		case HealthStatusUnclassified:
			summary.Unclassified++
		}

		if report.Severity == SeverityWarning {
			summary.NeedsAttention++
		} else if report.Severity == SeverityCritical {
			summary.CriticalProblems++
			summary.NeedsAttention++
		}
	}

	ha.logger.Info("Health summary generated",
		zap.Int("healthy", summary.Healthy),
		zap.Int("needs_attention", summary.NeedsAttention),
		zap.Int("critical", summary.CriticalProblems),
	)

	return summary
}

// isOrphanDownload checks if file is in qBittorrent but not in media library
func (ha *HealthAnalyzer) isOrphanDownload(file *MediaFileInfo) bool {
	// File is in qBittorrent downloads but not imported to Jellyfin
	return file.InQBittorrent && !file.InJellyfin
}

// isDeadTorrent checks if torrent is in error state or missing files
func (ha *HealthAnalyzer) isDeadTorrent(file *MediaFileInfo) bool {
	if !file.InQBittorrent {
		return false
	}

	// Check for error states
	errorStates := []string{"error", "missingFiles", "metaDL"}
	for _, state := range errorStates {
		if file.TorrentState == state {
			return true
		}
	}

	// Check if torrent is paused and not seeding (potential dead torrent)
	if file.TorrentState == "pausedUP" && !file.IsSeeding && file.SeedRatio < 0.1 {
		return true
	}

	return false
}

// isOnlyHardlink checks if file is only a hardlink with original deleted
func (ha *HealthAnalyzer) isOnlyHardlink(file *MediaFileInfo) bool {
	// File has hardlinks but original torrent is gone
	// This means we can safely remove the hardlink from downloads without losing data
	return file.IsHardlink &&
		file.HardlinkPaths != "" &&
		!file.InQBittorrent &&
		file.InJellyfin
}

// isNeverWatched checks if file has never been watched and is old
func (ha *HealthAnalyzer) isNeverWatched(file *MediaFileInfo) bool {
	// File is in Jellyfin but never watched and older than threshold
	if !file.InJellyfin || file.HasBeenWatched {
		return false
	}

	// Check if we have last watched date or total play count
	if file.TotalPlayCount > 0 {
		return false
	}

	// Calculate age (we would need AddedDate in MediaFileInfo, using a placeholder check)
	// TODO: Add AddedDate field to MediaFileInfo and use proper date comparison
	// For now, return false as we don't have the data yet
	return false
}

// isUnclassified checks if file has no service metadata
func (ha *HealthAnalyzer) isUnclassified(file *MediaFileInfo) bool {
	// File is not tracked by any service
	return !file.InRadarr &&
		!file.InSonarr &&
		!file.InJellyfin &&
		!file.InJellyseerr &&
		!file.InQBittorrent
}

// markAsOrphanDownload marks file as orphan download with suggestions
func (ha *HealthAnalyzer) markAsOrphanDownload(report *FileHealthReport, file *MediaFileInfo) {
	report.Status = HealthStatusOrphanDownload
	report.Severity = SeverityWarning
	report.Issues = []string{
		"Archivo descargado pero no está en la biblioteca de Jellyfin",
	}

	// Determine which service to import to
	if file.Type == "movie" || (!file.InSonarr && !file.InRadarr) {
		report.Suggestions = []string{
			"Importar a Radarr para agregarlo automáticamente a Jellyfin",
			"Si es una serie, importar a Sonarr en su lugar",
		}
		report.Actions = []string{"import_radarr", "import_sonarr", "delete", "ignore"}
	} else if file.Type == "series" || file.Type == "episode" {
		report.Suggestions = []string{
			"Importar a Sonarr para agregarlo automáticamente a Jellyfin",
		}
		report.Actions = []string{"import_sonarr", "delete", "ignore"}
	} else {
		report.Suggestions = []string{
			"Clasificar como película o serie e importar al servicio correspondiente",
		}
		report.Actions = []string{"import_radarr", "import_sonarr", "delete", "ignore"}
	}

	// Add service status to issues
	if !file.InRadarr && !file.InSonarr {
		report.Issues = append(report.Issues, "No gestionado por Radarr ni Sonarr")
	}
}

// markAsDeadTorrent marks file as having a dead torrent
func (ha *HealthAnalyzer) markAsDeadTorrent(report *FileHealthReport, file *MediaFileInfo) {
	report.Status = HealthStatusDeadTorrent
	report.Severity = SeverityCritical
	report.Issues = []string{
		"Torrent en estado de error: " + file.TorrentState,
	}

	if file.InJellyfin {
		report.Suggestions = []string{
			"El archivo está en tu biblioteca pero el torrent tiene errores",
			"Puedes eliminar el torrent de forma segura si el archivo en Jellyfin funciona correctamente",
		}
		report.Actions = []string{"delete", "ignore"}
	} else {
		report.Suggestions = []string{
			"Torrent con errores y archivo no está en biblioteca",
			"Recomendado: eliminar y re-descargar si es necesario",
		}
		report.Actions = []string{"delete"}
	}

	report.Issues = append(report.Issues, "Ratio: "+formatFloat(file.SeedRatio))
}

// markAsOnlyHardlink marks file as having only hardlinks
func (ha *HealthAnalyzer) markAsOnlyHardlink(report *FileHealthReport, file *MediaFileInfo) {
	report.Status = HealthStatusOnlyHardlink
	report.Severity = SeverityWarning
	report.Issues = []string{
		"El torrent original fue eliminado pero quedan hardlinks",
	}
	report.Suggestions = []string{
		"Puedes eliminar el hardlink de descargas sin perder el archivo en Jellyfin",
		"Esto liberará espacio en la carpeta de descargas (aunque el archivo real ya está solo en biblioteca)",
	}
	report.Actions = []string{"cleanup_hardlink", "ignore"}
}

// markAsNeverWatched marks file as never watched
func (ha *HealthAnalyzer) markAsNeverWatched(report *FileHealthReport, file *MediaFileInfo) {
	report.Status = HealthStatusNeverWatched
	report.Severity = SeverityWarning
	report.Issues = []string{
		"Archivo nunca reproducido en Jellyfin",
	}
	report.Suggestions = []string{
		"Considerar eliminarlo si lleva mucho tiempo sin verse",
		"Revisar si sigue siendo de interés",
	}
	report.Actions = []string{"delete", "ignore"}
}

// markAsUnclassified marks file as unclassified
func (ha *HealthAnalyzer) markAsUnclassified(report *FileHealthReport, file *MediaFileInfo) {
	report.Status = HealthStatusUnclassified
	report.Severity = SeverityWarning
	report.Issues = []string{
		"Archivo no está registrado en ningún servicio",
	}
	report.Suggestions = []string{
		"Clasificar e importar a Radarr/Sonarr según corresponda",
		"O eliminar si no es necesario",
	}
	report.Actions = []string{"import_radarr", "import_sonarr", "delete", "ignore"}
}

// Helper function to format float
func formatFloat(f float64) string {
	if f == 0 {
		return "0.00"
	}
	return fmt.Sprintf("%.2f", f)
}
