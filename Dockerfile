# Production Dockerfile - Multi-stage build for minimal final image

# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && \
    go mod verify

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY pkg/ ./pkg/
COPY web/ ./web/

# Build binary with optimizations
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -trimpath \
    -o /app/bin/mediacheky \
    ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    sqlite

# Create app directory
WORKDIR /app

# Copy binary
COPY --from=builder /app/bin/mediacheky /app/mediacheky

# Copy web assets
COPY --from=builder /app/web /app/web

# Create data directory
RUN mkdir -p /data && \
    chown -R nobody:nobody /app /data

# Use non-root user
USER nobody:nobody

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/app/mediacheky", "health"] || exit 1

# Set environment
ENV MEDIACHEKY_APP_ENVIRONMENT=production \
    MEDIACHEKY_SERVER_PORT=8080 \
    MEDIACHEKY_SERVER_HOST=0.0.0.0 \
    MEDIACHEKY_DATABASE_PATH=/data/mediacheky.db

# Run
CMD ["/app/mediacheky"]
