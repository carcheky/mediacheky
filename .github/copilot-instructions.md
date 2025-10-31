# GitHub Copilot Instructions - KeeperCheky Project

USE MCP MEMORY SERVER AND SEQUENTIAL THINKING

READ logs/keepercheky-dev.log AFTER CHANGES, VISIT URLS, AND READ logs/keepercheky-dev.log AND INSPECT FILES IF NEEDED


⛔️ ⛔️ ⛔️ CRITICAL RULE - NEVER VIOLATE ⛔️ ⛔️ ⛔️

**YOU MUST NEVER, UNDER ANY CIRCUMSTANCES:**
- never print the log, always read directly logs/keepercheky-dev.log
- Run `make dev` or `make run` or ANY make command that starts services
- Run `docker-compose up` or `docker-compose down` or `docker-compose restart` or `docker-compose stop`
- Run `docker start` or `docker stop` or `docker restart` or `docker kill` or `docker rm`
- Execute ANY command that starts, stops, restarts, kills, or removes Docker containers
- Use `run_in_terminal` with `isBackground: true` for ANY command that starts servers or services

**ONLY THE USER CAN START, STOP, OR RESTART SERVICES.**

**WHAT YOU CAN DO:**
- Read logs with `cat`, `tail`, `grep`, etc.
- Execute commands INSIDE running containers (docker exec) for debugging
- Inspect files and configurations
- Make code changes
- Run tests (but NOT start test servers)

**IF YOU NEED TO TEST SOMETHING, ASK THE USER TO START/RESTART THE SERVICE.**

⛔️ ⛔️ ⛔️ END OF CRITICAL RULE ⛔️ ⛔️ ⛔️

> **Language Note**: This document is in English for consistency with code and technical documentation. However, **always communicate with users in Spanish** when responding to issues, pull requests, or user interactions.

## 🎯 Project Overview

