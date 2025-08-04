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

## help: Show this help message
help:
	@echo "EngLog API - Available Commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
