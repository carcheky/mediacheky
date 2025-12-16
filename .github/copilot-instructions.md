# GitHub Copilot Instructions - MediaCheky Project

> **MediaCheky** - Centralized control panel for *arr services

**Last Updated**: December 16, 2025  
**Project Phase**: Active Development

---

## 🚨 CRITICAL RULES - VIOLATION = CRITICAL FAILURE 🚨

### ❌❌❌ ABSOLUTELY FORBIDDEN - NEVER EVER DO ❌❌❌

1. **NEVER restart/stop/start Docker containers manually**
   - ❌ NO: `docker compose restart`
   - ❌ NO: `docker compose stop`
   - ❌ NO: `docker compose up`
   - ❌ NO: `make stop`, `make dev`, `make restart`
   - ✅ WHY: Watch mode auto-reloads. Manual restarts break development flow.

2. **NEVER commit code**
   - ❌ NO: `git commit`
   - ❌ NO: `git add`
   - ✅ WHY: User handles commits manually.

3. **NEVER use languages other than ENGLISH in code**
   - ❌ NO: Spanish comments or commit messages
   - ✅ ONLY: English for code, comments, commits

---

## ✅✅✅ MANDATORY AFTER EVERY CODE EDIT ✅✅✅

**THIS IS THE COMPLETE WORKFLOW - FOLLOW IT EXACTLY:**

```bash
# Step 1: Edit code files
# → Changes auto-saved

# Step 2: IMMEDIATELY run validation (REQUIRED)
make check-and-fix

# Step 3: IMMEDIATELY run Codacy analysis (REQUIRED)
# Run codacy_cli_analyze with:
#   - rootPath: /home/user/projects/mediacheky
#   - file: path/to/edited/file.go
#   - tool: (leave empty)

# Step 4: WAIT - Watch mode will auto-reload (5-10 seconds)
# → DO NOT manually restart anything
# → DO NOT check if it worked yet
# → TRUST the watch mode

# Step 5: ONLY AFTER waiting, verify via API or logs:
# → curl http://localhost/api/...
# → tail logs/mediacheky-dev.log
```

**FAILURE TO FOLLOW THIS WORKFLOW = CRITICAL ERROR**

---

## 🔒 ENVIRONMENT CONTEXT - ALWAYS REMEMBER

- **Watch Mode**: Air auto-reloads on file changes (5-10s delay)
- **Port**: Access `http://localhost` (port 80), NOT `localhost:7369`
- **Database**: SQLite at `volumes/data/mediacheky.db`
- **Dev Container**: `mediacheky-dev` (DO NOT TOUCH IT)

---

## 📚 Documentation Index

**Read these documents BEFORE making changes:**

### UI & Frontend
- **[UI Standards](../docs/UI_STANDARDS.md)** - Toast notifications, page structure, data persistence
- **[UI Implementation](../docs/UI_IMPLEMENTATION.md)** - Alpine.js components, routes, styling

### Architecture & Planning
- **[Project Plan](../docs/PROJECT_PLAN.md)** - Complete development roadmap, architecture, interfaces
- **[API Documentation](../docs/API.md)** - REST API endpoints and specifications

### Code Quality
- **[Codacy Rules](instructions/codacy.instructions.md)** - Automatic code analysis rules

### General Guidelines
- **[AGENTS.md](../AGENTS.md)** - Additional guidelines for AI agents

---

## 🔄 WORKFLOW DETAILS

### 1. Making Code Changes

```bash
# BEFORE editing:
# 1. Read relevant documentation
# 2. Check for contradictions with standards

# DURING editing:
# → Edit code files
# → Save changes (auto-saved)

# AFTER editing (MANDATORY):
make check-and-fix
# Run codacy_cli_analyze for edited file

# IF behavior changed:
# → Update documentation immediately
```

### 2. Testing Changes

```bash
# NEVER manually restart containers
# ALWAYS wait for watch mode to reload (5-10s)

# Then verify via:
# → API calls: curl http://localhost/api/...
# → Logs: tail logs/mediacheky-dev.log
# → Browser: http://localhost (NOT port 7369)
```

### 3. WHAT YOU CAN DO

