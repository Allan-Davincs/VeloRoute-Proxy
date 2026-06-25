# VeloRoute

**High-performance reverse proxy and load balancer** with a real-time monitoring dashboard.

VeloRoute sits between clients and your backend servers. It distributes HTTP traffic using pluggable load balancing algorithms, monitors backend health, enforces per-IP rate limiting, and exposes Prometheus metrics plus a live React dashboard.

---

## Table of Contents

- [The Idea](#the-idea)
- [What Has Been Built](#what-has-been-built)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Load Balancing Algorithms](#load-balancing-algorithms)
- [Admin API Reference](#admin-api-reference)
- [Prometheus Metrics](#prometheus-metrics)
- [Dashboard Guide](#dashboard-guide)
- [Development](#development)
- [Docker Deployment](#docker-deployment)
- [Roadmap](#roadmap)

---

## The Idea

Modern applications need more than a simple reverse proxy. They need:

1. **Intelligent traffic distribution** — route requests across multiple backend instances
2. **Automatic failover** — detect and remove unhealthy backends without manual intervention
3. **Protection** — rate limit abusive clients before they reach your servers
4. **Observability** — metrics, structured logs, and a real-time dashboard for operators

VeloRoute was designed to solve all four problems in a single, lightweight Go binary with an optional React dashboard. It uses only the Go standard library for HTTP (no framework overhead), stores all state in memory (no database dependency), and streams access logs to the dashboard via Server-Sent Events (SSE).

### Design Principles

- **Single binary deployment** — compile once, run anywhere
- **Stdlib-first** — `net/http` + `httputil.ReverseProxy`, no external web framework
- **Thread-safe by default** — `sync.Mutex`, `sync.RWMutex`, and `sync/atomic` for all shared state
- **Graceful shutdown** — drain in-flight requests on SIGTERM/SIGINT (30s timeout)
- **SSE over WebSockets** — simpler, HTTP/1.1 compatible log streaming

---

## What Has Been Built

### Backend (Go)

| Component | Status | Description |
|-----------|--------|-------------|
| Config loader | Done | YAML configuration with validation |
| Round Robin balancer | Done | Even distribution across alive backends |
| Weighted Round Robin | Done | Proportional distribution by weight |
| Least Connections | Done | Routes to backend with fewest active connections |
| IP Hash | Done | Sticky sessions via FNV-1a hash |
| Rate limiter | Done | Token bucket per client IP with stale cleanup |
| Health checker | Done | Background goroutines with configurable interval/timeout |
| Reverse proxy | Done | `httputil.ReverseProxy` with forwarded headers |
| Access logger | Done | Structured JSON logs via `slog` |
| Prometheus metrics | Done | Counters, histograms, gauges, summaries on `:9091` |
| Admin REST API | Done | Runtime backend management + metrics snapshot |
| SSE log stream | Done | Real-time access log feed for dashboard |
| Graceful shutdown | Done | 30-second drain on signal |

### Frontend (React)

| Component | Status | Description |
|-----------|--------|-------------|
| Metrics cards | Done | Total requests, req/sec, error rate, P95 latency |
| Requests chart | Done | 60-second rolling line chart (Recharts) |
| Latency chart | Done | P50/P95/P99 percentile lines |
| Backend table | Done | Live status with expandable request bars |
| Access log feed | Done | SSE stream with auto-scroll and badges |
| Error boundary | Done | Catches render errors gracefully |
| Design system | Done | Dark DevOps theme per UI/UX Pro Max guidelines |

### Infrastructure

| Component | Status | Description |
|-----------|--------|-------------|
| Docker Compose | Done | Full stack with 3 echo backends + Prometheus |
| Makefile | Done | `dev`, `test`, `build`, `lint`, `run-*` targets |
| Prometheus config | Done | Scrapes VeloRoute metrics endpoint |

---

## Architecture

### System Overview

```
                    ┌─────────────────────────────────────────┐
                    │         Client HTTP Requests            │
                    └────────────────────┬────────────────────┘
                                         │
                                         ▼
                              ┌──────────────────┐
                              │  Proxy :8080     │
                              │  (proxy.go)      │
                              └────────┬─────────┘
                    ┌──────────────────┼──────────────────┐
                    ▼                  ▼                  ▼
            ┌─────────────┐   ┌─────────────┐   ┌─────────────┐
            │ Rate Limiter│   │  Balancer   │   │Access Logger│
            │ (429 if hit)│   │ (4 algos)   │   │ (slog JSON) │
            └─────────────┘   └──────┬──────┘   └──────┬──────┘
                                     │                  │
                                     ▼                  ▼
                            ┌──────────────┐    ┌──────────────┐
                            │ Backend Pool │    │ SSE Channel  │
                            │ (health chk) │    │ /api/logs/   │
                            └──────────────┘    └──────┬───────┘
                                                         │
    ┌────────────────────────────────────────────────────┼────────────┐
    │                                                    │            │
    ▼                                                    ▼            ▼
┌─────────────┐                                  ┌─────────────┐  ┌──────────┐
│ Admin API   │◄── poll /api/metrics ────────────│  Dashboard  │  │Prometheus│
│ :9090       │◄── SSE /api/logs/stream ─────────│  :3000      │  │ :9091    │
└─────────────┘                                  └─────────────┘  └──────────┘
```

### Request Lifecycle

1. Client sends HTTP request to VeloRoute on port `8080`
2. Rate limiter checks the client IP token bucket → `429` if exceeded
3. Load balancer selects the next alive backend → `503` if none available
4. Request is proxied via `httputil.ReverseProxy` with `X-Forwarded-For`, `X-Real-IP`, `X-VeloRoute-Backend` headers
5. Response is returned to the client
6. Access log entry is written (JSON to stdout) and broadcast to SSE subscribers
7. Prometheus metrics are updated (counters, histograms)

### Port Map

| Port | Service | Purpose |
|------|---------|---------|
| `8080` | Proxy | Main reverse proxy — route client traffic here |
| `9090` | Admin API | REST API + SSE log stream for dashboard |
| `9091` | Metrics | Prometheus `/metrics` scrape endpoint |
| `3000` | Dashboard | React monitoring UI |
| `9092` | Prometheus | Prometheus UI (Docker only) |

---

## Tech Stack

| Layer | Technologies |
|-------|-------------|
| **Backend** | Go 1.22+, `net/http`, `httputil.ReverseProxy`, `log/slog`, Prometheus client, YAML config |
| **Frontend** | React 18, TypeScript, Vite, Tailwind CSS, Recharts, Lucide React |
| **Ops** | Docker Compose, Makefile, Prometheus |
| **Design** | UI/UX Pro Max design system — Real-Time Monitoring / Data-Dense Dashboard pattern |

---

## Project Structure

```
VeloRoute-Proxy/
├── backend/                        # Go reverse proxy core
│   ├── cmd/veloroute/main.go       # Entry point — wires all services
│   ├── internal/
│   │   ├── proxy/proxy.go          # Core reverse proxy handler
│   │   ├── balancer/               # Load balancing (4 algorithms + tests)
│   │   ├── health/checker.go       # Active health check goroutines
│   │   ├── ratelimit/limiter.go    # Token bucket per IP
│   │   ├── metrics/prometheus.go   # Prometheus metrics registry
│   │   ├── logger/access.go        # Structured JSON access logs + SSE pub/sub
│   │   ├── admin/api.go            # REST admin API + SSE endpoint
│   │   └── config/                 # YAML config loader
│   ├── config.yaml                 # Local development config
│   ├── config.docker.yaml          # Docker Compose config
│   ├── Dockerfile
│   └── go.mod
│
├── frontend/                       # React TypeScript dashboard
│   ├── src/
│   │   ├── app/                    # App entry + error boundary
│   │   ├── components/Dashboard/   # Dashboard, charts, table, log feed
│   │   ├── hooks/                  # useMetrics, useBackends, useLogStream
│   │   ├── lib/api.ts              # API client (native fetch)
│   │   └── types/index.ts          # Shared TypeScript types
│   ├── Dockerfile
│   └── package.json
│
├── design-system/                  # UI/UX Pro Max design tokens
│   ├── MASTER.md                   # Global design system
│   └── pages/dashboard.md          # Dashboard-specific overrides
│
├── docker-compose.yml              # Full stack local development
├── prometheus.yml                  # Prometheus scrape config
├── Makefile                        # dev, test, build, lint
└── README.md
```

---

## Getting Started

### Prerequisites

- **Go 1.22+** — for building/running the backend
- **Node.js 18+** — for the frontend dashboard
- **Docker & Docker Compose** — for full-stack deployment (optional)

### Quick Start (Docker)

```bash
git clone git@github.com:Allan-Davincs/VeloRoute-Proxy.git
cd VeloRoute-Proxy
make dev
```

This starts VeloRoute, 3 test backends, Prometheus, and the dashboard.

| URL | What |
|-----|------|
| http://localhost:8080 | Proxied traffic |
| http://localhost:9090/api/metrics | Metrics JSON |
| http://localhost:3000 | Dashboard |
| http://localhost:9092 | Prometheus UI |

Test the proxy:

```bash
curl http://localhost:8080
# Rotates between "Hello from Backend 1/2/3"
```

### Local Development (without Docker)

**Terminal 1 — Start test backends** (or use any HTTP servers on ports 8001-8003):

```bash
# Example with Python
python3 -m http.server 8001 &
python3 -m http.server 8002 &
python3 -m http.server 8003 &
```

**Terminal 2 — Start VeloRoute:**

```bash
make run-backend
```

**Terminal 3 — Start dashboard:**

```bash
cd frontend && npm install && npm run dev
```

Open http://localhost:3000 and generate traffic:

```bash
for i in $(seq 1 20); do curl -s http://localhost:8080 > /dev/null; done
```

### Build

```bash
make build
# Backend binary: backend/bin/veloroute
# Frontend static files: frontend/dist/
```

---

## Configuration

Configuration is loaded from a YAML file (default: `./config.yaml`).

```yaml
veloroute:
  listen_addr: ":8080"          # Proxy listen address
  admin_addr: ":9090"           # Admin API listen address
  metrics_addr: ":9091"         # Prometheus metrics address

  load_balancing:
    algorithm: "round_robin"    # round_robin | weighted_round_robin | least_connections | ip_hash

  health_check:
    enabled: true
    interval_seconds: 5         # Check interval
    timeout_seconds: 2          # Per-check timeout
    path: "/"                   # Health check endpoint on each backend

  rate_limit:
    enabled: true
    requests_per_second: 10     # Sustained rate per IP
    burst: 20                   # Burst allowance per IP

  backends:
    - url: "http://localhost:8001"
      weight: 1
      name: "backend-1"
    - url: "http://localhost:8002"
      weight: 1
      name: "backend-2"
    - url: "http://localhost:8003"
      weight: 2
      name: "backend-3"
```

### Environment Variables (Frontend)

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:9090` | Admin API base URL for dashboard |

### Runtime Configuration

The admin API allows changing backends and algorithms without restarting:

```bash
# Switch algorithm
curl -X PUT http://localhost:9090/api/config/algorithm \
  -H "Content-Type: application/json" \
  -d '{"algorithm": "least_connections"}'

# Add a backend
curl -X POST http://localhost:9090/api/backends \
  -H "Content-Type: application/json" \
  -d '{"url": "http://localhost:8004", "name": "backend-4", "weight": 1}'
```

---

## Load Balancing Algorithms

| Algorithm | Key | When to Use |
|-----------|-----|-------------|
| **Round Robin** | `round_robin` | Default. Even distribution when all backends are equal |
| **Weighted Round Robin** | `weighted_round_robin` | Backends have different capacities (higher weight = more traffic) |
| **Least Connections** | `least_connections` | Long-lived connections or variable request durations |
| **IP Hash** | `ip_hash` | Sticky sessions — same client IP always hits the same backend |

All algorithms skip backends marked as dead by the health checker.

---

## Admin API Reference

Base URL: `http://localhost:9090`

### Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/backends` | List all backends with status |
| `POST` | `/api/backends` | Add a backend at runtime |
| `DELETE` | `/api/backends/:url` | Remove a backend (URL base64-encoded) |
| `PUT` | `/api/backends/:url/weight` | Update backend weight |
| `GET` | `/api/config` | Current running configuration |
| `PUT` | `/api/config/algorithm` | Switch load balancing algorithm |
| `GET` | `/api/metrics` | JSON metrics snapshot for dashboard |
| `GET` | `/api/logs/stream` | SSE stream of access log entries |

### `GET /api/metrics` Response

```json
{
  "total_requests": 142983,
  "requests_per_second": 234.5,
  "error_rate_percent": 0.12,
  "active_connections": 47,
  "p50_latency_ms": 12.3,
  "p95_latency_ms": 89.1,
  "p99_latency_ms": 234.5,
  "uptime_seconds": 86400,
  "backends": [
    {
      "name": "backend-1",
      "url": "http://localhost:8001",
      "alive": true,
      "weight": 1,
      "active_connections": 12,
      "total_requests": 47000,
      "total_errors": 3,
      "last_health_check": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### SSE Log Stream

```
GET /api/logs/stream
Content-Type: text/event-stream

data: {"time":"...","client_ip":"192.168.1.1","method":"GET","path":"/","status":200,"duration_ms":12.3,"backend":"backend-1",...}
```

---

## Prometheus Metrics

Scrape endpoint: `http://localhost:9091/metrics`

| Metric | Type | Labels |
|--------|------|--------|
| `veloroute_requests_total` | Counter | backend, method, status_code |
| `veloroute_errors_total` | Counter | backend, error_type |
| `veloroute_request_duration_seconds` | Histogram | backend, method |
| `veloroute_active_connections` | Gauge | backend |
| `veloroute_backend_alive` | Gauge | backend, name |
| `veloroute_request_duration_summary` | Summary | backend |

---

## Dashboard Guide

The React dashboard provides real-time visibility into VeloRoute operations.

### Layout

```
┌─────────────────────────────────────────────────────┐
│  VeloRoute        Algorithm: Round Robin    Online  │  Header
├──────────┬──────────┬──────────┬───────────────────┤
│  Total   │ Req/sec  │ Error %  │  P95 Latency      │  Metrics Cards
├──────────┴──────────┴──────────┴───────────────────┤
│  [Requests/sec — Line Chart (last 60s)]              │
├────────────────────────┬────────────────────────────┤
│ P50/P95/P99 Latency    │  Backend Servers Table      │
├────────────────────────┴────────────────────────────┤
│  Live Access Log Feed (SSE, auto-scroll)             │
└─────────────────────────────────────────────────────┘
```

### Components

- **MetricsCards** — Polls `/api/metrics` every 2s. Color-coded thresholds for error rate and latency.
- **RequestsChart** — 60-point rolling window of requests per second.
- **LatencyChart** — P50 (green), P95 (amber), P99 (red) latency lines.
- **BackendTable** — Click a row to expand a mini request distribution bar. Alive backends show a pulsing green dot.
- **AccessLogFeed** — SSE connection with method/status badges, monospace font, fade-in animation.

### Design System

The dashboard follows the **Real-Time Monitoring** pattern from the [UI/UX Pro Max](https://github.com/nextlevelbuilder/ui-ux-pro-max-skill) design skill:

- Dark theme optimized for operations (`#0f1117` background)
- Indigo primary accent (`#6366f1`)
- Skeleton loaders on initial load (no spinners)
- Lucide React icons throughout
- `prefers-reduced-motion` respected

Design tokens are documented in `design-system/MASTER.md` and `design-system/pages/dashboard.md`.

---

## Development

### Commands

```bash
make dev            # Docker Compose full stack
make test           # Go tests (-race) + frontend lint
make build          # Compile backend binary + frontend bundle
make lint           # go vet + ESLint
make run-backend    # Run Go proxy locally
make run-frontend   # Run Vite dev server
```

### Coding Conventions

**Go:**
- No external web framework — stdlib `net/http` only
- All concurrent access via `sync.Mutex`, `sync.RWMutex`, or `sync/atomic`
- Godoc comments on all exported functions
- Table-driven tests for balancers and rate limiter
- No global mutable state — dependencies passed via constructors

**React:**
- Strict TypeScript (no `any`)
- Functional components + hooks only
- API calls in custom hooks, not components
- Tailwind CSS only (no inline styles)
- SSE connections cleaned up on unmount

---

## Docker Deployment

`docker-compose.yml` defines five services:

| Service | Image/Build | Ports |
|---------|-------------|-------|
| `veloroute` | `./backend` | 8080, 9090, 9091 |
| `backend1-3` | `hashicorp/http-echo` | internal |
| `prometheus` | `prom/prometheus` | 9092 |
| `dashboard` | `./frontend` | 3000 |

Docker uses `backend/config.docker.yaml` with internal service hostnames (`backend1:8001`, etc.).

---

## Roadmap

- [ ] TLS termination
- [ ] Admin API authentication
- [ ] Config hot-reload from file watch
- [ ] Weighted algorithm switch at runtime (rebuild balancer)
- [ ] Request tracing with OpenTelemetry
- [ ] WebSocket alternative for log streaming

---

## License

MIT
