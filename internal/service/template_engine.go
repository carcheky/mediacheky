package service

// TemplateEngine generates docker compose files from templates

import (
	"bytes"
	"context"
	"embed"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/carcheky/mediacheky/internal/models"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// TemplateEngine handles template parsing and docker compose generation
type TemplateEngine struct {
	logger       *zap.Logger
	templatesDir string
	servicesDir  string
	baseDir      string // Base directory where MediaCheky docker compose.yml is located
	configRepo   ConfigRepository
	templateRepo TemplateRepository
}

// ConfigRepository defines the interface for accessing global configuration
type ConfigRepository interface {
	GetAsMap() (map[string]string, error)
}

// TemplateRepository defines the interface for accessing templates
type TemplateRepository interface {
	GetByName(name string) (*models.Template, error)
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	// Service-specific configuration
	Image         string
	ContainerName string
	Port          int
	HostPort      int    // Port to expose on host (0 or empty = no exposure, >0 = expose this port)
	PublicUrl     string // URL for external access (used by some services like Jellyfin)
	Paths         map[string]string
	Umask         string
	Network       string
	RestartPolicy string

	// Global configuration
	Global GlobalConfig

	// Additional custom fields
	Custom map[string]interface{}
}

// GlobalConfig represents global configuration variables
type GlobalConfig struct {
	PUID     string
	PGID     string
	Timezone string
	// Absolute host path to the shared media library root
	// Source of truth: MEDIACHEKY_MEDIA_PATH (deprecated fallback: BASE_MEDIA_PATH)
	MediaPath string
}

// NewTemplateEngine creates a new TemplateEngine instance
func NewTemplateEngine(logger *zap.Logger, templatesDir, servicesDir string, configRepo ConfigRepository, templateRepo TemplateRepository) *TemplateEngine {
	// Ensure servicesDir is absolute path
	absServicesDir, err := filepath.Abs(servicesDir)
	if err != nil {
		logger.Warn("Failed to get absolute path for servicesDir, using as-is",
			zap.String("servicesDir", servicesDir),
			zap.Error(err))
		absServicesDir = servicesDir
	}

	// Get base directory from environment variable (host path) or fallback to working directory
	baseDir := os.Getenv("MEDIACHEKY_HOST_PATH")
	if baseDir == "" {
		// Fallback to current working directory (for development/testing)
		var err error
		baseDir, err = os.Getwd()
		if err != nil {
			logger.Warn("Failed to get working directory, using current dir",
				zap.Error(err))
			baseDir = "."
		}
		logger.Warn("MEDIACHEKY_HOST_PATH not set, using working directory",
			zap.String("baseDir", baseDir))
	} else {
		logger.Info("Using MEDIACHEKY_HOST_PATH for base directory",
			zap.String("baseDir", baseDir))
	}

	return &TemplateEngine{
		logger:       logger,
		templatesDir: templatesDir,
		servicesDir:  absServicesDir,
		baseDir:      baseDir,
		configRepo:   configRepo,
		templateRepo: templateRepo,
	}
}

// GenerateCompose always generates a dynamic compose file from templates.
// This ensures all paths and configuration are correctly applied from the database config.
func (te *TemplateEngine) GenerateCompose(serviceName string, config models.ServiceConfig) (string, error) {
	te.logger.Info("Generating dynamic compose for service", zap.String("service", serviceName))
	return te.generateComposeOld(serviceName, config)
}

// needsDynamicCompose returns true if config contains fields requiring template processing.
func needsDynamicCompose(config models.ServiceConfig) bool {
	// HostPort field exists and is > 0 (port exposure requested)
	if hostPortFloat, ok := config["HostPort"].(float64); ok && int(hostPortFloat) > 0 {
		return true
	}
	if hostPortInt, ok := config["HostPort"].(int); ok && hostPortInt > 0 {
		return true
	}
	// HostPort exists and is 0 or empty (no port exposure - also needs template to omit ports section)
	if _, ok := config["HostPort"].(float64); ok {
		return true
	}
	if _, ok := config["HostPort"].(int); ok {
		return true
	}
	// Paths field exists - need dynamic generation to resolve absolute paths
	if paths, ok := config["Paths"].(map[string]interface{}); ok && len(paths) > 0 {
		return true
	}
	// Future: add more conditional triggers here (environment overrides, optional volumes, etc.)
	return false
}

// Legacy: Keep unused code for reference (can be removed later)
func (te *TemplateEngine) generateComposeOld(serviceName string, config models.ServiceConfig) (string, error) {
	// Load global configuration
	globalConfig, err := te.loadGlobalConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load global config: %w", err)
	}

	// Load template content
	templateContent, err := te.loadTemplate(serviceName)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Build template data
	templateData, err := te.buildTemplateData(serviceName, config, globalConfig)
	if err != nil {
		return "", fmt.Errorf("failed to build template data: %w", err)
	}

	// Validate required fields AFTER buildTemplateData applies defaults
	if templateData.Image == "" {
		return "", fmt.Errorf("image is required but not specified in configuration after applying defaults")
	}

	// Parse and execute template
	tmpl, err := template.New(serviceName).Parse(templateContent)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// DEBUG: Log template data before execution
	te.logger.Info("Executing template",
		zap.String("service", serviceName),
		zap.Any("template_data", templateData),
		zap.Any("paths", templateData.Paths))

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	composeContent := buf.String()

	// DEBUG: Log generated compose content
	te.logger.Info("Generated compose content",
		zap.String("service", serviceName),
		zap.String("content", composeContent))

	// Validate generated YAML
	if err := te.validateYAML(composeContent); err != nil {
		return "", fmt.Errorf("generated invalid YAML: %w", err)
	}

	// Write compose file
	composePath, err := te.writeComposeFile(serviceName, composeContent)
	if err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	te.logger.Info("Successfully generated docker compose",
		zap.String("service", serviceName),
		zap.String("path", composePath))

	return composePath, nil
}

