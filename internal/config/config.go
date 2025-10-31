package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// EnvSourceMap tracks which config values come from environment variables
type EnvSourceMap struct {
	Radarr struct {
		Enabled bool `json:"enabled"`
		APIKey  bool `json:"api_key"`
		URL     bool `json:"url"`
	} `json:"radarr"`
	Sonarr struct {
		Enabled bool `json:"enabled"`
		APIKey  bool `json:"api_key"`
		URL     bool `json:"url"`
	} `json:"sonarr"`
	Jellyfin struct {
		Enabled bool `json:"enabled"`
		APIKey  bool `json:"api_key"`
		URL     bool `json:"url"`
	} `json:"jellyfin"`
	Jellyseerr struct {
		Enabled bool `json:"enabled"`
		APIKey  bool `json:"api_key"`
		URL     bool `json:"url"`
	} `json:"jellyseerr"`
	Jellystat struct {
		Enabled bool `json:"enabled"`
		URL     bool `json:"url"`
		APIKey  bool `json:"api_key"`
	} `json:"jellystat"`
	QBittorrent struct {
		Enabled  bool `json:"enabled"`
		Username bool `json:"username"`
		Password bool `json:"password"`
		URL      bool `json:"url"`
	} `json:"qbittorrent"`
	Bazarr struct {
		Enabled bool `json:"enabled"`
		APIKey  bool `json:"api_key"`
		URL     bool `json:"url"`
	} `json:"bazarr"`
}

// GetEnvSourceMap returns a map indicating which config values come from environment variables
func GetEnvSourceMap() *EnvSourceMap {
	envMap := &EnvSourceMap{}

	// Check Radarr
	envMap.Radarr.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_RADARR_ENABLED") != ""
	envMap.Radarr.URL = os.Getenv("KEEPERCHEKY_CLIENTS_RADARR_URL") != ""
	envMap.Radarr.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_RADARR_API_KEY") != ""

	// Check Sonarr
	envMap.Sonarr.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_SONARR_ENABLED") != ""
	envMap.Sonarr.URL = os.Getenv("KEEPERCHEKY_CLIENTS_SONARR_URL") != ""
	envMap.Sonarr.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_SONARR_API_KEY") != ""

	// Check Jellyfin
	envMap.Jellyfin.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYFIN_ENABLED") != ""
	envMap.Jellyfin.URL = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYFIN_URL") != ""
	envMap.Jellyfin.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYFIN_API_KEY") != ""

	// Check Jellyseerr
	envMap.Jellyseerr.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSEERR_ENABLED") != ""
	envMap.Jellyseerr.URL = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSEERR_URL") != ""
	envMap.Jellyseerr.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSEERR_API_KEY") != ""

	// Check Jellystat
	envMap.Jellystat.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSTAT_ENABLED") != ""
	envMap.Jellystat.URL = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSTAT_URL") != ""
	envMap.Jellystat.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_JELLYSTAT_API_KEY") != ""

	// Check qBittorrent
	envMap.QBittorrent.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_QBITTORRENT_ENABLED") != ""
	envMap.QBittorrent.URL = os.Getenv("KEEPERCHEKY_CLIENTS_QBITTORRENT_URL") != ""
	envMap.QBittorrent.Username = os.Getenv("KEEPERCHEKY_CLIENTS_QBITTORRENT_USERNAME") != ""
	envMap.QBittorrent.Password = os.Getenv("KEEPERCHEKY_CLIENTS_QBITTORRENT_PASSWORD") != ""

	// Check Bazarr
	envMap.Bazarr.Enabled = os.Getenv("KEEPERCHEKY_CLIENTS_BAZARR_ENABLED") != ""
	envMap.Bazarr.URL = os.Getenv("KEEPERCHEKY_CLIENTS_BAZARR_URL") != ""
	envMap.Bazarr.APIKey = os.Getenv("KEEPERCHEKY_CLIENTS_BAZARR_API_KEY") != ""

	return envMap
}

type Config struct {
	App        AppConfig        `mapstructure:"app" yaml:"app"`
	Server     ServerConfig     `mapstructure:"server" yaml:"server"`
	Database   DatabaseConfig   `mapstructure:"database" yaml:"database"`
	Clients    ClientsConfig    `mapstructure:"clients" yaml:"clients"`
	Cleanup    CleanupConfig    `mapstructure:"cleanup" yaml:"cleanup"`
	Filesystem FilesystemConfig `mapstructure:"filesystem" yaml:"filesystem"`
}

