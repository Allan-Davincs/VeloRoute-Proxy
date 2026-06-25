# VeloRoute – Master Cursor Agent Prompt
**High-Performance Reverse Proxy & Load Balancer**
**Stack: Go (Backend) · TypeScript/React (Frontend Dashboard)**

---

## 🎯 PROJECT OVERVIEW

Build **VeloRoute** — a production-ready, high-performance reverse proxy and load balancer written in Go, with a real-time web dashboard built in TypeScript/React.

VeloRoute sits between the internet and backend servers. It:
- Receives all incoming HTTP requests from clients
- Distributes traffic across multiple backend servers using pluggable load balancing algorithms
- Continuously monitors backend health and removes dead servers automatically
- Enforces rate limiting per IP
- Exposes Prometheus metrics, structured access logs, and a real-time React dashboard

---

## 🗂️ MONOREPO FOLDER STRUCTURE

```
veloroute/
├── backend/                        # Go reverse proxy core
│   ├── cmd/
│   │   └── veloroute/
│   │       └── main.go             # Entry point
│   ├── internal/
│   │   ├── proxy/
│   │   │   ├── proxy.go            # Core reverse proxy handler
│   │   │   └── proxy_test.go
│   │   ├── balancer/
│   │   │   ├── balancer.go         # Balancer interface
│   │   │   ├── roundrobin.go       # Round Robin implementation
│   │   │   ├── weighted.go         # Weighted Round Robin
│   │   │   ├── leastconn.go        # Least Connections
│   │   │   ├── iphash.go           # IP Hash (sticky sessions)
│   │   │   └── balancer_test.go
│   │   ├── health/
│   │   │   ├── checker.go          # Active health check goroutine
│   │   │   └── checker_test.go
│   │   ├── ratelimit/
│   │   │   ├── limiter.go          # Token bucket per IP
│   │   │   └── limiter_test.go
│   │   ├── metrics/
│   │   │   ├── prometheus.go       # Prometheus metrics registry
│   │   │   └── collector.go        # Custom collectors
│   │   ├── logger/
│   │   │   └── access.go           # Structured JSON access logs
│   │   ├── admin/
│   │   │   └── api.go              # REST admin API (add/remove backends at runtime)
│   │   └── config/
│   │       ├── config.go           # Config loader
│   │       └── schema.go           # Config structs
│   ├── config.yaml                 # Default configuration file
│   ├── go.mod
│   └── go.sum
│
├── frontend/                       # React TypeScript Dashboard
│   ├── src/
│   │   ├── app/
│   │   │   ├── App.tsx
│   │   │   └── main.tsx
│   │   ├── components/
│   │   │   ├── Dashboard/
│   │   │   │   ├── Dashboard.tsx       # Main layout
│   │   │   │   ├── MetricsCards.tsx    # Total requests, error rate, latency
│   │   │   │   ├── BackendTable.tsx    # Live backend servers table
│   │   │   │   ├── RequestsChart.tsx   # Requests/sec line chart
│   │   │   │   ├── LatencyChart.tsx    # P50/P95/P99 latency chart
│   │   │   │   └── AccessLogFeed.tsx   # Live scrolling log feed
│   │   │   └── ui/                     # Shared UI primitives
│   │   ├── hooks/
│   │   │   ├── useMetrics.ts           # Polling /api/metrics
│   │   │   ├── useBackends.ts          # Polling /api/backends
│   │   │   └── useLogStream.ts         # SSE log stream
│   │   ├── lib/
│   │   │   └── api.ts                  # API client
│   │   └── types/
│   │       └── index.ts                # Shared TypeScript types
│   ├── package.json
│   ├── tsconfig.json
│   ├── vite.config.ts
│   └── tailwind.config.ts
│
├── docker-compose.yml              # Full stack local dev
├── Makefile                        # make dev, make test, make build
└── README.md
```

---

## ⚙️ BACKEND — Go Implementation