// loadGlobalConfig loads global configuration from the repository
func (te *TemplateEngine) loadGlobalConfig() (GlobalConfig, error) {
	configMap, err := te.configRepo.GetAsMap()
	if err != nil {
		return GlobalConfig{}, err
	}

	// Set defaults if not found
	puid := configMap["PUID"]
	if puid == "" {
		puid = "1000"
	}

	pgid := configMap["PGID"]
	if pgid == "" {
		pgid = "1000"
	}

	timezone := configMap["TZ"]
	if timezone == "" {
		timezone = "UTC"
	}

	// Get MediaPath using the same logic as service config paths
	// This ensures we use the exact same path mounted in MediaCheky container
	mediaPath := getMediaLibraryPath()

	return GlobalConfig{
		PUID:      puid,
		PGID:      pgid,
		Timezone:  timezone,
		MediaPath: mediaPath,
	}, nil
}

// LoadGlobalConfigPublic is a public wrapper for loadGlobalConfig
// Returns both GlobalConfig and the raw config map for variable expansion
// New canonical key for shared media root (also supports legacy key)
func (te *TemplateEngine) LoadGlobalConfigPublic() (map[string]string, error) {
	configMap, err := te.configRepo.GetAsMap()
	if err != nil {
		return nil, err
	}

	// Set defaults if not found
	if configMap["PUID"] == "" {
		configMap["PUID"] = "1000"
	}
	if configMap["PGID"] == "" {
		configMap["PGID"] = "1000"
	}
	if configMap["TZ"] == "" {
		configMap["TZ"] = "UTC"
	}

	// Path defaults - convert relative paths to absolute based on MediaCheky base directory
	if configMap["CONFIG_BASE_PATH"] == "" {
		configMap["CONFIG_BASE_PATH"] = "./volumes"
	}
	// Convert to absolute path if relative (relative to MediaCheky base dir)
	configMap["CONFIG_BASE_PATH"] = te.toAbsolutePath(configMap["CONFIG_BASE_PATH"])

	// Media library path - always use the global function to ensure consistency
	// This reads from MEDIACHEKY_MEDIA_PATH env var and converts relative to absolute
	configMap["MEDIACHEKY_MEDIA_PATH"] = getMediaLibraryPath()

	te.logger.Info("Using media library path",
		zap.String("path", configMap["MEDIACHEKY_MEDIA_PATH"]))

	if configMap["DOWNLOADS_PATH"] == "" {
		configMap["DOWNLOADS_PATH"] = "downloads"
	}
	if configMap["MOVIES_PATH"] == "" {
		configMap["MOVIES_PATH"] = "movies"
	}
	if configMap["SERIES_PATH"] == "" {
		configMap["SERIES_PATH"] = "series"
	}

	return configMap, nil
}

