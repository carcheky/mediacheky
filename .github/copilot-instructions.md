# GitHub Copilot Instructions - MediaCheky Project

> **MediaCheky** - Centralized control panel for *arr services

**Last Updated**: December 18, 2025  
**Project Phase**: Active Development - API-First Approach

---

## 🚨 CRITICAL RULES - VIOLATION = CRITICAL FAILURE 🚨

### Always follow these rules without exception:
   - use mcp servers always as possible
   - edit code files directly, not by terminal commands


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

## 🔄 DEVELOPMENT METHOD: API-FIRST

**THIS IS THE MANDATORY DEVELOPMENT APPROACH:**

### 1. API Development FIRST

For any important feature (enable, disable, save, update, start, stop, delete, defaults, notify, progress, load, configs or related functions):

1. **Check if API exists** in [docs/API.md](../docs/API.md)
2. **If NO**: Develop the API endpoint FIRST
   - Implement handler in `internal/handler/`
   - Add route in `cmd/server/main.go`
   - Update [docs/API.md](../docs/API.md) with endpoint documentation
3. **If YES**: Proceed to next step
4. **Write tests** for the API endpoint - must pass `make check-and-fix`
5. **Only THEN**: Implement UI/Frontend

**NO EXCEPTIONS**: UI is implemented ONLY when API exists and tests pass.

### 2. Documentation BEFORE Development

**MANDATORY WORKFLOW:**

```
1. READ documentation for the feature/service
   ├─ docs/API.md (for API details)
   ├─ docs/PROJECT_PLAN.md (for architecture/context)
   ├─ docs/UI_STANDARDS.md (for UI patterns)
   └─ service-specific docs in docs/

2. DEVELOP the feature
   ├─ API endpoints (with handlers + tests)
   ├─ Update API documentation
   ├─ Update PROJECT_PLAN.md if behavior changed
   └─ Run: make check-and-fix

3. UPDATE documentation AFTER development
   ├─ docs/API.md (new endpoints or changes)
   ├─ docs/PROJECT_PLAN.md (if architecture changed)
   ├─ docs/UI_STANDARDS.md (if UI patterns changed)
   └─ Service-specific docs (if applicable)
```

### 3. Tests MUST Pass and Be Consistent

- **ALWAYS run**: `make check-and-fix` after code edits
- **NO obsolete tests**: Remove or update any failing tests
- **Tests must reflect current code**: No outdated test assertions
- **Tests are documentation**: They show how the API/feature works

### ✅✅✅ MANDATORY AFTER EVERY CODE EDIT ✅✅✅

```bash
# Step 1: Edit code files → Changes auto-saved

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

## 📚 Documentation Structure & Hierarchy

**Documentation is NOT redundant. Clear separation of concerns:**

### Global Documentation (Read FIRST before any development)

- **[PROJECT_PLAN.md](../docs/PROJECT_PLAN.md)** 
  - Complete development roadmap
  - Architecture overview
  - Phase progress & milestone tracking
  - Long-term vision
  - **Update when**: Feature architecture changes or phase status updates

- **[API.md](../docs/API.md)**
  - **GLOBAL section**: Base URL, response format, authentication, CORS, error handling, validation rules
  - **ENDPOINTS section (grouped by category)**:
    - Service Management endpoints
    - Global Configuration endpoints
    - Docker Information endpoints
    - Dashboard endpoints
  - **Update when**: New API endpoint added or existing endpoint changes

### UI Documentation

- **[UI_STANDARDS.md](../docs/UI_STANDARDS.md)**
  - UI patterns, component design, toast notifications
  - Color scheme, form design, spacing conventions
  - DO NOT include API details here
  - **Update when**: UI component patterns change or new patterns added

- **[UI_IMPLEMENTATION.md](../docs/UI_IMPLEMENTATION.md)**
  - Technical implementation details
  - Alpine.js components structure
  - File organization
  - **Update when**: Implementation approach changes

### Service-Specific Documentation

- **docs/SERVICE_NAME.md** (optional, if service needs detailed docs)
  - Service-specific configuration details
  - Service-specific quirks or requirements
  - **Update when**: Service-specific behavior changes

### Important Rules

✅ **DO**:
- Keep API documentation organized by endpoint category
- Document global configuration separately from service-specific APIs
- Update docs IMMEDIATELY after implementation
- Link between documents when relevant

❌ **DON'T**:
- Duplicate information across documents
- Keep obsolete or outdated documentation
- Mix service-specific and global documentation
- Leave "TODO" sections without action

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
