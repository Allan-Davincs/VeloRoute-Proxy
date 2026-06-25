const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:9090'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    throw new Error(`API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export function getMetrics() {
  return request<import('../types').MetricsSnapshot>('/api/metrics')
}

export function getBackends() {
  return request<import('../types').Backend[]>('/api/backends')
}

export function getConfig() {
  return request<{ algorithm: string }>('/api/config')
}

export function getLogStreamURL(): string {
  return `${API_BASE}/api/logs/stream`
}

export { API_BASE }
