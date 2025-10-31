# File Actions API - Testing Guide

This document provides examples for testing the File Actions API endpoints.

## Base URL

All endpoints are under: `http://localhost:8000/api/files`

## Endpoints

### 1. Ignore File

Mark a file as ignored for future analysis.

**Endpoint:** `POST /api/files/:id/ignore`

**Request:**
```bash
curl -X POST http://localhost:8000/api/files/1/ignore \
  -H "Content-Type: application/json" \
  -d '{
    "reason": "keep_seeding",
    "permanent": true
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Archivo marcado como ignorado"
}
```

### 2. Delete File

Delete a file from the system and optionally from services.

**Endpoint:** `POST /api/files/:id/delete`

**Request:**
```bash
curl -X POST http://localhost:8000/api/files/1/delete \
  -H "Content-Type: application/json" \
  -d '{
    "delete_from_services": true,
    "delete_torrent": true,
    "confirm": true
  }'
```

**Response:**
```json
{
  "success": true,
  "deleted_from": ["radarr", "qbittorrent", "filesystem"],
  "message": "Archivo eliminado exitosamente de todos los servicios"
}
```

**Notes:**
- `confirm` must be `true` for the operation to proceed
- `delete_from_services` will remove from Radarr/Sonarr/Jellyfin
- `delete_torrent` will remove from qBittorrent
- File will always be deleted from filesystem if operation succeeds

### 3. Cleanup Hardlink

Remove a specific hardlink while keeping the original file.

**Endpoint:** `POST /api/files/:id/cleanup-hardlink`

**Request:**
```bash
curl -X POST http://localhost:8000/api/files/1/cleanup-hardlink \
  -H "Content-Type: application/json" \
  -d '{
    "keep_path": "/jellyfin/movies/Inception.2010.mkv",
    "remove_path": "/downloads/Inception.2010.mkv"
  }'
```

**Response:**
```json
{
  "success": true,
  "space_freed": 0,
  "message": "Hardlink eliminado sin afectar el archivo original"
}
```

**Notes:**
- Both paths must exist
- Paths must be hardlinks of the same inode
- `space_freed` is always 0 for hardlinks (they share the same data)

### 4. Bulk Action

Execute an action on multiple files at once.

**Endpoint:** `POST /api/files/bulk-action`

**Request (Bulk Delete):**
```bash
curl -X POST http://localhost:8000/api/files/bulk-action \
  -H "Content-Type: application/json" \
  -d '{
    "file_ids": [1, 2, 3, 4, 5],
    "action": "delete",
    "params": {
      "delete_from_services": true,
      "delete_torrent": true
    }
  }'
```

**Request (Bulk Ignore):**
```bash
curl -X POST http://localhost:8000/api/files/bulk-action \
  -H "Content-Type: application/json" \
  -d '{
    "file_ids": [1, 2, 3],
    "action": "ignore",
    "params": {}
  }'
```

**Response:**
```json
{
  "success": true,
  "results": [
    {"id": 1, "success": true},
    {"id": 2, "success": false, "error": "File in use"},
    {"id": 3, "success": true},
    {"id": 4, "success": true},
    {"id": 5, "success": true}
  ],
  "summary": {
    "total": 5,
    "succeeded": 4,
    "failed": 1
  }
}
```

**Allowed Actions:**
- `delete` - Delete multiple files
- `ignore` - Ignore multiple files

### 5. Import to Radarr (Stub)

**Endpoint:** `POST /api/files/:id/import-to-radarr`

**Status:** Not yet implemented (returns HTTP 501)

**Request:**
```bash
curl -X POST http://localhost:8000/api/files/1/import-to-radarr \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/downloads/Inception.2010.mkv",
    "quality_profile_id": 1,
    "root_folder_path": "/movies"
  }'
```

**Response:**
```json
{
  "success": false,
  "error": "Import to Radarr not yet implemented - requires metadata parsing",
  "message": "Esta funcionalidad está en desarrollo"
}
```

### 6. Import to Sonarr (Stub)

**Endpoint:** `POST /api/files/:id/import-to-sonarr`

**Status:** Not yet implemented (returns HTTP 501)

**Request:**
```bash
curl -X POST http://localhost:8000/api/files/1/import-to-sonarr \
  -H "Content-Type: application/json" \
  -d '{
    "file_path": "/downloads/Breaking.Bad.S01E01.mkv",
    "quality_profile_id": 1,
    "root_folder_path": "/series"
  }'
```

**Response:**
```json
{
  "success": false,
  "error": "Import to Sonarr not yet implemented - requires metadata parsing",
  "message": "Esta funcionalidad está en desarrollo"
}
```

## Error Responses

All endpoints return consistent error responses:

**400 Bad Request:**
```json
{
  "success": false,
  "error": "Invalid request body"
}
```

**404 Not Found:**
```json
{
  "success": false,
  "error": "File not found"
}
```

**500 Internal Server Error:**
```json
{
  "success": false,
  "error": "Failed to delete file: permission denied"
}
```

**501 Not Implemented:**
```json
{
  "success": false,
  "error": "Feature not yet implemented",
  "message": "Esta funcionalidad está en desarrollo"
}
```

## Testing Checklist

- [ ] Test ignore endpoint with valid file ID
- [ ] Test delete endpoint with all flags enabled
- [ ] Test delete endpoint without confirmation (should fail)
- [ ] Test cleanup-hardlink with valid hardlinks
- [ ] Test cleanup-hardlink with non-hardlinks (should fail)
- [ ] Test bulk-action with delete action
- [ ] Test bulk-action with ignore action
- [ ] Test all endpoints with invalid file IDs (should return 404)
- [ ] Test all endpoints with malformed JSON (should return 400)
- [ ] Verify history table receives entries for all operations
- [ ] Verify files are actually deleted from filesystem
- [ ] Verify hardlinks are properly removed without affecting original

## Notes

- All destructive operations require explicit confirmation
- History table logs all actions for audit purposes
- Service deletions continue even if one service fails
- Bulk operations process all files even if some fail
- Import endpoints are stubbed for future implementation