// loadTemplate loads template content from database or filesystem
func (te *TemplateEngine) loadTemplate(serviceName string) (string, error) {
	// First try to load from database
	if te.templateRepo != nil {
		tmpl, err := te.templateRepo.GetByName(serviceName)
		if err == nil && tmpl != nil && tmpl.Content != "" {
			te.logger.Debug("Loaded template from database",
				zap.String("service", serviceName))
			return tmpl.Content, nil
		}
	}

	// Fall back to filesystem
	templatePath := filepath.Join(te.templatesDir, serviceName+".yml")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("template not found: %s: %w", templatePath, err)
	}

	te.logger.Debug("Loaded template from filesystem",
		zap.String("service", serviceName),
		zap.String("path", templatePath))

	return string(content), nil
}

// toAbsolutePath converts a relative path to absolute based on MediaCheky's base directory
// If the path is already absolute, it returns it unchanged
func (te *TemplateEngine) toAbsolutePath(path string) string {
	// Check if path is already absolute
	if filepath.IsAbs(path) {
		return path
	}

	// Convert relative path to absolute by joining with base directory
	absPath := filepath.Join(te.baseDir, path)

	// Clean the path to remove any .. or . components
	absPath = filepath.Clean(absPath)

	te.logger.Debug("Converted relative path to absolute",
		zap.String("relative", path),
		zap.String("absolute", absPath),
		zap.String("base_dir", te.baseDir))

	return absPath
}

// buildTemplateData builds the template data from service config and global config
func (te *TemplateEngine) buildTemplateData(serviceName string, config models.ServiceConfig, globalConfig GlobalConfig) (TemplateData, error) {
	// Debug: Log input config
	te.logger.Info("Building template data",
		zap.Any("config", config),
		zap.Any("config_paths", config["Paths"]))

	data := TemplateData{
		Global: globalConfig,
		Custom: make(map[string]interface{}),
	}

	// Extract standard fields from config
	if image, ok := config["Image"].(string); ok && image != "" {
		data.Image = image
	} else if image, ok := config["image"].(string); ok && image != "" {
		// Try lowercase variant
		data.Image = image
	} else {
		// Set empty to detect missing image later
		data.Image = ""
	}
	if containerName, ok := config["ContainerName"].(string); ok && containerName != "" {
		data.ContainerName = containerName
	} else {
		// Fallback to serviceName if ContainerName not provided
		data.ContainerName = serviceName
	}

	// Fallback: if Image is still empty, infer from ContainerName
	if data.Image == "" && data.ContainerName != "" {
		// Default to linuxserver/<container>:latest when missing
		data.Image = fmt.Sprintf("linuxserver/%s:latest", data.ContainerName)
		te.logger.Warn("Image missing in config; defaulting based on container name",
			zap.String("container", data.ContainerName),
			zap.String("inferred_image", data.Image))
	}

	if port, ok := config["Port"].(float64); ok {
		data.Port = int(port)
	} else if port, ok := config["Port"].(int); ok {
		data.Port = port
	}

	// Extract host port configuration (0 = no exposure, >0 = expose this port)
	if hostPort, ok := config["HostPort"].(float64); ok {
		data.HostPort = int(hostPort)
	} else if hostPort, ok := config["HostPort"].(int); ok {
		data.HostPort = hostPort
	} else {
		data.HostPort = 0 // 0 means no port exposure (internal network only)
	}

	// Extract paths and convert to absolute paths
	if paths, ok := config["Paths"].(map[string]interface{}); ok {
		data.Paths = make(map[string]string)
		for k, v := range paths {
			if strVal, ok := v.(string); ok {
				// Convert to absolute path if relative
				data.Paths[k] = te.toAbsolutePath(strVal)
			}
		}
		te.logger.Info("Extracted and converted paths to absolute",
			zap.Any("paths_interface", paths),
			zap.Any("paths_absolute", data.Paths))
	} else {
		te.logger.Warn("Paths field missing or wrong type",
			zap.Any("paths_value", config["Paths"]),
			zap.String("paths_type", fmt.Sprintf("%T", config["Paths"])))
	}

	// Validate required paths for compose generation
	if len(data.Paths) == 0 {
		return data, fmt.Errorf("paths configuration is missing - at least Config path is required")
	}
	if data.Paths["Config"] == "" {
		return data, fmt.Errorf("Config path is required but was empty")
	}

	// Extract optional fields
	if umask, ok := config["Umask"].(string); ok {
		data.Umask = umask
	}

	// Extract PublicUrl (used by Jellyfin and similar services)
	if publicUrl, ok := config["PublicUrl"].(string); ok {
		data.PublicUrl = publicUrl
	}

	// Always auto-detect MediaCheky's network (not configurable by user)
	detectedNetwork := te.detectSelfNetwork()
	data.Network = detectedNetwork
	te.logger.Info("Service will be created on MediaCheky network",
		zap.String("network", detectedNetwork),
		zap.String("service", data.ContainerName))

	if restartPolicy, ok := config["RestartPolicy"].(string); ok {
		data.RestartPolicy = restartPolicy
	} else {
		data.RestartPolicy = "unless-stopped" // Default
	}

	// Store any additional custom fields
	standardFields := map[string]bool{
		"Image": true, "ContainerName": true, "Port": true, "Paths": true,
		"Umask": true, "Network": true, "RestartPolicy": true,
		"ExposePort": true, "HostPort": true, "PublicUrl": true,
	}
	for k, v := range config {
		if !standardFields[k] {
			data.Custom[k] = v
		}
	}

	return data, nil
}