type AppConfig struct {
	Environment      string `mapstructure:"environment" yaml:"environment"`
	LogLevel         string `mapstructure:"log_level" yaml:"log_level"`
	SchedulerEnabled bool   `mapstructure:"scheduler_enabled" yaml:"scheduler_enabled"`
}

type CleanupConfig struct {
	DryRun            bool     `mapstructure:"dry_run" yaml:"dry_run"`
	DaysToKeep        int      `mapstructure:"days_to_keep" yaml:"days_to_keep"`
	LeavingSoonDays   int      `mapstructure:"leaving_soon_days" yaml:"leaving_soon_days"`
	ExclusionTags     []string `mapstructure:"exclusion_tags" yaml:"exclusion_tags"`
	DeleteUnmonitored bool     `mapstructure:"delete_unmonitored" yaml:"delete_unmonitored"`
}

type FilesystemConfig struct {
	ScanEnabled     bool     `mapstructure:"scan_enabled" yaml:"scan_enabled"`         // Enable filesystem scanning
	RootPaths       []string `mapstructure:"root_paths" yaml:"root_paths"`             // Paths to scan
	LibraryPaths    []string `mapstructure:"library_paths" yaml:"library_paths"`       // Priority paths (libraries)
	DownloadPaths   []string `mapstructure:"download_paths" yaml:"download_paths"`     // Download paths (lower priority)
	VideoExtensions []string `mapstructure:"video_extensions" yaml:"video_extensions"` // Video file extensions
	MinSizeMB       int64    `mapstructure:"min_size_mb" yaml:"min_size_mb"`           // Minimum file size in MB
	SkipHidden      bool     `mapstructure:"skip_hidden" yaml:"skip_hidden"`           // Skip hidden files/dirs
}

type ServerConfig struct {
	Port string `mapstructure:"port" yaml:"port"`
	Host string `mapstructure:"host" yaml:"host"`
}

type DatabaseConfig struct {
	Type string `mapstructure:"type" yaml:"type"` // sqlite or postgres
	Path string `mapstructure:"path" yaml:"path"` // for sqlite
	Host string `mapstructure:"host" yaml:"host"` // for postgres
	Port string `mapstructure:"port" yaml:"port"`
	User string `mapstructure:"user" yaml:"user"`
	Pass string `mapstructure:"password" yaml:"password"`
	Name string `mapstructure:"name" yaml:"name"`
}

type ClientsConfig struct {
	Radarr      ServiceClient     `mapstructure:"radarr" yaml:"radarr"`
	Sonarr      ServiceClient     `mapstructure:"sonarr" yaml:"sonarr"`
	Jellyfin    ServiceClient     `mapstructure:"jellyfin" yaml:"jellyfin"`
	Jellyseerr  ServiceClient     `mapstructure:"jellyseerr" yaml:"jellyseerr"`
	Jellystat   JellystatClient   `mapstructure:"jellystat" yaml:"jellystat"`
	QBittorrent QBittorrentClient `mapstructure:"qbittorrent" yaml:"qbittorrent"`
	Bazarr      ServiceClient     `mapstructure:"bazarr" yaml:"bazarr"`
}

type ServiceClient struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	URL     string `mapstructure:"url" yaml:"url"`
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
}

