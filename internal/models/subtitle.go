package models

// Subtitle represents subtitle information for media
type Subtitle struct {
	Language string `json:"language"`
	Path     string `json:"path"`
	Forced   bool   `json:"forced"`
	Default  bool   `json:"default"`
}
