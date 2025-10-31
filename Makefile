.PHONY: help dev dev-watch build test clean docker-build docker-run shell logs stop check-deps

# Default target
help:
	@echo "MediaCheky - Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make dev          - Start development server (checks deps & sets up env)"
	@echo "  make dev-watch    - Start with Docker Compose Watch (auto-rebuild)"
	@echo "  make logs         - Show development logs"
	@echo "  make shell        - Open shell in development container"
	@echo "  make stop         - Stop development server"
	@echo "  make clean-media  - Clean and recreate mock media library"
	@echo ""
	@echo "Build:"
	@echo "  make build        - Build production binary"
	@echo "  make docker-build - Build production Docker image (default: production target)"
	@echo "  make docker-build-dev - Build development Docker image (with hot-reload)"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run all tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter"
	@echo "  make check-deps   - Check required dependencies"
	@echo ""

# Check dependencies
check-deps:
	@echo "🔍 Checking dependencies..."
	@command -v docker >/dev/null 2>&1 || { \
		echo "❌ Docker is not installed"; \
		echo "📥 Install from: https://docs.docker.com/get-docker/"; \
		echo ""; \
		exit 1; \
	}
	@echo "✅ Docker found: $$(docker --version)"
	@docker compose version >/dev/null 2>&1 || { \
		echo "❌ Docker Compose is not available"; \
		echo "📥 Install Docker Desktop or Docker Compose plugin"; \
		echo "   https://docs.docker.com/compose/install/"; \
		echo ""; \
		exit 1; \
	}
	@echo "✅ Docker Compose found: $$(docker compose version)"
	@command -v go >/dev/null 2>&1 || { \
		echo "⚠️  Go is not installed (optional for local development)"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		echo "   Note: Docker will handle Go compilation, but local Go is useful for IDE support"; \
		echo ""; \
	}
	@if command -v go >/dev/null 2>&1; then \
		echo "✅ Go found: $$(go version)"; \
	fi
	@echo "✅ All required dependencies are installed"
	@echo ""

