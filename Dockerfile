# Multi-stage Dockerfile for both development and production
# By default builds production image. Use --target=development for dev.
# Target: <25MB production image, <50MB RAM usage

# ============================================================================
# Base stage - Common dependencies
# ============================================================================
FROM golang:1.25-alpine AS base

WORKDIR /app

# Install common dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    gcc \
    musl-dev \
    sqlite-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go mod verify

# ============================================================================
# Development stage - With hot-reload support
# ============================================================================
FROM base AS development

# Install Air for hot-reload
RUN go install github.com/air-verse/air@latest

# Copy source code
COPY . .

# Expose port
EXPOSE 8000

# Set development environment
ENV KEEPERCHEKY_APP_ENVIRONMENT=development \
    KEEPERCHEKY_SERVER_PORT=8000 \
    KEEPERCHEKY_SERVER_HOST=0.0.0.0

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
    -o /app/bin/keepercheky \
    ./cmd/server

# Verify binary exists and is executable
RUN ls -lh /app/bin/keepercheky

# ============================================================================
# Production stage - Minimal final image (default)
# ============================================================================
FROM alpine:3.19 AS production

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    wget

# Create necessary directories with proper permissions
RUN mkdir -p /app/data /app/config /app/logs && \
    chown -R 65534:65534 /app

# Set working directory
WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/keepercheky /app/keepercheky

# Copy web assets
COPY --from=builder /app/web /app/web

# Create non-root user and switch to it
USER 65534:65534

# Expose port
EXPOSE 8000

# Set environment
ENV KEEPERCHEKY_APP_ENVIRONMENT=production \
    KEEPERCHEKY_SERVER_PORT=8000 \
    KEEPERCHEKY_SERVER_HOST=0.0.0.0

# Run
ENTRYPOINT ["/app/keepercheky"]
