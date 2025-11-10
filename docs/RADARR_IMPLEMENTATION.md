# Radarr Service Implementation - Complete Guide

## Overview

This document describes the complete implementation of the Radarr service as the MVP proof of concept for MediaCheky's service management system.

## Architecture

### Components Implemented

1. **Database Layer**
   - Service record seeding with default configuration
   - Template and schema association
   - Service logs for audit trail

2. **Service Manager** (`internal/service/service_manager.go`)
   - Orchestrates service lifecycle
   - Integrates Template Engine, Docker Compose, and Docker Client
   - Handles enable/disable, start/stop/restart operations
   - Manages configuration updates and compose file regeneration

3. **API Layer** (`internal/handler/service.go`)
   - RESTful endpoints for all service operations
   - Integration with Service Manager
   - Error handling and logging

4. **Frontend**
   - Dashboard with services section
   - Real-time status updates (auto-refresh every 15s)
   - Quick action buttons (Enable, Start, Stop, Restart)
   - Configuration page
   - Settings integration

## Files Modified/Created

### Backend
- `internal/database/migrations.go` - Added Radarr service seeding
- `internal/service/service_manager.go` - **NEW** Service lifecycle manager
- `internal/handler/handler.go` - Service Manager initialization
- `internal/handler/service.go` - Updated to use Service Manager
- `internal/handler/service_test.go` - Updated tests

### Frontend
- `web/templates/pages/dashboard.html` - Added services section

### Configuration
- `templates/radarr.yml` - Docker Compose template (already existed)
- `templates/radarr.schema.json` - JSON Schema (already existed)

## How It Works

### 1. Service Initialization (Database Seeding)

On first startup, the `seedServices()` function creates the Radarr service:

```go
Service{
    Name:        "radarr",
    DisplayName: "Radarr",
    Icon:        "🎬",
    Enabled:     false,
    Status:      "stopped",
    Image:       "linuxserver/radarr:latest",
    Port:        7878,
    Config: {
        "Image":         "linuxserver/radarr:latest",
        "ContainerName": "radarr",
        "Port":          7878,
        "Paths": {
            "Config":    "/data/config/radarr",
            "Movies":    "/data/media/movies",
            "Downloads": "/data/downloads",
        },
        "RestartPolicy": "unless-stopped",
    },
}
```

### 2. Enable Service Flow

**User Action**: Click "Enable" in Dashboard or Settings

**Backend Process**:
1. `ServiceHandler.EnableService()` receives HTTP POST
2. Calls `ServiceManager.EnableService()`
3. Service Manager:
   - Retrieves service from database
   - Calls `TemplateEngine.GenerateCompose()` with service config
   - Template Engine:
     - Loads global config (PUID, PGID, TZ)
     - Loads template from `templates/radarr.yml`
     - Merges service config with global config
     - Generates docker-compose.yml at `volumes/services/radarr/docker-compose.yml`
   - Updates service.Enabled = true in database
   - Logs action to ServiceLog
4. Returns success response

**Generated File**: `volumes/services/radarr/docker-compose.yml`

```yaml
version: '3.8'
services:
  radarr:
    image: linuxserver/radarr:latest
    container_name: radarr
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
    volumes:
      - /data/config/radarr:/config
      - /data/media/movies:/movies
      - /data/downloads:/downloads
    ports:
      - "7878:7878"
    restart: unless-stopped
```

### 3. Start Service Flow

**User Action**: Click "Start" button

**Backend Process**:
1. `ServiceHandler.StartContainer()` receives HTTP POST
2. Calls `ServiceManager.StartService()`
3. Service Manager:
   - Gets compose file path
   - Calls `DockerComposeClient.ComposeUp()`
   - Docker Compose Client executes: `docker compose -f /path/to/docker-compose.yml up -d`
   - Waits for container creation
   - Finds container by name using Docker Client
   - Updates service.Status = "running" and service.ContainerID in database
   - Logs action to ServiceLog
4. Returns success response

**Docker Command Executed**:
```bash
docker compose -f volumes/services/radarr/docker-compose.yml up -d
```

### 4. Stop Service Flow

**User Action**: Click "Stop" button

**Backend Process**:
1. `ServiceHandler.StopContainer()` receives HTTP POST
2. Calls `ServiceManager.StopService()`
3. Service Manager:
   - Gets compose file path
   - Calls `DockerComposeClient.ComposeDown()`
   - Docker Compose Client executes: `docker compose -f /path/to/docker-compose.yml down`
   - Updates service.Status = "stopped" and clears ContainerID in database
   - Logs action to ServiceLog
4. Returns success response

### 5. Update Configuration Flow

**User Action**: Modify settings in `/services/radarr` page and click "Save"

**Backend Process**:
1. `ServiceHandler.UpdateServiceConfig()` receives HTTP PUT with new config
2. Calls `ServiceManager.UpdateServiceConfig()`
3. Service Manager:
   - Updates service.Config in database
   - If service is enabled, regenerates docker-compose.yml with new config
   - Logs action to ServiceLog
4. Returns success response

**Note**: Container needs to be restarted for changes to take effect

### 6. Restart Service Flow

**User Action**: Click "Restart" button

**Backend Process**:
1. `ServiceHandler.RestartContainer()` receives HTTP POST
2. Calls `ServiceManager.RestartService()`
3. Service Manager:
   - Calls `StopService()`
   - Calls `StartService()`
   - Logs action to ServiceLog
4. Returns success response

