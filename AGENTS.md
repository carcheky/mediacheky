# AGENTS Guidelines for MediaCheky

**MediaCheky** is a centralized web control panel for managing *arr services and related multimedia services.

**Stack:** Go 1.25+, Fiber v2, Alpine.js 3.x, Tailwind CSS, Docker Socket API

---

## 🎯 Project Vision

**MediaCheky is a centralized control panel for managing multimedia services**, NOT a library cleanup tool.

See [PROJECT_PLAN.md](docs/PROJECT_PLAN.md) for complete development plan.

---

## 🔄 MANDATORY Development Workflow

### 1. API-First Approach

For ANY important feature (enable, disable, save, update, start, stop, delete, etc.):

```text
1. Check if API exists in docs/API.md
   ├─ NO  → Develop API endpoint FIRST (handler + tests)
   │       └─ Update docs/API.md after implementation
   └─ YES → Proceed to step 2

2. Ensure tests pass → make check-and-fix

3. Run Codacy analysis → codacy_cli_analyze

4. Only THEN → Implement UI/Frontend
```

**Critical Rule**: UI is implemented ONLY when API exists and tests pass.

### 2. Documentation Workflow

Read → Develop → Update:

```text
BEFORE development:
├─ Read docs/PROJECT_PLAN.md (architecture context)
├─ Read docs/API.md (existing endpoints)
├─ Read docs/UI_STANDARDS.md (UI patterns)
└─ Read service-specific docs (if applicable)

AFTER development:
├─ Update docs/API.md (new/modified endpoints)
├─ Update docs/PROJECT_PLAN.md (if architecture changed)
├─ Update docs/UI_STANDARDS.md (if UI patterns changed)
└─ Update service docs (if applicable)
```

### 3. Documentation Organization

**NO REDUNDANCY** - Clear separation of concerns:

| Document | Purpose | Update When |
| --- | --- | --- |
| `docs/API.md` | REST API endpoints (global + category) | New endpoint or API change |
| `docs/UI_STANDARDS.md` | UI patterns, design, components | UI pattern change |
| `docs/PROJECT_PLAN.md` | Architecture, roadmap, phases | Architecture or phase update |
| `docs/SERVICE_NAME.md` | Service-specific details (optional) | Service behavior change |

✅ **DO**: Keep docs organized by category, update immediately, link between docs
❌ **DON'T**: Duplicate info, leave obsolete docs, mix service-specific with global

### 4. Tests Must Be Consistent

- **ALWAYS run**: `make check-and-fix` after code edits
- **NO obsolete tests**: Remove or update failing tests immediately
- **Tests reflect code**: Test assertions must match current implementation
- **Tests are docs**: They show how the API/feature actually works

### 5. Complete Workflow After Each Edit

```bash
# 1. Edit code files (auto-saved)
# 2. Run validation
make check-and-fix

# 3. Run Codacy analysis
codacy_cli_analyze --rootPath /home/user/projects/mediacheky --file path/to/file.go

# 4. Wait for watch mode to reload (5-10s)
# 5. Verify via API calls or logs
# 6. Update documentation if behavior changed
```

---

### Mandatory References

- **Copilot Instructions**: [.github/copilot-instructions.md](.github/copilot-instructions.md)
- **Codacy MCP Rules**: [.github/instructions/codacy.instructions.md](.github/instructions/codacy.instructions.md)
- **UI Standards**: [docs/UI_STANDARDS.md](docs/UI_STANDARDS.md)
- **Project Plan**: [docs/PROJECT_PLAN.md](docs/PROJECT_PLAN.md)
- **API Documentation**: [docs/API.md](docs/API.md)

---

**Last Updated**: December 18, 2025  
**Project Phase**: Active Development - API-First Approach  
**Key Principle**: Documentation → API → Tests → UI
