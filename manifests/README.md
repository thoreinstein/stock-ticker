# Pre-rendered Kubernetes Manifests

This directory contains pre-rendered Kubernetes manifests generated from the Helm chart.

## stock-ticker-rendered.yaml

This file contains all the Kubernetes resources for deploying the Stock Ticker API with default configuration values. It was generated using:

```bash
helm template stock-ticker ../helm/stock-ticker --set apiKey=PLACEHOLDER_API_KEY > stock-ticker-rendered.yaml
```

### Usage

1. Replace the placeholder API key:
   ```bash
   sed -i 's/PLACEHOLDER_API_KEY/your_actual_api_key/g' stock-ticker-rendered.yaml
   ```

2. Apply to your Kubernetes cluster:
   ```bash
   kubectl create namespace stock-ticker
   kubectl apply -f stock-ticker-rendered.yaml -n stock-ticker
   ```

3. Check deployment status:
   ```bash
   kubectl get pods -n stock-ticker
   ```

### Customization

For more customization options, it's recommended to use the Helm chart directly. See the `../helm/stock-ticker/README.md` for details.

If you need to customize these static manifests, you can:

1. Modify this file directly
2. Or regenerate it with different values:
   ```bash
   helm template stock-ticker ../helm/stock-ticker \
     --set apiKey=PLACEHOLDER_API_KEY \
     --set replicaCount=5 \
     --set image.tag=v1.2.3 \
     > stock-ticker-custom.yaml
   ```