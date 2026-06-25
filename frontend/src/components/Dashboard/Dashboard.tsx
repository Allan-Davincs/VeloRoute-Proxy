import { useEffect, useState } from 'react'
import { Activity } from 'lucide-react'
import { getConfig } from '../../lib/api'
import { useMetrics } from '../../hooks/useMetrics'
import { useLogStream } from '../../hooks/useLogStream'
import MetricsCards from './MetricsCards'
import BackendTable from './BackendTable'
import RequestsChart from './RequestsChart'
import LatencyChart from './LatencyChart'
import AccessLogFeed from './AccessLogFeed'

function formatAlgorithm(algo: string): string {
  return algo
    .split('_')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ')
}

export default function Dashboard() {
  const { data, loading, error } = useMetrics()
  const { logs, connected } = useLogStream()
  const [algorithm, setAlgorithm] = useState('round_robin')

  useEffect(() => {
    getConfig()
      .then((cfg) => setAlgorithm(cfg.algorithm))
      .catch(() => {})
  }, [])

  const backends = data?.backends ?? []
  const aliveCount = backends.filter((b) => b.alive).length

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <Activity className="w-6 h-6 text-primary" />
            <h1 className="text-xl font-bold">VeloRoute</h1>
            <span className="flex items-center gap-1.5 text-xs text-success">
              <span className="w-2 h-2 rounded-full bg-success animate-pulse-dot" />
              Online
            </span>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-text-muted">
              Algorithm:{' '}
              <span className="text-primary font-medium">{formatAlgorithm(algorithm)}</span>
            </span>
            <span className="text-sm text-text-muted">
              {aliveCount}/{backends.length} backends alive
            </span>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 py-6 space-y-6">
        {error && (
          <div className="bg-error/10 border border-error/30 text-error text-sm rounded-lg px-4 py-3">
            {error}
          </div>
        )}

        <MetricsCards data={data} loading={loading} />

        <RequestsChart rps={data?.requests_per_second ?? 0} />

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <LatencyChart
            p50={data?.p50_latency_ms ?? 0}
            p95={data?.p95_latency_ms ?? 0}
            p99={data?.p99_latency_ms ?? 0}
          />
          <BackendTable backends={backends} loading={loading} />
        </div>

        <AccessLogFeed logs={logs} connected={connected} />
      </main>
    </div>
  )
}
