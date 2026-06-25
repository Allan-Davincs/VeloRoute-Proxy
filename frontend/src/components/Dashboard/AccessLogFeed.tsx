import { useEffect, useRef } from 'react'
import type { LogEntry } from '../../types'

interface AccessLogFeedProps {
  logs: LogEntry[]
  connected: boolean
}

const methodColors: Record<string, string> = {
  GET: 'bg-primary/20 text-primary',
  POST: 'bg-success/20 text-success',
  PUT: 'bg-warning/20 text-warning',
  DELETE: 'bg-error/20 text-error',
}

function statusColor(status: number): string {
  if (status >= 500) return 'bg-error/20 text-error'
  if (status >= 400) return 'bg-warning/20 text-warning'
  if (status >= 300) return 'bg-blue-500/20 text-blue-400'
  return 'bg-success/20 text-success'
}

export default function AccessLogFeed({ logs, connected }: AccessLogFeedProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const userScrolledRef = useRef(false)

  useEffect(() => {
    const el = containerRef.current
    if (!el || userScrolledRef.current) return
    el.scrollTop = el.scrollHeight
  }, [logs])

  function handleScroll() {
    const el = containerRef.current
    if (!el) return
    const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40
    userScrolledRef.current = !atBottom
  }

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="px-5 py-4 border-b border-border flex items-center justify-between">
        <h2 className="text-lg font-semibold">Live Access Log</h2>
        <span className="flex items-center gap-1.5 text-xs text-text-muted">
          <span
            className={`w-2 h-2 rounded-full ${connected ? 'bg-success animate-pulse-dot' : 'bg-error'}`}
          />
          {connected ? 'Connected' : 'Disconnected'}
        </span>
      </div>
      <div
        ref={containerRef}
        onScroll={handleScroll}
        className="h-64 overflow-y-auto font-mono text-xs"
      >
        {logs.length === 0 ? (
          <p className="p-5 text-text-muted">Waiting for log entries...</p>
        ) : (
          logs.map((log) => (
            <div
              key={log.request_id + log.time}
              className="flex items-center gap-3 px-5 py-2 border-b border-border/30 hover:bg-border/10 animate-fade-in transition-colors duration-150"
            >
              <span className="text-text-muted w-20 shrink-0">
                {new Date(log.time).toLocaleTimeString()}
              </span>
              <span
                className={`px-1.5 py-0.5 rounded text-[10px] font-semibold shrink-0 ${methodColors[log.method] ?? 'bg-border text-text-muted'}`}
              >
                {log.method}
              </span>
              <span className="truncate flex-1 text-text-primary">{log.path}</span>
              <span
                className={`px-1.5 py-0.5 rounded text-[10px] font-semibold shrink-0 ${statusColor(log.status)}`}
              >
                {log.status}
              </span>
              <span className="text-text-muted w-16 text-right shrink-0">
                {log.duration_ms.toFixed(1)}ms
              </span>
              <span className="text-text-muted w-24 truncate shrink-0">{log.backend}</span>
              <span className="text-text-muted w-28 truncate shrink-0 hidden lg:inline">
                {log.client_ip}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
