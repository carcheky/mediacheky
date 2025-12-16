# Jellyfin Default Configuration Implementation

**Date**: December 16, 2025  
**Branch**: feat/jellyfin  
**Related PR**: #39

## Summary

This implementation captures the user's manual Jellyfin configuration and applies it as the default configuration for all new Jellyfin deployments, automatically on first run only.

## Changes Made

### 1. Default Configuration Files (New)
- **Location**: `templates/jellyfin.defaults/`
  - `system.xml`: Server configuration with:
    - `IsStartupWizardCompleted=true`
    - `PreferredMetadataLanguage=es`, `MetadataCountryCode=ES`, `UICulture=es`
    - `ServerName=mediacheky`
    - All other settings preserved from user's configuration
  - `database.xml`: SQLite configuration (`Jellyfin-SQLite`, `LockingBehavior=NoLock`)
  - `encoding.xml`: Transcoding settings (VAAPI support, HW encoding enabled, CRF H264=23/H265=28)
  - `logging.default.json`: Serilog configuration (Information level, 3-day retention)

### 2. First-Run Initialization Script (New)
- **Location**: `scripts/jellyfin-entrypoint.sh`
- **Purpose**: Initialize Jellyfin configuration only on first run
- **Logic**:
  - Checks if `/config/system.xml` exists
  - If NOT found (first run): copies all files from `/defaults` to `/config`
  - If found: skips initialization and proceeds normally
  - Delegates to original Jellyfin entrypoint (`/init`)

### 3. Docker Compose Template Update
- **File**: `templates/jellyfin.yml`
- **Changes**:
  - Added `entrypoint: /jellyfin-entrypoint.sh` to use custom initialization
  - Added volume mounts:
    - `/defaults` (read-only): mounted from `{{ .DefaultsPath }}`
    - `/jellyfin-entrypoint.sh` (read-only): mounted from `{{ .EntrypointPath }}`
  - Added environment variables: `LANG=es_ES.UTF-8`, `LC_ALL=es_ES.UTF-8` to align with locale defaults

## How It Works

### First Run (New Deployment)
```
1. Container starts
2. Custom entrypoint runs
3. Checks for /config/system.xml → NOT FOUND
4. Copies /defaults/*.xml and /defaults/logging.default.json to /config
5. Proceeds to Jellyfin initialization with pre-configured settings
6. User sees wizard already completed, Spanish UI, mediacheky as server name
```

### Subsequent Runs (Existing Deployment)
```
1. Container starts
2. Custom entrypoint runs
3. Checks for /config/system.xml → FOUND
4. Skips copy, proceeds directly to Jellyfin
5. Existing configuration preserved
```

## Integration - Already Implemented ✅

### Template Variables Auto-Injected
The `TemplateEngine` automatically provides these variables in `TemplateData`:
- `{{ .DefaultsPath }}` - Calculated as `{project-root}/templates/{serviceName}.defaults`
- `{{ .EntrypointPath }}` - Calculated as `{project-root}/scripts/{serviceName}-entrypoint.sh`

### Implementation Location
**File**: `internal/service/template_engine.go`

**Changes Made**:
1. Added `DefaultsPath` and `EntrypointPath` fields to `TemplateData` struct (line ~46-47)
2. Updated `buildTemplateData()` function to auto-calculate paths (line ~375-376):
   ```go
   // Set paths for service defaults and entrypoint (for first-run initialization like Jellyfin)
   data.DefaultsPath = filepath.Join(filepath.Dir(filepath.Dir(te.servicesDir)), "templates", data.ContainerName+".defaults")
   data.EntrypointPath = filepath.Join(data.ScriptsPath, data.ContainerName+"-entrypoint.sh")
   ```

### How It Works
1. User creates Jellyfin service via API with `{ "Image": "linuxserver/jellyfin:latest", ... }`
2. Handler calls `ServiceManager.EnableService("jellyfin")`
3. Service Manager calls `TemplateEngine.GenerateCompose("jellyfin", config)`
4. Template Engine:
   - Loads `templates/jellyfin.yml`
   - Sets `DefaultsPath = /app/templates/jellyfin.defaults`
   - Sets `EntrypointPath = /app/scripts/jellyfin-entrypoint.sh`
   - Renders template with these paths
5. Generated `docker-compose.yml` includes:
   ```yaml
   volumes:
     - /app/templates/jellyfin.defaults:/defaults:ro
     - /app/scripts/jellyfin-entrypoint.sh:/jellyfin-entrypoint.sh:ro
   entrypoint: /jellyfin-entrypoint.sh
   ```
6. Container starts with custom entrypoint that copies defaults on first run

## Testing

### Test First-Run Initialization
1. Fresh deployment (new `{{ .Paths.Config }}` volume)
2. Start Jellyfin container
3. Verify `/config/system.xml` contains Spanish settings, wizard completed
4. Verify `/config/encoding.xml` has correct transcoding settings

### Test Subsequent Runs
1. Container with existing `/config/system.xml`
2. Modify config (e.g., change server name)
3. Restart container
4. Verify modified settings are preserved (not overwritten)

## Files Modified/Created

```
templates/
├── jellyfin.defaults/          (NEW)
│   ├── system.xml              (NEW)
│   ├── database.xml            (NEW)
│   ├── encoding.xml            (NEW)
│   └── logging.default.json    (NEW)
└── jellyfin.yml                (MODIFIED - added entrypoint + volumes)

scripts/
└── jellyfin-entrypoint.sh      (NEW - initialization script)

internal/service/
└── template_engine.go          (MODIFIED - added DefaultsPath, EntrypointPath to TemplateData)

docs/
└── JELLYFIN_DEFAULTS.md        (THIS FILE - NEW)
```

## Related Documentation

- [UI_IMPLEMENTATION.md](UI_IMPLEMENTATION.md) - Jellyfin service configuration details
- [PROJECT_PLAN.md](PROJECT_PLAN.md) - Jellyfin implementation roadmap

## Future Enhancements

- Extend pattern to other services (Radarr, Sonarr, etc.)
- Allow customization of defaults via configuration
- Support per-library default metadata preferences
- Automated configuration backup/restore

