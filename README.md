# gke-microservice-app

Go microservice + CI for the Production-grade GitOps on GKE project.

- REST API with `/healthz`, `/readyz`, `/metrics`
- Custom Prometheus metrics + structured JSON logs (stdout) + OpenTelemetry traces
- Multi-stage, non-root, distroless container
- GitHub Actions: lint → test → Trivy scan → multi-arch build → push to Artifact Registry → bump image tag in `gke-gitops-infra`

Deployment manifests live in the **gke-gitops-infra** repo (single source of truth for ArgoCD).