### Tech Stack
- **Language**: Go 1.22+
- **HTTP**: `net/http` + `httputil.ReverseProxy` (stdlib)
- **Metrics**: `github.com/prometheus/client_golang`
- **Config**: `gopkg.in/yaml.v3`
- **Rate Limiting**: Token bucket — implement manually using `sync.Mutex` + `time.Ticker`
- **Logging**: `log/slog` (structured JSON logs, stdlib Go 1.21+)
- **Testing**: `testing` package + `net/http/httptest`

### 1. Configuration Schema (`config.yaml`)

```yaml
veloroute:
  listen_addr: ":8080"          # Port VeloRoute listens on
  admin_addr: ":9090"           # Admin REST API port
  metrics_addr: ":9091"         # Prometheus /metrics port
  dashboard_addr: ":3000"       # React dashboard dev server (proxied in prod)

  load_balancing:
    algorithm: "round_robin"    # round_robin | weighted_round_robin | least_connections | ip_hash

  health_check:
    enabled: true
    interval_seconds: 5         # Check every 5 seconds
    timeout_seconds: 2          # Mark dead if no response in 2s
    path: "/health"             # Endpoint to ping on each backend

  rate_limit:
    enabled: true
    requests_per_second: 10     # Max requests per IP per second
    burst: 20                   # Allow burst of 20

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

---

### 2. Load Balancer Interface

```go
// internal/balancer/balancer.go
package balancer

type Backend struct {
    URL             string
    Name            string
    Weight          int
    ActiveConns     int64       // atomic
    IsAlive         bool        // atomic via sync/atomic
    TotalRequests   int64       // atomic
    TotalErrors     int64       // atomic
}

type Balancer interface {
    Next(clientIP string) (*Backend, error)
    AddBackend(b *Backend)
    RemoveBackend(url string)
    GetBackends() []*Backend
    MarkAlive(url string, alive bool)
}
```

**Implement all 4 algorithms:**

#### Round Robin (`roundrobin.go`)
```go
// Use sync/atomic counter. Cycle through alive backends only.
// Thread-safe: atomic increment of counter, modulo len(aliveBackends)
```

#### Weighted Round Robin (`weighted.go`)
```go
// Expand backends by weight into a slot slice at initialization.
// [B1, B2, B2, B3, B3, B3] for weights [1,2,3]
// Atomic counter cycles through slots.
```

#### Least Connections (`leastconn.go`)
```go
// On each Next() call, iterate alive backends and pick the one
// with lowest ActiveConns (atomic read). Increment on pick, decrement
// via defer in the proxy handler after response is complete.
```

#### IP Hash (`iphash.go`)
```go
// Hash clientIP using FNV-1a. Modulo len(aliveBackends).
// Same IP always routes to same backend (sticky sessions).
```

---

### 3. Core Proxy Handler (`internal/proxy/proxy.go`)

```go
// Use httputil.ReverseProxy as the underlying transport.
// 
// For EACH incoming request:
// 1. Check rate limit → 429 if exceeded
// 2. Call balancer.Next(clientIP) → 503 if no backends available
// 3. Clone request, set correct Host/X-Forwarded-For headers
// 4. If LeastConn balancer: increment ActiveConns, defer decrement
// 5. Record start time
// 6. Proxy request using httputil.ReverseProxy
// 7. After response: write access log entry (structured JSON via slog)
// 8. Update Prometheus counters and histograms

