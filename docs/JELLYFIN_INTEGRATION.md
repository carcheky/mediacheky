# Jellyfin Integration - New Features

## Overview

This document describes the new Jellyfin API integration features added to KeeperCheky, providing enhanced monitoring and statistics capabilities.

## New API Functions

### 1. Active Sessions (`GetActiveSessions`)

**Purpose**: Monitor real-time streaming activity on your Jellyfin server.

**API Endpoint**: `GET /api/jellyfin/sessions`

**Response Example**:
```json
{
  "sessions": [
    {
      "Id": "session-123",
      "UserId": "user-456",
      "UserName": "john_doe",
      "Client": "Web",
      "DeviceName": "Chrome",
      "DeviceId": "device-789",
      "ApplicationVersion": "10.8.13",
      "RemoteEndPoint": "192.168.1.100",
      "LastActivityDate": "2024-01-15T20:30:00Z",
      "SupportsRemoteControl": true,
      "NowPlayingItem": {
        "Id": "movie-123",
        "Name": "Inception",
        "Type": "Movie",
        "MediaType": "Video"
      },
      "PlayState": {
        "PositionTicks": 36000000000,
        "CanSeek": true,
        "IsPaused": false,
        "IsMuted": false,
        "VolumeLevel": 100,
        "AudioStreamIndex": 0,
        "SubtitleStreamIndex": -1,
        "MediaSourceId": "source-123",
        "PlayMethod": "DirectPlay"
      },
      "TranscodingInfo": null
    }
  ],
  "count": 1
}
```

**Use Cases**:
- Monitor who is currently watching content
- Identify transcoding sessions (performance impact)
- Track playback methods (DirectPlay vs Transcode)
- Display real-time streaming activity on dashboard

**Dashboard Integration**:
- Shows active sessions with user, device, and content information
- Displays playback status (playing/paused)
- Indicates playback method (DirectPlay, DirectStream, Transcode)
- Auto-refreshes every 15 seconds

---

### 2. Library Statistics (`GetLibraryStats`)

**Purpose**: Get comprehensive statistics about your Jellyfin media library.

**API Endpoint**: `GET /api/jellyfin/stats`

**Response Example**:
```json
{
  "total_items": 1250,
  "movie_count": 450,
  "series_count": 80,
  "episode_count": 720,
  "album_count": 0,
  "song_count": 0,
  "total_size": 5497558138880,
  "library_folders": [
    {
      "Name": "Movies",
      "Locations": ["/media/movies"],
      "ItemId": "folder-123"
    },
    {
      "Name": "TV Shows",
      "Locations": ["/media/tv"],
      "ItemId": "folder-456"
    }
  ]
}
```

**Use Cases**:
- Display total library size and item counts
- Monitor library growth over time
- Verify library folder configurations
- Generate reports for library statistics

**Integration Points**:
- Available via API for custom dashboards
- Can be used for monitoring library health
- Useful for capacity planning

---

### 3. Recently Added Items (`GetRecentlyAdded`)

**Purpose**: Track newly added content to your Jellyfin library.

**API Endpoint**: `GET /api/jellyfin/recently-added?limit=20`

**Query Parameters**:
- `limit` (optional, default: 20): Number of items to return

**Response Example**:
```json
{
  "items": [
    {
      "id": "movie-789",
      "name": "The Matrix",
      "type": "Movie",
      "date_created": "2024-01-15T18:00:00Z",
      "poster_url": "http://jellyfin:8096/Items/movie-789/Images/Primary?tag=abc123",
      "overview": "A computer hacker learns..."
    }
  ],
  "count": 20
}
```

**Use Cases**:
- Display recently added content on dashboard
- Create "New Arrivals" collections
- Track library additions for auditing
- Notify users of new content

**Dashboard Integration**:
- Gallery view with poster images (up to 12 items)
- Shows relative dates ("Hoy", "Ayer", "Hace 3 días")
- Visual indicator of content type (🎬 for movies, 📺 for series)

---

### 4. Activity Log (`GetActivityLog`)

**Purpose**: Retrieve server activity and event logs from Jellyfin.

**API Endpoint**: `GET /api/jellyfin/activity?limit=50`

**Query Parameters**:
- `limit` (optional, default: 50): Number of entries to return

