# Stock Ticker Helm Chart Quick Start Guide

This guide provides simple instructions for deploying the Stock Ticker Go application to Kubernetes using Helm.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+
- Alpha Vantage API key

## Setup GitHub Container Registry Access

Since the container image is stored in a private GitHub Container Registry, you need to set up authentication:

```bash
# Set your GitHub credentials
export GITHUB_USERNAME=your_github_username
export GITHUB_TOKEN=your_github_token  # Token with 'read:packages' scope

# Create the pull secret
cd helm/
./create-github-secret.sh --namespace=your-namespace
```

## Deploy the Go Application

### Option 1: Using the Deploy Script (Recommended)

```bash
# Deploy the Go app with default settings (Ingress)
./deploy.sh --api-key=your_api_key --namespace=your-namespace

# Deploy with LoadBalancer instead of Ingress
./deploy.sh --api-key=your_api_key --namespace=your-namespace --values=ping-values-lb.yaml

```

### Option 2: Manual Helm Installation

```bash
# Deploy with default settings (Ingress)
helm install stock-ticker ./helm/stock-ticker \
  --namespace your-namespace \
  --set apiKey=your_api_key

# Or deploy with LoadBalancer
helm install stock-ticker ./helm/stock-ticker \
  --namespace your-namespace \
  --values ./ping-values-lb.yaml \
  --set apiKey=your_api_key

```

## Verify Deployment

```bash
# Check if pods are running
kubectl get pods -n your-namespace

# Check the service
kubectl get svc -n your-namespace
```

## Access the Stock Ticker API

Depending on your deployment option:

### For Ingress Deployment

```bash
# Access via hostname
curl https://stock-ticker.thoreinstein.com/
```

### For LoadBalancer Deployment

```bash
# Get the LoadBalancer external IP
kubectl get svc -n your-namespace stock-ticker

# Access via IP
curl http://<EXTERNAL-IP>/
```

## Uninstall

```bash
helm uninstall stock-ticker -n your-namespace
```