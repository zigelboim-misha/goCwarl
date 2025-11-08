# goCrawl Helm Chart

A Helm chart for deploying goCrawl - a Go service that dynamically manages crawl4ai pods for web crawling.

## Overview

goCrawl is a microservice that:
- Accepts crawl requests via REST API
- Dynamically creates crawl4ai pods in Kubernetes
- Forwards crawl requests to the temporary pods
- Returns filtered results
- Automatically cleans up pods after completion

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- RBAC enabled cluster

## Installation

### Build and Push Docker Image

First, build and push the goCrawl Docker image:

```bash
cd services/gocrawl

# Build the image
docker build -t your-registry/gocrawl:latest .

# Push to your registry
docker push your-registry/gocrawl:latest
```

### Install the Helm Chart

```bash
# Install with default values
helm install gocrawl ./helm/gocrawl

# Install with custom image
helm install gocrawl ./helm/gocrawl \
  --set image.repository=your-registry/gocrawl \
  --set image.tag=latest

# Install in specific namespace
helm install gocrawl ./helm/gocrawl -n your-namespace --create-namespace
```

## Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `gocrawl` |
| `image.tag` | Container image tag | `latest` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |
| `service.targetPort` | Container target port | `8080` |
| `serviceAccount.create` | Create service account | `true` |
| `serviceAccount.name` | Service account name | `gocrawl` |
| `rbac.create` | Create RBAC resources | `true` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `250m` |
| `resources.requests.memory` | Memory request | `256Mi` |

## Usage

### API Endpoint

Once deployed, the service exposes the following endpoint:

**POST** `/api/v1/crawl`

Request body:
```json
{
  "model": "anthropic/claude-3-opus"
}
```

Response:
```json
{
  "url": "https://openrouter.ai/anthropic/claude-3-opus",
  "success": true,
  "markdown": "# Model documentation..."
}
```

### Access the Service

```bash
# Port forward to access locally
kubectl port-forward svc/gocrawl 8080:8080

# Make a crawl request
curl -X POST http://localhost:8080/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{"model": "anthropic/claude-3-opus"}'
```

### Swagger Documentation

Access the Swagger UI at:
```
http://<service-url>:8080/swagger/index.html
```

### Health Checks

- Health endpoint: `GET /health`
- Readiness endpoint: `GET /ready`

## RBAC Permissions

The service requires the following permissions to manage pods:
- `create`, `get`, `list`, `watch`, `delete` on pods
- `get` on pod logs

These are automatically configured when `rbac.create=true`.

## Upgrading

```bash
# Upgrade with new values
helm upgrade gocrawl ./helm/gocrawl --set replicaCount=2

# Upgrade with values file
helm upgrade gocrawl ./helm/gocrawl -f custom-values.yaml
```

## Uninstalling

```bash
helm uninstall gocrawl
```

## Troubleshooting

### Pods not being created

Check RBAC permissions:
```bash
kubectl get role,rolebinding -l app.kubernetes.io/name=gocrawl
```

### Service not responding

Check logs:
```bash
kubectl logs -l app.kubernetes.io/name=gocrawl
```

### Failed crawl requests

The service automatically cleans up failed pods. Check the API response for error details.

## Architecture

```
User Request → goCrawl Service → Creates crawl4ai Pod → Crawls URL → Returns Filtered Data → Deletes Pod
```

Each crawl request:
1. Creates a unique crawl4ai pod with shared memory (3Gi)
2. Waits for pod to be ready (max 2 minutes)
3. Sends crawl request to pod's IP
4. Filters response to return only URL, success, and markdown
5. Deletes the pod (even on error)
