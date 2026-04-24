import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../shared/api/client'
import {
  Cpu, Wifi, WifiOff, Plus, Terminal, UploadCloud,
  Activity, RefreshCw, Shield, Settings
} from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface Device {
  id: string
  imei: string
  sim_number: string
  device_name: string
  device_type: string
  manufacturer: string
  model: string
  protocol: string
  status: 'active' | 'inactive' | 'paused' | 'lost' | 'stolen'
  is_connected: boolean
  firmware_version: string
  vehicle_id?: string
  registration?: string
  last_seen_at?: string
  signal_strength?: number
  battery_percent?: number
}

interface Protocol {
  name: string
  port: number
  models: string[]
}

/* ─── Status badge ───────────────────────────────────────────────────────── */
const statusColor: Record<string, string> = {
  active: 'badge-success', inactive: 'badge-muted',
  paused: 'badge-warning', lost: 'badge-danger', stolen: 'badge-danger',
}

/* ─── Provision Modal ────────────────────────────────────────────────────── */
function ProvisionModal({ protocols, onClose }: { protocols: Protocol[]; onClose: () => void }) {
  const qc = useQueryClient()
  const [form, setForm] = useState({
    imei: '', sim_number: '', device_name: '', manufacturer: '',
    model: '', protocol: '', vehicle_id: '',
  })
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const mutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/devices', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['devices'] }); onClose() },
  })

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{ width: 500 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>Provision New Device</h2>
          <button id="device-modal-close" className="btn btn-ghost" onClick={onClose}>✕</button>
        </div>
        <div className="modal-body" style={{ display: 'grid', gap: 'var(--space-4)' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">IMEI *</label>
              <input className="input font-mono" placeholder="356307042441993"
                value={form.imei} onChange={e => set('imei', e.target.value)} />
            </div>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">SIM Number</label>
              <input className="input font-mono" placeholder="8991101200003204869"
                value={form.sim_number} onChange={e => set('sim_number', e.target.value)} />
            </div>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            <label className="text-sm font-medium">Device Name</label>
            <input className="input" placeholder="Truck-Fleet-01"
              value={form.device_name} onChange={e => set('device_name', e.target.value)} />
          </div>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">Manufacturer *</label>
              <select className="input" value={form.manufacturer} onChange={e => set('manufacturer', e.target.value)}>
                <option value="">Select…</option>
                {protocols.map(p => <option key={p.name} value={p.name}>{p.name}</option>)}
                <option value="Custom">Custom</option>
              </select>
            </div>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">Protocol *</label>
              <select className="input" value={form.protocol} onChange={e => set('protocol', e.target.value)}>
                <option value="">Select…</option>
                {protocols.map(p => (
                  <option key={p.name} value={p.name.toLowerCase()}>
                    {p.name} (port {p.port})
                  </option>
                ))}
              </select>
            </div>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            <label className="text-sm font-medium">Model</label>
            <input className="input" placeholder="FMB920"
              value={form.model} onChange={e => set('model', e.target.value)} />
          </div>
          {form.protocol && (
            <div className="card" style={{ background: 'var(--color-surface-raised)', padding: 'var(--space-3)' }}>
              <p className="text-sm text-muted">
                TCP Listener: <code className="font-mono" style={{ color: 'var(--color-primary)' }}>
                  {protocols.find(p => p.name.toLowerCase() === form.protocol)?.port ?? '—'}
                </code> · Configure device server IP to point to this platform.
              </p>
            </div>
          )}
        </div>
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>Cancel</button>
          <button id="device-save" className="btn btn-primary"
            disabled={!form.imei || !form.manufacturer || !form.protocol || mutation.isPending}
            onClick={() => mutation.mutate(form)}>
            {mutation.isPending ? 'Provisioning…' : 'Provision Device'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* ─── Command Console Modal ──────────────────────────────────────────────── */
function CommandModal({ device, onClose }: { device: Device; onClose: () => void }) {
  const [cmd, setCmd] = useState('restart')
  const [log, setLog] = useState<string[]>([])
  const mutation = useMutation({
    mutationFn: () => api.post(`/devices/${device.id}/command`, { command: cmd }),
    onSuccess: (res) => setLog(l => [`> ${cmd}`, `← ${JSON.stringify(res.data.data)}`, ...l]),
    onError: () => setLog(l => [`> ${cmd}`, '← ERROR: Command failed', ...l]),
  })
  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{ width: 520 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>
            <Terminal size={18} style={{ marginRight: 8 }} />
            Command Console — {device.imei}
          </h2>
          <button className="btn btn-ghost" onClick={onClose}>✕</button>
        </div>
        <div className="modal-body" style={{ display: 'grid', gap: 'var(--space-4)' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr auto', gap: 8 }}>
            <select className="input" value={cmd} onChange={e => setCmd(e.target.value)}>
              <option value="restart">restart — Reboot device</option>
              <option value="request_config">request_config — Fetch current settings</option>
              <option value="set_interval_30">set_interval_30 — Set 30s GPS interval</option>
              <option value="set_interval_60">set_interval_60 — Set 60s GPS interval</option>
              <option value="immobilize">immobilize — Cut engine relay</option>
              <option value="unlock">unlock — Restore engine relay</option>
            </select>
            <button id="device-send-cmd" className="btn btn-primary"
              onClick={() => mutation.mutate()} disabled={mutation.isPending}>
              Send
            </button>
          </div>
          <div style={{
            background: 'var(--color-bg)', borderRadius: 8, padding: 12, minHeight: 160,
            fontFamily: 'monospace', fontSize: 12, color: 'var(--color-text)', overflowY: 'auto',
          }}>
            {log.length === 0 ? (
              <span style={{ color: 'var(--color-text-muted)' }}>Select a command and click Send…</span>
            ) : log.map((l, i) => (
              <div key={i} style={{ color: l.startsWith('>') ? 'var(--color-primary)' : l.includes('ERROR') ? 'var(--color-danger)' : 'var(--color-success)' }}>
                {l}
              </div>
            ))}
          </div>
        </div>
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>Close</button>
        </div>
      </div>
    </div>
  )
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function DevicesPage() {
  const [showAdd, setShowAdd] = useState(false)
  const [cmdDevice, setCmdDevice] = useState<Device | null>(null)
  const [search, setSearch] = useState('')

  const { data: devices = [], isLoading, refetch } = useQuery<Device[]>({
    queryKey: ['devices'],
    queryFn: async () => (await api.get('/devices')).data.data ?? [],
    refetchInterval: 30_000,
  })

  const { data: protocols = [] } = useQuery<Protocol[]>({
    queryKey: ['device-protocols'],
    queryFn: async () => (await api.get('/device-protocols')).data.data ?? [],
  })

  const filtered = devices.filter(d =>
    `${d.imei} ${d.device_name} ${d.manufacturer} ${d.registration}`.toLowerCase().includes(search.toLowerCase())
  )

  const connected = devices.filter(d => d.is_connected).length
  const active = devices.filter(d => d.status === 'active').length

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {showAdd && <ProvisionModal protocols={protocols} onClose={() => setShowAdd(false)} />}
      {cmdDevice && <CommandModal device={cmdDevice} onClose={() => setCmdDevice(null)} />}

      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Device Management</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>
            M12 · Provision, monitor, and command GPS devices
          </p>
        </div>
        <div className="flex gap-2">
          <button id="devices-refresh" className="btn btn-secondary" onClick={() => refetch()}>
            <RefreshCw size={14} /> Refresh
          </button>
          <button id="devices-provision" className="btn btn-primary" onClick={() => setShowAdd(true)}>
            <Plus size={14} /> Provision Device
          </button>
        </div>
      </div>

      {/* KPI row */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4,1fr)', gap: 'var(--space-4)', marginBottom: 'var(--space-6)' }}>
        {[
          { label: 'Total Devices', value: devices.length, icon: Cpu, color: 'var(--color-primary)' },
          { label: 'Connected', value: connected, icon: Wifi, color: 'var(--color-success)' },
          { label: 'Offline', value: devices.length - connected, icon: WifiOff, color: 'var(--color-danger)' },
          { label: 'Active', value: active, icon: Activity, color: 'var(--color-info)' },
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

      {/* Supported Protocols */}
      <div className="card" style={{ padding: 'var(--space-4)', marginBottom: 'var(--space-4)' }}>
        <p className="text-sm font-medium" style={{ marginBottom: 'var(--space-3)' }}>Supported Protocols</p>
        <div className="flex flex-wrap gap-2">
          {protocols.map(p => (
            <span key={p.name} className="badge badge-info">
              <Shield size={10} /> {p.name} :{p.port}
            </span>
          ))}
        </div>
      </div>

      {/* Search */}
      <div style={{ marginBottom: 'var(--space-4)' }}>
        <input className="input" style={{ maxWidth: 360 }} placeholder="Search by IMEI, name, protocol…"
          value={search} onChange={e => setSearch(e.target.value)} />
      </div>

      {/* Table */}
      <div className="card animate-fade-in">
        {isLoading ? (
          <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}><div className="spinner" /></div>
        ) : filtered.length === 0 ? (
          <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
            <Cpu size={40} color="var(--color-text-muted)" style={{ margin: '0 auto 12px' }} />
            <p className="text-muted">No devices found. Provision your first device.</p>
          </div>
        ) : (
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>IMEI</th>
                  <th>Name</th>
                  <th>Protocol</th>
                  <th>Manufacturer</th>
                  <th>Vehicle</th>
                  <th>Signal</th>
                  <th>Battery</th>
                  <th>Status</th>
                  <th>Last Seen</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {filtered.map(d => (
                  <tr key={d.id}>
                    <td>
                      <div className="flex items-center gap-2">
                        <div style={{
                          width: 8, height: 8, borderRadius: '50%',
                          background: d.is_connected ? 'var(--color-success)' : 'var(--color-text-muted)',
                          flexShrink: 0,
                        }} />
                        <span className="font-mono text-sm">{d.imei}</span>
                      </div>
                    </td>
                    <td className="font-medium">{d.device_name || '—'}</td>
                    <td><span className="badge badge-info text-xs">{d.protocol}</span></td>
                    <td className="text-sm text-muted">{d.manufacturer}</td>
                    <td><span className="font-mono text-sm">{d.registration || '—'}</span></td>
                    <td>
                      {d.signal_strength != null ? (
                        <div className="flex items-center gap-1">
                          <div style={{
                            width: 40, height: 6, borderRadius: 3, background: 'var(--color-surface-raised)',
                            overflow: 'hidden',
                          }}>
                            <div style={{
                              width: `${d.signal_strength}%`, height: '100%',
                              background: d.signal_strength > 60 ? 'var(--color-success)' : d.signal_strength > 30 ? 'var(--color-warning)' : 'var(--color-danger)',
                            }} />
                          </div>
                          <span className="text-xs text-muted">{d.signal_strength}%</span>
                        </div>
                      ) : '—'}
                    </td>
                    <td>
                      {d.battery_percent != null ? (
                        <span style={{ color: d.battery_percent < 20 ? 'var(--color-danger)' : 'var(--color-text)' }}>
                          {d.battery_percent}%
                        </span>
                      ) : '—'}
                    </td>
                    <td><span className={`badge ${statusColor[d.status] ?? 'badge-muted'}`}>{d.status}</span></td>
                    <td className="text-xs text-muted">
                      {d.last_seen_at ? new Date(d.last_seen_at).toLocaleString() : '—'}
                    </td>
                    <td>
                      <div className="flex gap-1">
                        <button
                          id={`device-cmd-${d.id}`}
                          className="btn btn-secondary btn-sm"
                          title="Send Command"
                          onClick={() => setCmdDevice(d)}
                        >
                          <Terminal size={12} />
                        </button>
                        <button
                          id={`device-ota-${d.id}`}
                          className="btn btn-secondary btn-sm"
                          title="OTA Update"
                          onClick={() => api.post(`/devices/${d.id}/ota`, {})}
                        >
                          <UploadCloud size={12} />
                        </button>
                        <button className="btn btn-ghost btn-sm" title="Settings">
                          <Settings size={12} />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  )
}