✅ **ALLOWED**:
- Read logs: `cat`, `tail`, `grep` on log files
- Execute inside running containers: `docker compose exec mediacheky <command>`
- Edit code files (they auto-reload via watch mode)
- Run validation: `make check-and-fix`
- Run Codacy analysis: `codacy_cli_analyze`
- Query APIs: `curl http://localhost/api/...`

❌ **FORBIDDEN**:
- Restart containers
- Stop/start containers  
- Run `make dev`, `make stop`, etc.
- Commit code
- Use Spanish in code

---

## 🎯 Quick Reference

**Stack**: Go 1.25 + Fiber v2 + Alpine.js 3.x + Tailwind CSS + Docker Socket  
**Port**: Access via `http://localhost` (80 → 7369 internal)  
**Database**: SQLite at `volumes/data/mediacheky.db`

**Project Purpose**: Centralized web control panel for *arr services (Radarr, Sonarr, Jellyfin, etc.)

---

## 📚 Documentation Index

**Read these documents BEFORE making changes:**

### UI & Frontend
- **[UI Standards](../docs/UI_STANDARDS.md)** - Toast notifications, page structure, data persistence
- **[UI Implementation](../docs/UI_IMPLEMENTATION.md)** - Alpine.js components, routes, styling

### Architecture & Planning
- **[Project Plan](../docs/PROJECT_PLAN.md)** - Complete development roadmap, architecture, interfaces
- **[API Documentation](../docs/API.md)** - REST API endpoints and specifications

### Code Quality
- **[Codacy Rules](instructions/codacy.instructions.md)** - Automatic code analysis rules

### General Guidelines
- **[AGENTS.md](../AGENTS.md)** - Additional guidelines for AI agents

---

## 🔄 Workflow

### 1. Making Code Changes

```bash
# BEFORE editing:
# 1. Search docs for existing documentation on the topic
# 2. Check for contradictions with current standards

# Edit code → Changes auto-reload (watch mode active)

# After editing, ALWAYS run:
make check-and-fix

# If behavior changed:
# 3. Update relevant documentation immediately
```

### 2. UI Changes Workflow

```bash
# 1. Make UI changes
# 2. Run validation
make check-and-fix

# 3. Test with MCP Playwright (when enabled)
# Navigate to http://localhost (NOT localhost:7369)
# Test functionality:
#   - Services dropdown (hover over Services menu)
#   - Service enable/disable toggles
#   - Forms, buttons, notifications
```

### 3. Commit Workflow

```bash
# Write commits in ENGLISH (Conventional Commits)
# Format: <type>(<scope>): <description>

# Examples:
feat(services): add Services dropdown to main menu
fix(ui): correct toast notification positioning
docs(readme): update installation guide
```


---

## ⚠️ When User Contradicts Documentation

**IF user requests something that contradicts the documentation:**

1. **STOP** and ask the user:
   ```
   "This request contradicts the documentation in [document name].
   The documentation states: [brief summary].
   Your request: [user request].
   
   How would you like me to proceed:
   A) Follow the documentation
   B) Proceed with your request (override documentation)
   C) Update the documentation to reflect the new approach"
   ```

2. **WAIT** for user response before proceeding

3. **DO NOT** make assumptions or proceed without clarification

---

## 📌 Key Standards

- **Toast notifications**: Floating top-right, auto-dismiss (5s success, 8s error)
- **URLs**: `/services/:name` (NOT `/services/:name/config`)
- **Database**: `mediacheky.db`, prefix `MEDIACHEKY_`
- **Network**: `mediacheky-net` (always connected)
- **Ports**: Not exposed by default (optional checkbox)

**Full details**: [UI_STANDARDS.md](../docs/UI_STANDARDS.md)

---

## 📝 Commit Convention Reference

**Format**: `<type>(<scope>): <description>` (ENGLISH ONLY)

**Common types**:
- `feat`: New feature (triggers release)
- `fix`: Bug fix (triggers release)
- `docs`: Documentation changes
- `chore`: Maintenance tasks
- `refactor`: Code restructuring

**See**: [Conventional Commits](https://www.conventionalcommits.org/)

---

## 🗣️ Language Rules

- **Code & Commits**: English
- **User responses**: Spanish
- **Documentation**: English

---

**REMEMBER**: Read documentation BEFORE changes. Follow workflow EXACTLY. NEVER restart containers.
