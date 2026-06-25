import { useEffect, useState } from 'react'
import { getBackends } from '../lib/api'
import type { Backend } from '../types'

const POLL_INTERVAL = 2000

export function useBackends() {
  const [backends, setBackends] = useState<Backend[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let active = true

    async function poll() {
      try {
        const list = await getBackends()
        if (active) {
          setBackends(list)
          setError(null)
          setLoading(false)
        }
      } catch (err) {
        if (active) {
          setError(err instanceof Error ? err.message : 'Failed to fetch backends')
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

  return { backends, loading, error }
}
