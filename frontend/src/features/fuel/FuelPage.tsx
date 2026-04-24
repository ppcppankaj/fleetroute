import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../shared/api/client'
import {
  Fuel, TrendingDown, TrendingUp, AlertTriangle,
  Plus, Download, Filter, BarChart3
} from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface FuelLog {
  id: string
  vehicle_id: string
  registration: string
  driver_name?: string
  liters: number
  cost_per_liter: number
  total_cost: number
  odometer_km: number
  station_name: string
  filled_at: string
  fill_type: 'full' | 'partial'
}

interface FuelAnomaly {
  id: string
  registration: string
  type: string
  drop_liters: number
  drop_percent: number
  detected_at: string
  confirmed: boolean
}

interface FuelStats {
  total_liters: number
  total_cost: number
  avg_efficiency_km_per_l: number
  vehicles_count: number
}

/* ─── Helpers ────────────────────────────────────────────────────────────── */
const fmt = (n: number, decimals = 1) => n?.toFixed(decimals) ?? '—'
const fmtCurrency = (n: number) =>
  new Intl.NumberFormat('en-IN', { style: 'currency', currency: 'INR', maximumFractionDigits: 0 }).format(n)
const fmtDate = (s: string) => s ? new Date(s).toLocaleDateString('en-IN') : '—'

