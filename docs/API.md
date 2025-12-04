# MediaCheky REST API Documentation

This document describes the REST API endpoints available in MediaCheky v1.0+.

## Base URL

```
http://localhost
```

## Response Format

All API endpoints return responses in the following JSON format:

### Success Response

```json
{
  "success": true,
  "data": { ... }
}
```

### Error Response

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

## Authentication

Currently, MediaCheky does not require authentication for API access. This may change in future versions.

## CORS

CORS is enabled on all API endpoints with the following configuration:

- **Access-Control-Allow-Origin**: `*`

- **Access-Control-Allow-Methods**: `GET, POST, PUT, DELETE, OPTIONS`

- **Access-Control-Allow-Headers**: `Content-Type, Authorization, X-Request-ID`

---

## Endpoints

### Service Management

#### List All Services

```http
GET /api/services
```

Returns a list of all configured services.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "name": "radarr",
      "display_name": "Radarr",
      "enabled": true,
      "status": "running",
      "image": "linuxserver/radarr",
      "container_id": "abc123...",
      "port": 7878,
      "config": {
        "api_key": "...",
        "url": "http://localhost:7878"
      }
    }
  ]
}
```

#### Get Specific Service

```http
GET /api/services/:name
```

Returns details for a specific service.

**Parameters:**

- `name` (path) - Service name (e.g., `radarr`, `sonarr`)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "radarr",
    "display_name": "Radarr",
    "enabled": true,
    "status": "running",
    "image": "linuxserver/radarr",
    "container_id": "abc123...",
    "port": 7878,
    "config": { ... }
  }
}
```

**Error Codes:**

- `404` - Service not found

- `400` - Invalid service name

#### Enable Service

```http
POST /api/services/:name/enable
```

Enables a service.

**Parameters:**

- `name` (path) - Service name

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Service enabled successfully"
  }
}
```

#### Disable Service

```http
POST /api/services/:name/disable
```

Disables a service.

**Parameters:**

- `name` (path) - Service name

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Service disabled successfully"
  }
}
```

#### Start Container

```http
POST /api/services/:name/start
```

Starts a service's Docker container.

**Parameters:**

- `name` (path) - Service name

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Container started successfully"
  }
}
```

**Error Codes:**

- `400` - Service has no container associated

- `500` - Failed to start container

#### Stop Container

```http
POST /api/services/:name/stop
```

Stops a service's Docker container.

**Parameters:**

- `name` (path) - Service name

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Container stopped successfully"
  }
}
```

**Error Codes:**

- `400` - Service has no container associated

- `500` - Failed to stop container

#### Restart Container

```http
POST /api/services/:name/restart
```

Restarts a service's Docker container.

**Parameters:**

- `name` (path) - Service name

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Container restarted successfully"
  }
}
```

#### Update Service Configuration

```http
PUT /api/services/:name/config
```

Updates a service's configuration.

**Parameters:**

- `name` (path) - Service name

**Request Body:**
```json
{
  "api_key": "new_key_here",
  "url": "http://localhost:7878",
  "custom_setting": "value"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Service configuration updated successfully"
  }
}
```

#### Get Container Logs

```http
GET /api/services/:name/logs?tail=100
```

Retrieves logs from a service's Docker container.

**Parameters:**

- `name` (path) - Service name

- `tail` (query, optional) - Number of lines to retrieve (default: 100)

**Response:**
```json
{
  "success": true,
  "data": {
    "logs": "2024-01-01T00:00:00Z Log line 1\n2024-01-01T00:00:01Z Log line 2\n..."
  }
}
```

---

### Global Configuration

#### Get All Global Configuration

```http
GET /api/config/global
```

Returns all global configuration entries, grouped by category.

**Response:**
```json
{
  "success": true,
  "data": {
    "system": {
      "PUID": "1000",
      "PGID": "1000",
      "TZ": "Europe/Madrid"
    },
    "paths": {
      "CONFIG_PATH": "/config",
      "DATA_PATH": "/data"
    }
  }
}
```

#### Update Global Configuration

```http
PUT /api/config/global
```

Updates multiple global configuration entries at once.

**Request Body:**
```json
[
  {
    "key": "PUID",
    "value": "1001",
    "category": "system"
  },
  {
    "key": "PGID",
    "value": "1001",
    "category": "system"
  }
]
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Global configuration updated successfully"
  }
}
```

#### Get Specific Configuration Value

```http
GET /api/config/global/:key
```

Returns a specific configuration entry.

**Parameters:**

- `key` (path) - Configuration key (e.g., `PUID`, `TZ`)

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "key": "PUID",
    "value": "1000",
    "category": "system"
  }
}
```

