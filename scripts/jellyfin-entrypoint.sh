#!/bin/bash
set -e

# Check if this is the first run (system.xml doesn't exist)
if [ ! -f /config/system.xml ]; then
  echo "[jellyfin-entrypoint] First run detected. Copying default configuration files..."
  
  # Copy defaults from /defaults (mounted as read-only volume)
  if [ -d /defaults ] && [ -n "$(ls -A /defaults)" ]; then
    cp /defaults/* /config/ 2>/dev/null || true
    echo "[jellyfin-entrypoint] Default configuration files copied successfully."
  else
    echo "[jellyfin-entrypoint] WARNING: /defaults directory not found or empty. Jellyfin will use its built-in defaults."
  fi
else
  echo "[jellyfin-entrypoint] Configuration already exists. Skipping defaults copy."
fi

# Execute the original Jellyfin entrypoint
exec /init