type QBittorrentClient struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	URL      string `mapstructure:"url" yaml:"url"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
}

type JellystatClient struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	URL     string `mapstructure:"url" yaml:"url"`
	APIKey  string `mapstructure:"api_key" yaml:"api_key"`
}

func Load() (*Config, error) {
	// 1. Set defaults FIRST (lowest priority)
	setDefaults()

	// 2. Configure environment variable handling
	// CRITICAL: SetEnvKeyReplacer allows KEEPERCHEKY_CLIENTS_RADARR_URL to map to clients.radarr.url
	viper.SetEnvPrefix("KEEPERCHEKY")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// Bind environment variables WITHOUT prefix (for external service credentials)
	// These also take precedence over config file values
	_ = viper.BindEnv("clients.radarr.api_key", "RADARR_API_KEY")
	_ = viper.BindEnv("clients.sonarr.api_key", "SONARR_API_KEY")
	_ = viper.BindEnv("clients.jellyfin.api_key", "JELLYFIN_API_KEY")
	_ = viper.BindEnv("clients.jellyseerr.api_key", "JELLYSEERR_API_KEY")
	_ = viper.BindEnv("clients.jellystat.api_key", "JELLYSTAT_API_KEY")
	_ = viper.BindEnv("clients.qbittorrent.username", "QBITTORRENT_USERNAME")
	_ = viper.BindEnv("clients.qbittorrent.password", "QBITTORRENT_PASSWORD")
	_ = viper.BindEnv("clients.bazarr.api_key", "BAZARR_API_KEY")

	// 3. Configure config file paths
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/app/config") // Primary location (mounted volume)
	viper.AddConfigPath("./config")    // Fallback for local development
	viper.AddConfigPath(".")

	// 4. Try to read config file (medium priority - overrides defaults, but ENV vars override this)
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found is OK, we'll use defaults + env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		fmt.Println("No config file found, using defaults and environment variables")
	} else {
		fmt.Printf("Loaded config from: %s\n", viper.ConfigFileUsed())
	}

	// 5. Unmarshal into Config struct
	// This automatically merges: Defaults < Config File < Environment Variables
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// 6. CRITICAL: Check if environment variables have DIFFERENT values than config.yaml
	// If so, PHYSICALLY overwrite config.yaml with .env values
	envSources := GetEnvSourceMap()
	if hasAnyEnvVars(envSources) {
		fmt.Println("🔧 Configuration overrides from environment variables:")
		logEnvOverrides(envSources)

		// Read the ORIGINAL config.yaml values (before env override)
		originalConfig := &Config{}

		// Create a separate viper instance to read ONLY the file (no env vars)
		fileViper := viper.New()
		fileViper.SetConfigName("config")
		fileViper.SetConfigType("yaml")
		fileViper.AddConfigPath("/app/config")
		fileViper.AddConfigPath("./config")
		fileViper.AddConfigPath(".")

		needsSave := false

		// Try to read the original file
		if err := fileViper.ReadInConfig(); err == nil {
			if err := fileViper.Unmarshal(originalConfig); err == nil {
				// Compare and check if ANY value is different
				needsSave = configHasChanges(originalConfig, &config, envSources)
			} else {
				// If we can't read the file, assume we need to save
				needsSave = true
			}
		} else {
			// No config file exists, so we need to create it
			needsSave = true
		}

		if needsSave {
			fmt.Println("💾 Detected changes from .env, syncing to config.yaml...")
			if err := Save(&config); err != nil {
				fmt.Printf("⚠️  Warning: Could not save config.yaml: %v\n", err)
			} else {
				fmt.Println("✅ Config synced: .env values written to config.yaml")
			}
		} else {
			fmt.Println("ℹ️  Config.yaml already up-to-date, no save needed")
		}
	}

	return &config, nil
}

// configHasChanges compares original config.yaml values with merged config (including env vars)
// Returns true if ANY value from environment variables is DIFFERENT from the file
func configHasChanges(original, merged *Config, envSources *EnvSourceMap) bool {
	// Check each field that has an env var override
	if envSources.Radarr.Enabled && original.Clients.Radarr.Enabled != merged.Clients.Radarr.Enabled {
		fmt.Printf("  📝 Change detected: radarr.enabled (%v → %v)\n", original.Clients.Radarr.Enabled, merged.Clients.Radarr.Enabled)
		return true
	}
	if envSources.Radarr.URL && original.Clients.Radarr.URL != merged.Clients.Radarr.URL {
		fmt.Printf("  📝 Change detected: radarr.url (%s → %s)\n", original.Clients.Radarr.URL, merged.Clients.Radarr.URL)
		return true
	}
	if envSources.Radarr.APIKey && original.Clients.Radarr.APIKey != merged.Clients.Radarr.APIKey {
		fmt.Printf("  📝 Change detected: radarr.api_key (****** → ******)\n")
		return true
	}

	if envSources.Sonarr.Enabled && original.Clients.Sonarr.Enabled != merged.Clients.Sonarr.Enabled {
		fmt.Printf("  📝 Change detected: sonarr.enabled (%v → %v)\n", original.Clients.Sonarr.Enabled, merged.Clients.Sonarr.Enabled)
		return true
	}
	if envSources.Sonarr.URL && original.Clients.Sonarr.URL != merged.Clients.Sonarr.URL {
		fmt.Printf("  📝 Change detected: sonarr.url (%s → %s)\n", original.Clients.Sonarr.URL, merged.Clients.Sonarr.URL)
		return true
	}
	if envSources.Sonarr.APIKey && original.Clients.Sonarr.APIKey != merged.Clients.Sonarr.APIKey {
		fmt.Printf("  📝 Change detected: sonarr.api_key (****** → ******)\n")
		return true
	}

	if envSources.Jellyfin.Enabled && original.Clients.Jellyfin.Enabled != merged.Clients.Jellyfin.Enabled {
		fmt.Printf("  📝 Change detected: jellyfin.enabled (%v → %v)\n", original.Clients.Jellyfin.Enabled, merged.Clients.Jellyfin.Enabled)
		return true
	}
	if envSources.Jellyfin.URL && original.Clients.Jellyfin.URL != merged.Clients.Jellyfin.URL {
		fmt.Printf("  📝 Change detected: jellyfin.url (%s → %s)\n", original.Clients.Jellyfin.URL, merged.Clients.Jellyfin.URL)
		return true
	}
	if envSources.Jellyfin.APIKey && original.Clients.Jellyfin.APIKey != merged.Clients.Jellyfin.APIKey {
		fmt.Printf("  📝 Change detected: jellyfin.api_key (****** → ******)\n")
		return true
	}

	if envSources.Jellyseerr.Enabled && original.Clients.Jellyseerr.Enabled != merged.Clients.Jellyseerr.Enabled {
		fmt.Printf("  📝 Change detected: jellyseerr.enabled (%v → %v)\n", original.Clients.Jellyseerr.Enabled, merged.Clients.Jellyseerr.Enabled)
		return true
	}
	if envSources.Jellyseerr.URL && original.Clients.Jellyseerr.URL != merged.Clients.Jellyseerr.URL {
		fmt.Printf("  📝 Change detected: jellyseerr.url (%s → %s)\n", original.Clients.Jellyseerr.URL, merged.Clients.Jellyseerr.URL)
		return true
	}
	if envSources.Jellyseerr.APIKey && original.Clients.Jellyseerr.APIKey != merged.Clients.Jellyseerr.APIKey {
		fmt.Printf("  📝 Change detected: jellyseerr.api_key (****** → ******)\n")
		return true
	}

	if envSources.Jellystat.Enabled && original.Clients.Jellystat.Enabled != merged.Clients.Jellystat.Enabled {
		fmt.Printf("  📝 Change detected: jellystat.enabled (%v → %v)\n", original.Clients.Jellystat.Enabled, merged.Clients.Jellystat.Enabled)
		return true
	}
	if envSources.Jellystat.URL && original.Clients.Jellystat.URL != merged.Clients.Jellystat.URL {
		fmt.Printf("  📝 Change detected: jellystat.url (%s → %s)\n", original.Clients.Jellystat.URL, merged.Clients.Jellystat.URL)
		return true
	}
	if envSources.Jellystat.APIKey && original.Clients.Jellystat.APIKey != merged.Clients.Jellystat.APIKey {
		fmt.Printf("  📝 Change detected: jellystat.api_key (****** → ******)\n")
		return true
	}

	if envSources.QBittorrent.Enabled && original.Clients.QBittorrent.Enabled != merged.Clients.QBittorrent.Enabled {
		fmt.Printf("  📝 Change detected: qbittorrent.enabled (%v → %v)\n", original.Clients.QBittorrent.Enabled, merged.Clients.QBittorrent.Enabled)
		return true
	}
	if envSources.QBittorrent.URL && original.Clients.QBittorrent.URL != merged.Clients.QBittorrent.URL {
		fmt.Printf("  📝 Change detected: qbittorrent.url (%s → %s)\n", original.Clients.QBittorrent.URL, merged.Clients.QBittorrent.URL)
		return true
	}
	if envSources.QBittorrent.Username && original.Clients.QBittorrent.Username != merged.Clients.QBittorrent.Username {
		fmt.Printf("  📝 Change detected: qbittorrent.username (%s → %s)\n", original.Clients.QBittorrent.Username, merged.Clients.QBittorrent.Username)
		return true
	}
	if envSources.QBittorrent.Password && original.Clients.QBittorrent.Password != merged.Clients.QBittorrent.Password {
		fmt.Printf("  📝 Change detected: qbittorrent.password (****** → ******)\n")
		return true
	}

	if envSources.Bazarr.Enabled && original.Clients.Bazarr.Enabled != merged.Clients.Bazarr.Enabled {
		fmt.Printf("  📝 Change detected: bazarr.enabled (%v → %v)\n", original.Clients.Bazarr.Enabled, merged.Clients.Bazarr.Enabled)
		return true
	}
	if envSources.Bazarr.URL && original.Clients.Bazarr.URL != merged.Clients.Bazarr.URL {
		fmt.Printf("  📝 Change detected: bazarr.url (%s → %s)\n", original.Clients.Bazarr.URL, merged.Clients.Bazarr.URL)
		return true
	}
	if envSources.Bazarr.APIKey && original.Clients.Bazarr.APIKey != merged.Clients.Bazarr.APIKey {
		fmt.Printf("  📝 Change detected: bazarr.api_key (****** → ******)\n")
		return true
	}

	// No changes detected
	return false
} // hasAnyEnvVars checks if any environment variables are set
func hasAnyEnvVars(envMap *EnvSourceMap) bool {
	return envMap.Radarr.Enabled || envMap.Radarr.URL || envMap.Radarr.APIKey ||
		envMap.Sonarr.Enabled || envMap.Sonarr.URL || envMap.Sonarr.APIKey ||
		envMap.Jellyfin.Enabled || envMap.Jellyfin.URL || envMap.Jellyfin.APIKey ||
		envMap.Jellyseerr.Enabled || envMap.Jellyseerr.URL || envMap.Jellyseerr.APIKey ||
		envMap.Jellystat.Enabled || envMap.Jellystat.URL || envMap.Jellystat.APIKey ||
		envMap.QBittorrent.Enabled || envMap.QBittorrent.URL || envMap.QBittorrent.Username || envMap.QBittorrent.Password ||
		envMap.Bazarr.Enabled || envMap.Bazarr.URL || envMap.Bazarr.APIKey
}

// logEnvOverrides prints which config values come from environment variables
func logEnvOverrides(envMap *EnvSourceMap) {
	if envMap.Radarr.Enabled {
		fmt.Println("  - clients.radarr.enabled (from KEEPERCHEKY_CLIENTS_RADARR_ENABLED)")
	}
	if envMap.Radarr.URL {
		fmt.Println("  - clients.radarr.url (from KEEPERCHEKY_CLIENTS_RADARR_URL)")
	}
	if envMap.Radarr.APIKey {
		fmt.Println("  - clients.radarr.api_key (from RADARR_API_KEY)")
	}
	if envMap.Sonarr.Enabled {
		fmt.Println("  - clients.sonarr.enabled (from KEEPERCHEKY_CLIENTS_SONARR_ENABLED)")
	}
	if envMap.Sonarr.URL {
		fmt.Println("  - clients.sonarr.url (from KEEPERCHEKY_CLIENTS_SONARR_URL)")
	}
	if envMap.Sonarr.APIKey {
		fmt.Println("  - clients.sonarr.api_key (from SONARR_API_KEY)")
	}
	if envMap.Jellyfin.Enabled {
		fmt.Println("  - clients.jellyfin.enabled (from KEEPERCHEKY_CLIENTS_JELLYFIN_ENABLED)")
	}
	if envMap.Jellyfin.URL {
		fmt.Println("  - clients.jellyfin.url (from KEEPERCHEKY_CLIENTS_JELLYFIN_URL)")
	}
	if envMap.Jellyfin.APIKey {
		fmt.Println("  - clients.jellyfin.api_key (from JELLYFIN_API_KEY)")
	}
	if envMap.Jellyseerr.Enabled {
		fmt.Println("  - clients.jellyseerr.enabled (from KEEPERCHEKY_CLIENTS_JELLYSEERR_ENABLED)")
	}
	if envMap.Jellyseerr.URL {
		fmt.Println("  - clients.jellyseerr.url (from KEEPERCHEKY_CLIENTS_JELLYSEERR_URL)")
	}
	if envMap.Jellyseerr.APIKey {
		fmt.Println("  - clients.jellyseerr.api_key (from JELLYSEERR_API_KEY)")
	}
	if envMap.Jellystat.Enabled {
		fmt.Println("  - clients.jellystat.enabled (from KEEPERCHEKY_CLIENTS_JELLYSTAT_ENABLED)")
	}
	if envMap.Jellystat.URL {
		fmt.Println("  - clients.jellystat.url (from KEEPERCHEKY_CLIENTS_JELLYSTAT_URL)")
	}
	if envMap.Jellystat.APIKey {
		fmt.Println("  - clients.jellystat.api_key (from JELLYSTAT_API_KEY)")
	}
	if envMap.QBittorrent.Enabled {
		fmt.Println("  - clients.qbittorrent.enabled (from KEEPERCHEKY_CLIENTS_QBITTORRENT_ENABLED)")
	}
	if envMap.QBittorrent.URL {
		fmt.Println("  - clients.qbittorrent.url (from KEEPERCHEKY_CLIENTS_QBITTORRENT_URL)")
	}
	if envMap.QBittorrent.Username {
		fmt.Println("  - clients.qbittorrent.username (from QBITTORRENT_USERNAME)")
	}
	if envMap.QBittorrent.Password {
		fmt.Println("  - clients.qbittorrent.password (from QBITTORRENT_PASSWORD)")
	}
	if envMap.Bazarr.Enabled {
		fmt.Println("  - clients.bazarr.enabled (from KEEPERCHEKY_CLIENTS_BAZARR_ENABLED)")
	}
	if envMap.Bazarr.URL {
		fmt.Println("  - clients.bazarr.url (from KEEPERCHEKY_CLIENTS_BAZARR_URL)")
	}
	if envMap.Bazarr.APIKey {
		fmt.Println("  - clients.bazarr.api_key (from BAZARR_API_KEY)")
	}
}

// Save writes the current configuration to a YAML file
// This PHYSICALLY overwrites config.yaml with values from environment variables
func Save(cfg *Config) error {
	// Update viper's internal state with the merged config
	// This includes all values from .env that were loaded via AutomaticEnv()
	viper.Set("app", cfg.App)
	viper.Set("server", cfg.Server)
	viper.Set("database", cfg.Database)
	viper.Set("clients", cfg.Clients)
	viper.Set("cleanup", cfg.Cleanup)

	// Determine the config file path
	configPaths := []string{
		"/app/config/config.yaml", // Docker container path
		"./config/config.yaml",    // Local development
		"./config.yaml",           // Fallback
	}

	var configPath string
	var writeErr error

	// Try each path until one works
	for _, path := range configPaths {
		// Check if directory exists and is writable
		dir := filepath.Dir(path)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Try to create directory
			if err := os.MkdirAll(dir, 0755); err != nil {
				continue
			}
		}

		// Try to write to this path
		writeErr = viper.WriteConfigAs(path)
		if writeErr == nil {
			configPath = path
			fmt.Printf("📝 Config saved to: %s\n", configPath)
			break
		}
	}

	if writeErr != nil {
		return fmt.Errorf("failed to write config to any path: %w", writeErr)
	}

	return nil
}

func setDefaults() {
	// App defaults
	viper.SetDefault("app.environment", "development")
	viper.SetDefault("app.log_level", "info")
	viper.SetDefault("app.scheduler_enabled", false)

	// Server defaults
	viper.SetDefault("server.port", "8000")
	viper.SetDefault("server.host", "0.0.0.0")

	// Database defaults
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./data/keepercheky.db")

	// Cleanup defaults
	viper.SetDefault("cleanup.dry_run", true)
	viper.SetDefault("cleanup.days_to_keep", 90)
	viper.SetDefault("cleanup.leaving_soon_days", 7)
	viper.SetDefault("cleanup.delete_unmonitored", false)

	// Clients defaults (all disabled by default)
	viper.SetDefault("clients.radarr.enabled", false)
	viper.SetDefault("clients.sonarr.enabled", false)
	viper.SetDefault("clients.jellyfin.enabled", false)
	viper.SetDefault("clients.jellyseerr.enabled", false)
	viper.SetDefault("clients.jellystat.enabled", false)
	viper.SetDefault("clients.qbittorrent.enabled", false)
	viper.SetDefault("clients.bazarr.enabled", false)

	// Filesystem defaults
	viper.SetDefault("filesystem.scan_enabled", false)
	viper.SetDefault("filesystem.root_paths", []string{"/BibliotecaMultimedia"})
	viper.SetDefault("filesystem.library_paths", []string{
		"/BibliotecaMultimedia/Peliculas",
		"/BibliotecaMultimedia/Series",
	})
	viper.SetDefault("filesystem.download_paths", []string{
		"/BibliotecaMultimedia/Descargas/Peliculas",
		"/BibliotecaMultimedia/Descargas/Series",
	})
	viper.SetDefault("filesystem.video_extensions", []string{".mkv", ".mp4", ".avi", ".m4v", ".ts", ".m2ts"})
	viper.SetDefault("filesystem.min_size_mb", 100)
	viper.SetDefault("filesystem.skip_hidden", true)
}
