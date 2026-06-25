import { useEffect, useRef, useState } from 'react'
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'
import type { ChartPoint } from '../../types'

interface RequestsChartProps {
  rps: number
}

const MAX_POINTS = 60

export default function RequestsChart({ rps }: RequestsChartProps) {
  const [data, setData] = useState<ChartPoint[]>([])
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null)

  useEffect(() => {
    function addPoint() {
      const now = new Date().toLocaleTimeString()
      setData((prev) => {
        const next = [...prev, { time: now, value: rps }]
        return next.length > MAX_POINTS ? next.slice(-MAX_POINTS) : next
      })
    }

    addPoint()
    intervalRef.current = setInterval(addPoint, 1000)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [rps])

  return (
    <div className="bg-card border border-border rounded-lg p-5">
      <h2 className="text-lg font-semibold mb-4">Requests / sec</h2>
      <ResponsiveContainer width="100%" height={220}>
        <LineChart data={data}>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a2d3e" />
          <XAxis dataKey="time" stroke="#64748b" fontSize={11} tickLine={false} />
          <YAxis stroke="#64748b" fontSize={11} tickLine={false} />
          <Tooltip
            contentStyle={{
              backgroundColor: '#1a1d27',
              border: '1px solid #2a2d3e',
              borderRadius: '8px',
              color: '#f1f5f9',
            }}
          />
          <Line
            type="monotone"
            dataKey="value"
            stroke="#6366f1"
            strokeWidth={2}
            dot={false}
            isAnimationActive={true}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  )
}