## API Endpoints

### Service Operations

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/services` | List all services |
| GET | `/api/services/:name` | Get service details |
| POST | `/api/services/:name/enable` | Enable service |
| POST | `/api/services/:name/disable` | Disable service |
| POST | `/api/services/:name/start` | Start service container |
| POST | `/api/services/:name/stop` | Stop service container |
| POST | `/api/services/:name/restart` | Restart service container |
| PUT | `/api/services/:name/config` | Update service configuration |
| GET | `/api/services/:name/logs` | Get container logs |

### Web Pages

| Path | Description |
|------|-------------|
| `/` | Dashboard with services section |
| `/settings` | Settings page with enable/disable toggles |
| `/services/:name` | Service configuration page |

## Testing Guide

### Prerequisites

1. Docker and Docker Compose installed
2. MediaCheky running
3. Access to ports 7369 (MediaCheky) and 7878 (Radarr)

### Test Scenario 1: Enable and Start Radarr

**Steps**:
1. Open browser at `http://localhost:7369/`
2. In the Services section, find Radarr card
3. Click "Enable" button
4. Wait for success response (service card updates)
5. Click "Start" button
6. Wait ~5 seconds for container to start
7. Status should change to "Running ✅"
8. Click "Open UI →" link
9. Radarr UI should open in new tab at `http://localhost:7878`

**Verify**:
```bash
# Check that compose file was generated
cat volumes/services/radarr/docker-compose.yml

# Check that container is running
docker ps | grep radarr

# Check service logs
docker logs radarr
```

### Test Scenario 2: Configure Radarr

**Steps**:
1. In Services section, click "Configure" button on Radarr card
2. Modify port to 7879
3. Change movies path to `/data/media/movies-test`
4. Click "Save Configuration"
5. Go back to Dashboard
6. Click "Restart" on Radarr card
7. Wait for restart to complete
8. Click "Open UI →" - should now open at `http://localhost:7879`

**Verify**:
```bash
# Check updated compose file
cat volumes/services/radarr/docker-compose.yml | grep 7879

# Check container port mapping
docker ps | grep radarr | grep 7879
```

### Test Scenario 3: Stop and Disable Radarr

**Steps**:
1. In Dashboard, click "Stop" on Radarr card
2. Wait for container to stop
3. Status should change to "Stopped 📴"
4. Click "Enable" button (to toggle to disabled)
5. Service should show as disabled

**Verify**:
```bash
# Container should not be running
docker ps | grep radarr
# (should return nothing)

# Compose file should still exist
ls -la volumes/services/radarr/docker-compose.yml
```

### Test Scenario 4: Service Logs

**Steps**:
1. Make some actions (enable, start, stop)
2. Check ServiceLog entries in database

**Verify**:
```bash
# View the database
sqlite3 data/mediacheky.db

# Query service logs
SELECT * FROM service_logs WHERE service_id = 1 ORDER BY created_at DESC LIMIT 10;
```

Expected log entries:
- enable/success
- start/success
- stop/success
- config_update/success

## Default Configuration

### Radarr Service Defaults

```json
{
  "Image": "linuxserver/radarr:latest",
  "ContainerName": "radarr",
  "Port": 7878,
  "Paths": {
    "Config": "/data/config/radarr",
    "Movies": "/data/media/movies",
    "Downloads": "/data/downloads"
  },
  "RestartPolicy": "unless-stopped"
}
```

### Global Configuration

```json
{
  "PUID": "1000",
  "PGID": "1000",
  "TZ": "UTC"
}
```

## Troubleshooting

### Service won't start

**Check**:
1. Docker daemon is running: `docker info`
2. Paths exist and have correct permissions
3. Port 7878 is not already in use: `lsof -i :7878`
4. Container logs: `docker logs radarr`

### Compose file not generated

**Check**:
1. Template exists: `ls templates/radarr.yml`
2. Service is in database: Check via API `/api/services/radarr`
3. Application logs: `tail -f logs/mediacheky-dev.log`

### Container status not updating

**Check**:
1. Docker socket is accessible: `ls -la /var/run/docker.sock`
2. User has permissions to access Docker
3. Refresh the page (auto-refresh is every 15s)

## Security Considerations

### Docker Socket Access

⚠️ **IMPORTANT**: MediaCheky requires access to the Docker socket, which is equivalent to root access.

**Security measures implemented**:
1. Input validation on all service operations
2. Path sanitization to prevent directory traversal
3. Container name validation (regex)
4. Whitelisted images (configurable)
5. Audit logging of all actions

### Recommendations

1. Run MediaCheky in a dedicated network segment
2. Use firewall rules to restrict access
3. Review service logs regularly
4. Keep Docker and MediaCheky updated
5. Use secrets management for sensitive data

## Future Enhancements

### Planned Features

1. **Network Configuration**
   - VPN routing support
   - Custom networks
   - Network isolation

2. **Health Checks**
   - Automatic health monitoring
   - Restart on failure
   - Notifications

3. **Resource Limits**
   - CPU limits
   - Memory limits
   - Storage quotas

4. **Backup/Restore**
   - Configuration backup
   - Service state snapshots
   - Disaster recovery

5. **Multi-Service Dependencies**
   - Service startup order
   - Dependency management
   - Coordinated restarts

## References

- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [LinuxServer Radarr Image](https://hub.docker.com/r/linuxserver/radarr)
- [Radarr Documentation](https://wiki.servarr.com/radarr)
