import { useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { addBackend, removeBackend } from '../../lib/api'
import type { Backend } from '../../types'

interface BackendManagerProps {
  backends: Backend[]
  loading: boolean
  onMutate: () => void
}

function SkeletonRow() {
  return (
    <tr className="animate-pulse">
      {Array.from({ length: 9 }).map((_, i) => (
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
      <span className={alive ? 'text-success' : 'text-error'}>{alive ? 'Alive' : 'Dead'}</span>
    </span>
  )
}

export default function BackendManager({ backends, loading, onMutate }: BackendManagerProps) {
  const [name, setName] = useState('')
  const [url, setUrl] = useState('')
  const [weight, setWeight] = useState('1')
  const [formError, setFormError] = useState<string | null>(null)
  const [busy, setBusy] = useState<string | null>(null)

  async function handleAdd(e: React.FormEvent) {
    e.preventDefault()
    setFormError(null)
    const w = parseInt(weight, 10)
    if (!name.trim() || !url.trim()) {
      setFormError('Name and URL are required')
      return
    }
    if (Number.isNaN(w) || w <= 0) {
      setFormError('Weight must be a positive number')
      return
    }
    setBusy('add')
    try {
      await addBackend({ name: name.trim(), url: url.trim(), weight: w })
      setName('')
      setUrl('')
      setWeight('1')
      onMutate()
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to add backend')
    } finally {
      setBusy(null)
    }
  }

  async function handleRemove(backendUrl: string) {
    setBusy(backendUrl)
    try {
      await removeBackend(backendUrl)
      onMutate()
    } catch {
      // metrics poll will reflect state; avoid noisy UI
    } finally {
      setBusy(null)
    }
  }

  return (
    <div className="bg-card border border-border rounded-lg overflow-hidden">
      <div className="px-5 py-4 border-b border-border flex items-center justify-between gap-4">
        <h2 className="text-lg font-semibold">Backend Servers</h2>
      </div>

      <form onSubmit={handleAdd} className="px-5 py-4 border-b border-border bg-background/30">
        <p className="text-xs text-text-muted mb-3">Add a backend at runtime</p>
        <div className="grid grid-cols-1 sm:grid-cols-4 gap-3">
          <input
            type="text"
            placeholder="Name (e.g. backend-4)"
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="bg-card border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
          />
          <input
            type="url"
            placeholder="http://localhost:8004"
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            className="sm:col-span-2 bg-card border border-border rounded-lg px-3 py-2 text-sm font-mono focus:outline-none focus:ring-2 focus:ring-primary/30"
          />
          <div className="flex gap-2">
            <input
              type="number"
              min={1}
              placeholder="Weight"
              value={weight}
              onChange={(e) => setWeight(e.target.value)}
              className="w-20 bg-card border border-border rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/30"
            />
            <button
              type="submit"
              disabled={busy === 'add'}
              className="flex-1 inline-flex items-center justify-center gap-1.5 bg-primary hover:bg-primary/90 text-white rounded-lg px-3 py-2 text-sm font-medium transition-colors duration-200 disabled:opacity-50 cursor-pointer"
            >
              <Plus className="w-4 h-4" />
              Add
            </button>
          </div>
        </div>
        {formError && <p className="text-xs text-error mt-2">{formError}</p>}
      </form>

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
              <th className="px-4 py-3 text-right font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {loading
              ? Array.from({ length: 3 }).map((_, i) => <SkeletonRow key={i} />)
              : backends.length === 0
                ? (
                    <tr>
                      <td colSpan={9} className="px-4 py-8 text-center text-text-muted">
                        No backends registered yet.
                      </td>
                    </tr>
                  )
                : backends.map((b) => (
                    <tr key={b.url} className="border-b border-border/50 hover:bg-border/20 transition-colors duration-150">
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
                      <td className="px-4 py-3 text-right">
                        <button
                          type="button"
                          onClick={() => handleRemove(b.url)}
                          disabled={busy === b.url}
                          className="inline-flex items-center gap-1 text-error hover:text-error/80 text-xs font-medium transition-colors duration-150 disabled:opacity-50 cursor-pointer"
                          title="Remove backend"
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                          Remove
                        </button>
                      </td>
                    </tr>
                  ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
