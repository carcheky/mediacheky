# MediaCheky Development Guide

> **Estado Actual**: v1.0.0-dev.17 - Desarrollo activo con la mayoría de features implementadas

## 🚀 Quick Start

## ✅ What's Implemented

### Backend Services (Go)
- ✅ **Core Application** - Fiber v2 web server con hot-reload
- ✅ **Database** - GORM v2 con SQLite/PostgreSQL
- ✅ **Configuration** - Sistema de config con Viper (YAML + env vars)
- ✅ **Logging** - Structured logging con niveles configurables

### Service Integrations
- ✅ **Radarr** - System info, queue, history, calendar, quality profiles
- ✅ **Sonarr** - System info, queue, history, calendar, quality profiles
- ✅ **Jellyfin** - Stats, sessions, recently added, activity
- ✅ **Jellyseerr** - Request statistics and management
- ✅ **Jellystst** - User activity, library stats, dashboard widgets
- ✅ **Bazarr** - Subtitle management, history, wanted lists
- ✅ **qBittorrent** - Transfer info, server state, active torrents, torrent properties

### Frontend (Alpine.js + Tailwind)
- ✅ **Dashboard** - Real-time stats, service health, activity timeline
- ✅ **Files/Health** - Storage health analysis, orphan detection, bulk actions
- ✅ **Media** - Library browser, filters, bulk delete, detailed views
- ✅ **Settings** - Service configuration, connection testing, two-tab layout
- 🚧 **Schedules** - Template ready, logic pending
- 🚧 **Logs** - Template ready, real-time streaming pending

### API Endpoints
- ✅ `/api/media` - Media CRUD operations
- ✅ `/api/files` - File operations and health analysis
- ✅ `/api/stats` - Dashboard statistics
- ✅ `/api/config` - Settings management
- ✅ `/api/sync` - Media library synchronization
- ✅ `/api/{service}/system` - Service system info for all integrations
- ✅ `/health` - Application health check

### Features
- ✅ Dry-run mode (safe testing)
- ✅ Exclusion tags
- ✅ File health analysis
- ✅ Bulk operations
- ✅ Service health monitoring
- ✅ Real-time statistics
- 🚧 Scheduled cleanups (in progress)
- 🚧 Leaving Soon collections (in progress)


### Prerequisites

- Docker 28.5+ installed
- Docker Compose V2
- 2GB RAM available
- ~500MB disk space

### First-Time Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/carcheky/mediacheky.git
   cd mediacheky
   ```

2. **Initialize development environment**
   ```bash
   make init
   ```

3. **Start development server**
   ```bash
   make dev
   ```

4. **Access the application**
   - Open http://localhost:8000
   - You should see the MediaCheky dashboard

## 🛠️ Development Workflow

### Using Hot-Reload (Recommended)

The development environment uses [Air](https://github.com/cosmtrek/air) for automatic hot-reload:

```bash
# Start with hot-reload
make dev

# The server will automatically restart when you edit:
# - Go files (*.go)
# - Templates (*.html)
# - Config files (*.yaml)
```

### Using Docker Compose Watch (Docker 28+)

Docker 28+ includes a `watch` feature that syncs file changes:

```bash
# Start with compose watch
make dev-watch
```

**What gets synced:**
- `./cmd` → Real-time sync
- `./internal` → Real-time sync
- `./pkg` → Real-time sync
- `./web` → Real-time sync
- `go.mod/go.sum` → Triggers rebuild

### Useful Commands

```bash
# View logs
make logs

# Open shell in container
make shell

# Run tests
make test

# Format code
make fmt

# Stop server
make stop

# Clean everything
make clean
```

## 📁 Project Structure

```
mediacheky/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/                     # Private application code
│   ├── config/                   # Configuration management
│   │   └── config.go
│   ├── models/                   # Database models (GORM)
│   │   └── models.go
│   ├── repository/               # Data access layer
│   │   └── repository.go
│   ├── service/                  # Business logic
│   │   ├── clients/              # External service clients
│   │   ├── cleanup/              # Cleanup strategies
│   │   └── scheduler/            # Job scheduling
│   ├── handler/                  # HTTP handlers (Fiber)
│   │   ├── handler.go
│   │   ├── dashboard.go
│   │   ├── media.go
│   │   └── ...
│   └── middleware/               # HTTP middleware
│       └── middleware.go
├── web/
│   ├── templates/                # Go html/template files
│   │   ├── layouts/
│   │   │   └── main.html
│   │   └── pages/
│   │       ├── dashboard.html
│   │       └── ...
│   └── static/                   # CSS, JS, images
│       ├── css/
│       ├── js/
│       └── images/
├── pkg/                          # Public/shared packages
│   ├── logger/
│   │   └── logger.go
│   └── utils/
├── data/                         # Runtime data (gitignored)
│   └── dev.db                    # SQLite database
├── config/                       # Configuration files
│   └── config.example.yaml
├── .air.toml                     # Air configuration
├── docker-compose.yml        # Development compose file
├── Dockerfile.dev                # Development Dockerfile
├── Makefile                      # Development commands
└── README.md
```

## 🔧 Configuration

### Environment Variables

The application uses environment variables with the prefix `KEEPERCHEKY_`:

```bash
# App
KEEPERCHEKY_APP_ENVIRONMENT=development
KEEPERCHEKY_APP_LOG_LEVEL=debug
KEEPERCHEKY_APP_DRY_RUN=true
KEEPERCHEKY_APP_LEAVING_SOON_DAYS=7