**KeeperCheky** is a modern web-based media library cleanup manager - a complete rewrite of [Janitorr](https://github.com/Schaka/janitorr) with a beautiful UI.

### Core Technology Stack

- **Backend**: Go 1.22+ with Fiber v2 framework
- **Frontend**: Alpine.js 3.x with Tailwind CSS
- **Database**: GORM v2 with SQLite (development) / PostgreSQL (production)
- **Scheduler**: robfig/cron v3
- **HTTP Client**: go-resty/resty v2
- **Templates**: Go html/template
- **Docker**: Multi-stage builds targeting scratch/alpine
- **Target**: Single binary deployment, 15-25MB Docker image, 20-50MB RAM usage

## 🏗️ Architecture Principles

### 1. Project Structure

**ALWAYS** follow this exact structure:

```
keepercheky/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/                       # Private application code
│   ├── config/                     # Configuration management
│   ├── models/                     # Database models (GORM)
│   ├── repository/                 # Data access layer
│   ├── service/                    # Business logic
│   │   ├── clients/                # External service clients
│   │   ├── cleanup/                # Cleanup strategies
│   │   └── scheduler/              # Job scheduling
│   ├── handler/                    # HTTP handlers (Fiber)
│   └── middleware/                 # HTTP middleware
├── web/
│   ├── templates/                  # Go html/template files
│   │   ├── layouts/
│   │   ├── pages/
│   │   └── components/
│   └── static/                     # CSS, JS, images
│       ├── css/
│       ├── js/
│       └── images/
├── pkg/                            # Public/shared packages
│   ├── filesystem/
│   ├── logger/
│   └── utils/
├── migrations/                     # Database migrations
├── scripts/                        # Build and utility scripts
├── docs/                           # Documentation
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

**Key Rules:**
- ✅ Use `internal/` for application-specific code (not importable by other projects)
- ✅ Use `pkg/` only for truly reusable, public packages
- ✅ Keep `cmd/` minimal - only main.go and setup
- ✅ Never put business logic in handlers - use services
- ✅ Templates go in `web/templates/`, static files in `web/static/`

### 2. Code Organization Patterns

#### Repository Pattern
```go
// internal/repository/media_repo.go
type MediaRepository struct {
    db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) *MediaRepository {
    return &MediaRepository{db: db}
}

func (r *MediaRepository) GetAll() ([]models.Media, error) {
    var media []models.Media
    result := r.db.Find(&media)
    return media, result.Error
}
```

#### Service Layer
```go
// internal/service/cleanup_service.go
type CleanupService struct {
    mediaRepo   *repository.MediaRepository
    radarrClient clients.MediaClient
    logger      *logger.Logger
}

func NewCleanupService(
    mediaRepo *repository.MediaRepository,
    radarrClient clients.MediaClient,
    logger *logger.Logger,
) *CleanupService {
    return &CleanupService{
        mediaRepo:   mediaRepo,
        radarrClient: radarrClient,
        logger:      logger,
    }
}
```

#### Handler Pattern (Fiber)
```go
// internal/handler/media_handler.go
type MediaHandler struct {
    cleanupSvc *service.CleanupService
    mediaRepo  *repository.MediaRepository
}

func (h *MediaHandler) GetMedia(c *fiber.Ctx) error {
    media, err := h.mediaRepo.GetAll()
    if err != nil {
        return c.Status(500).JSON(fiber.Map{"error": err.Error()})
    }
    return c.JSON(media)
}

func (h *MediaHandler) RenderMediaPage(c *fiber.Ctx) error {
    media, err := h.mediaRepo.GetAll()
    if err != nil {
        return err
    }
    return c.Render("pages/media", fiber.Map{
        "Title": "Media Management",
        "Media": media,
    })
}
```

## 🔑 Critical Development Rules

### Error Handling

**ALWAYS** handle errors explicitly. Never ignore them.

```go
// ❌ BAD
result, _ := someFunction()

// ✅ GOOD
result, err := someFunction()
if err != nil {
    log.Printf("Error in someFunction: %v", err)
    return fmt.Errorf("failed to execute: %w", err)
}
```

**For Fiber handlers:**
```go
// ✅ GOOD
func (h *Handler) Example(c *fiber.Ctx) error {
    data, err := h.service.GetData()
    if err != nil {
        h.logger.Error("Failed to get data", "error", err)
        return c.Status(500).JSON(fiber.Map{
            "error": "Internal server error",
        })
    }
    return c.JSON(data)
}
```

### Logging

**ALWAYS** use structured logging with context:

```go
// ✅ GOOD
logger.Info("Starting cleanup process",
    "media_count", len(mediaList),
    "dry_run", config.DryRun,
)

logger.Error("Failed to delete media",
    "media_id", media.ID,
    "title", media.Title,
    "error", err,
)
```

**Avoid:**
```go
// ❌ BAD
log.Println("Starting cleanup")
fmt.Printf("Error: %v\n", err)
```

### Context Propagation

**ALWAYS** pass context through the call chain:

```go
// ✅ GOOD
func (s *Service) ProcessMedia(ctx context.Context, mediaID int) error {
    media, err := s.repo.GetByID(ctx, mediaID)
    if err != nil {
        return err
    }
    return s.client.DeleteMedia(ctx, media)
}
```

### Goroutines and Concurrency

**ALWAYS** use proper synchronization and context cancellation:

```go
// ✅ GOOD
func (s *Service) ProcessBatch(ctx context.Context, items []Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(item Item) {
            defer wg.Done()
            if err := s.processItem(ctx, item); err != nil {
                errChan <- err
            }
        }(item)
    }
    
    wg.Wait()
    close(errChan)
    
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    return nil
}
```

### Database Operations

**ALWAYS** use transactions for multi-step operations:

```go
// ✅ GOOD
func (r *Repository) UpdateMediaWithHistory(media *models.Media) error {
    return r.db.Transaction(func(tx *gorm.DB) error {
        // Update media
        if err := tx.Save(media).Error; err != nil {
            return err
        }
        
        // Create history record
        history := &models.History{
            MediaID: media.ID,
            Action:  "updated",
        }
        if err := tx.Create(history).Error; err != nil {
            return err
        }
        
        return nil
    })
}
```

## 🎨 Frontend Guidelines (Alpine.js)

### Component Structure

**ALWAYS** create reusable Alpine.js components:

```html
<!-- web/templates/components/media_card.html -->
<div x-data="mediaCard({{ .Media.ID }})" class="card">
    <img :src="media.poster_url" :alt="media.title">
    <h4 x-text="media.title"></h4>
    
    <button @click="exclude()" class="btn">Exclude</button>
    <button @click="deleteMedia()" class="btn btn-danger">Delete</button>
