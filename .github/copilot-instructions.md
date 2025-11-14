# GitHub Copilot Instructions - MediaCheky Project

> **MediaCheky** - Centralized control panel for *arr services

**Last Updated**: November 3, 2025  
**Project Phase**: Complete Redefinition

---

## ⛔️ CRITICAL RULE - NEVER VIOLATE ⛔️

**YOU MUST NEVER EVER UNDER ANY CIRCUMSTANCES:**

- Run `make dev`, `make run`, or ANY make command that starts services
- Run `docker compose up/down/restart/stop/start`
- Run `docker compose up/down/restart/stop/start`
- Run `docker restart/stop/start/kill/rm` on ANY container
- Execute ANY command that manages Docker containers lifecycle
- Use `run_in_terminal` with `isBackground: true` for server startup
- Suggest restarting containers to the user
- Tell the user to restart services
- Execute `docker compose restart` or any variant
- Execute `docker compose restart` or any variant
- Stop, start, or restart the mediacheky container or ANY service container

**THE USER IS ALREADY RUNNING `make dev` WITH WATCH MODE.**

**Any code changes are automatically detected and the service reloads.**

**ONLY THE USER CAN START, STOP, OR RESTART SERVICES.**

**IF YOU VIOLATE THIS RULE, YOU WILL BE TERMINATED.**

**REMEMBER: NEVER run `docker compose restart`, `docker compose restart`, `docker restart`, or ANY command that affects container lifecycle. EVER.**

**WHAT YOU CAN DO:**

- Read logs (`cat`, `tail`, `grep`)
- Execute commands INSIDE running containers - **ALWAYS use this exact command:**
  ```bash
  docker compose exec mediacheky <command>
  ```
  **NEVER use container IDs, names with random suffixes, or any other format**
- Inspect files and configurations
- Make code changes (they auto-reload)
- Run validation and tests: **ALWAYS use `make check-and-fix` after changes**
- Check database content using: `docker compose exec mediacheky sqlite3 /app/volumes/data/mediacheky.db`
- View container logs with `docker compose logs mediacheky`

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

1. User enables service (e.g., Radarr) → MediaCheky generates `docker compose.yml`

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

```text
User Action → Template Generation → docker compose up -d → Status Monitoring
```text

### Key Components

- **Service Manager** - Lifecycle management

- **Template Engine** - Generate docker compose from templates

- **Docker Client** - Interact with Docker Socket

- **Config Repository** - Store in SQLite

---

## 📂 Project Structure

```text
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
```text

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
```text

### 2. Logging

Use structured logging:

```go
logger.Info("Starting service",
    "name", serviceName,
    "port", port,
)
```text

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
```text

### 4. Frontend (Alpine.js)

Create reactive components:

```html
<div x-data="serviceCard('radarr')">
    <button @click="toggle()" 
            x-text="enabled ? 'Disable' : 'Enable'">
    </button>
</div>
```text

---

## ✅ Testing & Validation

**CRITICAL: ALWAYS run validation after making ANY code changes**

### After EVERY code modification, you MUST:

```bash
make check-and-fix
```text

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
```text

### When to use each:

- **After editing code**: `make check-and-fix` ✅

- **Before committing**: `make validate` (full)

- **Quick check**: `make validate-quick`

---

## 🔄 Git Commit Conventions

**⚠️ CRITICAL: ALL COMMITS MUST BE IN ENGLISH - NO EXCEPTIONS ⚠️**

**Format**: `<type>(<scope>): <description>`

**Use Conventional Commits specification (https://www.conventionalcommits.org/)**

### Commit Message Rules

1. **ALWAYS write commit messages in ENGLISH** 🇬🇧
2. **NEVER use Spanish or any other language** ❌
3. Use present tense: "add feature" not "added feature"
4. Use imperative mood: "fix bug" not "fixes bug"
5. Keep subject line under 72 characters
6. Capitalize first letter after type
7. No period at the end of subject line

### Types that TRIGGER releases

- `feat`: New feature
- `fix`: Bug fix
- `perf`: Performance improvement

### Types that DO NOT trigger releases

- `docs`: Documentation only changes
- `chore`: Maintenance tasks
- `refactor`: Code restructuring without feature/fix
- `test`: Adding or updating tests
- `style`: Code formatting, whitespace
- `ci`: CI/CD pipeline changes
- `build`: Build system or dependencies

### Examples (ALL IN ENGLISH)

```bash
# ✅ CORRECT - Triggers release
feat(services): add Radarr configuration panel
fix(docker): correct container status check
perf(db): optimize query performance

# ✅ CORRECT - Does not trigger release
docs(readme): update installation guide
chore(deps): update Go dependencies to latest
refactor(handler): simplify error handling logic
test(service): add unit tests for template engine
style(format): fix code formatting issues
ci(github): update workflow to use Go 1.25

# ❌ WRONG - Spanish (NEVER DO THIS)
feat(servicios): agregar panel de configuración
fix(docker): corregir verificación de estado
docs(readme): actualizar guía de instalación

# ❌ WRONG - Past tense
feat(services): added Radarr configuration panel
fix(docker): corrected container status

# ❌ WRONG - Missing type
add Radarr configuration panel
update installation guide
```

---

## 🗣️ Communication & Language Rules

### Commit Messages & Code
- � **Git commits**: **ALWAYS ENGLISH** (Conventional Commits) - NO EXCEPTIONS
- 📝 **Code & comments**: English
- � **Documentation files**: English
- 🏷️ **Variable/function names**: English

### User Interaction
- � **User responses**: Spanish (when talking to user)
- 🐛 **Issue/PR titles**: Spanish
- 💬 **PR descriptions**: Spanish
- � **Code review comments**: Spanish

### Summary
- **Write code in English** ✅
- **Write commits in English** ✅
- **Talk to user in Spanish** ✅
- **NEVER mix languages in commits** ❌

---

## 🎨 UI Standards (CRITICAL)

**ALL user feedback MUST use floating toast notifications**

### Toast Notifications Rules

- **Position**: Fixed top-right (`fixed top-4 right-4 z-50`)
- **Success**: Auto-dismiss after 5 seconds
- **Error**: Auto-dismiss after 8 seconds
- **Structure**: Icon + Message + Close button
- **Animations**: Fade + slide transitions

**NEVER use inline messages or static positioned feedback**

### Page Structure

- **Settings page has TWO tabs**: Services + Global Variables
- **NO separate `/global` route** - it's a tab in Settings
- **Service config URLs**: `/services/:name` (NOT `/services/:name/config`)

### Data Standards

- **Database**: `mediacheky.db` (NOT `keepercheky.db`)
- **Env Prefix**: `MEDIACHEKY_` (NOT `KEEPERCHEKY_`)
- **Network**: `mediacheky-net` (ALWAYS connected, NOT conditional)
- **Ports**: NOT exposed by default (optional via checkbox)

See [docs/UI_STANDARDS.md](../docs/UI_STANDARDS.md) for complete guidelines.

---

## 📚 Resources

- [UI Standards](../docs/UI_STANDARDS.md) - **READ THIS FIRST** for UI changes
- [Development Plan](../docs/PROJECT_PLAN.md) - Complete roadmap
- [UI Implementation](../docs/UI_IMPLEMENTATION.md) - Technical details
- [Docker API](https://docs.docker.com/engine/api/)
- [Fiber Framework](https://docs.gofiber.io/)
- [Alpine.js](https://alpinejs.dev/)

---

**Stack**: Go 1.25 + Fiber v2 + Alpine.js + Docker Socket  
**Port**: 7369 🦊
