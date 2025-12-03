#!/bin/sh
set -e

# Default values
PUID=${PUID:-1000}
PGID=${PGID:-1000}

echo "Starting MediaCheky with PUID=${PUID} and PGID=${PGID}"

# Create group if it doesn't exist
if ! getent group appgroup >/dev/null; then
    addgroup -g ${PGID} appgroup
fi

# Create user if it doesn't exist
if ! id appuser >/dev/null 2>&1; then
    adduser -D -u ${PUID} -G appgroup appuser
fi

# Add user to docker group for socket access
DOCKER_GID=$(stat -c '%g' /var/run/docker.sock 2>/dev/null || echo 999)
if ! getent group docker >/dev/null; then
    addgroup -g ${DOCKER_GID} docker
fi
addgroup appuser docker 2>/dev/null || true

# Ensure directories exist and have proper ownership
mkdir -p /app/data /app/config /app/logs /app/tmp
chown -R ${PUID}:${PGID} /app/data /app/config /app/logs /app/tmp

# Execute command as appuser
exec su-exec appuser "$@"