</div>

<script>
function mediaCard(mediaId) {
    return {
        media: null,
        
        async init() {
            await this.fetchMedia();
        },
        
        async fetchMedia() {
            const response = await fetch(`/api/media/${mediaId}`);
            this.media = await response.json();
        },
        
        async exclude() {
            if (!confirm('Exclude this media?')) return;
            await fetch(`/api/media/${mediaId}/exclude`, { method: 'POST' });
            await this.fetchMedia();
        },
        
        async deleteMedia() {
            if (!confirm('Delete this media permanently?')) return;
            await fetch(`/api/media/${mediaId}`, { method: 'DELETE' });
            window.location.reload();
        }
    }
}
</script>
```

### State Management

**ALWAYS** keep state local unless it needs to be global:

```html
<!-- Local state -->
<div x-data="{ open: false, loading: false }">
    <!-- Component content -->
</div>

<!-- Global state (when needed) -->
<script>
document.addEventListener('alpine:init', () => {
    Alpine.store('config', {
        dryRun: true,
        
        toggle() {
            this.dryRun = !this.dryRun;
        }
    });
});
</script>
```

### API Calls

**ALWAYS** handle loading and error states:

```javascript
async function fetchWithState() {
    return {
        data: null,
        loading: false,
        error: null,
        
        async fetch() {
            this.loading = true;
            this.error = null;
            
            try {
                const response = await fetch('/api/endpoint');
                if (!response.ok) {
                    throw new Error(`HTTP ${response.status}`);
                }
                this.data = await response.json();
            } catch (err) {
                this.error = err.message;
                console.error('Fetch error:', err);
            } finally {
                this.loading = false;
            }
        }
    }
}
```

## 🔌 External Service Clients

### Client Interface Pattern

**ALWAYS** define interfaces for external clients:

```go
// internal/service/clients/client.go
type MediaClient interface {
    TestConnection(ctx context.Context) error
    GetLibrary(ctx context.Context) ([]*models.Media, error)
    DeleteItem(ctx context.Context, id int) error
}

// internal/service/clients/radarr.go
type RadarrClient struct {
    client  *resty.Client
    baseURL string
    apiKey  string
    logger  *logger.Logger
}

func NewRadarrClient(baseURL, apiKey string, logger *logger.Logger) *RadarrClient {
    client := resty.New()
    client.SetBaseURL(baseURL)
    client.SetHeader("X-Api-Key", apiKey)
    client.SetTimeout(30 * time.Second)
    
    return &RadarrClient{
        client:  client,
        baseURL: baseURL,
        apiKey:  apiKey,
        logger:  logger,
    }
}