// Required Headers to set on forwarded request:
// X-Forwarded-For: client_ip
// X-Forwarded-Host: original host
// X-Real-IP: client_ip
// X-VeloRoute-Backend: backend name (for debugging)
```

---

### 4. Health Checker (`internal/health/checker.go`)

```go
// Start a background goroutine per backend on startup.
// Every interval_seconds: send HTTP GET to backend_url + health_check.path
// Timeout: health_check.timeout_seconds
//
// If response: 200-399 → mark alive = true
// If timeout or non-2xx/3xx → mark alive = false, log warning via slog
//
// On alive→dead transition: log "Backend %s is DOWN" at WARN level
// On dead→alive transition: log "Backend %s is UP" at INFO level
//
// Use context.WithTimeout for each health check request.
// Use sync/atomic for IsAlive flag on Backend struct.
```

---

### 5. Rate Limiter (`internal/ratelimit/limiter.go`)

```go
// Token Bucket implementation per IP address.
// 
// Data structure: map[string]*tokenBucket protected by sync.RWMutex
// Each bucket: tokens float64, lastRefill time.Time, mu sync.Mutex
//
// On each request:
// 1. Get or create bucket for clientIP
// 2. Refill tokens based on elapsed time since lastRefill
//    newTokens = elapsed.Seconds() * requestsPerSecond
//    tokens = min(tokens + newTokens, burst)
// 3. If tokens >= 1.0: subtract 1.0, allow request
// 4. Else: return 429 Too Many Requests with Retry-After header
//
// Cleanup: background goroutine every 60s removes stale IPs
// (buckets not accessed in last 5 minutes)
```

---

### 6. Prometheus Metrics (`internal/metrics/prometheus.go`)

Register and expose these metrics on `:9091/metrics`:

```go
// COUNTERS
veloroute_requests_total              // labels: backend, method, status_code
veloroute_errors_total                // labels: backend, error_type (timeout|refused|rate_limited)

// HISTOGRAMS
veloroute_request_duration_seconds    // labels: backend, method
// Buckets: 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0

// GAUGES
veloroute_active_connections          // labels: backend
veloroute_backend_alive               // labels: backend, name (0=dead, 1=alive)

// SUMMARY (for accurate percentiles)
veloroute_request_duration_summary    // labels: backend
// Objectives: 0.5, 0.9, 0.95, 0.99
```

---

### 7. Access Logger (`internal/logger/access.go`)

```go
// Use log/slog with JSON handler writing to os.Stdout
// Write one log line per request after response is sent.
//
// Log fields (ALL required):
// {
//   "time":           "2024-01-15T10:30:00Z",     // RFC3339
//   "level":          "INFO",
//   "client_ip":      "192.168.1.100",
//   "method":         "GET",
//   "path":           "/api/users",
//   "query":          "page=1&limit=10",
//   "status":         200,
//   "duration_ms":    45.23,
//   "bytes_sent":     1024,
//   "backend":        "backend-2",
//   "backend_url":    "http://localhost:8002",
//   "user_agent":     "Mozilla/5.0...",
//   "request_id":     "abc123"                     // UUID per request
// }
```

---

### 8. Admin REST API (`internal/admin/api.go`)

Serve on `:9090`. All responses are `application/json`.

```
GET    /api/backends              → List all backends with status, connections, request count
POST   /api/backends              → Add a new backend at runtime
DELETE /api/backends/:url         → Remove a backend at runtime (base64-encoded URL as param)
PUT    /api/backends/:url/weight  → Update backend weight
GET    /api/config                → Return current running config
PUT    /api/config/algorithm      → Switch load balancing algorithm at runtime (no restart)
GET    /api/metrics               → JSON metrics snapshot (for dashboard polling)
GET    /api/logs/stream           → SSE endpoint — stream access log entries as they happen
```

**`GET /api/metrics` response shape:**
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

**`GET /api/logs/stream` — SSE Format:**
```
data: {"time":"...","client_ip":"...","method":"GET","path":"/","status":200,"duration_ms":12.3,"backend":"backend-1"}

