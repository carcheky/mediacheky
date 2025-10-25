.PHONY: help build run test clean docker-build docker-run dev

# Default target
help:
	@echo "MediaCheky - Makefile commands:"
	@echo "  make build         - Build the application"
	@echo "  make run           - Run the application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make dev           - Run in development mode"

# Build the application
build:
	@echo "Building MediaCheky..."
	go build -ldflags="-w -s" -o bin/mediacheky ./cmd/server
	@echo "Build complete: bin/mediacheky"

# Run the application
run:
	@echo "Running MediaCheky..."
	go run ./cmd/server

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f mediacheky
	@echo "Clean complete"

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t mediacheky:latest .
	@echo "Docker image built: mediacheky:latest"

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker-compose up -d
	@echo "Container started. Access http://localhost:8080"

# Run in development mode with live reload
dev:
	@echo "Running in development mode..."
	@echo "Make sure you have 'air' installed: go install github.com/cosmtrek/air@latest"
	air

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod verify
	@echo "Dependencies downloaded"

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	go mod tidy
	@echo "Dependencies tidied"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	golangci-lint run ./...
	@echo "Linting complete"

# Install development tools
install-tools:
	@echo "Installing development tools..."
	go install github.com/cosmtrek/air@latest
	@echo "Tools installed"
