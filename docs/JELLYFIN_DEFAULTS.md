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

### 2. First-Run Defaults Applied by Service Manager (New)

- **Location**: `internal/service/service_manager.go`
- **Purpose**: On first enable, copy defaults from `templates/jellyfin.defaults/` into the Jellyfin config volume.
- **Logic**:
  - Ensures `volumes/mediacheky-data/services-volumes/jellyfin/{data,cache,log}` exist
  - If `system.xml` is missing, copies all files from `templates/jellyfin.defaults/` into the config path
  - Skips copy when `system.xml` already exists (preserves user changes)

### 3. Docker Compose Template

- **File**: `templates/jellyfin.yml`
- **Changes**:
  - Standard compose (no custom entrypoint or defaults volume)
  - Uses global env (`PUID/PGID/TZ`) and optional host port exposure

## How It Works

### First Run (New Deployment)

```text
1. Service is enabled via API
2. ServiceManager initializes config directories
3. Detects missing /config/system.xml
4. Copies templates/jellyfin.defaults/* into /config
5. Container starts with defaults applied (wizard completed, Spanish UI, mediacheky server name)
```

### Subsequent Runs (Existing Deployment)

```text
1. Service is enabled or restarted
2. ServiceManager finds existing /config/system.xml
3. Skips copy to preserve user changes
4. Container starts using existing configuration
```

## Integration - Already Implemented ✅

### Implementation Location

- **Defaults dir helper**: `internal/service/template_engine.go` exposes `GetDefaultsDir()`
- **First-run copy**: `internal/service/service_manager.go` (`initializeJellyfinConfig`)

### Integration Flow

1. User enables Jellyfin via API (`/api/services/jellyfin/enable` or reset flow)
2. `initializeJellyfinConfig` creates directories and, if `system.xml` is missing, copies defaults from `templates/jellyfin.defaults/`
3. Docker Compose uses standard `templates/jellyfin.yml` (no entrypoint/extra volumes)

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

```text
templates/
├── jellyfin.defaults/          (NEW)
│   ├── system.xml              (NEW)
│   ├── database.xml            (NEW)
│   ├── encoding.xml            (NEW)
│   └── logging.default.json    (NEW)
└── jellyfin.yml                (MODIFIED - env + hosts; standard compose)

internal/service/
├── service_manager.go          (MODIFIED - first-run defaults copy)
└── template_engine.go          (MODIFIED - defaults dir helper)

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

