# AGENTS Guidelines for MediaCheky

**MediaCheky** is a modern web-based media library cleanup manager - a complete rewrite of Janitorr with a beautiful UI. Built with Go + Fiber backend and Alpine.js + Tailwind CSS frontend.

**Stack:** Go 1.22+, Fiber v2, GORM v2, Alpine.js 3.x, Tailwind CSS, Docker

---

## ⛔️ CRITICAL RULE - NEVER VIOLATE ⛔️

**YOU MUST NEVER, UNDER ANY CIRCUMSTANCES:**

- Run `make dev` or `make run` or ANY make command that starts services
- Run `docker-compose up`, `docker-compose down`, `docker-compose restart`, or `docker-compose stop`
- Run `docker start`, `docker stop`, `docker restart`, `docker kill`, or `docker rm`
- Execute ANY command that starts, stops, restarts, kills, or removes Docker containers
- Use `run_in_terminal` with `isBackground: true` for ANY command that starts servers or services

**ONLY THE USER CAN START, STOP, OR RESTART SERVICES.**

**WHAT YOU CAN DO:**

- Read logs with `cat`, `tail`, `grep`, etc.
- Execute commands INSIDE running containers (`docker exec`) for debugging
- Inspect files and configurations
- Make code changes
- Run tests (but NOT start test servers)

**IF YOU NEED TO TEST SOMETHING, ASK THE USER TO START/RESTART THE SERVICE.**

---

## 📂 Project Structure Quick Reference

```
mediacheky/
├── cmd/server/main.go              # Application entry point
├── internal/                       # Private application code (NOT importable)
│   ├── config/                     # Configuration management
│   ├── models/                     # Database models (GORM)
│   ├── repository/                 # Data access layer
│   ├── service/                    # Business logic
│   │   ├── clients/                # External service clients (Radarr, Sonarr, etc.)
│   │   ├── cleanup/                # Cleanup strategies
│   │   └── scheduler/              # Job scheduling (cron)
│   ├── handler/                    # HTTP handlers (Fiber)
│   └── middleware/                 # HTTP middleware
├── web/
│   ├── templates/                  # Go html/template files
│   │   ├── layouts/
│   │   ├── pages/
│   │   └── components/
│   └── static/                     # CSS, JS, images
├── pkg/                            # Public/shared packages (reusable)
│   ├── filesystem/
│   ├── logger/
│   └── utils/
├── migrations/                     # Database migrations
├── scripts/                        # Build and utility scripts
└── docs/                           # Documentation
```

**Key Architecture Patterns:**

- **Repository Pattern**: Data access abstraction (`internal/repository/`)
- **Service Layer**: Business logic (`internal/service/`)
- **Handler Pattern**: HTTP request handling (`internal/handler/`)
- **Client Interface**: External services abstraction (`internal/service/clients/`)

---

## 🔧 Development Environment Setup

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- Make (optional, for convenience)

### Initial Setup

```bash
# Clone and enter directory
cd /home/user/projects/mediacheky

# Install Go dependencies
go mod download

# Copy example configuration
cp config/config.example.yaml config/config.yaml
```

**IMPORTANT:** Do NOT run `make dev` or `docker-compose up` - the user manages services.

---

## 🏗️ Building and Testing

### Build the Binary

```bash
# Build for current OS
go build -o bin/mediacheky ./cmd/server

# Build with optimizations (same as Dockerfile)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
  -ldflags="-w -s" \
  -o bin/mediacheky \
  ./cmd/server
```

### Run Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests in a specific package
go test ./internal/service/...

# Run a specific test
go test -run TestCleanupService_GetMediaToDelete ./internal/service
```

### Code Quality

```bash
# Format code
gofmt -w .

# Run linter (if installed)
golangci-lint run

# Vet code
go vet ./...
```

---

## 🐛 Debugging and Inspection

### Read Logs

**ALWAYS read logs directly after making changes:**

```bash
# Read the entire log
cat logs/mediacheky-dev.log

