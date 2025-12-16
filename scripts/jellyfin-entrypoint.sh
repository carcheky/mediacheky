#!/bin/bash
set -e

# Check if this is the first run (system.xml doesn't exist)
if [ ! -f /config/system.xml ]; then
  echo "[jellyfin-entrypoint] First run detected. Letting Jellyfin create its own configuration..."
  
  # Skip copying defaults - let Jellyfin initialize with its built-in defaults
  # The database.xml from older versions may cause migration issues with newer Jellyfin
  echo "[jellyfin-entrypoint] Skipping defaults copy to avoid migration conflicts."
else
  echo "[jellyfin-entrypoint] Configuration already exists. Skipping defaults copy."
fi

# Execute the original Jellyfin entrypoint
exec /init