// detectSelfNetwork detects the Docker network MediaCheky is running on.
// Returns "mediacheky-net" as fallback if detection fails.
func (te *TemplateEngine) detectSelfNetwork() string {
	// Try to create a Docker client to query our own network
	dockerClient, err := NewDockerClient(te.logger)
	if err != nil {
		te.logger.Warn("Failed to create Docker client for network detection, using fallback",
			zap.Error(err))
		return "mediacheky-net" // Fallback
	}
	defer dockerClient.Close()

	// Detect network
	network, err := dockerClient.GetSelfNetwork(context.Background())
	if err != nil || network == "" {
		te.logger.Warn("Failed to detect MediaCheky network, using fallback",
			zap.Error(err))
		return "mediacheky-net" // Fallback
	}

	return network
}

// validateYAML validates that the generated content is valid YAML
func (te *TemplateEngine) validateYAML(content string) error {
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(content), &result); err != nil {
		return fmt.Errorf("invalid YAML: %w", err)
	}

	// Check for required services definition
	if _, hasServices := result["services"]; !hasServices {
		return fmt.Errorf("generated YAML does not contain services definition")
	}

	return nil
}

// writeComposeFile writes the compose content to the appropriate file location
func (te *TemplateEngine) writeComposeFile(serviceName, content string) (string, error) {
	// Get UID/GID for proper ownership
	uid, gid, err := te.getSystemUIDGID()
	if err != nil {
		te.logger.Warn("Failed to get UID/GID, files will be created with default ownership", zap.Error(err))
		uid, gid = -1, -1 // Use -1 to skip ownership change
	}

	// Create service directory
	serviceDir := filepath.Join(te.servicesDir, serviceName)
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create service directory: %w", err)
	}

	// Set directory ownership
	if uid != -1 && gid != -1 {
		if err := te.setOwnership(serviceDir, uid, gid); err != nil {
			te.logger.Warn("Could not set directory ownership", zap.String("path", serviceDir), zap.Error(err))
		}
	}

	composePath := filepath.Join(serviceDir, "docker-compose.yml")

	// Backup existing file if it exists
	if _, err := os.Stat(composePath); err == nil {
		if err := te.backupComposeFile(composePath); err != nil {
			te.logger.Warn("Failed to backup existing compose file",
				zap.String("path", composePath),
				zap.Error(err))
		}
	}

	// Write new compose file
	if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write compose file: %w", err)
	}

	// Set file ownership
	if uid != -1 && gid != -1 {
		if err := te.setOwnership(composePath, uid, gid); err != nil {
			te.logger.Warn("Could not set file ownership", zap.String("path", composePath), zap.Error(err))
		}
	}

	// Get absolute path before returning (required for docker compose)
	absPath, err := filepath.Abs(composePath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	return absPath, nil
}

// backupComposeFile creates a backup of an existing compose file
func (te *TemplateEngine) backupComposeFile(composePath string) error {
	timestamp := time.Now().Format("20060102-150405")
	backupPath := composePath + ".backup." + timestamp

	content, err := os.ReadFile(composePath)
	if err != nil {
		return err
	}

	if err := os.WriteFile(backupPath, content, 0644); err != nil {
		return err
	}

	te.logger.Info("Created backup of compose file",
		zap.String("original", composePath),
		zap.String("backup", backupPath))

	return nil
}