func (c *RadarrClient) TestConnection(ctx context.Context) error {
    resp, err := c.client.R().
        SetContext(ctx).
        Get("/api/v3/system/status")
    
    if err != nil {
        return fmt.Errorf("connection failed: %w", err)
    }
    
    if resp.StatusCode() != 200 {
        return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
    }
    
    c.logger.Info("Radarr connection successful")
    return nil
}
```

### Retry Logic

**ALWAYS** implement retry logic for external API calls:

```go
func (c *Client) callWithRetry(ctx context.Context, fn func() error) error {
    maxRetries := 3
    backoff := time.Second
    
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if i < maxRetries-1 {
            c.logger.Warn("API call failed, retrying",
                "attempt", i+1,
                "error", err,
            )
            time.Sleep(backoff)
            backoff *= 2
        }
    }
    
    return fmt.Errorf("max retries exceeded")
}
```

## 🧹 Cleanup Logic Implementation

### Safety First

**ALWAYS** implement safety checks:

```go
func (s *CleanupService) DeleteMedia(ctx context.Context, media *models.Media) error {
    // 1. Check dry-run mode
    if s.config.DryRun {
        s.logger.Info("[DRY RUN] Would delete media",
            "id", media.ID,
            "title", media.Title,
        )
        return nil
    }
    
    // 2. Check exclusion tags
    if s.isExcluded(media) {
        s.logger.Info("Media is excluded, skipping",
            "id", media.ID,
            "tags", media.Tags,
        )
        return nil
    }
    
    // 3. Check seeding status (if enabled)
    if s.config.ValidateSeeding {
        isSeeding, err := s.checkSeeding(ctx, media)
        if err != nil {
            return fmt.Errorf("failed to check seeding: %w", err)
        }
        if isSeeding {
            s.logger.Info("Media is still seeding, skipping",
                "id", media.ID,
            )
            return nil
        }
    }
    
    // 4. Execute deletion
    return s.executeDelete(ctx, media)
}
```

### Leaving Soon Collections

**ALWAYS** create symlinks safely:

```go
func (s *FilesystemService) CreateLeavingSoonSymlink(media *models.Media) error {
    sourcePath := media.FilePath
    destPath := filepath.Join(s.config.LeavingSoonDir, filepath.Base(sourcePath))
    
    // Check if source exists
    if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
        return fmt.Errorf("source file not found: %s", sourcePath)
    }
    
    // Remove existing symlink if present
    if _, err := os.Lstat(destPath); err == nil {
        if err := os.Remove(destPath); err != nil {
            return fmt.Errorf("failed to remove existing symlink: %w", err)
        }
    }
    
    // Create symlink
    if err := os.Symlink(sourcePath, destPath); err != nil {
        return fmt.Errorf("failed to create symlink: %w", err)
    }
    
    s.logger.Info("Created leaving soon symlink",
        "source", sourcePath,
        "dest", destPath,
    )
    
    return nil
}
```

## 🗄️ Database Best Practices

### Model Definitions

**ALWAYS** use proper GORM tags and conventions:

```go
// internal/models/media.go
type Media struct {
    ID        uint           `json:"id" gorm:"primaryKey"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
    
    Title         string    `json:"title" gorm:"not null;index"`
    Type          string    `json:"type" gorm:"not null;index"` // "movie" or "series"
    PosterURL     string    `json:"poster_url"`
    FilePath      string    `json:"file_path" gorm:"not null"`
    Size          int64     `json:"size"`
    AddedDate     time.Time `json:"added_date" gorm:"index"`
    LastWatched   *time.Time `json:"last_watched"`
    
    // Service IDs
    RadarrID      *int `json:"radarr_id" gorm:"index"`
    SonarrID      *int `json:"sonarr_id" gorm:"index"`
    JellyfinID    *string `json:"jellyfin_id" gorm:"index"`
    JellyseerrID  *int `json:"jellyseerr_id" gorm:"index"`
    
    // Metadata
    Tags          []string `json:"tags" gorm:"type:json"`
    Quality       string   `json:"quality"`
    Excluded      bool     `json:"excluded" gorm:"default:false;index"`
    
    // Relationships
    History       []History `json:"history,omitempty" gorm:"foreignKey:MediaID"`
}

func (Media) TableName() string {
    return "media"
}
```

### Migrations

**ALWAYS** use AutoMigrate for development, but prepare manual migrations for production:

```go
// internal/database/migrate.go
func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.Media{},
        &models.Schedule{},
        &models.History{},
        &models.Config{},
    )
}
```

## 🐳 Docker Best Practices

### Multi-stage Dockerfile

**ALWAYS** use multi-stage builds for minimal images:

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o /app/bin/keepercheky \
    ./cmd/server

# Final stage
FROM scratch

# Copy certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /app/bin/keepercheky /keepercheky

# Copy web assets
COPY --from=builder /app/web /web

# Expose port
EXPOSE 8000

# Run
ENTRYPOINT ["/keepercheky"]
```

### Health Checks

**ALWAYS** implement health check endpoints:

```go
// internal/handler/health.go
func (h *HealthHandler) Check(c *fiber.Ctx) error {
    // Check database
    sqlDB, err := h.db.DB()
    if err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "error":  "database connection failed",
        })
    }
    
    if err := sqlDB.Ping(); err != nil {
        return c.Status(503).JSON(fiber.Map{
            "status": "unhealthy",
            "error":  "database ping failed",
        })
    }
    
    return c.JSON(fiber.Map{
        "status": "healthy",
        "timestamp": time.Now().Unix(),
    })
}
```

## 🧪 Testing Requirements

### Unit Tests

**ALWAYS** write unit tests for services and repositories:

```go
// internal/service/cleanup_service_test.go
func TestCleanupService_GetMediaToDelete(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    repo := repository.NewMediaRepository(db)
    service := NewCleanupService(repo, mockClient, logger)
    
    // Create test data
    media := &models.Media{
        Title:     "Test Movie",
        Type:      "movie",
        AddedDate: time.Now().AddDate(0, 0, -100), // 100 days old
    }
    db.Create(media)
    
    // Test
    toDelete, err := service.GetMediaToDelete(context.Background())
    
    // Assert
    assert.NoError(t, err)
    assert.Len(t, toDelete, 1)
    assert.Equal(t, media.ID, toDelete[0].ID)
}
```

### Integration Tests

**ALWAYS** test external client integrations:

```go
func TestRadarrClient_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    client := NewRadarrClient(testURL, testAPIKey, logger)
    
    err := client.TestConnection(context.Background())
    assert.NoError(t, err)
}
```

## 📋 Configuration Management

### Environment Variables

**ALWAYS** use a structured config with validation:

```go
// internal/config/config.go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Clients  ClientsConfig  `mapstructure:"clients"`
    App      AppConfig      `mapstructure:"app"`
}