data: {"time":"...","client_ip":"...","method":"POST","path":"/api/data","status":201,"duration_ms":45.1,"backend":"backend-2"}
```

---

### 9. `main.go` Startup Sequence

```go
// 1. Load config.yaml (path from --config flag, default: ./config.yaml)
// 2. Initialize logger (slog JSON to stdout)
// 3. Initialize Prometheus metrics registry
// 4. Initialize rate limiter
// 5. Initialize chosen balancer from config (algorithm field)
// 6. Register all backends from config into balancer
// 7. Start health checker goroutines for each backend
// 8. Start admin API server on admin_addr (goroutine)
// 9. Start Prometheus metrics server on metrics_addr (goroutine)
// 10. Start main proxy server on listen_addr (blocking)
// 11. Handle OS signals (SIGTERM/SIGINT) — graceful shutdown:
//     - Stop accepting new connections
//     - Wait for in-flight requests (max 30s drain)
//     - Shutdown all servers
//     - Log "VeloRoute shutdown complete"
```

---

## 🖥️ FRONTEND — TypeScript/React Dashboard

### Tech Stack
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Charts**: Recharts
- **Icons**: Lucide React
- **HTTP Client**: Native `fetch` (no axios)
- **State**: React hooks only (`useState`, `useEffect`, `useRef`) — no external state library
- **Formatting**: Prettier + ESLint

### Dashboard Layout & Components

#### `Dashboard.tsx` — Main Layout
```
┌─────────────────────────────────────────────────────┐
│  🟢 VeloRoute        Algorithm: Round Robin    ⚙️    │  ← Header
├──────────┬──────────┬──────────┬───────────────────┤
│  Total   │ Req/sec  │ Error %  │  Avg Latency      │  ← MetricsCards
│  142,983 │  234.5   │  0.12%   │  45ms             │
├──────────┴──────────┴──────────┴───────────────────┤
│                                                      │
│  [Requests/sec — Line Chart (last 60s)]              │  ← RequestsChart
│                                                      │
├────────────────────────┬────────────────────────────┤
│ P50/P95/P99 Latency    │  Backend Servers Table      │  ← Split row
│ [Bar/Line Chart]       │  [BackendTable]             │
├────────────────────────┴────────────────────────────┤
│  Live Access Log Feed (last 50 entries, auto-scroll) │  ← AccessLogFeed
└─────────────────────────────────────────────────────┘
```

**Color Theme:**
- Background: `#0f1117` (near black)
- Card background: `#1a1d27`
- Border: `#2a2d3e`
- Primary accent: `#6366f1` (indigo)
- Success/Alive: `#22c55e` (green)
- Error/Dead: `#ef4444` (red)
- Warning: `#f59e0b` (amber)
- Text primary: `#f1f5f9`
- Text muted: `#64748b`

---

#### `MetricsCards.tsx`
```tsx
// 4 cards in a responsive grid (2 col mobile, 4 col desktop)
// Each card has: icon (Lucide), label, value, and a subtle trend indicator
// Cards:
// 1. Total Requests — Activity icon — formatted with commas
// 2. Requests/sec — Zap icon — 1 decimal place
// 3. Error Rate — AlertTriangle icon — red if > 1%, green if < 0.5%
// 4. P95 Latency — Timer icon — in ms, amber if > 200ms, red if > 500ms
```

#### `BackendTable.tsx`
```tsx
// Table columns: Name | URL | Status | Weight | Active Conns | Total Reqs | Errors | Last Check
// Status: green dot "Alive" or red dot "Dead" — animated pulse on "Alive"
// Refresh: poll GET /api/metrics every 2 seconds
// Show: loading skeleton on initial load
// On click of a row: expand to show mini request history bar chart for that backend
```

#### `RequestsChart.tsx`
```tsx
// Recharts LineChart — shows requests/sec over last 60 data points
// X axis: time labels (last 60 seconds)
// Y axis: requests per second
// Line color: indigo (#6366f1)
// Tooltip: exact value + timestamp
// Data: maintained in local state array, push new point every 1 second from /api/metrics poll
// Animate: smooth transitions between data points
```

#### `LatencyChart.tsx`
```tsx
// Recharts LineChart with 3 lines: P50, P95, P99
// Colors: P50=green, P95=amber, P99=red
// Same 60-point rolling window as RequestsChart
// Legend at top
```

