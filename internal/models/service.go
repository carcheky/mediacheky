package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ServiceConfig is a custom type for service configuration stored as JSON
type ServiceConfig map[string]interface{}

// Scan implements the sql.Scanner interface for ServiceConfig
func (sc *ServiceConfig) Scan(value interface{}) error {
	if value == nil {
		*sc = make(ServiceConfig)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to scan ServiceConfig: expected []byte or string, got %T", value)
	}

	return json.Unmarshal(bytes, sc)
}

// Value implements the driver.Valuer interface for ServiceConfig
func (sc ServiceConfig) Value() (driver.Value, error) {
	if len(sc) == 0 {
		return "{}", nil
	}
	return json.Marshal(sc)
}

// Service represents a managed service configuration and state
type Service struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Service identification
	Name        string `json:"name" gorm:"uniqueIndex;not null"` // e.g., "radarr", "sonarr"
	DisplayName string `json:"display_name" gorm:"not null"`     // e.g., "Radarr", "Sonarr"
	Icon        string `json:"icon"`                             // Icon name or URL

	// Service state
	Enabled bool   `json:"enabled" gorm:"default:false;index"` // Whether service is enabled
	Status  string `json:"status" gorm:"default:'stopped'"`    // running, stopped, error

	// Docker configuration
	Image       string `json:"image"`        // Docker image (e.g., "linuxserver/radarr")
	ContainerID string `json:"container_id"` // Docker container ID
	Port        int    `json:"port"`         // Service port

	// Service configuration (stored as JSON)
	Config ServiceConfig `json:"config" gorm:"type:json"`

	// Proxy configuration
	Subdomain string `json:"subdomain" gorm:"index"` // Custom subdomain for the service (e.g., "radarr")
	Domain    string `json:"domain"`                 // Domain to use (if different from global)

	// Relationship to template
	TemplateID uint      `json:"template_id" gorm:"index"`
	Template   *Template `json:"template,omitempty" gorm:"foreignKey:TemplateID"`
}

func (Service) TableName() string {
	return "services"
}

// RadarrConfig represents Radarr's config.xml settings
type RadarrConfig struct {
	BindAddress            string `json:"BindAddress" xml:"BindAddress"`
	Port                   int    `json:"Port" xml:"Port"`
	SslPort                int    `json:"SslPort" xml:"SslPort"`
	EnableSsl              bool   `json:"EnableSsl" xml:"EnableSsl"`
	LaunchBrowser          bool   `json:"LaunchBrowser" xml:"LaunchBrowser"`
	ApiKey                 string `json:"ApiKey" xml:"ApiKey"`
	AuthenticationMethod   string `json:"AuthenticationMethod" xml:"AuthenticationMethod"`
	AuthenticationRequired string `json:"AuthenticationRequired" xml:"AuthenticationRequired"`
	Username               string `json:"Username" xml:"Username"`
	Password               string `json:"Password" xml:"Password"`
	PasswordConfirmation   string `json:"PasswordConfirmation" xml:"PasswordConfirmation"`
	Branch                 string `json:"Branch" xml:"Branch"`
	LogLevel               string `json:"LogLevel" xml:"LogLevel"`
	SslCertPath            string `json:"SslCertPath" xml:"SslCertPath"`
	SslCertPassword        string `json:"SslCertPassword" xml:"SslCertPassword"`
	UrlBase                string `json:"UrlBase" xml:"UrlBase"`
	InstanceName           string `json:"InstanceName" xml:"InstanceName"`
	UpdateMechanism        string `json:"UpdateMechanism" xml:"UpdateMechanism"`
	UseProxy               bool   `json:"UseProxy" xml:"UseProxy"`
	SendAnonymousUsageData bool   `json:"SendAnonymousUsageData" xml:"SendAnonymousUsageData"`
}

// SonarrConfig represents Sonarr's config.xml settings
type SonarrConfig struct {
	BindAddress            string `json:"BindAddress" xml:"BindAddress"`
	Port                   int    `json:"Port" xml:"Port"`
	SslPort                int    `json:"SslPort" xml:"SslPort"`
	EnableSsl              bool   `json:"EnableSsl" xml:"EnableSsl"`
	LaunchBrowser          bool   `json:"LaunchBrowser" xml:"LaunchBrowser"`
	ApiKey                 string `json:"ApiKey" xml:"ApiKey"`
	AuthenticationMethod   string `json:"AuthenticationMethod" xml:"AuthenticationMethod"`
	AuthenticationRequired string `json:"AuthenticationRequired" xml:"AuthenticationRequired"`
	Username               string `json:"Username" xml:"Username"`
	Password               string `json:"Password" xml:"Password"`
	PasswordConfirmation   string `json:"PasswordConfirmation" xml:"PasswordConfirmation"`
	Branch                 string `json:"Branch" xml:"Branch"`
	LogLevel               string `json:"LogLevel" xml:"LogLevel"`
	SslCertPath            string `json:"SslCertPath" xml:"SslCertPath"`
	SslCertPassword        string `json:"SslCertPassword" xml:"SslCertPassword"`
	UrlBase                string `json:"UrlBase" xml:"UrlBase"`
	InstanceName           string `json:"InstanceName" xml:"InstanceName"`
	UpdateMechanism        string `json:"UpdateMechanism" xml:"UpdateMechanism"`
	UseProxy               bool   `json:"UseProxy" xml:"UseProxy"`
	SendAnonymousUsageData bool   `json:"SendAnonymousUsageData" xml:"SendAnonymousUsageData"`
}
