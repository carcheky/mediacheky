# Multi-stage Dockerfile for both development and production
# By default builds production image. Use --target=development for dev.
# Target: <25MB production image, <50MB RAM usage

# ============================================================================
# Base stage - Common dependencies
# ============================================================================
FROM golang:1.25-alpine AS base

WORKDIR /app

# Install common dependencies (this layer is cached)
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev \
    curl

# Install Docker CLI and Compose plugin (shared by dev and production)
RUN DOCKER_VERSION="29.0.0" && \
    COMPOSE_VERSION="v2.40.3" && \
    # Install Docker CLI
    curl -fsSL "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" | tar -xz -C /tmp && \
    mv /tmp/docker/docker /usr/local/bin/ && \
    rm -rf /tmp/docker && \
    # Install Docker Compose plugin
    mkdir -p /usr/local/lib/docker/cli-plugins && \
    curl -fsSL "https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-linux-x86_64" -o /usr/local/lib/docker/cli-plugins/docker-compose && \
    chmod +x /usr/local/lib/docker/cli-plugins/docker-compose

# Copy go mod files first (separate layer for better caching)
COPY go.mod go.sum ./

# Download dependencies (this layer is cached unless go.mod/go.sum change)
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# ============================================================================
# Development stage - With hot-reload support
# ============================================================================
FROM base AS development

# Install Air for hot-reload (cached layer)
RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/air-verse/air@latest

# Install su-exec for runtime user switching (lightweight alternative to gosu)
RUN apk add --no-cache su-exec

# Docker CLI already installed in base stage

# Don't copy source code here - it's mounted as volumes in docker compose.yml
# This makes the image build MUCH faster since it doesn't rebuild on code changes

# Create required directories
RUN mkdir -p /app/data /app/config /app/logs /app/tmp

# Copy entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Expose port
EXPOSE 7369

# Set development environment (hardcoded for dev build)
ENV MEDIACHEKY_APP_ENVIRONMENT=development \
    MEDIACHEKY_APP_LOG_LEVEL=debug \
    MEDIACHEKY_APP_DRY_RUN=false \
    MEDIACHEKY_SERVER_PORT=7369 \
    MEDIACHEKY_SERVER_HOST=0.0.0.0 \
    MEDIACHEKY_DATABASE_TYPE=sqlite \
    MEDIACHEKY_DATABASE_PATH=/app/data/mediacheky.db

# Use entrypoint to handle PUID/PGID at runtime
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Run with Air for hot-reload
CMD ["air", "-c", ".air.toml"]

# ============================================================================
# Builder stage - Compile production binary
# ============================================================================
FROM base AS builder

# Build arguments for versioning
ARG VERSION=dev
ARG COMMIT_SHA=unknown

# Platform-specific build arguments (automatically set by buildx)
ARG TARGETOS
ARG TARGETARCH

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY web/ ./web/
COPY templates/ ./templates/
COPY services/ ./services/

# Build binary with optimizations
# - CGO_ENABLED=1: Required for SQLite (but using musl for static linking)
# - TARGETOS/TARGETARCH: Set by buildx for multi-platform builds
# - -ldflags="-w -s": Strip debug information (-w) and symbol table (-s)
# - -trimpath: Remove file system paths from binary
# - -X: Inject version information at build time
# - -linkmode external: Use external linker for CGO
# - -extldflags '-static': Force static linking
RUN CGO_ENABLED=1 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-w -s -linkmode external -extldflags '-static' -X main.Version=${VERSION} -X main.CommitSHA=${COMMIT_SHA}" \
    -trimpath \
    -o /app/bin/mediacheky \
    ./cmd/server

# Verify binary exists and is executable
RUN ls -lh /app/bin/mediacheky

# ============================================================================
# Production stage - Minimal final image (default)
# ============================================================================
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget \
    su-exec

# Copy Docker CLI and Compose plugin from base stage (already built there)
COPY --from=base /usr/local/bin/docker /usr/local/bin/docker
COPY --from=base /usr/local/lib/docker/cli-plugins/docker-compose /usr/local/lib/docker/cli-plugins/docker-compose

# Copy entrypoint script
COPY entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

# Set working directory
WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/mediacheky /app/mediacheky

# Copy web assets
COPY --from=builder /app/web /app/web

# Copy templates and service definitions
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/services /app/services

# Create required directories
RUN mkdir -p /app/data /app/config /app/logs /app/tmp

# Expose port
EXPOSE 7369

# Set production environment (hardcoded for production build)
ENV MEDIACHEKY_APP_ENVIRONMENT=production \
    MEDIACHEKY_APP_LOG_LEVEL=info \
    MEDIACHEKY_APP_DRY_RUN=false \
    MEDIACHEKY_SERVER_PORT=7369 \
    MEDIACHEKY_SERVER_HOST=0.0.0.0 \
    MEDIACHEKY_DATABASE_TYPE=sqlite \
    MEDIACHEKY_DATABASE_PATH=/app/data/mediacheky.db

# Use entrypoint to handle PUID/PGID at runtime
ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]

# Run
CMD ["/app/mediacheky"]
