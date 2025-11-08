# goCrawl - Dynamic Crawl4AI Pod Management Service

A Go microservice that dynamically creates and manages crawl4ai pods in Kubernetes for web crawling operations.

## Features

- REST API endpoint for crawling AI model pages
- Dynamic Kubernetes pod creation and lifecycle management
- Automatic pod cleanup after request completion
- Swagger/OpenAPI documentation
- Health and readiness probes
- RBAC-based pod management

## Project Structure

```
gocrawl/
├── handlers/           # HTTP request handlers
│   └── crawl.go       # Main crawl endpoint handler
├── k8s/               # Kubernetes client operations
│   └── pod_manager.go # Pod lifecycle management
├── models/            # Data models
│   ├── request.go     # Request structures
│   └── response.go    # Response structures
├── helm/              # Helm chart for deployment
│   └── gocrawl/
├── main.go            # Application entry point
├── Dockerfile         # Multi-stage Docker build
└── go.mod            # Go module definition
```

## API Reference

### POST /api/v1/crawl

Crawls an AI model page from OpenRouter.

**Request:**
```json
{
  "model": "anthropic/claude-3-opus"
}
```

**Response:**
```json
{
  "url": "https://openrouter.ai/anthropic/claude-3-opus",
  "success": true,
  "markdown": "# Claude 3 Opus\n\n..."
}
```

**Model Format:** Must be in `vendor/model-name` format

### GET /health

Health check endpoint.

### GET /ready

Readiness check endpoint.

### GET /swagger/index.html

Interactive Swagger UI documentation.

## Development

### Prerequisites

- Go 1.21+
- Docker
- Kubernetes cluster (for deployment)
- kubectl configured

### Local Development

1. **Install dependencies:**
```bash
cd services/gocrawl
go mod download
```

2. **Generate Swagger docs:**
```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init
```

3. **Build:**
```bash
go build -o gocrawl .
```

4. **Run locally** (requires kubeconfig):
```bash
export POD_NAMESPACE=default
./gocrawl
```

The service will start on port 8080.

### Building Docker Image

```bash
docker build -t gocrawl:latest .
```

## Deployment

### Using Helm

See [helm/gocrawl/README.md](helm/gocrawl/README.md) for detailed deployment instructions.

Quick start:
```bash
# Build and push image
docker build -t your-registry/gocrawl:latest .
docker push your-registry/gocrawl:latest

# Install chart
helm install gocrawl ./helm/gocrawl \
  --set image.repository=your-registry/gocrawl \
  --set image.tag=latest
```

## How It Works

1. **Request received** at `/api/v1/crawl` with model name
2. **Pod creation**: Creates unique crawl4ai pod with:
   - Image: `unclecode/crawl4ai:0.7.5`
   - Shared memory: 3Gi (for headless browser)
   - Unique name based on timestamp
3. **Wait for ready**: Polls until pod is running (max 2 minutes)
4. **Get pod IP**: Retrieves the pod's cluster IP
5. **Send crawl request**: POSTs to `http://<pod-ip>:11235/crawl`
6. **Filter response**: Extracts only `url`, `success`, `markdown.raw_markdown`
7. **Cleanup**: Deletes pod (always, even on error)
8. **Return response**: Sends filtered data back to client

## RBAC Requirements

The service account needs permissions to:
- Create pods
- Get pod status and IP
- Delete pods
- Read pod logs

These are automatically configured by the Helm chart.

## Configuration

Environment variables:
- `POD_NAMESPACE`: Kubernetes namespace for pod operations (required)
- `PORT`: HTTP server port (default: 8080)

## Error Handling

- Invalid model format → 400 Bad Request
- Pod creation failure → 500 Internal Server Error
- Pod not ready timeout → 500 Internal Server Error
- Crawl request failure → 500 Internal Server Error
- Pods are always cleaned up, even on errors

## Testing

Example curl request:
```bash
curl -X POST http://localhost:8080/api/v1/crawl \
  -H "Content-Type: application/json" \
  -d '{
    "model": "anthropic/claude-3-5-sonnet"
  }'
```

## Dependencies

- `github.com/gin-gonic/gin` - HTTP framework
- `github.com/swaggo/swag` - Swagger documentation
- `k8s.io/client-go` - Kubernetes client
- `k8s.io/api` - Kubernetes API types

## License

See project root for license information.
