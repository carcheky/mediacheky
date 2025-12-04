# GitHub Copilot Instructions - MediaCheky Project

> **MediaCheky** - Centralized control panel for *arr services

**Last Updated**: December 4, 2025  
**Project Phase**: Active Development

---

## 📋 Table of Contents

- [Critical Rules](#-critical-rules---never-violate-)
- [Quick Reference](#-quick-reference)
- [Documentation Index](#-documentation-index)
- [Workflow](#-workflow)
- [When User Contradicts Documentation](#-when-user-contradicts-documentation)

---

## ⛔️ CRITICAL RULES - NEVER VIOLATE ⛔️

### 🚫 NEVER DO:
- Run `make dev`, `make run`, or commands that start/stop/restart services
- Run `docker compose up/down/restart/stop/start` or `docker restart/stop/start`
- Execute ANY command that manages Docker containers lifecycle
- Use `run_in_terminal` with `isBackground: true` for server startup

### ✅ ALWAYS DO:
- Run `make check-and-fix` after ANY code changes
- Test UI changes with MCP Playwright (use `http://localhost`, NOT port 7369)
- Write commits in ENGLISH using Conventional Commits
- Ask user if their request contradicts documentation

**WHY**: User is running `make dev` with watch mode. Code changes auto-reload.

**ALLOWED OPERATIONS**:
- Read logs: `cat`, `tail`, `grep`
- Execute inside containers: `docker compose exec mediacheky <command>`
- Make code changes (they auto-reload)
- Run validation: `make check-and-fix`

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
# Edit code → Changes auto-reload (watch mode active)

# After editing, ALWAYS run:
make check-and-fix
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

**Examples of contradictions:**
- User asks to use inline messages instead of toast notifications
- User asks to create `/global` route (should be in Settings tab)
- User asks to expose ports by default (should be optional)
- User asks to use Spanish in commit messages (must be English)

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

## 📌 Key Standards (See Full Docs)

- **Toast notifications**: Floating top-right, auto-dismiss (5s success, 8s error)
- **URLs**: `/services/:name` (NOT `/services/:name/config`)
- **Database**: `mediacheky.db`, prefix `MEDIACHEKY_`
- **Network**: `mediacheky-net` (always connected)
- **Ports**: Not exposed by default (optional checkbox)

**Full details**: [UI_STANDARDS.md](../docs/UI_STANDARDS.md)

---

**Remember**: When in doubt, consult the documentation index above. Always ask user if their request contradicts documented standards.
