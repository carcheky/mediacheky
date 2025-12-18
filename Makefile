.PHONY: help dev dev-watch build test clean docker-build docker-run shell logs stop check-deps validate validate-quick lint-check lint-fix vet check-and-fix mod-tidy install-hooks uninstall-hooks clean-branches clean-branches-force finaltest finaltest-stop finaltest-clean

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
	@echo "  make test          - Run all tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make finaltest     - 🧪 Build production & test with quickstart"
	@echo "  make finaltest-stop - Stop final test environment"
	@echo "  make finaltest-clean - Clean final test (volumes + image)"
	@echo ""
	@echo "Validation (run before commit):"
	@echo "  make validate      - 🔍 Full validation (format, vet, test, lint)"
	@echo "  make validate-quick - ⚡ Quick validation (format, vet, test)"
	@echo "  make check-and-fix - 🔧 Auto-fix + validate"
	@echo "  make lint-check    - Check code format"
	@echo "  make lint-fix      - Fix code format"
	@echo ""
	@echo "Utilities:"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make fmt          - Format code"
	@echo "  make lint         - Run linter (golangci-lint)"
	@echo "  make check-deps   - Check required dependencies"
	@echo "  make install-hooks - Install pre-commit git hook"
	@echo "  make clean-branches - Remove local branches that don't exist in remote"
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
# 	@echo "🧹 Cleaning up old volumes and directories..."
# 	@docker container prune -f >/dev/null 2>&1 || true
# 	@docker kill $$(docker ps -a -q) &> /dev/null >/dev/null 2>&1 || true
# 	@docker volume prune -f >/dev/null 2>&1 || true
# 	@docker network prune -f >/dev/null 2>&1 || true
	@echo "📁 Creating required directories..."
	@rm -fr volumes/mediacheky-data volumes/library logs/* || true
	@mkdir -p volumes/go-modules-cache
	@mkdir -p volumes/mediacheky-data
	@mkdir -p volumes/library/downloads
	@mkdir -p volumes/library/library/movies
	@mkdir -p volumes/library/library/tv
	@mkdir -p logs
# 	@mkdir -p data
# 	@mkdir -p config
# 	@mkdir -p tmp
	@chmod 777 volumes/mediacheky-data 
	@chmod 777 logs -R
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
	@echo "  • URL: http://localhost"
	@echo "  • Hot-reload: Enabled (Air watches file changes)"
	@echo ""
	@echo "💡 Tips:"
	@echo "  • Press Ctrl+C to stop"
	@echo "  • Run 'make logs' to see live logs"
	@echo "  • Run 'make shell' to open a shell in the container"
	@echo "  • Edit code and Air will auto-reload (no rebuild needed!)"
	@echo "  • First start builds image (slow), subsequent starts are instant"
	@echo ""
	@echo "Starting containers..."
	@DOCKER_BUILDKIT=1 docker compose up --remove-orphans 

# Rebuild development image (only needed after Dockerfile changes)
dev-rebuild:
	@echo "🔨 Rebuilding development image..."
	@DOCKER_BUILDKIT=1 docker compose build --no-cache development
	@echo "✅ Image rebuilt. Run 'make dev' to start"

# Development with Docker Compose Watch (Docker 28+)
dev-watch:
	@echo "🚀 Starting development server with Docker Compose Watch..."
	@DOCKER_BUILDKIT=1 docker compose watch

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
	@rm -rf volumes/library/downloads
	@rm -rf volumes/library/library
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
	@docker compose exec -T mediacheky go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@docker compose exec -T mediacheky go test -v -coverprofile=coverage.out ./...
	@docker compose exec -T mediacheky go tool cover -html=coverage.out -o coverage.html
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

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🔍 VALIDATION TARGETS - Run before committing
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

# Validate all code (format, vet, test, lint) - RUN BEFORE COMMIT
validate:
	@command -v go >/dev/null 2>&1 || { \
		echo "❌ Go is not installed"; \
		echo "📥 Install from: https://golang.org/dl/"; \
		exit 1; \
	}
	@chmod +x scripts/validate.sh
	@bash scripts/validate.sh

# Quick validation (format + vet + test) - Fast pre-commit check
validate-quick: lint-check vet test
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ Quick validation passed!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check code format (without modifying files)
lint-check:
	@echo "📝 Checking Go code format..."
	@docker compose exec -T mediacheky sh -c 'OUTPUT=$$(gofmt -s -l . 2>&1 | grep -v "^vendor/" | grep -v "^volumes/" | grep ".go$$" || true); \
	if [ -n "$$OUTPUT" ]; then \
		echo "❌ The following files need formatting:"; \
		echo "$$OUTPUT"; \
		echo ""; \
		echo "💡 Run \"make lint-fix\" to fix automatically"; \
		exit 1; \
	fi'
	@echo "✅ All files are properly formatted"

# Fix code format automatically
lint-fix:
	@echo "🔧 Fixing code format..."
	@docker compose exec -T mediacheky sh -c 'find . -name "*.go" -not -path "./volumes/*" -not -path "./vendor/*" -exec gofmt -s -w {} \;'
	@echo "✅ Format applied successfully"

# Run go vet
vet:
	@echo "🔍 Running go vet..."
	@docker compose exec -T mediacheky go vet ./...
	@echo "✅ Go vet passed"

# Check and fix common issues, then validate
check-and-fix: lint-fix mod-tidy validate-quick
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ All fixes applied and validated!"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "👍 Ready to commit"

# Tidy go modules
mod-tidy:
	@echo "📦 Tidying go modules..."
	@docker compose exec -T mediacheky go mod tidy
	@echo "✅ Dependencies cleaned"

# Install git hooks (optional)
install-hooks:
	@echo "📎 Installing git pre-commit hook..."
	@mkdir -p .git/hooks
	@echo '#!/bin/sh' > .git/hooks/pre-commit
	@echo 'echo "🔍 Running pre-commit validation..."' >> .git/hooks/pre-commit
	@echo 'make validate-quick' >> .git/hooks/pre-commit
	@echo 'if [ $$? -ne 0 ]; then' >> .git/hooks/pre-commit
	@echo '  echo ""' >> .git/hooks/pre-commit
	@echo '  echo "❌ Pre-commit validation failed!"' >> .git/hooks/pre-commit
	@echo '  echo "💡 Fix the issues or use: git commit --no-verify"' >> .git/hooks/pre-commit
	@echo '  exit 1' >> .git/hooks/pre-commit
	@echo 'fi' >> .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "✅ Pre-commit hook installed"
	@echo "   Will run 'make validate-quick' before each commit"
	@echo "   To skip: git commit --no-verify"

# Uninstall git hooks
uninstall-hooks:
	@echo "🗑️  Removing git pre-commit hook..."
	@rm -f .git/hooks/pre-commit
	@echo "✅ Hook removed"

# Clean local branches that don't exist in remote
clean-branches:
	@echo "🧹 Cleaning local branches that don't exist in remote..."
	@echo ""
	@echo "📡 Fetching from remote and pruning deleted branches..."
	@git fetch --prune
	@echo ""
	@echo "🔍 Finding local branches to delete..."
	@BRANCHES=$$(git branch -vv | grep ': gone]' | awk '{print $$1}'); \
	if [ -z "$$BRANCHES" ]; then \
		echo "✅ No stale branches found - all local branches are in sync"; \
	else \
		echo "Found the following branches to delete:"; \
		echo "$$BRANCHES" | sed 's/^/  - /'; \
		echo ""; \
		printf "Delete these branches? [y/N] "; \
		read REPLY; \
		case $$REPLY in \
			[Yy]*) \
				echo "$$BRANCHES" | xargs git branch -D; \
				echo ""; \
				echo "✅ Branches deleted successfully"; \
				;; \
			*) \
				echo "❌ Operation cancelled"; \
				;; \
		esac; \
	fi

# Force clean local branches (no confirmation)
clean-branches-force:
	@echo "🧹 Force cleaning local branches that don't exist in remote..."
	@echo ""
	@echo "📡 Fetching from remote and pruning deleted branches..."
	@git fetch --prune
	@echo ""
	@BRANCHES=$$(git branch -vv | grep ': gone]' | awk '{print $$1}'); \
	if [ -z "$$BRANCHES" ]; then \
		echo "✅ No stale branches found - all local branches are in sync"; \
	else \
		echo "Deleting the following branches:"; \
		echo "$$BRANCHES" | sed 's/^/  - /'; \
		echo ""; \
		echo "$$BRANCHES" | xargs git branch -D; \
		echo ""; \
		echo "✅ Branches deleted successfully"; \
	fi

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

# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
# 🧪 FINAL TEST - Production build + quickstart test
# ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

finaltest:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "🧪 FINAL TEST - Building production image and testing quickstart"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "📦 Step 1: Building production image with local tag..."
	@echo "   💡 Usando BuildKit cache - solo reconstruye capas modificadas"
	@echo ""
	@docker build \
		--target production \
		--cache-from mediacheky:finaltest \
		--build-arg VERSION=finaltest \
		--build-arg COMMIT_SHA=$$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
		-t mediacheky:finaltest \
		.
	@echo ""
	@echo "✅ Production image built: mediacheky:finaltest"
	@echo ""
	@echo "📋 Step 2: Preparing finaltest environment..."
	@mkdir -p quickstart-finaltest/volumes/mediacheky-data
	@mkdir -p quickstart-finaltest/volumes/library
	@echo "✅ Directories created"
	@echo ""
	@echo "🚀 Step 3: Starting finaltest with local image..."
	@echo ""
	@cd quickstart-finaltest && docker compose up -d
	@echo ""
	@echo "📝 Step 4: Capturing logs to finaltest.log..."
	@cd quickstart-finaltest && nohup docker compose logs -f &> finaltest.log & echo $$! > finaltest.pid
	@sleep 2
	@echo "✅ Logs being captured in background (PID: $$(cat quickstart-finaltest/finaltest.pid))"
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "✅ FINAL TEST READY"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo ""
	@echo "🌐 Access:"
	@echo "   URL: http://localhost:80"
	@echo ""
	@echo "📊 Status:"
	@cd quickstart-finaltest && docker compose ps
	@echo ""
	@echo "📝 Logs:"
	@echo "   File: quickstart-finaltest/finaltest.log"
	@echo "   Live: tail -f quickstart-finaltest/finaltest.log"
	@echo ""
	@echo "💡 Commands:"
	@echo "   View logs:  tail -f quickstart-finaltest/finaltest.log"
	@echo "   Stop test:  make finaltest-stop"
	@echo "   Clean all:  make finaltest-clean"
	@echo ""

finaltest-stop:
	@echo "🛑 Stopping final test environment..."
	@if [ -f quickstart-finaltest/finaltest.pid ]; then \
		echo "   Stopping log capture (PID: $$(cat quickstart-finaltest/finaltest.pid))..."; \
		kill $$(cat quickstart-finaltest/finaltest.pid) 2>/dev/null || true; \
		rm quickstart-finaltest/finaltest.pid; \
	fi
	@cd quickstart-finaltest && docker compose down
	@echo "✅ Final test stopped"

finaltest-clean:
	@echo "🧹 Cleaning final test environment..."
	@if [ -f quickstart-finaltest/finaltest.pid ]; then \
		kill $$(cat quickstart-finaltest/finaltest.pid) 2>/dev/null || true; \
		rm quickstart-finaltest/finaltest.pid; \
	fi
	@cd quickstart-finaltest && docker compose down -v
	@rm -rf quickstart-finaltest/volumes/mediacheky-data
	@rm -rf quickstart-finaltest/volumes/library
	@rm -f quickstart-finaltest/finaltest.log
	@docker rmi mediacheky:finaltest ghcr.io/carcheky/mediacheky:finaltest 2>/dev/null || true
	@echo "✅ Final test environment cleaned"

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
	@mkdir -p volumes/library/library/movies
	@mkdir -p volumes/library/library/tv
	@mkdir -p volumes/library/downloads
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
