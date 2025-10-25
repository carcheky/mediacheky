package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration
type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	Services ServicesConfig
}

// AppConfig holds general application settings
type AppConfig struct {
	Environment string
	LogLevel    string
}

// ServerConfig holds HTTP server settings
type ServerConfig struct {
	Host string
	Port string
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Type string
	Path string
}

// ServicesConfig holds configuration for all multimedia services
type ServicesConfig struct {
	Jellyfin         ServiceConfig
	Sonarr           ServiceConfig
	Radarr           ServiceConfig
	Prowlarr         ServiceConfig
	QBittorrent      ServiceConfig
	Jellyseerr       ServiceConfig
	QBitManager      ServiceConfig
	DockerController ServiceConfig
	Bazarr           ServiceConfig
	Jellystat        ServiceConfig
}

// ServiceConfig holds configuration for a single service
type ServiceConfig struct {
	Enabled     bool
	Image       string
	Port        string
	Environment map[string]string
	Volumes     []string
}

// Load loads configuration from environment variables and config file
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Configure environment variables
	v.SetEnvPrefix("MEDIACHEKY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Try to read config file
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("/config")
	v.AddConfigPath("./config")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// App defaults
	v.SetDefault("app.environment", "production")
	v.SetDefault("app.loglevel", "info")

	// Server defaults
	v.SetDefault("server.host", "0.0.0.0")
	v.SetDefault("server.port", "8080")

	// Database defaults
	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.path", "/data/mediacheky.db")

	// Service defaults
	setServiceDefaults(v, "jellyfin", "jellyfin/jellyfin:latest", "8096")
	setServiceDefaults(v, "sonarr", "linuxserver/sonarr:latest", "8989")
	setServiceDefaults(v, "radarr", "linuxserver/radarr:latest", "7878")
	setServiceDefaults(v, "prowlarr", "linuxserver/prowlarr:latest", "9696")
	setServiceDefaults(v, "qbittorrent", "linuxserver/qbittorrent:latest", "8080")
	setServiceDefaults(v, "jellyseerr", "fallenbagel/jellyseerr:latest", "5055")
	setServiceDefaults(v, "qbitmanager", "bobokun/qbit_manage:latest", "")
	setServiceDefaults(v, "dockercontroller", "carcheky/docker-controller-bot:latest", "")
	setServiceDefaults(v, "bazarr", "linuxserver/bazarr:latest", "6767")
	setServiceDefaults(v, "jellystat", "cyfershepard/jellystat:latest", "3000")
}

// setServiceDefaults sets default values for a service
func setServiceDefaults(v *viper.Viper, name, image, port string) {
	v.SetDefault(fmt.Sprintf("services.%s.enabled", name), false)
	v.SetDefault(fmt.Sprintf("services.%s.image", name), image)
	v.SetDefault(fmt.Sprintf("services.%s.port", name), port)
}