# Setup environment files and directories
setup-env:
	@echo "⚙️  Setting up environment..."
	@if [ ! -f .env ]; then \
		echo "📄 Creating .env from .env.example..."; \
		cp .env.example .env; \
		echo "✅ .env created - you can customize it if needed"; \
	else \
		echo "✅ .env already exists"; \
	fi
	@echo "📁 Creating required directories..."
	@mkdir -p volumes/mediacheky-go-modules
	@mkdir -p volumes/mediacheky-data
	@mkdir -p volumes/mediacheky-config
	@mkdir -p volumes/media-library/downloads
	@mkdir -p volumes/media-library/library/movies
	@mkdir -p volumes/media-library/library/tv
	@mkdir -p logs
	@mkdir -p data
	@mkdir -p config
	@echo "✅ All directories created"
	@if [ -x scripts/log-with-rotation.sh ]; then \
		echo "✅ Scripts are executable"; \
	else \
		echo "🔧 Making scripts executable..."; \
		chmod +x scripts/*.sh 2>/dev/null || true; \
		echo "✅ Scripts ready"; \
	fi
	@echo ""

# Development with hot-reload (Air + Docker Compose Watch)
dev: check-deps setup-env
	@echo "🚀 Starting MediaCheky development server..."
	@echo ""
	@echo "📊 Status:"
	@echo "  • Environment: Development"
	@echo "  • URL: http://localhost:8000"
	@echo "  • Logs: logs/mediacheky-dev.log"
	@echo "  • Hot-reload: Enabled (via Docker Compose Watch)"
	@echo ""
	@echo "💡 Tips:"
	@echo "  • Press Ctrl+C to stop"
	@echo "  • Run 'make logs' in another terminal to see logs"
	@echo "  • Run 'make shell' to open a shell in the container"
	@echo "  • Edit code and it will auto-reload"
	@echo ""
	@echo "Starting containers..."
	@docker compose up --build --watch

# Development with Docker Compose Watch (Docker 28+)
dev-watch:
	@echo "🚀 Starting development server with Docker Compose Watch..."
	@docker compose watch

# Show development logs
logs:
	@docker compose logs -f mediacheky

# Open shell in development container
shell:
	@docker compose exec mediacheky sh

# Stop development server
stop:
	@docker compose down

# Stop and remove volumes
stop-clean:
	@echo "🧹 Stopping and cleaning volumes..."
	@docker compose down -v
	@echo "✅ Containers and volumes removed"

# Clean mock media library
clean-media:
	@echo "🧹 Cleaning mock media library..."
	@rm -rf volumes/media-library/downloads
	@rm -rf volumes/media-library/library
	@echo "✅ Media library cleaned"
	@echo "   Run './scripts/create-mock-media.sh' or 'make dev' to recreate it"

# Build production binary
build:
	@echo "🔨 Building production binary..."
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		exit 1; \
	}
	@mkdir -p bin
	@CGO_ENABLED=1 go build -ldflags="-w -s" -o bin/mediacheky ./cmd/server
	@echo "✅ Binary built: bin/mediacheky"

# Build production Docker image
docker-build:
	@echo "🐳 Building production Docker image..."
	@command -v docker >/dev/null 2>&1 || { \
		echo "❌ Docker is not installed"; \
		echo "📥 Install from: https://docs.docker.com/get-docker/"; \
		exit 1; \
	}
	@docker build -t mediacheky:latest .
	@echo "✅ Production image built successfully"
	@echo "   - Uses multi-stage build with 'production' target (default)"
	@echo "   - Final image based on scratch (~25MB)"
	@echo "   - To run: make docker-run"

# Build development Docker image (for testing)
docker-build-dev:
	@echo "🐳 Building development Docker image..."
	@command -v docker >/dev/null 2>&1 || { \
		echo "❌ Docker is not installed"; \
		echo "📥 Install from: https://docs.docker.com/get-docker/"; \
		exit 1; \
	}
	@docker build --target=development -t mediacheky:dev .
	@echo "✅ Development image built successfully"
	@echo "   - Uses 'development' target with hot-reload"
	@echo "   - Based on golang:alpine with Air installed"

# Run production Docker image
docker-run:
	@echo "🚀 Running production Docker image..."
	@command -v docker >/dev/null 2>&1 || { \
		echo "❌ Docker is not installed"; \
		echo "📥 Install from: https://docs.docker.com/get-docker/"; \
		exit 1; \
	}
	@mkdir -p data config
	@docker run -p 8000:8000 \
		-v $(PWD)/data:/data \
		-v $(PWD)/config:/config \
		mediacheky:latest

# Run tests
test:
	@echo "🧪 Running tests..."
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		exit 1; \
	}
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		exit 1; \
	}
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Format code
fmt:
	@echo "✨ Formatting code..."
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		exit 1; \
	}
	@go fmt ./...

# Run linter
lint:
	@echo "🔍 Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "❌ golangci-lint is not installed"; \
		echo "📥 Install from: https://golangci-lint.run/usage/install/"; \
		echo ""; \
		echo "Quick install:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin"; \
		echo ""; \
		exit 1; \
	}
	@golangci-lint run ./...

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf bin/ tmp/ coverage.out coverage.html
	@echo "⚠️  Note: Docker volumes in ./volumes/ are NOT deleted"
	@echo "   Run 'make clean-all' to also remove volume data"
	@echo "✅ Clean complete"

# Clean everything including volumes
clean-all:
	@echo "🧹 Cleaning everything (including volumes)..."
	@rm -rf bin/ tmp/ coverage.out coverage.html
	@docker compose down -v
	@rm -rf volumes/
	@echo "✅ Complete cleanup done"

# Initialize development environment
init:
	@echo "🔧 Initializing development environment..."
	@mkdir -p data config
	@mkdir -p volumes/mediacheky-go-modules
	@mkdir -p volumes/radarr-config
	@mkdir -p volumes/sonarr-config
	@mkdir -p volumes/jellyfin-config
	@mkdir -p volumes/jellyseerr-config
	@mkdir -p volumes/qbittorrent-config
	@mkdir -p volumes/bazarr-config
	@mkdir -p volumes/jellystat-config
	@mkdir -p volumes/media-library/library/movies
	@mkdir -p volumes/media-library/library/tv
	@mkdir -p volumes/media-library/downloads
	@echo "✅ Development environment initialized"
	@echo "🎬 Creating mock media library..."
	@./scripts/create-mock-media.sh
	@echo ""
	@echo "📁 Directory structure:"
	@echo "  ├── data/              (app data & database)"
	@echo "  ├── config/            (configuration files)"
	@echo "  └── volumes/           (Docker volume mounts)"
	@echo "      ├── mediacheky-go-modules/"
	@echo "      ├── radarr-config/"
	@echo "      ├── sonarr-config/"
	@echo "      ├── jellyfin-config/"
	@echo "      ├── jellyseerr-config/"
	@echo "      ├── qbittorrent-config/"
	@echo "      ├── bazarr-config/"
	@echo "      ├── jellystat-config/"
	@echo "      └── media-library/"
	@echo "          ├── library/"
	@echo "          │   ├── movies/"
	@echo "          │   └── tv/"
	@echo "          └── downloads/"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make dev' to start the development server"
	@echo "  2. Visit http://localhost:8000"
	@echo "  3. Check the documentation in docs/DEVELOPMENT.md"
