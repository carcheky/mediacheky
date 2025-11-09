# GitHub Copilot Instructions - MediaCheky Project

> **MediaCheky** - Centralized control panel for *arr services

**Last Updated**: November 3, 2025  
**Project Phase**: Complete Redefinition

---

## ⛔️ CRITICAL RULE - NEVER VIOLATE ⛔️

**YOU MUST NEVER:**

- Run `make dev`, `make run`, or ANY make command that starts services
- Run `docker-compose up/down/restart/stop`
- Run `docker start/stop/restart/kill/rm`
- Execute ANY command that manages Docker containers
- Use `run_in_terminal` with `isBackground: true` for server startup

**ONLY THE USER CAN START, STOP, OR RESTART SERVICES.**

**WHAT YOU CAN DO:**

- Read logs (`cat`, `tail`, `grep`)
- Execute commands INSIDE running containers (`docker exec`)
- Inspect files and configurations
- Make code changes
- Run validation and tests: **ALWAYS use `make check-and-fix` after changes**

---

## 🎯 Project Overview

**MediaCheky** is a web control panel for managing multimedia services (*arr ecosystem).

### What It Does

Centralized interface to:

1. **Dashboard** - Monitor service status (Jellyfin, Radarr, Sonarr, etc.)
2. **Settings** - Enable/disable services
3. **Services** - Configure each service with forms
4. **Global Variables** - Share config (PUID, PGID, TZ, paths)

### How It Works

1. User enables service (e.g., Radarr) → MediaCheky generates `docker-compose.yml`
2. MediaCheky starts container via Docker Socket
3. Dashboard shows real-time status

---

## 🏗️ Architecture

### Technology Stack

- **Backend**: Go 1.25 + Fiber v2
- **Frontend**: Alpine.js 3.x + Tailwind CSS
- **Database**: GORM v2 + SQLite
- **Docker**: Docker Socket API + Compose Templates
- **Target**: <25MB image, ~30-50MB RAM

### Hybrid Approach

**Docker Socket + Docker Compose Templates**

```
User Action → Template Generation → docker compose up -d → Status Monitoring
```

### Key Components

- **Service Manager** - Lifecycle management
- **Template Engine** - Generate docker-compose from templates
- **Docker Client** - Interact with Docker Socket
- **Config Repository** - Store in SQLite

---

## 📂 Project Structure

```
mediacheky/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                 # App configuration
│   ├── models/                 # Service, Template, GlobalConfig
│   ├── repository/             # Data access
│   ├── service/                # Business logic
│   │   ├── service_manager.go  # Service lifecycle
│   │   ├── docker_client.go    # Docker API wrapper
│   │   └── template_engine.go  # Compose generator
│   ├── handler/                # HTTP handlers
│   └── middleware/             # Middleware
├── web/
│   ├── templates/              # HTML templates
│   └── static/                 # CSS, JS
├── templates/                  # Docker Compose templates
│   ├── radarr.yml
│   └── sonarr.yml
├── volumes/
│   ├── services/               # Generated compose files
│   ├── config/                 # MediaCheky config
│   └── data/                   # SQLite DB
└── docs/
    └── PROJECT_PLAN.md         # Complete roadmap
```

---

## 🔑 Development Principles

### 1. Error Handling

**ALWAYS** handle errors explicitly:

```go
// ❌ BAD
result, _ := someFunction()

// ✅ GOOD
result, err := someFunction()
if err != nil {
    return fmt.Errorf("failed: %w", err)
}
```

### 2. Logging

Use structured logging:

```go
logger.Info("Starting service",
    "name", serviceName,
    "port", port,
)
```

### 3. Security

MediaCheky requires Docker socket → **root equivalent access**

**Security measures:**

- ✅ Validate all inputs
- ✅ Whitelist allowed images
- ✅ Sanitize paths (prevent `../`)
- ✅ Log all actions

```go
var allowedImages = map[string]bool{
    "linuxserver/radarr": true,
    "linuxserver/sonarr": true,
}
```

### 4. Frontend (Alpine.js)

Create reactive components:

```html
<div x-data="serviceCard('radarr')">
    <button @click="toggle()" 
            x-text="enabled ? 'Disable' : 'Enable'">
    </button>
</div>
```

---

## ✅ Testing & Validation

**CRITICAL: ALWAYS run validation after making ANY code changes**

### After EVERY code modification, you MUST:

```bash
make check-and-fix
```

This command will:
1. 🔧 Auto-fix code formatting (`gofmt`)
2. 📦 Clean dependencies (`go mod tidy`)
3. 📝 Check format compliance
4. 🔍 Run `go vet` static analysis
5. 🧪 Execute all tests

**NEVER skip this step.** If it fails:
- Fix the reported issues
- Run `make check-and-fix` again
- Repeat until it passes

### Available validation commands:

```bash
make validate        # Full validation (includes linting)
make validate-quick  # Fast check (format + vet + test)
make check-and-fix   # 🔧 AUTO-FIX + validate (USE THIS)
make lint-fix        # Only fix formatting
make test            # Run tests only
```

### When to use each:

- **After editing code**: `make check-and-fix` ✅
- **Before committing**: `make validate` (full)
- **Quick check**: `make validate-quick`

---

## 🔄 Git Commit Conventions

**Format**: `<type>(<scope>): <description>`

**ALL commit messages in ENGLISH** (Conventional Commits)

### Types that TRIGGER releases

- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement

### Types that DO NOT trigger releases

- `docs`: Documentation
- `chore`: Maintenance
- `refactor`: Code restructuring
- `test`: Tests
- `style`: Formatting
- `ci`: CI/CD changes

### Examples

```bash
# TRIGGERS BUILD
feat(services): add Radarr configuration panel
fix(docker): correct container status check

# DOES NOT TRIGGER BUILD
docs(readme): update installation guide
chore(deps): update dependencies
```

---

## 🗣️ Communication

- 📢 **User responses**: Spanish
- 📝 **Code/docs**: English
- 🐛 **Issues/PRs**: Spanish (titles and descriptions)
- 💬 **GitHub interactions**: Spanish (comments, reviews)

---

## 📚 Resources

- [Development Plan](../docs/PROJECT_PLAN.md) - Complete roadmap
- [Docker API](https://docs.docker.com/engine/api/)
- [Fiber Framework](https://docs.gofiber.io/)
- [Alpine.js](https://alpinejs.dev/)

---

**Stack**: Go 1.25 + Fiber v2 + Alpine.js + Docker Socket  
**Port**: 7369 🦊
