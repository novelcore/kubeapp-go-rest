# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Repo Is

A production-ready **Go REST API template** for the KAOS platform. The app lives under `app/` (a self-contained Go module). All development commands run from inside `app/`.

## Commands

```bash
# Setup
cd app

# Run locally (port 8080)
go run ./cmd/api

# Or via Makefile
cd app && make run

# Run full observability stack (API + Loki + Prometheus + Tempo + Grafana)
docker-compose up -d

# Tests
cd app && go test ./...
cd app && go test ./... -v                           # verbose
cd app && go test -coverprofile=coverage.out ./...   # with coverage
cd app && go tool cover -html=coverage.out           # view coverage

# Build binary
cd app && make build   # produces ./bin/server

# Lint
cd app && golangci-lint run
```

## Architecture

```
app/
├── cmd/api/
│   └── main.go                   # Entry point — load config, create server, handle SIGTERM
├── internal/
│   ├── version/version.go        # const VERSION = "0.1.0" — source of truth for semver
│   ├── config/config.go          # AppConfig (env vars) with feature flag toggles
│   ├── auth/
│   │   ├── validator.go          # JWTValidator — validates Zitadel-issued RS256 tokens
│   │   ├── jwks.go               # JWKSProvider — fetches and caches JWKS with 1h TTL
│   │   └── claims.go             # ZitadelClaims, AuthContext, ParseRoles()
│   ├── middleware/
│   │   ├── logger.go             # StructuredLogger — logrus request logging with chi
│   │   ├── metrics.go            # Metrics() — Prometheus histogram + counter (opt-in)
│   │   └── jwt.go                # JWTAuth() / OptionalJWTAuth() (opt-in via ZITADEL_OIDC)
│   ├── handlers/health.go        # /healthz, /readyz, /api/v1/hello
│   ├── router/router.go          # chi/v5 router — CORS, middleware, route registration
│   └── server/server.go          # HTTP server + graceful shutdown
├── Dockerfile                    # Multi-stage: golang:1.24-alpine → distroless nonroot
├── go.mod                        # module github.com/novelcore/kubeapp-go-rest
└── Makefile
```

### Key Patterns

**Adding a new endpoint:**
1. Add handler in `internal/handlers/` (or a new file for a feature group)
2. Register in `internal/router/router.go` under `r.Route("/api/v1", ...)`

**Authentication (opt-in):**
Set `ZITADEL_OIDC=true` and configure `ZITADEL_DOMAIN`, `JWT_ISSUER`, `JWT_AUDIENCE`, `JWT_JWKS_URL`.
When enabled, `JWTAuth` middleware wraps all `/api/v1` routes and validates Bearer tokens.
When disabled (default), `/api/v1` routes are unauthenticated.

**Observability (opt-in):**
Set `OBSERVABILITY=true` to activate:
- Prometheus metrics at `/metrics`
- `Metrics()` middleware on `/api/v1` routes
- OpenTelemetry traces to Tempo via `OTLP_GRPC_ENDPOINT`

**Version management:**
`internal/version/version.go` is the single source of truth. CI reads the Docker image tag (set by the workflow). To bump manually, edit `VERSION` in that file.

## Feature Flags

| Variable | Default | Description |
|---|---|---|
| `ZITADEL_OIDC` | `false` | Enable Zitadel JWT authentication on `/api/v1` routes |
| `OBSERVABILITY` | `false` | Enable Prometheus metrics and OpenTelemetry tracing |

## Environment Variables

```bash
SERVER_PORT=8080
LOG_LEVEL=info          # debug, info, warn, error
LOG_FORMAT=text         # text for dev, json for prod

# Only when ZITADEL_OIDC=true
ZITADEL_DOMAIN=zitadel.example.com
JWT_ISSUER=https://zitadel.example.com
JWT_AUDIENCE=<project_id>
JWT_JWKS_URL=https://zitadel.example.com/oauth/v2/keys

# Only when OBSERVABILITY=true
OTLP_GRPC_ENDPOINT=http://tempo:4317
```

## Release & GitOps

All CI/CD is in `.github/workflows/`. The release flow is:

```
feature/* → PR → dev  →  dev-release.yaml  →  image: dev-vX.Y.Z-{ts}-{sha}
                                              →  dispatch: dev-environment-update → GitOps repo
[add create-rc🚀 label or run manual-create-rc workflow]
rc/vX.Y.Z → PR → main  →  rc-release.yaml   →  image: vX.Y.Z-rcN → staging
[merge RC PR]           →  prod-release.yaml →  image: vX.Y.Z + latest → prod + GitHub Release
hotfix/* → PR → main   →  hotfix-release.yaml → skips staging, direct to prod
```

GitOps repo (`KUBECORE_KUBEPROJECT_REPO`) receives `repository_dispatch` events. It updates `kubeapps/{app_name}/overlays/{env}/` and ArgoCD syncs the cluster.

**Required repo secrets/variables:** `KUBECORE_REGISTRY`, `KUBECORE_REGISTRY_UNAME`, `KUBECORE_REGISTRY_PWD`, `KUBECORE_KUBEPROJECT_REPO`, `KUBECORE_APP_ID`, `KUBECORE_APP_PKEY`.

## Claude Code Entrypoints

### `/release-manager` skill

Use `/release-manager` for release guidance. The skill handles status checks, tag formats, and pipeline explanations; delegates complex multi-step operations to the **release-manager sub-agent**.

### `release-manager` agent (`.cursor/agents/release-manager.md`)

Specialized sub-agent for autonomous release orchestration. Knows overlay paths, dispatch event types, and end-to-end verification with `gh` and `kubectl`.
