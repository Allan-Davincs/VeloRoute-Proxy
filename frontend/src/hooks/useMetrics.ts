import { useEffect, useRef, useState } from 'react'
import { getMetrics } from '../lib/api'
import type { MetricsSnapshot } from '../types'

const POLL_INTERVAL = 2000

export function useMetrics() {
  const [data, setData] = useState<MetricsSnapshot | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const dataRef = useRef<MetricsSnapshot | null>(null)

  useEffect(() => {
    let active = true

    async function poll() {
      try {
        const snapshot = await getMetrics()
        if (active) {
          dataRef.current = snapshot
          setData(snapshot)
          setError(null)
          setLoading(false)
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : 'Failed to fetch metrics')
          if (dataRef.current) {
            setData(dataRef.current)
          }
          setLoading(false)
        }
      }
    }

    poll()
    const id = setInterval(poll, POLL_INTERVAL)
    return () => {
      active = false
      clearInterval(id)
    }
  }, [])

  async function refresh() {
    try {
      const snapshot = await getMetrics()
      dataRef.current = snapshot
      setData(snapshot)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch metrics')
    }
  }

  return { data, loading, error, refresh }
}