**Error Codes:**

- `404` - Configuration key not found

- `400` - Invalid configuration key format

#### Update Specific Configuration Value

```http
PUT /api/config/global/:key
```

Updates a specific configuration entry.

**Parameters:**

- `key` (path) - Configuration key

**Request Body:**
```json
{
  "value": "1001",
  "category": "system"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Configuration value updated successfully"
  }
}
```

---

### Docker Information

#### Get Docker Host Information

```http
GET /api/docker/info
```

Returns information about the Docker host.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "...",
    "name": "hostname",
    "server_version": "24.0.7",
    "operating_system": "Ubuntu 22.04.3 LTS",
    "os_type": "linux",
    "architecture": "x86_64",
    "ncpu": 8,
    "memory_total": 16777216000,
    "docker_root_dir": "/var/lib/docker",
    "containers": 15,
    "containers_running": 10,
    "containers_paused": 0,
    "containers_stopped": 5,
    "images": 25
  }
}
```

#### List Managed Containers

```http
GET /api/docker/containers
```

Returns a list of containers managed by MediaCheky.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": "abc123...",
      "name": "radarr",
      "image": "linuxserver/radarr:latest",
      "status": "running",
      "state": "running",
      "created_at": "2024-01-01T00:00:00Z",
      "started_at": "2024-01-01T00:05:00Z",
      "labels": {
        "com.mediacheky.managed": "true",
        "com.mediacheky.service": "radarr"
      },
      "ports": [
        {
          "container_port": 7878,
          "host_port": 7878,
          "protocol": "tcp"
        }
      ]
    }
  ]
}
```

---

### Dashboard

#### Get Dashboard Statistics

```http
GET /api/dashboard/stats
```

Returns general statistics for the dashboard.

**Response:**
```json
{
  "success": true,
  "data": {
    "services_active": 5,
    "services_stopped": 2,
    "services_total": 7,
    "containers_running": 8,
    "containers_total": 10
  }
}
```

#### Health Check All Services

```http
GET /api/dashboard/health
```

Returns health status for all enabled services.

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "radarr",
      "display_name": "Radarr",
      "enabled": true,
      "status": "running",
      "healthy": true,
      "container_status": "running"
    },
    {
      "name": "sonarr",
      "display_name": "Sonarr",
      "enabled": true,
      "status": "stopped",
      "healthy": false,
      "container_status": "stopped"
    }
  ]
}
```

---

## Input Validation

### Service Names

Service names must:

- Be alphanumeric with hyphens and underscores allowed

- Not exceed 50 characters

- Match pattern: `^[a-zA-Z0-9_-]+$`

**Valid examples:** `radarr`, `my-service`, `service_1`
**Invalid examples:** `service.name`, `service@test`, `my service`

### Configuration Keys

Configuration keys must:

- Be uppercase alphanumeric with underscores allowed

- Not exceed 50 characters

- Match pattern: `^[A-Z0-9_]+$`

**Valid examples:** `PUID`, `MY_CONFIG`, `CONFIG_123`
**Invalid examples:** `puid`, `My-Config`, `config.key`

---

## Rate Limiting

Currently, there are no rate limits on API endpoints. This may be implemented in future versions.

---

## Error Handling

All errors follow the standard error response format:

```json
{
  "success": false,
  "error": "Descriptive error message"
}
```

Common HTTP status codes:

- `200` - Success

- `400` - Bad Request (invalid input)

- `404` - Not Found (resource doesn't exist)

- `500` - Internal Server Error

- `503` - Service Unavailable (Docker client not available)

---

## Request ID

All requests automatically receive a unique request ID in the `X-Request-ID` header. This ID is:

- Generated automatically if not provided

- Can be provided in the request header for request tracking

- Included in error responses for debugging

- Logged with all operations

Example:
```http
X-Request-ID: 550e8400-e29b-41d4-a716-446655440000
```

---

## Changelog

### Version 1.0 (Current)

- Initial REST API implementation

- Service management endpoints

- Global configuration endpoints

- Docker information endpoints

- Dashboard statistics and health endpoints

- Input validation middleware

- CORS support

- Standardized response format