#### `AccessLogFeed.tsx`
```tsx
// Connect to GET /api/logs/stream (SSE)
// Maintain array of last 100 log entries in state
// Auto-scroll to bottom on new entry (unless user has scrolled up — detect this)
// Each row: timestamp | method badge | path | status badge | duration | backend | client_ip
// Status badge colors: 2xx=green, 3xx=blue, 4xx=amber, 5xx=red
// Method badge colors: GET=indigo, POST=green, PUT=amber, DELETE=red
// Font: monospace for log feed
// Add subtle fade-in animation on new entries
```

---

### TypeScript Types (`src/types/index.ts`)

```typescript
export interface Backend {
  name: string
  url: string
  alive: boolean
  weight: number
  active_connections: number
  total_requests: number
  total_errors: number
  last_health_check: string
}

export interface MetricsSnapshot {
  total_requests: number
  requests_per_second: number
  error_rate_percent: number
  active_connections: number
  p50_latency_ms: number
  p95_latency_ms: number
  p99_latency_ms: number
  uptime_seconds: number
  backends: Backend[]
}

export interface LogEntry {
  time: string
  client_ip: string
  method: string
  path: string
  query: string
  status: number
  duration_ms: number
  bytes_sent: number
  backend: string
  backend_url: string
  request_id: string
}

export type Algorithm = 'round_robin' | 'weighted_round_robin' | 'least_connections' | 'ip_hash'
```

---

### Custom Hooks

#### `useMetrics.ts`
```typescript
// Poll GET http://localhost:9090/api/metrics every 2000ms
// Return: { data: MetricsSnapshot | null, loading: boolean, error: string | null }
// On error: keep last successful data, set error message
// Cleanup: clear interval on unmount
```

#### `useLogStream.ts`
```typescript
// Connect to SSE: GET http://localhost:9090/api/logs/stream
// Parse each SSE 'data' event as LogEntry JSON
// Maintain rolling array of last 100 entries (newest at bottom)
// Return: { logs: LogEntry[], connected: boolean }
// Auto-reconnect on disconnect: exponential backoff (1s, 2s, 4s, max 30s)
// Cleanup: close EventSource on unmount
```

---

## 🐳 docker-compose.yml

```yaml
version: '3.9'
services:

  veloroute:
    build: ./backend
    ports:
      - "8080:8080"   # Proxy
      - "9090:9090"   # Admin API
      - "9091:9091"   # Prometheus metrics
    volumes:
      - ./backend/config.yaml:/app/config.yaml
    depends_on:
      - backend1
      - backend2
      - backend3

  backend1:
    image: hashicorp/http-echo
    command: ["-text=Hello from Backend 1", "-listen=:8001"]
    ports: ["8001:8001"]

  backend2:
    image: hashicorp/http-echo
    command: ["-text=Hello from Backend 2", "-listen=:8001"]
    ports: ["8002:8001"]

  backend3:
    image: hashicorp/http-echo
    command: ["-text=Hello from Backend 3", "-listen=:8001"]
    ports: ["8003:8001"]

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml
    ports: ["9092:9090"]

  dashboard:
    build: ./frontend
    ports: ["3000:3000"]
    environment:
      - VITE_API_BASE_URL=http://localhost:9090
```

---

## 📋 Makefile

```makefile
.PHONY: dev test build lint

dev:
	docker-compose up --build

test:
	cd backend && go test ./... -v -race
	cd frontend && npm run test

build:
	cd backend && go build -o bin/veloroute ./cmd/veloroute
	cd frontend && npm run build

lint:
	cd backend && golangci-lint run
	cd frontend && npm run lint

run-backend:
	cd backend && go run ./cmd/veloroute --config config.yaml

run-frontend:
	cd frontend && npm run dev
```

---

## 🔐 IMPLEMENTATION RULES (Lazima Zifuatwe)

