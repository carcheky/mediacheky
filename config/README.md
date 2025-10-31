# Configuration Directory

This directory contains configuration examples for KeeperCheky.

## Files

- **`config.example.yaml`**: Example configuration file with all available options
- **`config.yaml`**: Your actual configuration (not tracked in git)

## Setup

### Docker (Recommended)

When running with Docker, your config is stored in:
```
./volumes/keepercheky-config/config.yaml
```

The container automatically:
1. Loads environment variables from `.env` (in project root)
2. Reads configuration from `/app/config/config.yaml` (mounted volume)
3. Environment variables take precedence over config file values

### Local Development

1. Copy the example config:
   ```bash
   cp config/config.example.yaml config/config.yaml
   ```

2. Edit `config.yaml` with your service URLs and API keys

3. Run the application:
   ```bash
   go run ./cmd/server
   ```

## Configuration Priority

KeeperCheky uses [Viper](https://github.com/spf13/viper) for configuration management.

**Order of precedence** (highest to lowest):
1. Environment variables (e.g., `KEEPERCHEKY_CLIENTS_RADARR_API_KEY`)
2. Configuration file (`config.yaml`)
3. Default values (defined in code)

## Environment Variable Format

Environment variables follow this pattern:
```
KEEPERCHEKY_<SECTION>_<SUBSECTION>_<KEY>
```

Examples:
```bash
KEEPERCHEKY_APP_ENVIRONMENT=production
KEEPERCHEKY_APP_LOG_LEVEL=info
KEEPERCHEKY_CLIENTS_RADARR_ENABLED=true
KEEPERCHEKY_CLIENTS_RADARR_URL=http://radarr:7878
KEEPERCHEKY_CLIENTS_RADARR_API_KEY=your_api_key_here
```

## Getting API Keys

### Radarr
1. Open Radarr web UI
2. Go to **Settings → General**
3. Copy the **API Key**

### Sonarr
1. Open Sonarr web UI
2. Go to **Settings → General**
3. Copy the **API Key**

### Jellyfin
1. Open Jellyfin Dashboard
2. Go to **API Keys**
3. Click **+** to create new key
4. Name it "KeeperCheky"
5. Copy the generated key

### Jellyseerr
1. Open Jellyseerr
2. Go to **Settings → General**
3. Copy the **API Key**

### qBittorrent
Default credentials are `admin` / `adminadmin`.

To change:
1. Open qBittorrent Web UI
2. Go to **Tools → Options → Web UI**
3. Change username and password
4. Update your `.env` or `config.yaml`

## Security Notes

- **Never commit** `config.yaml` or `.env` files containing real API keys
- Use `.env.example` and `config.example.yaml` as templates
- Store sensitive credentials in environment variables when deploying to production
- Consider using Docker secrets or Kubernetes secrets for production deployments