// GetComposePath returns the path to a service's docker-compose.yml file
func (te *TemplateEngine) GetComposePath(serviceName string) string {
	// Prefer dynamically generated compose if it exists
	dynamicPath := filepath.Join(te.servicesDir, serviceName, "docker-compose.yml")
	if _, err := os.Stat(dynamicPath); err == nil {
		abs, err2 := filepath.Abs(dynamicPath)
		if err2 == nil {
			return abs
		}
		return dynamicPath
	}
	// Fallback to static tracked file (services/<name>.yml)
	staticPath := filepath.Join("services", serviceName+".yml")
	absStatic, err := filepath.Abs(staticPath)
	if err == nil {
		return absStatic
	}
	te.logger.Warn("Returning relative static compose path (abs resolution failed)",
		zap.String("service", serviceName),
		zap.String("path", staticPath),
		zap.Error(err))
	return staticPath
}

// LoadTemplatesFromFS loads templates from filesystem into the database
// This is useful for initializing the database with default templates
func (te *TemplateEngine) LoadTemplatesFromFS(fs embed.FS, templatesPath string) error {
	entries, err := fs.ReadDir(templatesPath)
	if err != nil {
		return fmt.Errorf("failed to read templates directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		serviceName := entry.Name()[:len(entry.Name())-4] // Remove .yml extension

		te.logger.Info("Found template file",
			zap.String("service", serviceName),
			zap.String("file", entry.Name()))
	}

	return nil
}

// ExtractVolumePaths parses a docker compose file and extracts host volume paths
// It expands environment variables using the provided config map
func (te *TemplateEngine) ExtractVolumePaths(composePath string, configMap map[string]string) ([]string, error) {
	content, err := os.ReadFile(composePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read compose file: %w", err)
	}

	// Expand environment variables in the content
	contentStr := string(content)
	for key, value := range configMap {
		contentStr = strings.ReplaceAll(contentStr, "${"+key+"}", value)
	}

	var composeData map[string]interface{}
	if err := yaml.Unmarshal([]byte(contentStr), &composeData); err != nil {
		return nil, fmt.Errorf("failed to parse compose YAML: %w", err)
	}

	var volumePaths []string

	// Extract services section
	services, ok := composeData["services"].(map[string]interface{})
	if !ok {
		return volumePaths, nil
	}

	// Iterate through each service
	for _, serviceData := range services {
		service, ok := serviceData.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract volumes section
		volumes, ok := service["volumes"].([]interface{})
		if !ok {
			continue
		}

		// Parse each volume definition
		for _, vol := range volumes {
			volStr, ok := vol.(string)
			if !ok {
				continue
			}

			// Parse volume string (format: "host_path:container_path" or "host_path:container_path:options")
			parts := strings.Split(volStr, ":")
			if len(parts) < 2 {
				continue
			}

			hostPath := parts[0]

			// Skip named volumes (no path separator)
			if !filepath.IsAbs(hostPath) && !strings.HasPrefix(hostPath, ".") {
				continue
			}

			// Convert relative paths to absolute
			if !filepath.IsAbs(hostPath) {
				composeDir := filepath.Dir(composePath)
				hostPath = filepath.Join(composeDir, hostPath)
			}

			volumePaths = append(volumePaths, hostPath)
		}
	}

	return volumePaths, nil
}

// getSystemUIDGID returns the UID and GID from environment or current user
func (te *TemplateEngine) getSystemUIDGID() (int, int, error) {
	// Try to get from environment first (PUID/PGID)
	puidStr := os.Getenv("PUID")
	pgidStr := os.Getenv("PGID")

	var uid, gid int
	var err error

	if puidStr != "" && pgidStr != "" {
		uid, err = strconv.Atoi(puidStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid PUID: %w", err)
		}
		gid, err = strconv.Atoi(pgidStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid PGID: %w", err)
		}
		return uid, gid, nil
	}

	// Fallback to current user
	currentUser, err := user.Current()
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get current user: %w", err)
	}

	uid, err = strconv.Atoi(currentUser.Uid)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid user UID: %w", err)
	}

	gid, err = strconv.Atoi(currentUser.Gid)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid user GID: %w", err)
	}

	return uid, gid, nil
}

