# EngLog - Makefile for Development
.PHONY: run build test clean help

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=englog-api
MAIN_PATH=./cmd/api
BUILD_DIR=./bin

## run: Run the application in development mode
run:
	@echo "🚀 Starting EngLog API server..."
	go run $(MAIN_PATH)

## build: Build the application binary
build:
	@echo "🔨 Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "✅ Binary built: $(BUILD_DIR)/$(BINARY_NAME)"

## test: Run all tests
test:
	@echo "🧪 Running tests..."
	go test -v ./...

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	go clean

## mod: Download and tidy Go modules
mod:
	@echo "📦 Downloading and tidying Go modules..."
	go mod download
	go mod tidy

## fmt: Format Go code
fmt:
	@echo "📝 Formatting Go code..."
	go fmt ./...

## vet: Run go vet
vet:
	@echo "🔍 Running go vet..."
	go vet ./...

## check: Run formatting, vetting, and tests
check: fmt vet test
	@echo "✅ All checks passed!"

## docker-build: Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t englog:latest .
	@echo "✅ Docker image built: englog:latest"

## docker-run: Run with Docker Compose (production mode)
docker-run:
	@echo "🐳 Starting EngLog with Docker Compose..."
	./scripts/docker-setup.sh

## docker-dev: Run with Docker Compose (development mode)
docker-dev:
	@echo "🐳 Starting EngLog in development mode..."
	./scripts/docker-setup.sh --dev

## docker-stop: Stop Docker services
docker-stop:
	@echo "⏹️  Stopping Docker services..."
	docker-compose down

## docker-logs: Show Docker logs
docker-logs:
	@echo "📋 Docker logs:"
	docker-compose logs -f

## docker-clean: Clean Docker resources
docker-clean: docker-stop
	@echo "🧹 Cleaning Docker resources..."
	docker-compose down -v
	docker system prune -f

## help: Show this help message
help:
	@echo "EngLog API - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
