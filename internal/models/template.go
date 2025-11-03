package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// JSONSchema is a custom type for JSON Schema stored as JSON
type JSONSchema map[string]interface{}

// Scan implements the sql.Scanner interface for JSONSchema
func (js *JSONSchema) Scan(value interface{}) error {
	if value == nil {
		*js = make(JSONSchema)
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("failed to scan JSONSchema: expected []byte or string, got %T", value)
	}

	return json.Unmarshal(bytes, js)
}

// Value implements the driver.Valuer interface for JSONSchema
func (js JSONSchema) Value() (driver.Value, error) {
	if len(js) == 0 {
		return "{}", nil
	}
	return json.Marshal(js)
}

// Template represents a docker-compose template for a service
type Template struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Template identification
	Name    string `json:"name" gorm:"uniqueIndex;not null"` // e.g., "radarr", "sonarr"
	Version string `json:"version" gorm:"not null"`          // Template version
	
	// Template content
	Content string `json:"content" gorm:"type:text;not null"` // Docker compose YAML template
	
	// Configuration schema (JSON Schema for validation)
	Schema JSONSchema `json:"schema" gorm:"type:json"` // JSON Schema for config validation
}

func (Template) TableName() string {
	return "templates"
}
