import { Fragment, useState } from 'react'
import type { Backend } from '../../types'

interface BackendTableProps {
  backends: Backend[]
  loading: boolean
}

function SkeletonRow() {
  return (
    <tr className="animate-pulse">
      {Array.from({ length: 8 }).map((_, i) => (
        <td key={i} className="px-4 py-3">
          <div className="h-4 bg-border rounded w-full" />
        </td>
      ))}
    </tr>
  )
}

function StatusBadge({ alive }: { alive: boolean }) {
  return (
    <span className="inline-flex items-center gap-1.5">
      <span
        className={`w-2 h-2 rounded-full ${alive ? 'bg-success animate-pulse-dot' : 'bg-error'}`}
      />
      <span className={alive ? 'text-success' : 'text-error'}>
        {alive ? 'Alive' : 'Dead'}
      </span>
    </span>
  )
}

function MiniBar({ requests }: { requests: number }) {
  const width = Math.min(100, (requests % 100))
  return (
    <div className="mt-2 h-2 bg-border rounded-full overflow-hidden">
      <div
        className="h-full bg-primary rounded-full transition-all duration-300"
        style={{ width: `${width}%` }}
      />
    </div>
  )
}

export default function BackendTable({ backends, loading }: BackendTableProps) {
  const [expanded, setExpanded] = useState<string | null>(null)

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="px-5 py-4 border-b border-border">
        <h2 className="text-lg font-semibold">Backend Servers</h2>
      </div>
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-text-muted border-b border-border">
              <th className="px-4 py-3 text-left font-medium">Name</th>
              <th className="px-4 py-3 text-left font-medium">URL</th>
              <th className="px-4 py-3 text-left font-medium">Status</th>
              <th className="px-4 py-3 text-left font-medium">Weight</th>
              <th className="px-4 py-3 text-left font-medium">Conns</th>
              <th className="px-4 py-3 text-left font-medium">Requests</th>
              <th className="px-4 py-3 text-left font-medium">Errors</th>
              <th className="px-4 py-3 text-left font-medium">Last Check</th>
            </tr>
          </thead>
          <tbody>
            {loading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} />)
              : backends.map((b) => (
                  <Fragment key={b.url}>
                    <tr
                      className="border-b border-border/50 hover:bg-border/20 cursor-pointer transition-colors duration-150"
                      onClick={() => setExpanded(expanded === b.url ? null : b.url)}
                    >
                      <td className="px-4 py-3 font-medium">{b.name}</td>
                      <td className="px-4 py-3 text-text-muted font-mono text-xs">{b.url}</td>
                      <td className="px-4 py-3">
                        <StatusBadge alive={b.alive} />
                      </td>
                      <td className="px-4 py-3">{b.weight}</td>
                      <td className="px-4 py-3">{b.active_connections}</td>
                      <td className="px-4 py-3">{b.total_requests.toLocaleString()}</td>
                      <td className="px-4 py-3">{b.total_errors}</td>
                      <td className="px-4 py-3 text-text-muted text-xs">
                        {b.last_health_check
                          ? new Date(b.last_health_check).toLocaleTimeString()
                          : '—'}
                      </td>
                    </tr>
                    {expanded === b.url && (
                      <tr key={`${b.url}-expanded`}>
                        <td colSpan={8} className="px-4 py-2 bg-background/50">
                          <p className="text-xs text-text-muted mb-1">Request distribution</p>
                          <MiniBar requests={b.total_requests} />
                        </td>
                      </tr>
                    )}
                  </Fragment>
                ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
