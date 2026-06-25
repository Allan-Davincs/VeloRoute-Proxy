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

export type Algorithm =
  | 'round_robin'
  | 'weighted_round_robin'
  | 'least_connections'
  | 'ip_hash'

export interface ChartPoint {
  time: string
  value: number
  p50?: number
  p95?: number
  p99?: number
}
