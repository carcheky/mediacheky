# Example: Generated Docker Compose File

This document shows an example of what the template engine generates.

## Input Configuration

```json
{
  "Image": "linuxserver/radarr:latest",
  "ContainerName": "radarr",
  "Port": 7878,
  "Paths": {
    "Config": "/data/radarr/config",
    "Movies": "/media/movies",
    "Downloads": "/data/downloads"
  },
  "Umask": "022",
  "Network": "media_network",
  "RestartPolicy": "unless-stopped"
}
```

## Global Configuration

```json
{
  "PUID": "1000",
  "PGID": "1000",
  "Timezone": "Europe/Madrid"
}
```

## Generated docker-compose.yml

```yaml
version: '3.8'
services:
  radarr:
    image: linuxserver/radarr:latest
    container_name: radarr
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=Europe/Madrid
      - UMASK=022
    volumes:
      - /data/radarr/config:/config
      - /media/movies:/movies
      - /data/downloads:/downloads
    ports:
      - "7878:7878"
    networks:
      - media_network
    restart: unless-stopped

networks:
  media_network:
    external: true
```

## Minimal Configuration Example

Input without optional fields:

```json
{
  "Image": "linuxserver/radarr:latest",
  "ContainerName": "radarr",
  "Port": 7878,
  "Paths": {
    "Config": "/data/radarr/config",
    "Movies": "/media/movies"
  },
  "RestartPolicy": "unless-stopped"
}
```

Generated output:

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
      - /data/radarr/config:/config
      - /media/movies:/movies
    ports:
      - "7878:7878"
    restart: unless-stopped
```

## File Location

Generated files are saved to:
```
volumes/services/radarr/docker-compose.yml
```

## Backup Behavior

When regenerating an existing compose file, the previous version is backed up:
```
volumes/services/radarr/docker-compose.yml.backup.20231108-153045
```

## Usage in Code

```go
package main

import (
    "github.com/carcheky/mediacheky/internal/models"
    "github.com/carcheky/mediacheky/internal/service"
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewDevelopment()
    
    // Create template engine
    engine := service.NewTemplateEngine(
        logger,
        "templates",
        "volumes/services",
        configRepo,
        templateRepo,
    )
    
    // Create validator
    validator := service.NewConfigValidator(logger, "templates")
    
    // Service configuration
    config := models.ServiceConfig{
        "Image":         "linuxserver/radarr:latest",
        "ContainerName": "radarr",
        "Port":          7878,
        "Paths": map[string]interface{}{
            "Config": "/data/radarr/config",
            "Movies": "/media/movies",
        },
        "RestartPolicy": "unless-stopped",
    }
    
    // Validate configuration
    if err := validator.Validate("radarr", config); err != nil {
        logger.Fatal("Validation failed", zap.Error(err))
    }
    
    // Generate docker-compose file
    composePath, err := engine.GenerateCompose("radarr", config)
    if err != nil {
        logger.Fatal("Generation failed", zap.Error(err))
    }
    
    logger.Info("Generated compose file", zap.String("path", composePath))
}
```