# Tail the last 100 lines
tail -n 100 logs/mediacheky-dev.log

# Follow logs in real-time
tail -f logs/mediacheky-dev.log

# Search for errors
grep -i error logs/mediacheky-dev.log

# Search for specific patterns
grep -i "media" logs/mediacheky-dev.log
```

### Inspect Running Containers

```bash
# List running containers
docker ps

# View logs of a container
docker logs mediacheky-app

# Execute commands inside a running container
docker exec -it mediacheky-app sh

# Inside container, you can:
ls -la /app
cat /app/config/config.yaml
ps aux
```

### Check Service Health

```bash
# Check if the service is responding
curl http://localhost:8000/health

# Check API endpoints
curl http://localhost:8000/api/media
```

### Inspect Database

```bash
# Access SQLite database (development)
docker exec -it mediacheky-app sqlite3 /app/data/mediacheky.db

# Inside SQLite:
.tables                 # List tables
.schema media           # Show schema for media table
SELECT * FROM media;    # Query data
.exit                   # Exit
```

---

## 🎨 Frontend Development (Alpine.js)

### Alpine.js Component Structure

Components are defined in `web/templates/` as inline Alpine.js components:

```html
<!-- Example: Media Card Component -->
<div x-data="mediaCard({{ .Media.ID }})" class="card">
    <img :src="media.poster_url" :alt="media.title">
    <h4 x-text="media.title"></h4>
    <button @click="exclude()" class="btn">Exclude</button>
</div>

<script>
function mediaCard(mediaId) {
    return {
        media: null,
        async init() { await this.fetchMedia(); },
        async fetchMedia() { /* ... */ },
        async exclude() { /* ... */ }
    }
}
</script>
```

### State Management

- **Local state**: Use `x-data` for component-specific state
- **Global state**: Use `Alpine.store()` for shared state (config, user settings)

### API Calls Best Practices

Always handle loading and error states:

```javascript
{
    data: null,
    loading: false,
    error: null,
    
    async fetch() {
        this.loading = true;
        this.error = null;
        try {
            const response = await fetch('/api/endpoint');
            if (!response.ok) throw new Error(`HTTP ${response.status}`);
            this.data = await response.json();
        } catch (err) {
            this.error = err.message;
        } finally {
            this.loading = false;
        }
    }
}
```

---

## 📝 Code Conventions

### Error Handling

**ALWAYS handle errors explicitly:**

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

### Logging

Use structured logging with context:

```go
logger.Info("Starting cleanup process",
    "media_count", len(mediaList),
    "dry_run", config.DryRun,
)
```

### Database Operations

Use transactions for multi-step operations:

```go
return r.db.Transaction(func(tx *gorm.DB) error {
    if err := tx.Save(media).Error; err != nil {
        return err
    }
    // ... more operations
    return nil
})
```

### Context Propagation

Always pass context through the call chain:

```go
func (s *Service) ProcessMedia(ctx context.Context, mediaID int) error {
    media, err := s.repo.GetByID(ctx, mediaID)
    // ...
}
```

---

## 🔄 Git Workflow

### Commit Message Format

**ALL commit messages MUST be in English** following Conventional Commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Commit Types

**⚠️ IMPORTANT: Only `feat`, `fix`, and `perf` trigger releases and Docker builds!**

**Types that TRIGGER builds (use sparingly):**

- `feat`: New user-facing feature or significant functionality
- `fix`: Bug fix that affects runtime behavior
- `perf`: Performance improvement that affects runtime

**Types that DO NOT trigger builds (use for maintenance):**

- `docs`: Documentation-only changes (README, comments, .env.example)
- `chore`: Maintenance tasks, config changes, dependencies
- `refactor`: Code restructuring without changing behavior
- `test`: Adding or updating tests only
- `style`: Code style/formatting changes (gofmt, linting)
- `ci`: CI/CD configuration changes

### Examples

```bash
# TRIGGERS BUILD
feat(api): add endpoint for bulk media deletion
fix(sync): correct torrent hash matching algorithm
perf(db): add index on media.created_at for faster queries