/* ─── Add Fuel Log Modal ─────────────────────────────────────────────────── */
function AddFuelModal({ onClose }: { onClose: () => void }) {
  const qc = useQueryClient()
  const [form, setForm] = useState({
    vehicle_id: '', liters: '', cost_per_liter: '', odometer_km: '',
    station_name: '', fill_type: 'full', filled_at: new Date().toISOString().slice(0, 16),
  })

  const { data: vehicles = [] } = useQuery<any[]>({
    queryKey: ['vehicles'],
    queryFn: async () => (await api.get('/vehicles')).data.data ?? [],
  })

  const mutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/fuel/logs', {
      ...data,
      liters: parseFloat(data.liters),
      cost_per_liter: parseFloat(data.cost_per_liter),
      odometer_km: parseInt(data.odometer_km),
    }),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['fuel'] }); onClose() },
  })

  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{ width: 480 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>Add Fuel Log</h2>
          <button id="fuel-modal-close" className="btn btn-ghost" onClick={onClose}>✕</button>
        </div>
        <div className="modal-body" style={{ display: 'grid', gap: 'var(--space-4)' }}>
          <div style={{ display: 'grid', gap: 8 }}>
            <label className="text-sm font-medium">Vehicle *</label>
            <select className="input" value={form.vehicle_id} onChange={e => set('vehicle_id', e.target.value)}>
              <option value="">Select vehicle…</option>
              {vehicles.map((v: any) => (
                <option key={v.id} value={v.id}>{v.registration} — {v.make} {v.model}</option>
              ))}
            </select>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'grid', gap: 8 }}>
              <label className="text-sm font-medium">Litres *</label>
              <input className="input" type="number" step="0.001" placeholder="45.5"
                value={form.liters} onChange={e => set('liters', e.target.value)} />
            </div>
            <div style={{ display: 'grid', gap: 8 }}>
              <label className="text-sm font-medium">Cost / Litre (₹) *</label>
              <input className="input" type="number" step="0.01" placeholder="92.50"
                value={form.cost_per_liter} onChange={e => set('cost_per_liter', e.target.value)} />
            </div>
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'grid', gap: 8 }}>
              <label className="text-sm font-medium">Odometer (km)</label>
              <input className="input" type="number" placeholder="45680"
                value={form.odometer_km} onChange={e => set('odometer_km', e.target.value)} />
            </div>
            <div style={{ display: 'grid', gap: 8 }}>
              <label className="text-sm font-medium">Fill Type</label>
              <select className="input" value={form.fill_type} onChange={e => set('fill_type', e.target.value)}>
                <option value="full">Full Tank</option>
                <option value="partial">Partial</option>
              </select>
            </div>
          </div>
          <div style={{ display: 'grid', gap: 8 }}>
            <label className="text-sm font-medium">Station Name</label>
            <input className="input" placeholder="HP Petrol Pump, MG Road"
              value={form.station_name} onChange={e => set('station_name', e.target.value)} />
          </div>
          <div style={{ display: 'grid', gap: 8 }}>
            <label className="text-sm font-medium">Date & Time *</label>
            <input className="input" type="datetime-local"
              value={form.filled_at} onChange={e => set('filled_at', e.target.value)} />
          </div>
          {form.liters && form.cost_per_liter && (
            <div className="card" style={{ background: 'var(--color-surface-raised)', padding: 'var(--space-3)' }}>
              <span className="text-sm text-muted">Total Cost: </span>
              <span className="font-semibold" style={{ color: 'var(--color-primary)' }}>
                {fmtCurrency(parseFloat(form.liters) * parseFloat(form.cost_per_liter))}
              </span>
            </div>
          )}
        </div>
        <div className="modal-footer">
          <button id="fuel-cancel" className="btn btn-secondary" onClick={onClose}>Cancel</button>
          <button id="fuel-save" className="btn btn-primary"
            disabled={!form.vehicle_id || !form.liters || mutation.isPending}
            onClick={() => mutation.mutate(form)}>
            {mutation.isPending ? 'Saving…' : 'Save Log'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function FuelPage() {
  const [tab, setTab] = useState<'logs' | 'efficiency' | 'anomalies'>('logs')
  const [showAdd, setShowAdd] = useState(false)

  const { data: logs = [], isLoading } = useQuery<FuelLog[]>({
    queryKey: ['fuel', 'logs'],
    queryFn: async () => (await api.get('/fuel/logs?limit=100')).data.data ?? [],
    refetchInterval: 60_000,
  })

  const { data: anomalies = [] } = useQuery<FuelAnomaly[]>({
    queryKey: ['fuel', 'anomalies'],
    queryFn: async () => (await api.get('/fuel/anomalies')).data.data ?? [],
  })

  // Compute summary stats from logs
  const stats: FuelStats = {
    total_liters: logs.reduce((s, l) => s + (l.liters || 0), 0),
    total_cost: logs.reduce((s, l) => s + (l.total_cost || 0), 0),
    avg_efficiency_km_per_l: 12.4, // TODO real calculation
    vehicles_count: new Set(logs.map(l => l.vehicle_id)).size,
  }

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {showAdd && <AddFuelModal onClose={() => setShowAdd(false)} />}

      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Fuel Management</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>
            M09 · Track fuel consumption, efficiency & theft detection
          </p>
        </div>
        <div className="flex gap-2">
          <button id="fuel-export" className="btn btn-secondary">
            <Download size={14} /> Export
          </button>
          <button id="fuel-add" className="btn btn-primary" onClick={() => setShowAdd(true)}>
            <Plus size={14} /> Add Fuel Log
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: 'var(--space-4)', marginBottom: 'var(--space-6)' }}>
        {[
          { label: 'Total Litres', value: `${fmt(stats.total_liters, 0)} L`, icon: Fuel, color: 'var(--color-primary)' },
          { label: 'Total Cost', value: fmtCurrency(stats.total_cost), icon: TrendingDown, color: 'var(--color-danger)' },
          { label: 'Avg Efficiency', value: `${fmt(stats.avg_efficiency_km_per_l)} km/L`, icon: TrendingUp, color: 'var(--color-success)' },
          { label: 'Anomalies', value: `${anomalies.filter(a => !a.confirmed).length}`, icon: AlertTriangle, color: 'var(--color-warning)' },
        ].map(({ label, value, icon: Icon, color }) => (
          <div key={label} className="card animate-fade-in" style={{ padding: 'var(--space-5)' }}>
            <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-3)' }}>
              <span className="text-sm text-muted">{label}</span>
              <div style={{ width: 36, height: 36, borderRadius: 8, background: `${color}18`, display: 'grid', placeItems: 'center' }}>
                <Icon size={18} color={color} />
              </div>
            </div>
            <div style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>{value}</div>
          </div>
        ))}
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)' }}>
        {(['logs', 'efficiency', 'anomalies'] as const).map(t => (
          <button
            key={t}
            id={`fuel-tab-${t}`}
            className={`btn ${tab === t ? 'btn-primary' : 'btn-ghost'}`}
            style={{ borderRadius: '6px 6px 0 0', textTransform: 'capitalize' }}
            onClick={() => setTab(t)}
          >
            {t === 'anomalies' ? `Anomalies${anomalies.length ? ` (${anomalies.length})` : ''}` : t === 'efficiency' ? <><BarChart3 size={13} /> Efficiency</> : 'Fuel Logs'}
          </button>
        ))}
      </div>

      {/* ── Tab: Logs ── */}
      {tab === 'logs' && (
        <div className="card animate-fade-in">
          {isLoading ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <div className="spinner" />
            </div>
          ) : logs.length === 0 ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <Fuel size={40} color="var(--color-text-muted)" style={{ margin: '0 auto 12px' }} />
              <p className="text-muted">No fuel logs yet. Add your first log.</p>
            </div>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Vehicle</th>
                    <th>Date</th>
                    <th>Litres</th>
                    <th>₹/L</th>
                    <th>Total Cost</th>
                    <th>Odometer</th>
                    <th>Station</th>
                    <th>Type</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map(log => (
                    <tr key={log.id}>
                      <td><span className="font-mono text-sm">{log.registration}</span></td>
                      <td><span className="text-sm">{fmtDate(log.filled_at)}</span></td>
                      <td><span className="font-medium">{fmt(log.liters)} L</span></td>
                      <td>₹{fmt(log.cost_per_liter, 2)}</td>
                      <td><span style={{ color: 'var(--color-danger)' }}>{fmtCurrency(log.total_cost)}</span></td>
                      <td className="text-muted text-sm">{log.odometer_km?.toLocaleString()} km</td>
                      <td className="text-sm text-muted">{log.station_name || '—'}</td>
                      <td>
                        <span className={`badge ${log.fill_type === 'full' ? 'badge-success' : 'badge-info'}`}>
                          {log.fill_type}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ── Tab: Efficiency ── */}
      {tab === 'efficiency' && (
        <div className="card animate-fade-in" style={{ padding: 'var(--space-8)', textAlign: 'center' }}>
          <BarChart3 size={48} color="var(--color-primary)" style={{ margin: '0 auto 16px' }} />
          <p className="text-muted">Efficiency analysis charts — coming in next sprint.</p>
          <p className="text-sm text-muted" style={{ marginTop: 8 }}>
            Will show km/L per vehicle, monthly trends, and fleet comparison.
          </p>
        </div>
      )}

      {/* ── Tab: Anomalies ── */}
      {tab === 'anomalies' && (
        <div className="card animate-fade-in">
          {anomalies.length === 0 ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <AlertTriangle size={40} color="var(--color-success)" style={{ margin: '0 auto 12px' }} />
              <p className="text-muted">No fuel anomalies detected.</p>
            </div>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Vehicle</th>
                    <th>Type</th>
                    <th>Drop (L)</th>
                    <th>Drop (%)</th>
                    <th>Detected At</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {anomalies.map(a => (
                    <tr key={a.id}>
                      <td className="font-mono text-sm">{a.registration}</td>
                      <td><span className="badge badge-danger">{a.type}</span></td>
                      <td className="font-medium">{fmt(a.drop_liters)} L</td>
                      <td style={{ color: 'var(--color-danger)' }}>{fmt(a.drop_percent)}%</td>
                      <td className="text-sm text-muted">{new Date(a.detected_at).toLocaleString()}</td>
                      <td>
                        <span className={`badge ${a.confirmed ? 'badge-danger' : 'badge-warning'}`}>
                          {a.confirmed ? 'Confirmed Theft' : 'Under Review'}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
