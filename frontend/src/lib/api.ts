import type { Algorithm, Backend, MetricsSnapshot } from '../types'

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:9090'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json', ...options?.headers },
    ...options,
  })
  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `API error: ${res.status} ${res.statusText}`)
  }
  return res.json() as Promise<T>
}

export function getMetrics() {
  return request<MetricsSnapshot>('/api/metrics')
}

export function getBackends() {
  return request<Backend[]>('/api/backends')
}

export function getConfig() {
  return request<{ algorithm: Algorithm }>('/api/config')
}

export function setAlgorithm(algorithm: Algorithm) {
  return request<{ algorithm: string }>('/api/config/algorithm', {
    method: 'PUT',
    body: JSON.stringify({ algorithm }),
  })
}

export function addBackend(payload: { url: string; name: string; weight: number }) {
  return request<{ status: string }>('/api/backends', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export function encodeBackendURL(url: string): string {
  return btoa(url).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
}

export function removeBackend(url: string) {
  return request<{ status: string }>(`/api/backends/${encodeBackendURL(url)}`, {
    method: 'DELETE',
  })
}

export function getLogStreamURL(): string {
  return `${API_BASE}/api/logs/stream`
}

export { API_BASE }
