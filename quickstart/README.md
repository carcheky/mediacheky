# MediaCheky Quickstart 🚀

Quick and easy way to test the compiled MediaCheky Docker image.

## 📋 Prerequisites

- Docker and Docker Compose installed
- Access to your media library path
- API keys for your services (Radarr, Sonarr, etc.)

## 🚀 Quick Start

### 1. Configure Environment

```bash
# Copy and edit the .env file
cp .env .env.local  # Optional: keep original as reference
nano .env  # or your preferred editor
```

**Important settings to update:**

- `MEDIA_PATH`: Path to your actual media library
- `RADARR_API_KEY`, `SONARR_API_KEY`, etc.: Your actual API keys
- `APP_DRY_RUN`: Keep as `true` until you're confident!

### 2. Start MediaCheky

```bash
docker-compose up -d
```

### 3. Access the UI

Open your browser and navigate to:

```
http://localhost:8780
```

### 4. Check Logs

```bash
# Follow logs
docker-compose logs -f

# View specific number of lines
docker-compose logs --tail=100
```

### 5. Stop MediaCheky

```bash
docker-compose down
```

## 🔧 Configuration

### Port Comparison

- **Quickstart (production image)**: `http://localhost:8780`
- **Dev environment**: `http://localhost:8000`

This allows you to run both simultaneously for comparison!

### Enable Services

Edit `.env` and set to `true`:

```env
RADARR_ENABLED=true
RADARR_URL=http://your-radarr-host:7878
RADARR_API_KEY=your_actual_api_key
```

Repeat for other services (Sonarr, Jellyfin, Jellyseerr, qBittorrent).

### Database Options

**SQLite (Default - Recommended for testing):**
```env
DB_TYPE=sqlite
DB_PATH=/data/mediacheky.db
```

**PostgreSQL (For production):**
```env
DB_TYPE=postgres
DB_HOST=postgres
DB_PORT=5432
DB_USER=mediacheky
DB_PASSWORD=your_secure_password
DB_NAME=mediacheky
```

## 📁 Directory Structure

After starting, you'll have:

```
quickstart/
├── docker-compose.yml
├── .env
├── config/          # Configuration files
├── data/            # SQLite database (if using SQLite)
└── logs/            # Application logs
```

## ⚠️ Safety Features

### Dry Run Mode

**Always test with dry run first!**

```env
APP_DRY_RUN=true  # No actual deletions
```

When you're confident:

```env
APP_DRY_RUN=false  # Enable actual deletions
```

### Exclusion Tags

Protect specific media with tags:

```env
APP_EXCLUSION_TAGS=keep,favorite,archive
```

Tag your media in Radarr/Sonarr, and MediaCheky will skip them.

## 🔍 Testing Workflow

### 1. Compare Dev vs Production

```bash
# Terminal 1: Run dev environment
cd /home/user/projects/mediacheky
make dev  # Runs on port 8000

# Terminal 2: Run quickstart
cd /home/user/projects/mediacheky/quickstart
docker-compose up  # Runs on port 8780
```

Now you can compare:
- Dev: http://localhost:8000
- Production: http://localhost:8780

### 2. Test Configuration Changes

```bash
# Edit .env
nano .env

# Restart to apply changes
docker-compose restart

# Check logs
docker-compose logs -f
```

### 3. Verify Health

```bash
# Check health endpoint
curl http://localhost:8780/health

# Check container health
docker-compose ps
```

## 🐛 Troubleshooting

### Container Won't Start

```bash
# Check logs
docker-compose logs

# Check if port is already in use
sudo lsof -i :8780

# Change port in .env if needed
KEEPERCHEKY_PORT=8081
```

### Can't Connect to Services

```bash
# Test network connectivity from container
docker-compose exec mediacheky wget -O- http://radarr:7878/api/v3/system/status

# Check if services are accessible from host
curl http://your-radarr-host:7878/api/v3/system/status
```

### Permission Issues

```bash
# Check volume permissions
ls -la ./data ./config ./logs

# Fix permissions if needed
sudo chown -R $USER:$USER ./data ./config ./logs
```

### Database Issues

```bash
# Reset database (WARNING: Deletes all data!)
docker-compose down
rm -rf ./data/mediacheky.db
docker-compose up -d
```

## 🔄 Updating

```bash
# Pull latest image
docker-compose pull

# Recreate container
docker-compose up -d --force-recreate
```

## 📊 Monitoring

### View Resource Usage

```bash
docker stats mediacheky-quickstart
```

### Check Disk Usage

```bash
du -sh ./data ./logs ./config
```

## 🧹 Cleanup

### Remove Everything

```bash
# Stop and remove containers
docker-compose down

# Remove volumes (WARNING: Deletes all data!)
docker-compose down -v

# Remove all files
cd ..
rm -rf quickstart/
```

### Keep Configuration

```bash
# Stop containers but keep volumes
docker-compose down

# Data persists in ./data, ./config, ./logs
```

## 📚 Next Steps

1. ✅ Test with `APP_DRY_RUN=true`
2. ✅ Verify connections to all services
3. ✅ Review logs for any warnings
4. ✅ Configure exclusion tags
5. ✅ Test leaving soon collections
6. ✅ When confident, set `APP_DRY_RUN=false`
7. ✅ Monitor first real cleanup closely

## 🆘 Need Help?

- 📖 Full documentation: [/docs](../docs/)
- 🐛 Report issues: [GitHub Issues](https://github.com/carcheky/mediacheky/issues)
- 💬 Discussions: [GitHub Discussions](https://github.com/carcheky/mediacheky/discussions)

## ⚡ Pro Tips

1. **Use symlinks for media**: Mount media as read-only (`:ro`) for safety
2. **Backup your database**: Copy `./data/mediacheky.db` before enabling deletions
3. **Start conservative**: Use high thresholds initially, then adjust
4. **Monitor logs**: Check `./logs/mediacheky.log` regularly
5. **Test incrementally**: Enable one service at a time

---

**Happy cleaning! 🧹**