**Response Example**:
```json
{
  "entries": [
    {
      "id": 12345,
      "name": "AuthenticationSucceeded",
      "type": "AuthenticationSuccess",
      "user_id": "user-123",
      "date": "2024-01-15T20:30:00Z",
      "severity": "Info",
      "short_overview": "User john_doe logged in"
    },
    {
      "id": 12346,
      "name": "SubtitleDownloadFailure",
      "type": "SubtitleDownload",
      "user_id": "",
      "date": "2024-01-15T19:00:00Z",
      "severity": "Error",
      "short_overview": "Failed to download subtitle for Movie XYZ"
    }
  ],
  "count": 50
}
```

**Use Cases**:
- Monitor server events and errors
- Track user authentication activity
- Debug subtitle download issues
- Audit server operations

**Available for Integration**:
- Can be displayed in logs/monitoring page
- Useful for troubleshooting issues
- Security audit trail

---

## Dashboard Features

### Active Sessions Widget

**Location**: Dashboard page (appears when sessions are active)

**Features**:
- Real-time display of active streaming sessions
- User information (name, device, client)
- Currently playing content details
- Playback status (playing/paused position)
- Playback method indicator:
  - ⚡ **Direct Play**: No transcoding (best performance)
  - 🌊 **Direct Stream**: Container remuxing only
  - 🔄 **Transcode**: Full transcoding (CPU intensive)
- Auto-refresh every 15 seconds

**Visual Indicators**:
- Green "En reproducción" badge for active playback
- Yellow pause indicator for paused content
- Colored badges for playback methods
- Relative timestamps for playback position

---

### Recently Added Widget

**Location**: Dashboard page (appears when items exist)

**Features**:
- Grid gallery of up to 12 recently added items
- Poster image display (with fallback emoji)
- Content type indicator (🎬/📺)
- Relative date display in Spanish
- Hover effects for better UX

**Visual Design**:
- 2-column grid on mobile
- 4-column grid on medium screens
- 6-column grid on large screens
- 2:3 aspect ratio poster containers
- Smooth hover transitions

---

## Settings Page Integration

The Settings page automatically displays detailed Jellyfin system information when testing the connection:

**Displayed Information**:
- Server version
- Server ID and name
- Operating system details
- Product name
- Local address

This information is retrieved via the existing `GetSystemInfo()` function and displayed in the standard system info card format.

---

## Files Page Integration

The Files (Health) page already integrates Jellyfin playback data:

**Existing Features**:
- ✅ **Watched status**: Shows if content has been viewed
- 👁️ **Unwatched indicator**: Highlights content never played
- 🟢 **Health classification**: Considers watch status in health scoring

**Data Points Used**:
- `has_been_watched`: Boolean indicating if content was viewed
- `last_watched`: Timestamp of last playback (if available)
- `in_jellyfin`: Whether the file exists in Jellyfin library

---

## Technical Implementation

### Client Layer

**File**: `internal/service/clients/jellyfin.go`

All new functions follow the same patterns:
1. Use `callWithRetry()` for resilient API calls
2. Implement exponential backoff on failures
3. Respect context cancellation
4. Log detailed information at appropriate levels
5. Return structured Go types (not raw JSON)

### Service Layer

**File**: `internal/service/sync_service.go`

Type-safe wrapper methods:
```go
func (s *SyncService) GetJellyfinActiveSessions(ctx context.Context) ([]clients.SessionInfo, error)
func (s *SyncService) GetJellyfinLibraryStats(ctx context.Context) (*clients.LibraryStats, error)
func (s *SyncService) GetJellyfinRecentlyAdded(ctx context.Context, limit int) ([]clients.RecentlyAddedItem, error)
func (s *SyncService) GetJellyfinActivityLog(ctx context.Context, limit int) ([]clients.ActivityLogEntry, error)
```

### Handler Layer

**File**: `internal/handler/settings.go`

RESTful HTTP handlers with proper:
- Timeout configuration (10-30s based on operation)
- Error handling and logging
- Standardized JSON responses
- Query parameter validation

### Routes

**File**: `cmd/server/main.go`

```go
// Jellyfin endpoints
api.Get("/jellyfin/stats", h.Settings.GetJellyfinStats)
api.Get("/jellyfin/sessions", h.Settings.GetJellyfinSessions)
api.Get("/jellyfin/recently-added", h.Settings.GetJellyfinRecentlyAdded)
api.Get("/jellyfin/activity", h.Settings.GetJellyfinActivity)
```

---

## Testing

