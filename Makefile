.PHONY: test test-unit test-integration test-coverage test-all test-benchmark lint golangci-lint lint-errcheck build run docker-build docker-run docker-logs docker-push helm-lint helm-template helm-package helm-test k8s-test install-tools

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

# Build multi-platform Docker image (AMD64 and ARM64)
docker-build-multiplatform:
	@echo "Building multi-platform Docker image..."
	@docker buildx create --use --name multiplatform-builder >/dev/null 2>&1 || true
	@docker buildx build --platform linux/amd64,linux/arm64 -t ping-sre:multiplatform -f Dockerfile.multiplatform . --load
	@echo "Multi-platform images built for Linux AMD64 and ARM64"

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
	@rm -rf bin/ coverage.out coverage.html *.tgz

# Development workflow - test then run
dev: test run

# Run linting
lint:
	@echo "Running linters..."
	@go vet ./...
	@go fmt ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	elif [ -f $$HOME/go/bin/golangci-lint ]; then \
		$$HOME/go/bin/golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not found, skipping"; \
	fi

# Test cross-platform compatibility
cross-platform-test:
	@echo "Testing cross-platform compatibility (Linux AMD64)..."
	@docker build -t stock-ticker-test-amd64 -f Dockerfile.test . --platform linux/amd64
	@docker run --platform linux/amd64 stock-ticker-test-amd64
	@echo "Success! Code works on Linux AMD64 architecture."

# Run golangci-lint directly
golangci-lint:
	@echo "Running golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	elif [ -f $$HOME/go/bin/golangci-lint ]; then \
		$$HOME/go/bin/golangci-lint run --timeout=5m; \
	else \
		echo "golangci-lint not found, skipping"; \
	fi

# Run benchmarks
test-benchmark:
	@echo "Running benchmarks..."
	@go test -bench=. ./...

# Lint Helm chart
helm-lint:
	@echo "Linting Helm chart..."
	@helm lint ./charts/stock-ticker

# Template Helm chart
helm-template:
	@echo "Templating Helm chart..."
	@helm template stock-ticker ./charts/stock-ticker --set apiKey=dummy-api-key --debug

# Package Helm chart
helm-package:
	@echo "Packaging Helm chart..."
	@helm package ./charts/stock-ticker

# Test Helm chart with all checks
helm-test: helm-lint helm-template helm-package
	@echo "Helm chart validation complete"

# Test Kubernetes manifests with dry-run
k8s-test:
	@echo "Testing Kubernetes manifests..."
	@kubectl apply --dry-run=client -f manifests/stock-ticker-rendered.yaml

# CI/CD pipeline simulation
ci: lint test build docker-build cross-platform-test
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

# Install development tools
install-tools:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/kisielk/errcheck@latest

# Run errcheck specifically
lint-errcheck:
	@echo "Running errcheck..."
	@if command -v errcheck >/dev/null 2>&1; then \
		errcheck ./...; \
	elif [ -f $$HOME/go/bin/errcheck ]; then \
		$$HOME/go/bin/errcheck ./...; \
	else \
		echo "errcheck not found, installing..."; \
		go install github.com/kisielk/errcheck@latest; \
		$$HOME/go/bin/errcheck ./...; \
	fi
