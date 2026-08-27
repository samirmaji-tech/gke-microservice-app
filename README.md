# gke-microservice-app

A small Go REST service built to be operated properly on Kubernetes: health and readiness probes, custom Prometheus metrics, structured JSON logs, OpenTelemetry tracing, and graceful shutdown — shipped as a minimal non-root container.

Deployment manifests live in the companion GitOps repository, [gke-gitops-infra](https://github.com/samirmaji-tech/gke-gitops-infra). This repo contains only the application and its CI pipeline.

---

## Endpoints

| Path | Purpose |
|---|---|
| `GET /` | Service identity and timestamp |
| `GET /api/v1/items/{id}` | Example REST resource |
| `GET /healthz` | Liveness — is the process alive? |
| `GET /readyz` | Readiness — can it serve traffic right now? |
| `GET /metrics` | Prometheus scrape endpoint |

---

## Running locally

```bash
go mod tidy
go run ./cmd/server

curl localhost:8080/
curl localhost:8080/healthz
curl localhost:8080/readyz          # 503 during warmup, then 200
curl localhost:8080/metrics
```

Build the container:

```bash
docker build -f build/Dockerfile -t gke-microservice:local .
docker run --rm -p 8080:8080 gke-microservice:local
```

Configuration is environment-driven: `HTTP_ADDR`, `OTEL_SERVICE_NAME`, `OTEL_EXPORTER_OTLP_ENDPOINT`.

---

## Observability design

Three custom collectors map to the golden signals, all updated in a single middleware:

- `http_requests_total` — counter labelled by method, route, status (traffic and errors)
- `http_request_duration_seconds` — histogram (latency; p50/p95/p99 derived in Grafana)
- `http_in_flight_requests` — gauge (saturation)

Metrics are labelled by the **route pattern** (`/api/v1/items/{id}`) rather than the raw path. Labelling by raw path would create a new time series per unique ID and blow up Prometheus cardinality.

Logs are structured JSON written to stdout via `slog` — no log files, no rotation, nothing for the container to manage. Tracing is optional and activates only when an OTLP endpoint is configured, so the service still runs cleanly on a laptop.

---

## Probes and zero-downtime behaviour

Liveness and readiness answer different questions, and conflating them causes outages. `/healthz` failing means the process is wedged and should be restarted. `/readyz` failing means "don't send me traffic right now" — the pod stays alive but leaves the load-balancer pool.

On `SIGTERM` the service marks itself unready *first*, letting the load balancer drain it, then finishes in-flight requests before exiting. Combined with a `maxUnavailable: 0` rolling update, this is what makes deployments zero-downtime.

---

## Container

Two-stage build: the Go image compiles a static, stripped binary with `CGO_ENABLED=0`; the runtime stage is `gcr.io/distroless/static:nonroot`. The result has no shell, no package manager, and runs as UID 65532 — a few MB with minimal attack surface, which also means faster pulls when the autoscaler adds pods.

---

## CI pipeline

`.github/workflows/ci-cd.yaml` runs on every push to `main`:

1. `go vet`, `golangci-lint`, and race-enabled tests
2. Trivy filesystem/dependency scan — fails on HIGH/CRITICAL
3. Multi-arch build (`linux/amd64`, `linux/arm64`) pushed to Artifact Registry
4. Trivy scan of the built image
5. Commit the new image tag into the GitOps repo

Authentication to GCP uses Workload Identity Federation, so no service-account key is stored in GitHub. The pipeline never touches the cluster — it only changes Git, and ArgoCD does the deploying.

---

## Layout

```
cmd/server/          entrypoint, wiring, graceful shutdown
internal/handlers/   API + health handlers
internal/metrics/    Prometheus collectors
internal/middleware/ per-request metrics and structured logging
internal/telemetry/  OpenTelemetry tracer provider
build/Dockerfile     multi-stage, distroless, non-root
```
