# Ping SRE Exercise - Stock Ticker Service

A web service that fetches and returns closing stock prices for a specified number of days.

## Part 1: Stock Ticker Service

### Building the Docker Image

```bash
# Using make
make docker-build

# Or using Docker directly
docker build -t stock-ticker:latest .
```

### Running the Service

The service requires the following environment variables:
- `SYMBOL`: Stock symbol to fetch (e.g., MSFT)
- `NDAYS`: Number of days of data to return
- `APIKEY`: Alpha Vantage API key

```bash
# Using make (reads variables from your environment)
export SYMBOL=MSFT
export NDAYS=7
export APIKEY=your_api_key
make docker-run

# Or using Docker directly
docker run -p 8080:8080 \
  -e SYMBOL=MSFT \
  -e NDAYS=7 \
  -e APIKEY=your_api_key \
  stock-ticker:latest
```

### Testing the Service

```bash
curl http://localhost:8080
```

## Testing and Development

This project follows Inside-Out TDD with comprehensive test coverage (currently 85.7%), along with robust linting and code quality checks.

### Testing Commands

```bash
# Run all tests
make test

# Run unit tests only
make test-unit

# Run integration tests only
make test-integration

# Run tests with coverage
make test-coverage

# Run tests and build
make all
```

### Linting

```bash
# Install golangci-lint
make install-tools

# Add golangci-lint to your PATH
export PATH=$HOME/go/bin:$PATH

# Run linters
make lint
```

### Testing Details

This project follows Inside-Out TDD principles with unit tests in the `tests/unit` directory 
and integration tests in the `tests/integration` directory.

## CI/CD Pipeline

This project uses GitHub Actions for CI/CD with separate workflows for different purposes:

### Enhanced CI/CD Pipeline (`enhanced-pipeline.yaml`)

This workflow runs on pushes to main and provides a comprehensive CI/CD pipeline:

1. **Setup**: Configures the Go environment and caches dependencies
2. **Lint**: Runs static code analysis and formatting checks
3. **Test**: Executes unit and integration tests with coverage reporting
4. **Build**: Builds and pushes the Docker image to GitHub Container Registry
5. **Security Scan**: Scans the container image for vulnerabilities using Trivy
6. **Helm Validation**: Validates and packages the Helm chart

### Pull Request Checks (`pr-checks.yaml`)

This workflow runs on pull requests to ensure code quality before merging:

1. **Lint**: Validates code style and formatting
2. **Test**: Runs unit and integration tests
3. **Docker Validation**: Ensures the Docker build is successful
4. **Helm Validation**: Validates Helm chart templates

### Manual Publishing

```bash
# Using make (to GitHub Container Registry)
make docker-build
make docker-push

# Or using Docker directly
docker tag stock-ticker:latest ghcr.io/username/stock-ticker:latest
docker push ghcr.io/username/stock-ticker:latest
```

## Part 2: Kubernetes Deployment

### Option 1: Using Helm (Recommended)

A Helm chart is available in the `charts/` directory for more flexible and configurable deployment.

#### Documentation

Detailed information about the Helm chart can be found in the chart's values.yaml file.

#### GitHub Container Registry Access

The container image is stored in a public GitHub Container Registry, so no authentication is required to pull the image.

#### Deploy the Application

```bash
# Deploy with default settings (Ingress enabled)
cd charts/
helm install stock-ticker ./stock-ticker --namespace=stock-ticker --create-namespace --set apiKey=your_api_key

# Deploy with LoadBalancer instead of Ingress
cd charts/
helm install stock-ticker ./stock-ticker --namespace=stock-ticker --create-namespace -f ping-values-lb.yaml --set apiKey=your_api_key

# Deploy to production environment
cd charts/
helm install stock-ticker ./stock-ticker --namespace=stock-ticker-prod --create-namespace --set apiKey=your_api_key
```

#### Accessing the Service

Depending on your deployment option:

- **LoadBalancer**: Access directly via the external IP
  ```bash
  # Get the external IP
  kubectl get svc -n ping stock-ticker-stock-ticker
  # Access the service
  curl http://<EXTERNAL-IP>/
  ```

- **Ingress**: Access via the hostname
  ```bash
  curl https://stock-ticker.thoreinstein.com/
  ```

#### Testing the Helm Chart Locally

To test the Helm chart deployment locally:

```bash
# Lint the chart
make helm-lint

# Render templates for validation
make helm-template

# Package the chart
make helm-package

# Run all Helm validation in sequence
make helm-test

# Install with dry-run to check what would be applied
helm install stock-ticker ./charts/stock-ticker --dry-run --set apiKey=your-api-key

# Install the chart
helm install stock-ticker ./charts/stock-ticker --set apiKey=your-api-key

# Test with port-forwarding
kubectl port-forward svc/stock-ticker-stock-ticker 8080:80
curl http://localhost:8080
```

The values.yaml file contains detailed information about all configurable options.

### Option 2: Using Pre-rendered Manifests

A pre-rendered Kubernetes manifest generated from the Helm chart is available in the `manifests/` directory:

```bash
# Replace the placeholder API key with your actual key
sed -i 's/PLACEHOLDER_API_KEY/your_actual_api_key/g' manifests/stock-ticker-rendered.yaml

# Apply the manifest
kubectl create namespace stock-ticker
kubectl apply -f manifests/stock-ticker-rendered.yaml -n stock-ticker
```


### Testing the Deployment

```bash
# For Helm-based deployment
kubectl port-forward -n stock-ticker svc/stock-ticker-stock-ticker 8080:80

# Test the API
curl http://localhost:8080
```

For production deployments with Ingress enabled, the service would be accessible via the configured hostname.