### Go Rules
1. **No external web framework** — use `net/http` stdlib only for the proxy
2. **All concurrent access** to shared data must use `sync.Mutex`, `sync.RWMutex`, or `sync/atomic`
3. **Every exported function** must have a godoc comment
4. **Error handling**: never ignore errors — log with slog and return appropriate HTTP status
5. **Graceful shutdown**: context cancellation must propagate to all goroutines
6. **Tests**: write table-driven tests for all balancer algorithms and the rate limiter
7. **No global variables** — pass dependencies via struct constructors
8. **Interfaces first**: define the `Balancer` interface before implementing any algorithm
9. Metrics **must** be updated atomically — use Prometheus client's built-in thread safety

### TypeScript/React Rules
1. **No `any` types** — strict TypeScript throughout
2. **No class components** — functional components + hooks only
3. **No `useEffect` for derived state** — compute from existing state directly
4. **All API calls** in custom hooks, not in components
5. **No inline styles** — Tailwind classes only
6. **Icons**: Lucide React only — no other icon library
7. **No `console.log`** in production code — use a `logger` utility if needed
8. **SSE connection** must be cleaned up with `EventSource.close()` in useEffect cleanup
9. **Loading states**: show skeleton loaders on initial data fetch, not spinners
10. **Error boundaries**: wrap Dashboard in an ErrorBoundary component

---

## 🚀 IMPLEMENTATION ORDER (AI Afuate Hii Sequence)

```
Phase 1 — Go Core
  Step 1: go.mod initialization + config loader
  Step 2: Backend struct + Balancer interface
  Step 3: Round Robin implementation + tests
  Step 4: Weighted Round Robin + Least Connections + IP Hash + tests
  Step 5: Rate limiter (token bucket) + tests
  Step 6: Prometheus metrics registry
  Step 7: Access logger (slog JSON)
  Step 8: Health checker goroutine
  Step 9: Core proxy handler (httputil.ReverseProxy)
  Step 10: Admin REST API + SSE log stream endpoint
  Step 11: main.go — wire everything + graceful shutdown

Phase 2 — Frontend Dashboard
  Step 12: Vite + React + TypeScript scaffold
  Step 13: Tailwind CSS config + global styles
  Step 14: TypeScript types
  Step 15: API client + useMetrics hook
  Step 16: useLogStream hook (SSE)
  Step 17: MetricsCards component
  Step 18: BackendTable component
  Step 19: RequestsChart + LatencyChart (Recharts)
  Step 20: AccessLogFeed component
  Step 21: Dashboard layout — assemble all components
  Step 22: Error boundary + loading skeletons

Phase 3 — Integration & Polish
  Step 23: docker-compose.yml + Makefile
  Step 24: prometheus.yml scrape config
  Step 25: README.md with architecture diagram + setup instructions
  Step 26: End-to-end test — start all services, generate load, verify dashboard
```

---

## 📝 ADDITIONAL CONTEXT FOR AI AGENT

- VeloRoute should compile and run as a **single binary** (`./veloroute --config config.yaml`)
- The proxy core must handle **concurrent requests** — do not use any global mutable state without proper synchronization
- The SSE endpoint (`/api/logs/stream`) is the bridge between Go backend logs and React frontend — implement it as a Go channel-based pub/sub: access logger writes to a channel, SSE handler reads from it and writes to `http.ResponseWriter` with `text/event-stream` content type
- All durations in logs and metrics must be in **milliseconds** for the dashboard, but Prometheus histograms use **seconds** (Prometheus convention)
- The frontend polls `/api/metrics` every **2 seconds** for charts and backend table
- Do not use WebSockets — use **SSE only** for the log stream (simpler, HTTP/1.1 compatible)
- Backend health check HTTP client must have its own `http.Client` with a short timeout — **never** reuse the main proxy client
- The `X-Forwarded-For` header must correctly **append** to existing value (not overwrite) to support multi-hop proxies

---

*VeloRoute Master Prompt v1.0 — Generated for Cursor Agent*