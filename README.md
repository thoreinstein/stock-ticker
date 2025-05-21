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

## Testing

This project follows Inside-Out TDD with comprehensive test coverage (currently 85.7%).

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

### Testing Documentation

Detailed testing information is available in these documents:

- `README_TESTING.md` - Testing strategy and methodology
- `TESTING_SUMMARY.md` - Current implementation details and coverage
- `TEST_PLAN.md` - Test execution checklist and procedures

## CI/CD and Publishing

This project uses GitHub Actions for CI/CD. On push to main or when a tag is created,
the workflow will automatically:
1. Run tests
2. Build the Docker image
3. Push to GitHub Container Registry (ghcr.io)

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

A Helm chart is available in the `helm/` directory for more flexible and configurable deployment.

#### Documentation

For detailed information about the Helm chart, see:
- [HELM_CHART.md](./HELM_CHART.md) - Full documentation for the Helm chart
- [Values Configuration](./helm/stock-ticker/values.yaml) - Configurable values with comments

#### Setup GitHub Container Registry Access

Since the container image is stored in a private GitHub Container Registry, you need to set up authentication:

```bash
# Set your GitHub credentials
export GITHUB_USERNAME=your_github_username
export GITHUB_TOKEN=your_github_token  # Token with 'read:packages' scope

# Create the pull secret
cd helm/
./create-github-secret.sh --namespace=ping
```

#### Deploy the Application

```bash
# Deploy with default settings (Ingress enabled)
./deploy.sh --api-key=your_api_key --namespace=ping

# Deploy with LoadBalancer instead of Ingress
./deploy.sh --api-key=your_api_key --namespace=ping --values=ping-values-lb.yaml


# Deploy to production environment
./deploy.sh --env=prod --namespace=stock-ticker-prod --api-key=your_api_key

# For more options
./deploy.sh --help
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
helm lint ./helm/stock-ticker

# Do a dry-run template rendering
helm template stock-ticker ./helm/stock-ticker --set apiKey=your-api-key

# Install with dry-run to check what would be applied
helm install stock-ticker ./helm/stock-ticker --dry-run --set apiKey=your-api-key

# Install the chart
helm install stock-ticker ./helm/stock-ticker --set apiKey=your-api-key

# Test with port-forwarding
kubectl port-forward svc/stock-ticker-stock-ticker 8080:80
curl http://localhost:8080
```

See the [HELM_CHART.md](./HELM_CHART.md) file for more details on configuring and using the Helm chart.

### Option 2: Using Pre-rendered Manifests

A pre-rendered Kubernetes manifest generated from the Helm chart is available in the `manifests/` directory:

```bash
# Replace the placeholder API key with your actual key
sed -i 's/PLACEHOLDER_API_KEY/your_actual_api_key/g' manifests/stock-ticker-rendered.yaml

# Apply the manifest
kubectl create namespace stock-ticker
kubectl apply -f manifests/stock-ticker-rendered.yaml -n stock-ticker
```

### Option 3: Using Raw Kubernetes Manifests

Basic Kubernetes manifests are in the `k8s/` directory.

```bash
kubectl create namespace stock-ticker

kubectl apply -f k8s/ -n stock-ticker

kubectl get all -n stock-ticker
```

### Testing the Deployment

```bash
# For Helm-based deployment
kubectl port-forward -n stock-ticker svc/stock-ticker-stock-ticker 8080:80

# For manifest-based deployment
kubectl port-forward -n stock-ticker service/stock-ticker 8080:80

# Test the API
curl http://localhost:8080
```

For production deployments, the service is accessible via the Ingress at `https://stock-ticker.thoreinstein.com`.
