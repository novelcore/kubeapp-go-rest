# kubeapp-go-rest

A production-ready **Go REST API template** for KAOS KubeApp deployments.

## Features

- **Go 1.24** with `chi/v5` router
- **Opt-in Zitadel OIDC** JWT authentication (`ZITADEL_OIDC=true`)
- **Opt-in observability** — Prometheus metrics + OpenTelemetry traces (`OBSERVABILITY=true`)
- **Distroless container** — minimal attack surface (nonroot, read-only)
- **Full GitOps CI/CD** — dev → staging → prod via GitHub Actions + KubeCore
- **Local observability stack** — Loki + Prometheus + Tempo + Grafana via docker-compose

## Quick Start

```bash
# Run locally
cd app && go run ./cmd/api
# → http://localhost:8080/healthz

# Run with full observability stack
cp .env.example .env
docker-compose up -d
# → http://localhost:8080/healthz
# → http://localhost:3000 (Grafana)
```

## Project Structure

```
app/              — Go application (self-contained module)
etc/              — Observability configs (Prometheus, Grafana datasources, dashboards)
.github/          — CI/CD workflows
docker-compose.yaml
.env.example
```

## Configuration

Copy `.env.example` to `.env` and adjust as needed.

| Variable | Default | Description |
|---|---|---|
| `ZITADEL_OIDC` | `false` | Enable JWT auth on `/api/v1` routes |
| `OBSERVABILITY` | `false` | Enable Prometheus + OpenTelemetry |
| `SERVER_PORT` | `8080` | HTTP listen port |
| `LOG_LEVEL` | `info` | Log level |
| `LOG_FORMAT` | `text` | `text` or `json` |

## Endpoints

| Method | Path | Auth | Description |
|---|---|---|---|
| `GET` | `/healthz` | None | Liveness probe |
| `GET` | `/readyz` | None | Readiness probe |
| `GET` | `/metrics` | None | Prometheus metrics (when `OBSERVABILITY=true`) |
| `GET` | `/api/v1/hello` | Optional JWT | Example endpoint |

## Release Flow

```
feature/* → dev  →  rc/vX.Y.Z → main  →  prod
```

See [CLAUDE.md](CLAUDE.md) for full CI/CD documentation.

## Adding Endpoints

1. Add handler in `app/internal/handlers/`
2. Register in `app/internal/router/router.go` under `/api/v1`

## Authentication

To enable Zitadel OIDC authentication:

```bash
ZITADEL_OIDC=true
ZITADEL_DOMAIN=your-instance.zitadel.cloud
JWT_AUDIENCE=your-project-id
```

All `/api/v1` routes will then require a valid Zitadel Bearer token.
