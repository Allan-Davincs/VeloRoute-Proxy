import { useEffect, useRef, useState } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from 'recharts'
import type { ChartPoint } from '../../types'

interface LatencyChartProps {
  p50: number
  p95: number
  p99: number
}

const MAX_POINTS = 60

export default function LatencyChart({ p50, p95, p99 }: LatencyChartProps) {
  const [data, setData] = useState<ChartPoint[]>([])
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    function addPoint() {
      const now = new Date().toLocaleTimeString()
      setData((prev) => {
        const next = [...prev, { time: now, value: 0, p50, p95, p99 }]
        return next.length > MAX_POINTS ? next.slice(-MAX_POINTS) : next
      })
    }

    addPoint()
    intervalRef.current = setInterval(addPoint, 1000)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [p50, p95, p99])

  return (
    <div className="bg-card border border-border rounded-lg p-5">
      <h2 className="text-lg font-semibold mb-4">Latency Percentiles</h2>
      <ResponsiveContainer width="100%" height={220}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a2d3e" />
          <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} />
          <YAxis stroke="#64748b" fontSize={11} tickLine={false} unit="ms" />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1a1d27',
              border: '1px solid #2a2d3e',
              borderRadius: '8px',
              color: '#f1f5f9',
            }}
          />
          <Legend />
          <Line type="monotone" dataKey="p50" stroke="#22c55e" strokeWidth={2} dot={false} name="P50" />
          <Line type="monotone" dataKey="p95" stroke="#f59e0b" strokeWidth={2} dot={false} name="P95" />
          <Line type="monotone" dataKey="p99" stroke="#ef4444" strokeWidth={2} dot={false} name="P99" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
