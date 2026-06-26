import { useCallback, useEffect, useState } from 'react'
import { getConfig } from '../../lib/api'
import { VELO_ROUTE_LOGO_URL } from '../../lib/brand'
import { useMetrics } from '../../hooks/useMetrics'
import { useLogStream } from '../../hooks/useLogStream'
import type { Algorithm } from '../../types'
import MetricsCards from './MetricsCards'
import BackendManager from './BackendManager'
import RequestsChart from './RequestsChart'
import LatencyChart from './LatencyChart'
import AccessLogFeed from './AccessLogFeed'
import AlgorithmSwitcher from './AlgorithmSwitcher'

export default function Dashboard() {
  const { data, loading, error, refresh } = useMetrics()
  const { logs, connected } = useLogStream()
  const [algorithm, setAlgorithm] = useState<Algorithm>('round_robin')

  useEffect(() => {
    getConfig()
      .then((cfg) => setAlgorithm(cfg.algorithm))
      .catch(() => {})
  }, [])

  const handleMutate = useCallback(() => {
    refresh()
  }, [refresh])

  const backends = data?.backends ?? []
  const aliveCount = backends.filter((b) => b.alive).length

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b border-border bg-card/50 backdrop-blur-sm sticky top-0 z-10">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 py-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div className="flex items-center gap-3">
            <img
              src={VELO_ROUTE_LOGO_URL}
              alt="VeloRoute"
              className="h-8 w-auto"
            />
            <span className="flex items-center gap-1.5 text-xs text-success">
              <span className="w-2 h-2 rounded-full bg-success animate-pulse-dot" />
              Online
            </span>
          </div>
          <div className="flex items-center gap-6">
            <AlgorithmSwitcher value={algorithm} onChange={setAlgorithm} />
            <span className="text-sm text-text-muted whitespace-nowrap">
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
          <BackendManager backends={backends} loading={loading} onMutate={handleMutate} />
        </div>

        <AccessLogFeed logs={logs} connected={connected} />
      </main>
    </div>
  )
}
