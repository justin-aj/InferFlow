# InferFlow

InferFlow is a Week 1 MVP for a scalable LLM inference router project. This repository bootstraps the local development path, the Go router, a mock inference backend, starter Kubernetes assets, and Terraform scaffolding for AWS EKS so the team can move from proposal to implementation quickly.

## Current MVP Status

Implemented now:

- Go router with `POST /v1/chat/completions`
- Health endpoints: `GET /healthz` and `GET /readyz`
- Round-robin routing across configured backends
- Background backend health probing
- Local mock backend for end-to-end development
- Python load generator that writes proposal-aligned CSV output
- Docker Compose for local router + mock backend
- Starter Terraform for AWS/EKS baseline
- Starter Kubernetes manifests and Helm values placeholders

Planned next:

- Triton integration and model repository wiring
- Streaming SSE responses
- Kubernetes endpoint discovery
- Least-pending and cost-aware routing
- Observability stack wiring (Prometheus, Tempo, Grafana, OTel)
- Autoscaling and experiment automation

## Architecture

The first pass keeps the public API stable while swapping the backend implementation later:

1. Clients send OpenAI-compatible chat completion requests to the Go router.
2. The router chooses a healthy backend using round-robin.
3. The router forwards the request to the local mock backend today.
4. Later, the same router surface can forward to Triton-backed inference adapters.

## Repository Layout

- `cmd/router`: router entrypoint
- `cmd/mock-backend`: local fake inference backend
- `internal/router`: strategy and backend registry
- `internal/proxy`: backend client and translation layer
- `internal/metrics`: in-memory request metrics
- `internal/otel`: lightweight tracing hooks placeholder
- `internal/server`: HTTP handlers and server wiring
- `loadgen`: Python request generator
- `analysis`: starter analysis scaffold
- `triton`: placeholder Triton model repository layout
- `k8s`: starter manifests and Helm values
- `scripts`: local and cluster orchestration helpers
- `terraform`: AWS/EKS baseline infrastructure
- `results`: generated experiment artifacts

## Local Quick Start

### Option 1: Native processes

Start the mock backend:

```bash
go run ./cmd/mock-backend
```

In another terminal, start the router:

```bash
$env:INFERFLOW_BACKENDS="http://localhost:9000"
go run ./cmd/router
```

Send a sample request:

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d "{\"model\":\"mock-llm\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello from InferFlow\"}]}"
```

Run tests:

```bash
go test ./...
```

Generate a sample CSV:

```bash
python loadgen/generator.py --requests 5 --output results/sample.csv
```

### Option 2: Docker Compose

```bash
docker compose up --build
```

The router listens on `http://localhost:8080` and the mock backend is internal to Compose.

## Router API

### `POST /v1/chat/completions`

Accepts a minimal OpenAI-compatible request body:

```json
{
  "model": "mock-llm",
  "messages": [
    { "role": "user", "content": "Hello" }
  ],
  "stream": false
}
```

Returns a minimal OpenAI-compatible response shape containing:

- `id`
- `object`
- `created`
- `model`
- `choices`
- `usage`

### `GET /healthz`

Returns process liveness.

### `GET /readyz`

Returns success only when at least one backend is currently healthy.

## Infrastructure Workflow

Terraform is the source of truth for AWS infrastructure. Helm and `kubectl` will remain the source of truth for in-cluster services such as Triton, Prometheus, Grafana, Tempo, and the router deployment.

### Terraform scope for Week 1

Implemented as starter baseline:

- VPC
- Public and private subnets
- EKS cluster
- Managed GPU node group baseline
- IAM roles required for cluster and node group creation

Not wired yet:

- Production-ready networking hardening
- Observability add-ons
- Autoscaling extensions
- Remote Terraform state

Validate the baseline:

```bash
cd terraform/environments/dev
terraform init -backend=false
terraform validate
```

## Scripts

- `scripts/local-run.ps1`: starts mock backend and router locally
- `scripts/setup-cluster.sh`: Terraform + Helm/kubectl orchestration stub
- `scripts/teardown-cluster.sh`: Terraform destroy orchestration stub

These scripts are intentionally honest about current readiness. Local development is runnable today. AWS and cluster scripts are a structured starting point for the next implementation pass.
