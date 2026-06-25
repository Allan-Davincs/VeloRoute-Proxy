import { Activity, Zap, AlertTriangle, Timer } from 'lucide-react'
import type { MetricsSnapshot } from '../../types'

interface MetricsCardsProps {
  data: MetricsSnapshot | null
  loading: boolean
}

function SkeletonCard() {
  return (
    <div className="bg-card border border-border rounded-lg p-5 animate-pulse">
      <div className="h-4 w-24 bg-border rounded mb-3" />
      <div className="h-8 w-32 bg-border rounded" />
    </div>
  )
}

function formatNumber(n: number): string {
  return n.toLocaleString()
}

export default function MetricsCards({ data, loading }: MetricsCardsProps) {
  if (loading || !data) {
    return (
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
        {Array.from({ length: 4 }).map((_, i) => (
          <SkeletonCard key={i} />
        ))}
      </div>
    )
  }

  const errorColor =
    data.error_rate_percent > 1
      ? 'text-error'
      : data.error_rate_percent < 0.5
        ? 'text-success'
        : 'text-text-primary'

  const latencyColor =
    data.p95_latency_ms > 500
      ? 'text-error'
      : data.p95_latency_ms > 200
        ? 'text-warning'
        : 'text-text-primary'

  const cards = [
    {
      icon: Activity,
      label: 'Total Requests',
      value: formatNumber(data.total_requests),
      color: 'text-text-primary',
    },
    {
      icon: Zap,
      label: 'Requests/sec',
      value: data.requests_per_second.toFixed(1),
      color: 'text-text-primary',
    },
    {
      icon: AlertTriangle,
      label: 'Error Rate',
      value: `${data.error_rate_percent.toFixed(2)}%`,
      color: errorColor,
    },
    {
      icon: Timer,
      label: 'P95 Latency',
      value: `${data.p95_latency_ms.toFixed(1)}ms`,
      color: latencyColor,
    },
  ]

  return (
    <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
      {cards.map((card) => (
        <div
          key={card.label}
          className="bg-card border border-border rounded-lg p-5 transition-colors duration-200 hover:border-primary/30"
        >
          <div className="flex items-center gap-2 mb-2">
            <card.icon className="w-4 h-4 text-text-muted" />
            <span className="text-sm text-text-muted">{card.label}</span>
          </div>
          <p className={`text-2xl font-semibold ${card.color}`}>{card.value}</p>
        </div>
      ))}
    </div>
  )
}
