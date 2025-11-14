# Docker Compose Templates

This directory contains templates for generating docker compose.yml files for MediaCheky services.

## Template Structure

Each service has two files:

1. **`{service}.yml`** - Docker Compose template with Go template placeholders
2. **`{service}.schema.json`** - JSON Schema for validating service configuration

## Template Variables

Templates have access to:

### Service Configuration
- `Image` - Docker image name
- `ContainerName` - Container name
- `Port` - External port number
- `Paths` - Map of volume paths (Config, Movies, Downloads, etc.)
- `Umask` - Optional umask value
- `Network` - Optional custom network name
- `RestartPolicy` - Container restart policy (unless-stopped, always, on-failure, no)

### Global Configuration
- `Global.PUID` - User ID (default: 1000)
- `Global.PGID` - Group ID (default: 1000)
- `Global.Timezone` - Timezone (default: UTC)

### Custom Fields
Any additional fields in the service configuration are available in the `Custom` map.

## Example: Radarr Template

```yaml
version: '3.8'
services:
  radarr:
    image: {{ .Image }}
    container_name: {{ .ContainerName }}
    environment:
      - PUID={{ .Global.PUID }}
      - PGID={{ .Global.PGID }}
      - TZ={{ .Global.Timezone }}
    volumes:
      - {{ .Paths.Config }}:/config
      - {{ .Paths.Movies }}:/movies
    ports:
      - "{{ .Port }}:7878"
    restart: {{ .RestartPolicy }}
```

## JSON Schema Validation

Each template has a corresponding JSON schema file that validates:

- Required fields
- Field types and formats
- Value constraints (enums, ranges, patterns)
- Optional fields

Example schema requirements for Radarr:
- `Image`: Must be one of the allowed images (latest, develop, nightly)
- `Port`: Must be between 1024-65535
- `ContainerName`: Must match pattern `^[a-zA-Z0-9_-]+$`
- `Paths.Config` and `Paths.Movies`: Required absolute paths

## Adding New Service Templates

1. Create `{service}.yml` template with Go template syntax
2. Create `{service}.schema.json` with validation rules
3. Update service models to support the new service type
4. Add tests for the new template

## Template Engine

The template engine (`internal/service/template_engine.go`) provides:

- Template loading from filesystem or database
- Variable injection (service config + global config)
- YAML validation
- File generation with automatic backup
- Error handling with descriptive messages

## Generated Files

Docker compose files are generated in:
```
volumes/services/{service}/docker compose.yml
```

When a compose file is regenerated, the previous version is automatically backed up to:
```
volumes/services/{service}/docker compose.yml.backup.{timestamp}
```

## Validation

The validator (`internal/service/validator.go`) provides:

- JSON Schema validation
- Port range validation (1024-65535)
- Path validation (absolute paths, no traversal)
- Image whitelist validation
- Container name validation
- Required fields checking

## Testing

Tests cover:
- Template parsing and execution
- Variable injection
- YAML validation
- File generation and backup
- Schema validation (valid/invalid configs)
- Edge cases and error handling

Run tests:
```bash
go test -v ./internal/service/... -run "Template|Validator"
```
