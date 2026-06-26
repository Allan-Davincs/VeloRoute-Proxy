import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { setAlgorithm } from '../../lib/api'
import type { Algorithm } from '../../types'

const ALGORITHMS: { value: Algorithm; label: string }[] = [
  { value: 'round_robin', label: 'Round Robin' },
  { value: 'weighted_round_robin', label: 'Weighted Round Robin' },
  { value: 'least_connections', label: 'Least Connections' },
  { value: 'ip_hash', label: 'IP Hash' },
]

interface AlgorithmSwitcherProps {
  value: Algorithm
  onChange: (algo: Algorithm) => void
}

export default function AlgorithmSwitcher({ value, onChange }: AlgorithmSwitcherProps) {
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleChange(next: Algorithm) {
    if (next === value) return
    setSaving(true)
    setError(null)
    try {
      await setAlgorithm(next)
      onChange(next)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update algorithm')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="flex flex-col items-end gap-1">
      <label className="text-xs text-text-muted">Load balancing</label>
      <div className="relative">
        <select
          value={value}
          disabled={saving}
          onChange={(e) => handleChange(e.target.value as Algorithm)}
          className="appearance-none bg-card border border-border rounded-lg pl-3 pr-9 py-2 text-sm text-text-primary cursor-pointer hover:border-primary/40 focus:outline-none focus:ring-2 focus:ring-primary/30 transition-colors duration-200 disabled:opacity-50"
        >
          {ALGORITHMS.map((a) => (
            <option key={a.value} value={a.value}>
              {a.label}
            </option>
          ))}
        </select>
        <ChevronDown className="w-4 h-4 text-text-muted absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none" />
      </div>
      {error && <span className="text-xs text-error max-w-[200px] text-right">{error}</span>}
    </div>
  )
}
