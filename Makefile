.PHONY: test test-unit test-integration test-coverage test-all test-benchmark lint build run docker-build docker-run docker-logs docker-push helm-test helm-lint k8s-test

# Default target
all: test build

# Run unit tests
test-unit:
	@echo "Running unit tests..."
	@go test -v ./tests/unit/...

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	@go test -v ./tests/integration/...

# Run all tests
test: test-unit test-integration

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build the application
build:
	@echo "Building application..."
	@go build -o bin/stock-ticker cmd/main.go

# Run the application
run:
	@echo "Running application..."
	@go run cmd/main.go

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t ping-sre .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker stop ping-stock-ticker >/dev/null 2>&1 || true
	@docker rm ping-stock-ticker >/dev/null 2>&1 || true
	@docker run --name ping-stock-ticker -d -p 8080:8080 -e SYMBOL -e NDAYS -e APIKEY ping-sre
	@echo "Container started. Check logs with: docker logs ping-stock-ticker"
	@echo "Test the API with: curl http://localhost:8080"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/ coverage.out coverage.html

# Development workflow - test then run
dev: test run

# Run linting
lint:
	@echo "Running linters..."
	@go vet ./...
	@go fmt ./...

# Run benchmarks
test-benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Test Helm chart with dry-run
helm-test:
	@echo "Testing Helm chart..."
	@helm lint ./helm/stock-ticker
	@helm template --debug stock-ticker ./helm/stock-ticker

# Lint Helm chart
helm-lint:
	@echo "Linting Helm chart..."
	@helm lint ./helm/stock-ticker

# Test Kubernetes manifests with dry-run
k8s-test:
	@echo "Testing Kubernetes manifests..."
	@kubectl apply --dry-run=client -f manifests/stock-ticker-rendered.yaml

# CI/CD pipeline simulation
ci: test build docker-build
	@echo "CI pipeline complete!"

# Check Docker container logs
docker-logs:
	@docker logs ping-stock-ticker

# Push Docker image to GitHub Container Registry
docker-push:
	@echo "Pushing Docker image to GitHub Container Registry..."
	@docker tag ping-sre ghcr.io/$(shell git config --get remote.origin.url | cut -d: -f2 | cut -d. -f1)/stock-ticker:latest
	@docker push ghcr.io/$(shell git config --get remote.origin.url | cut -d: -f2 | cut -d. -f1)/stock-ticker:latest
	@echo "Image pushed successfully"