// setOwnership sets the owner and group of a file or directory
func (te *TemplateEngine) setOwnership(path string, uid, gid int) error {
	if err := os.Chown(path, uid, gid); err != nil {
		te.logger.Warn("Failed to set ownership",
			zap.String("path", path),
			zap.Int("uid", uid),
			zap.Int("gid", gid),
			zap.Error(err))
		return fmt.Errorf("failed to set ownership: %w", err)
	}
	return nil
}

// EnsureDirectoryExists creates a directory if it doesn't exist and sets proper ownership
func (te *TemplateEngine) EnsureDirectoryExists(path string) error {
	// Translate absolute host paths to container mount points when applicable
	containerPath := te.translateToContainerPath(path)

	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		te.logger.Debug("Creating volume directory", zap.String("path", containerPath))
		if err := os.MkdirAll(containerPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Set ownership
		uid, gid, err := te.getSystemUIDGID()
		if err != nil {
			te.logger.Warn("Failed to get UID/GID, skipping ownership change", zap.Error(err))
			return nil
		}

		if err := te.setOwnership(containerPath, uid, gid); err != nil {
			// Don't fail if we can't set ownership, just warn
			te.logger.Warn("Could not set directory ownership", zap.String("path", containerPath), zap.Error(err))
		}
	}
	return nil
}

// translateToContainerPath maps absolute host paths (based on MEDIACHEKY_HOST_PATH)
// to the corresponding container mount points. This prevents permission errors
// when MediaCheky tries to create directories using host paths from inside the container.
func (te *TemplateEngine) translateToContainerPath(path string) string {
	// If path is already pointing to a container mount, return as-is
	if strings.HasPrefix(path, "/app/data/") || strings.HasPrefix(path, "/MEDIACHEKY_LIBRARY/") {
		return path
	}

	// Resolve baseDir (host root of the project) from the TemplateEngine
	hostBase := te.baseDir
	if hostBase == "" {
		// Fallback to working directory if not set
		if wd, err := os.Getwd(); err == nil {
			hostBase = wd
		}
	}

	// Normalize for consistent prefix checks
	hostBase = filepath.Clean(hostBase)
	path = filepath.Clean(path)

	// Map host app data dir to /app/data
	hostAppData := filepath.Join(hostBase, "volumes", "mediacheky-data")
	if strings.HasPrefix(path, hostAppData) {
		// Remainder after the host app data prefix
		remainder := strings.TrimPrefix(path, hostAppData)
		// Ensure leading slash is preserved appropriately
		remainder = strings.TrimPrefix(remainder, string(os.PathSeparator))
		return filepath.Join("/app/data", remainder)
	}

	// Map host media library path to /MEDIACHEKY_LIBRARY
	mediaPath := os.Getenv("MEDIACHEKY_MEDIA_PATH")
	if mediaPath == "" {
		mediaPath = filepath.Join(hostBase, "volumes", "library")
	} else if !filepath.IsAbs(mediaPath) {
		mediaPath = filepath.Join(hostBase, mediaPath)
	}
	mediaPath = filepath.Clean(mediaPath)
	if strings.HasPrefix(path, mediaPath) {
		remainder := strings.TrimPrefix(path, mediaPath)
		remainder = strings.TrimPrefix(remainder, string(os.PathSeparator))
		return filepath.Join("/MEDIACHEKY_LIBRARY", remainder)
	}

	// No translation needed
	return path
}

// RemoveDirectory removes a directory and all its contents
func (te *TemplateEngine) RemoveDirectory(path string) error {
	// Security check: ensure path is within allowed directories
	if !strings.HasPrefix(path, "/app/data/services/") &&
		!strings.HasPrefix(path, "./volumes/mediacheky-data/services/") &&
		!strings.HasPrefix(path, "/app/data/services-volumes/") &&
		!strings.HasPrefix(path, "./volumes/mediacheky-data/services-volumes/") {
		return fmt.Errorf("refusing to delete directory outside of services config area: %s", path)
	}

	// Check if directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		te.logger.Debug("Directory does not exist, nothing to remove", zap.String("path", path))
		return nil
	}

	te.logger.Info("Removing directory", zap.String("path", path))
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove directory: %w", err)
	}

	te.logger.Info("Directory removed successfully", zap.String("path", path))
	return nil
}