### Unit Tests

**File**: `internal/service/clients/jellyfin_test.go`

Comprehensive test coverage:
- ✅ System info retrieval
- ✅ Active sessions (with and without playback)
- ✅ Library statistics calculation
- ✅ Recently added items sorting
- ✅ Activity log retrieval
- ✅ Error handling scenarios
- ✅ Empty result sets
- ✅ Default parameter values

**Run tests**:
```bash
go test -v ./internal/service/clients/ -run TestJellyfin
```

**Test Coverage**: All new functions have unit tests with mocked HTTP responses.

---

## Configuration

No additional configuration is required. The new features automatically work when Jellyfin is enabled:

```yaml
clients:
  jellyfin:
    enabled: true
    url: "http://jellyfin:8096"
    api_key: "your-api-key-here"
```

Or via environment variables:
```bash
JELLYFIN_ENABLED=true
JELLYFIN_URL=http://jellyfin:8096
JELLYFIN_API_KEY=your-api-key-here
```

---

## Performance Considerations

### API Call Optimization

1. **Sessions**: Lightweight endpoint, safe for frequent polling (15s refresh)
2. **Library Stats**: More expensive, should be cached or called on-demand
3. **Recently Added**: Limited by query parameter, reasonable for dashboard
4. **Activity Log**: Paginated, defaults to 50 entries

### Caching Strategy

Currently, no server-side caching is implemented. Consider:
- Browser-side caching for recently added items (5-10 minutes)
- In-memory cache for library stats (30-60 minutes)
- Real-time updates for active sessions (no caching)

### Resource Usage

- **Memory**: Minimal impact, responses are streamed
- **CPU**: Direct JSON serialization, no heavy processing
- **Network**: Typical API response sizes:
  - Sessions: ~500B per active session
  - Library Stats: ~2-5KB
  - Recently Added: ~10-50KB (with poster URLs)
  - Activity Log: ~5-20KB

---

## Future Enhancements

Potential additions based on Jellyfin API capabilities:

1. **User Activity Reports**: Weekly/monthly viewing statistics
2. **Playback Statistics**: Most watched content
3. **Server Health Metrics**: CPU, memory, disk usage
4. **Plugin Information**: Installed plugins and versions
5. **Scheduled Tasks**: View and manage server tasks
6. **Collection Management**: Create/update Jellyfin collections
7. **Notification Integration**: Server events to KeeperCheky notifications

---

## Troubleshooting

### Sessions Not Appearing

**Symptoms**: Dashboard shows no sessions despite active playback

**Checks**:
1. Verify Jellyfin configuration is correct
2. Check that sessions endpoint is accessible: `GET /api/jellyfin/sessions`
3. Review browser console for JavaScript errors
4. Check Jellyfin server logs for API errors

**Debug**:
```bash
curl http://localhost:8000/api/jellyfin/sessions
```

### Recently Added Empty

**Symptoms**: No items shown in recently added widget

**Checks**:
1. Verify content was actually added to Jellyfin recently
2. Check the limit parameter: `/api/jellyfin/recently-added?limit=20`
3. Verify Jellyfin library scan has completed
4. Check Jellyfin permissions for the API key user

### Library Stats Incorrect

**Symptoms**: Item counts don't match Jellyfin UI

**Possible Causes**:
- Library scan in progress
- Different item type filtering
- Jellyfin cache not refreshed

**Solution**: Trigger a library refresh in Jellyfin and wait for completion

---

## API Reference Summary

| Endpoint | Method | Parameters | Response |
|----------|--------|------------|----------|
| `/api/jellyfin/stats` | GET | None | Library statistics |
| `/api/jellyfin/sessions` | GET | None | Active sessions list |
| `/api/jellyfin/recently-added` | GET | `limit` (int) | Recently added items |
| `/api/jellyfin/activity` | GET | `limit` (int) | Activity log entries |

All endpoints return JSON and require Jellyfin to be configured in settings.

---

## Related Documentation

- [Jellyfin API Documentation](https://api.jellyfin.org/)
- [Jellyfin Sessions API](https://api.jellyfin.org/#tag/Session)
- [Jellyfin System API](https://api.jellyfin.org/#tag/System)
- [Jellyfin Items API](https://api.jellyfin.org/#tag/Items)

---

**Last Updated**: January 2025
**Version**: 1.0.0
**Author**: KeeperCheky Development Team