type AppConfig struct {
    DryRun          bool          `mapstructure:"dry_run"`
    LeavingSoonDays int           `mapstructure:"leaving_soon_days"`
    ExclusionTags   []string      `mapstructure:"exclusion_tags"`
    LogLevel        string        `mapstructure:"log_level"`
}

func Load() (*Config, error) {
    viper.SetConfigName("config")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")
    viper.AddConfigPath("/config")
    
    // Environment variables
    viper.SetEnvPrefix("KEEPERCHEKY")
    viper.AutomaticEnv()
    
    if err := viper.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read config: %w", err)
    }
    
    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, nil
}
```

## 🚨 Security Checklist

**ALWAYS** implement these security measures:

- ✅ Validate all user input
- ✅ Sanitize file paths to prevent directory traversal
- ✅ Use parameterized queries (GORM handles this)
- ✅ Implement rate limiting for API endpoints
- ✅ Never log sensitive data (API keys, passwords)
- ✅ Use HTTPS in production (via reverse proxy)
- ✅ Implement CORS properly for API endpoints
- ✅ Validate file types and sizes for uploads

```go
// Example: Path validation
func validatePath(path string) error {
    cleanPath := filepath.Clean(path)
    if strings.Contains(cleanPath, "..") {
        return fmt.Errorf("invalid path: contains '..'")
    }
    return nil
}
```

## 📝 Documentation Requirements

**ALWAYS** document:

1. **Package-level comments**
```go
// Package cleanup provides media cleanup strategies and execution logic.
// It implements multiple cleanup approaches: time-based, tag-based, and episode-based.
package cleanup
```

2. **Exported function comments**
```go
// DeleteMedia removes a media item from all configured services.
// It respects dry-run mode and checks exclusion tags before deletion.
// Returns an error if any service deletion fails.
func (s *CleanupService) DeleteMedia(ctx context.Context, media *models.Media) error {
    // Implementation
}
```

3. **Complex logic comments**
```go
// We need to check seeding status before deletion to prevent
// removing files that are still being seeded in the torrent client.
// This check is skipped if ValidateSeeding is disabled in config.
if s.config.ValidateSeeding {
    // Check implementation
}
```

## 🎯 Performance Optimization

### Database Query Optimization

**ALWAYS** use proper indexing and eager loading:

```go
// ✅ GOOD - With preloading
func (r *MediaRepository) GetAllWithHistory() ([]models.Media, error) {
    var media []models.Media
    result := r.db.Preload("History").Find(&media)
    return media, result.Error
}