# DOES NOT TRIGGER BUILD
docs(config): update .env.example with Bazarr configuration
chore(deps): update Go dependencies to latest versions
refactor(handler): extract validation logic to separate function
test(repository): add unit tests for media queries
style(models): format code with gofmt
```

---

## 🚀 Common Development Tasks

| Task | Command | Notes |
|------|---------|-------|
| Build binary | `go build -o bin/mediacheky ./cmd/server` | Development build |
| Run tests | `go test ./...` | All tests |
| Format code | `gofmt -w .` | Before committing |
| Vet code | `go vet ./...` | Static analysis |
| Read logs | `cat logs/mediacheky-dev.log` | After changes |
| Tail logs | `tail -f logs/mediacheky-dev.log` | Real-time |
| Check health | `curl http://localhost:8000/health` | Service status |
| List containers | `docker ps` | Running services |
| Exec in container | `docker exec -it mediacheky-app sh` | Debug inside |

---

## 📊 Makefile Commands Reference

**⚠️ DO NOT RUN THESE - Only for reference:**

| Command | Purpose | **DO NOT USE** |
|---------|---------|----------------|
| `make dev` | Start development environment | ❌ User only |
| `make stop` | Stop services | ❌ User only |
| `make restart` | Restart services | ❌ User only |
| `make logs` | View logs | ❌ Use `cat` instead |
| `make build` | Build Docker image | ❌ User only |

**What you CAN use:**

```bash
# Read files
cat logs/mediacheky-dev.log
cat config/config.yaml

# Search files
grep -r "pattern" internal/

# Format and test
gofmt -w .
go test ./...
```

---

## 🔐 Security Best Practices

- ✅ Validate all user input
- ✅ Sanitize file paths to prevent directory traversal
- ✅ Use parameterized queries (GORM handles this)
- ✅ Never log sensitive data (API keys, passwords)
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

---

## 📚 Key Dependencies

```go
// Primary dependencies
github.com/gofiber/fiber/v2          // Web framework
github.com/gofiber/template/html/v2  // Template engine
gorm.io/gorm                         // ORM
gorm.io/driver/sqlite                // SQLite driver (dev)
gorm.io/driver/postgres              // PostgreSQL driver (prod)
github.com/go-resty/resty/v2         // HTTP client
github.com/robfig/cron/v3            // Job scheduler
github.com/spf13/viper               // Configuration
go.uber.org/zap                      // Structured logging
```

---

## 🗣️ Communication Guidelines

**REMEMBER:**

- 📢 Always respond to users in **Spanish**
- 📝 Technical documentation and code comments in **English**
- 🐛 Issue titles and descriptions in **Spanish**
- 💬 Pull request descriptions in **Spanish**
- ✅ **ALL GitHub interactions MUST be in Spanish** (PR comments, code reviews, etc.)

---

## 🎯 Before Committing - Checklist

- [ ] Code runs (if needed, ask user to test)
- [ ] Tests pass: `go test ./...`
- [ ] Code is formatted: `gofmt -w .`
- [ ] No linter errors: `golangci-lint run` (if available)
- [ ] No sensitive data in code
- [ ] Error messages are descriptive
- [ ] Logs use structured logging
- [ ] Commit message follows Conventional Commits format
- [ ] Commit message is in **English**

---

## 📖 Additional Documentation

For more detailed guidelines, see:

- `.github/copilot-instructions.md` - Full project guidelines and philosophy
- `.vscode/copilot-commit-message-instructions.md` - Detailed commit message rules
- `docs/` - Additional documentation and analysis

---

**Last Updated:** 2025-01-25  
**Format:** AGENTS.md v1.0 (OpenAI standard)  
**Project:** MediaCheky - Media Library Cleanup Manager
