import { useEffect, useRef, useState } from 'react'
import { getLogStreamURL } from '../lib/api'
import type { LogEntry } from '../types'

const MAX_LOGS = 100

export function useLogStream() {
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [connected, setConnected] = useState(false)
  const backoffRef = useRef(1000)

  useEffect(() => {
    let source: EventSource | null = null
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let active = true

    function connect() {
      if (!active) return

      source = new EventSource(getLogStreamURL())

      source.onopen = () => {
        setConnected(true)
        backoffRef.current = 1000
      }

      source.onmessage = (event) => {
        try {
          const entry = JSON.parse(event.data) as LogEntry
          setLogs((prev) => {
            const next = [...prev, entry]
            return next.length > MAX_LOGS ? next.slice(-MAX_LOGS) : next
          })
        } catch {
          // ignore malformed entries
        }
      }

      source.onerror = () => {
        setConnected(false)
        source?.close()
        source = null
        if (active) {
          const delay = backoffRef.current
          backoffRef.current = Math.min(delay * 2, 30000)
          reconnectTimer = setTimeout(connect, delay)
        }
      }
    }

    connect()

    return () => {
      active = false
      if (reconnectTimer) clearTimeout(reconnectTimer)
      source?.close()
    }
  }, [])

  return { logs, connected }
}