# Server
KEEPERCHEKY_SERVER_PORT=8000
KEEPERCHEKY_SERVER_HOST=0.0.0.0

# Database
KEEPERCHEKY_DATABASE_TYPE=sqlite
KEEPERCHEKY_DATABASE_PATH=./data/dev.db
```

### Configuration File

Alternatively, create `config/config.yaml`:

```yaml
app:
  environment: development
  log_level: debug
  dry_run: true
  leaving_soon_days: 7
  scheduler_enabled: false

server:
  port: "8000"
  host: "0.0.0.0"

database:
  type: sqlite
  path: ./data/dev.db

clients:
  radarr:
    enabled: false
    url: "http://radarr:7878"
    api_key: ""
  
  sonarr:
    enabled: false
    url: "http://sonarr:8989"
    api_key: ""
  
  jellyfin:
    enabled: false
    url: "http://jellyfin:8096"
    api_key: ""
    username: ""
    password: ""
```

## 🧪 Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test -v ./internal/service/...
```

### Writing Tests

Follow the standard Go testing conventions:

```go
// internal/repository/repository_test.go
package repository

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMediaRepository_GetAll(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()
    
    repo := NewMediaRepository(db)
    
    // Test
    media, err := repo.GetAll()
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, media)
}
```

## 🎨 Frontend Development

### Alpine.js Components

Create reactive components in templates:

```html
<div x-data="mediaList()" x-init="init()">
    <template x-for="item in media" :key="item.id">
        <div x-text="item.title"></div>
    </template>
</div>

<script>
function mediaList() {
    return {
        media: [],
        
        async init() {
            this.media = await this.fetchMedia();
        },
        
        async fetchMedia() {
            const response = await fetch('/api/media');
            return await response.json();
        }
    }
}
</script>
```

### Tailwind CSS

Use Tailwind utility classes directly in templates:

```html
<div class="bg-white shadow rounded-lg p-6">
    <h2 class="text-2xl font-bold text-gray-900">Title</h2>
</div>
```

## 🐛 Debugging

### VS Code Debugging

1. Install the Go extension for VS Code
2. Set breakpoints in your code
3. Press `F5` to start debugging
4. Or use the "Launch Package" configuration

### Docker Debugging

```bash
# View container logs
make logs

# Open shell in container
make shell

# Inside container, you can:
go run ./cmd/server
```

### Common Issues

**Port already in use:**
```bash
# Find and kill process using port 8000
lsof -ti:8000 | xargs kill -9

# Or use a different port
KEEPERCHEKY_SERVER_PORT=8001 make dev
```

**Database locked:**
```bash
# Stop all containers and clean
make clean
make init
make dev
```

## 📚 Additional Resources

- [Fiber Documentation](https://docs.gofiber.io/)
- [GORM Documentation](https://gorm.io/docs/)
- [Alpine.js Documentation](https://alpinejs.dev/)
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [Air Documentation](https://github.com/cosmtrek/air)

## 🤝 Contributing

1. Create a new branch: `git checkout -b feature/my-feature`
2. Make your changes
3. Run tests: `make test`
4. Format code: `make fmt`
5. Commit: `git commit -m "feat: add my feature"`
6. Push: `git push origin feature/my-feature`
7. Create Pull Request

## 📝 Coding Guidelines

See [.github/copilot-instructions.md](../.github/copilot-instructions.md) for detailed coding standards.

**Key principles:**
- ✅ ALWAYS handle errors explicitly
- ✅ Use structured logging
- ✅ Follow Repository Pattern
- ✅ Write tests for new features
- ✅ Document exported functions
- ✅ Use context for cancellation

---

**Happy coding!** 🚀

If you encounter any issues, check the [main documentation](../docs/README.md) or open an issue on GitHub.