// ✅ GOOD - With pagination
func (r *MediaRepository) GetPaginated(page, pageSize int) ([]models.Media, int64, error) {
    var media []models.Media
    var total int64
    
    r.db.Model(&models.Media{}).Count(&total)
    
    offset := (page - 1) * pageSize
    result := r.db.Limit(pageSize).Offset(offset).Find(&media)
    
    return media, total, result.Error
}
```

### Caching Strategy

**ALWAYS** cache expensive operations:

```go
type CachedService struct {
    service *Service
    cache   *cache.Cache
}

func (s *CachedService) GetStats(ctx context.Context) (*Stats, error) {
    // Check cache
    if cached, found := s.cache.Get("stats"); found {
        return cached.(*Stats), nil
    }
    
    // Compute
    stats, err := s.service.ComputeStats(ctx)
    if err != nil {
        return nil, err
    }
    
    // Cache for 5 minutes
    s.cache.Set("stats", stats, 5*time.Minute)
    
    return stats, nil
}
```

## 🔄 Git Commit Conventions

**ALWAYS** use conventional commits **IN ENGLISH**:

**CRITICAL**: All commit messages MUST be written in English, following the Conventional Commits specification.

> **📄 For detailed commit message generation rules, see [`.vscode/copilot-commit-message-instructions.md`](../.vscode/copilot-commit-message-instructions.md)**

**Format**: `<type>(<scope>): <description>`

### Commit Types and When to Use Them

**⚠️ IMPORTANT: Only `feat`, `fix`, and `perf` trigger releases and Docker builds!**

Use these types strategically to avoid unnecessary builds:

#### Types that TRIGGER releases/builds (use sparingly):
- **`feat`**: New user-facing feature or significant functionality
  - ✅ New API endpoint
  - ✅ New UI component or page
  - ✅ New integration with external service
  - ❌ Adding a comment to code
  - ❌ Updating .env.example

- **`fix`**: Bug fix that affects runtime behavior
  - ✅ Fix crash or error in application
  - ✅ Fix incorrect data processing
  - ✅ Fix API response issue
  - ❌ Fix typo in comment
  - ❌ Fix README formatting

- **`perf`**: Performance improvement that affects runtime
  - ✅ Optimize database query
  - ✅ Reduce memory usage
  - ✅ Improve API response time
  - ❌ Code cleanup without measurable impact

#### Types that DO NOT trigger releases (use for maintenance):
- **`docs`**: Documentation-only changes
  - ✅ Update README.md
  - ✅ Update .env.example
  - ✅ Add code comments
  - ✅ Update documentation files in /docs
  - ✅ Update copilot-instructions.md

- **`chore`**: Maintenance tasks, config changes, dependencies
  - ✅ Update dependencies in go.mod
  - ✅ Update .gitignore
  - ✅ Update Makefile
  - ✅ Update docker-compose.yml (non-functional)
  - ✅ Cleanup temporary files

- **`refactor`**: Code restructuring without changing behavior
  - ✅ Extract function/method
  - ✅ Rename variables for clarity
  - ✅ Move code between files
  - ❌ If it changes behavior, use `feat` or `fix`

- **`test`**: Adding or updating tests only
  - ✅ Add unit tests
  - ✅ Add integration tests
  - ✅ Update test fixtures

- **`style`**: Code style/formatting changes
  - ✅ Run gofmt
  - ✅ Fix linting issues
  - ✅ Adjust indentation

- **`ci`**: CI/CD configuration changes
  - ✅ Update GitHub Actions workflows
  - ✅ Update release configuration

### Decision Tree: Which Commit Type?

```
Does it change runtime behavior?
├─ YES → Does it add functionality?
│         ├─ YES → feat
│         └─ NO → Does it fix a bug?
│                  ├─ YES → fix
│                  └─ NO → Does it improve performance?
│                           ├─ YES → perf
│                           └─ NO → refactor
│
└─ NO → What does it change?
         ├─ Documentation → docs
         ├─ Tests → test
         ├─ Dependencies/Config → chore
         ├─ CI/CD → ci
         └─ Code formatting → style
```

### Examples - GOOD:

```bash
# TRIGGERS BUILD (runtime changes)
feat(api): add endpoint for bulk media deletion
fix(sync): correct torrent hash matching algorithm
perf(db): add index on media.created_at for faster queries

