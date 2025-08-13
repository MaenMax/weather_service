.PHONY: help build run test clean lint vet coverage docker-build docker-run

# Default target
help: ## Show this help message
	@echo "Weather Service - Available Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build the application
build: ## Build the weather service binary
	@echo "Building weather service..."
	@go build -o bin/weather_service cmd/server/main.go
	@echo "Build complete: bin/weather_service"

# Run the application
run: ## Run the weather service
	@echo "Starting weather service..."
	@go run cmd/server/main.go

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Lint the code
lint: ## Run golangci-lint (if installed)
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "Running linter..."; \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Run go vet
vet: ## Run go vet for static analysis
	@echo "Running go vet..."
	@go vet ./...

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	@go fmt ./...

# Check for security vulnerabilities
security: ## Check for security vulnerabilities
	@echo "Checking for security vulnerabilities..."
	@go list -json -deps ./... | nancy sleuth

# Install dependencies
deps: ## Install Go dependencies
	@echo "Installing dependencies..."
	@go mod tidy
	@go mod download

# Docker build
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t weather-service:latest .
	@echo "Docker image built: weather-service:latest"

# Docker run
docker-run: ## Run Docker container
	@echo "Running Docker container..."
	@docker run -p 8080:8080 weather-service:latest

# Development setup
dev-setup: deps ## Setup development environment
	@echo "Development environment setup complete"

# Production build
prod-build: ## Build optimized production binary
	@echo "Building production binary..."
	@CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/weather_service cmd/server/main.go
	@echo "Production build complete: bin/weather_service"

# Check if all tests pass
check: test vet ## Run tests and static analysis
	@echo "All checks passed!"

# Install development tools
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/go-delve/delve/cmd/dlv@latest
	@echo "Development tools installed"

# Show project info
info: ## Show project information
	@echo "Weather Service - Project Information"
	@echo "===================================="
	@echo "Go version: $(shell go version)"
	@echo "Go modules: $(shell go list -m)"
	@echo "Project path: $(shell pwd)"
	@echo "Binary size: $(shell if [ -f bin/weather_service ]; then ls -lh bin/weather_service | awk '{print $$5}'; else echo "Not built"; fi)"

# Default target
all: clean build test ## Clean, build, and test