# DOES NOT TRIGGER BUILD (maintenance)
docs(config): update .env.example with Bazarr configuration
chore(deps): update Go dependencies to latest versions
refactor(handler): extract validation logic to separate function
test(repository): add unit tests for media queries
style(models): format code with gofmt
ci(release): update semantic-release configuration
```

### Examples - BAD (Spanish - DO NOT USE):

```bash
❌ feat(sync): implementar matching inteligente de torrents
❌ fix: arreglar tooltip en móviles
❌ actualizar dependencias
```

### Breaking Changes

If a commit introduces breaking changes, add `!` after the type or add `BREAKING CHANGE:` in the footer:

```bash
feat!: change API response format for /api/media
fix(api)!: remove deprecated endpoint /api/v1/cleanup

# Or in commit body:
feat(config): migrate to YAML configuration

BREAKING CHANGE: .env configuration is no longer supported, use config.yaml
```

## 🐛 Debugging Guidelines

**ALWAYS** add debug logging for troubleshooting:

```go
if s.logger.IsDebugEnabled() {
    s.logger.Debug("Processing media item",
        "media_id", media.ID,
        "title", media.Title,
        "age_days", media.AgeDays(),
        "tags", media.Tags,
    )
}
```

## ✅ Pre-commit Checklist

Before committing, **ALWAYS** verify:

- [ ] Code runs with `go run ./cmd/server`
- [ ] Tests pass with `go test ./...`
- [ ] Code is formatted with `gofmt -w .`
- [ ] No linter errors with `golangci-lint run`
- [ ] No sensitive data in code (API keys, passwords)
- [ ] Comments are clear and helpful
- [ ] Error messages are descriptive
- [ ] Logs use structured logging

## 🚀 Deployment Checklist

Before releasing, **ALWAYS** verify:

- [ ] Docker image builds successfully
- [ ] Health check endpoint works
- [ ] All environment variables documented
- [ ] Database migrations tested
- [ ] Configuration examples provided
- [ ] README.md updated
- [ ] CHANGELOG.md updated
- [ ] Version tagged in git

## 📚 Key Dependencies

**ALWAYS** use these specific versions (or newer compatible):

```go
// go.mod
require (
    github.com/gofiber/fiber/v2 v2.52.0
    github.com/gofiber/template/html/v2 v2.1.0
    gorm.io/gorm v1.25.5
    gorm.io/driver/sqlite v1.5.4
    gorm.io/driver/postgres v1.5.4
    github.com/go-resty/resty/v2 v2.11.0
    github.com/robfig/cron/v3 v3.0.1
    github.com/spf13/viper v1.18.2
    go.uber.org/zap v1.26.0
)
```

## 🎓 Learning Resources

When implementing features, refer to:

- **Fiber Framework**: https://docs.gofiber.io/
- **GORM**: https://gorm.io/docs/
- **Alpine.js**: https://alpinejs.dev/
- **Tailwind CSS**: https://tailwindcss.com/docs
- **Go Best Practices**: https://go.dev/doc/effective_go

## 🗣️ Communication Guidelines

**REMEMBER**: 
- 📢 Always respond to users in **Spanish**
- 📝 Technical documentation and code comments in **English**
- 🐛 Issue titles and descriptions in **Spanish**
- 💬 Pull request descriptions in **Spanish**
- 📖 User-facing documentation in **Spanish**

**⚠️ IMPORTANT - GitHub Communication:**
- ✅ **ALL GitHub interactions MUST be in Spanish**
- ✅ This includes: PR comments, code review comments, issue comments, discussions
- ✅ When approving/reviewing PRs, write comments in Spanish
- ✅ When requesting changes, explain them in Spanish
- ❌ NEVER write GitHub comments in English unless interacting with external contributors

**Examples:**
```markdown
✅ GOOD - GitHub PR Review in Spanish:
"Excelente trabajo implementando los componentes Alpine.js. El código se ve limpio y bien estructurado. ✅"

❌ BAD - GitHub PR Review in English:
"Great work implementing the Alpine.js components. Code looks clean and well structured. ✅"
```

---

**Last Updated**: October 25, 2025  
**Project Phase**: Initial Development  
**Stack Decision**: Go + Alpine.js (Proposal 3)
