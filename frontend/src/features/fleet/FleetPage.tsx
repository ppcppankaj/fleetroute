import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Truck, Users, Cpu, Grid, Plus, ExternalLink, Signal } from 'lucide-react'
import api from '../../shared/api/client'

type Tab = 'vehicles' | 'drivers' | 'devices' | 'groups'

interface Vehicle {
  id: string
  registration: string
  make: string
  model: string
  year: number
  device_id?: string
  created_at: string
}

interface Driver {
  id: string
  name: string
  license_number?: string
  phone?: string
  created_at: string
}

interface Device {
  id: string
  imei: string
  name: string
  protocol: string
  online: boolean
  last_seen_at?: string
}

export default function FleetPage() {
  const [tab, setTab] = useState<Tab>('vehicles')
  const navigate = useNavigate()

  const { data: vehicles = [] } = useQuery<Vehicle[]>({
    queryKey: ['vehicles'],
    queryFn: async () => { const r = await api.get('/vehicles'); return r.data.data ?? [] },
  })

  const { data: drivers = [] } = useQuery<Driver[]>({
    queryKey: ['drivers'],
    queryFn: async () => { const r = await api.get('/drivers'); return r.data.data ?? [] },
  })

  const { data: devices = [] } = useQuery<Device[]>({
    queryKey: ['devices-fleet'],
    queryFn: async () => { const r = await api.get('/devices'); return r.data.data ?? [] },
  })

  const online = devices.filter((d) => d.online).length

  const tabs = [
    { key: 'vehicles', icon: Truck,  label: 'Vehicles', count: vehicles.length },
    { key: 'drivers',  icon: Users,  label: 'Drivers',  count: drivers.length },
    { key: 'devices',  icon: Cpu,    label: 'Devices',  count: devices.length },
    { key: 'groups',   icon: Grid,   label: 'Groups',   count: 0 },
  ] as const

  return (
    <div className="page" style={{ padding: 'var(--space-6)', height: '100%', overflowY: 'auto' }}>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Fleet Management</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>
            {vehicles.length} vehicles · {drivers.length} drivers · {online}/{devices.length} devices online
          </p>
        </div>
        <button id="add-fleet-item" className="btn btn-primary">
          <Plus size={15} /> Add {tab.slice(0, -1).replace(/^\w/, c => c.toUpperCase())}
        </button>
      </div>

      {/* Stat grid */}
      <div className="stat-grid" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="stat-card success">
          <div className="stat-label">Online Devices</div>
          <div className="stat-value" style={{ color: 'var(--color-success)' }}>{online}</div>
        </div>
        <div className="stat-card accent">
          <div className="stat-label">Vehicles</div>
          <div className="stat-value">{vehicles.length}</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Drivers</div>
          <div className="stat-value">{drivers.length}</div>
        </div>
        <div className="stat-card warning">
          <div className="stat-label">Offline Devices</div>
          <div className="stat-value" style={{ color: devices.length - online > 0 ? 'var(--color-warning)' : undefined }}>
            {devices.length - online}
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)' }}>
        {tabs.map(({ key, icon: Icon, label, count }) => (
          <button
            key={key}
            id={`fleet-tab-${key}`}
            className={`btn btn-sm ${tab === key ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setTab(key as Tab)}
          >
            <Icon size={13} /> {label}
            <span className="nav-badge" style={{ position: 'static', marginLeft: 4 }}>{count}</span>
          </button>
        ))}
      </div>

      {/* Vehicles */}
      {tab === 'vehicles' && (
        <div className="card animate-fade-in">
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Registration</th>
                  <th>Make / Model</th>
                  <th>Year</th>
                  <th>Device</th>
                  <th>Added</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {vehicles.length === 0 && (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No vehicles registered</td></tr>
                )}
                {vehicles.map((v) => (
                  <tr key={v.id} id={`vehicle-row-${v.id}`}>
                    <td><strong style={{ color: 'var(--color-accent)' }}>{v.registration}</strong></td>
                    <td>{[v.make, v.model].filter(Boolean).join(' ') || '—'}</td>
                    <td>{v.year || '—'}</td>
                    <td>
                      {v.device_id
                        ? <span className="badge badge-success">Assigned</span>
                        : <span className="badge badge-muted">None</span>}
                    </td>
                    <td className="text-xs text-muted">{new Date(v.created_at).toLocaleDateString()}</td>
                    <td>
                      <button
                        id={`view-vehicle-${v.id}`}
                        className="btn btn-ghost btn-sm"
                        onClick={() => navigate(`/fleet/vehicles/${v.id}`)}
                      >
                        <ExternalLink size={13} /> Details
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Drivers */}
      {tab === 'drivers' && (
        <div className="card animate-fade-in">
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Name</th>
                  <th>License</th>
                  <th>Phone</th>
                  <th>Added</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {drivers.length === 0 && (
                  <tr><td colSpan={5} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No drivers registered</td></tr>
                )}
                {drivers.map((d) => (
                  <tr key={d.id} id={`driver-row-${d.id}`}>
                    <td>
                      <div className="flex items-center gap-2">
                        <div className="user-avatar" style={{ width: 28, height: 28, fontSize: 12 }}>
                          {d.name[0].toUpperCase()}
                        </div>
                        <strong>{d.name}</strong>
                      </div>
                    </td>
                    <td className="font-mono text-sm">{d.license_number ?? '—'}</td>
                    <td>{d.phone ?? '—'}</td>
                    <td className="text-xs text-muted">{new Date(d.created_at).toLocaleDateString()}</td>
                    <td>
                      <button
                        id={`view-driver-${d.id}`}
                        className="btn btn-ghost btn-sm"
                        onClick={() => navigate(`/fleet/drivers/${d.id}`)}
                      >
                        <ExternalLink size={13} /> Details
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Devices */}
      {tab === 'devices' && (
        <div className="card animate-fade-in">
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Name</th>
                  <th>IMEI</th>
                  <th>Protocol</th>
                  <th>Last Seen</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {devices.length === 0 && (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No devices registered</td></tr>
                )}
                {devices.map((dev) => (
                  <tr key={dev.id} id={`device-row-fleet-${dev.id}`}>
                    <td>
                      <div className="flex items-center gap-2">
                        <div className={`dot ${dev.online ? 'dot-online' : 'dot-offline'}`} />
                        <span className={`badge ${dev.online ? 'badge-success' : 'badge-muted'}`}>
                          {dev.online ? 'online' : 'offline'}
                        </span>
                      </div>
                    </td>
                    <td><strong>{dev.name}</strong></td>
                    <td className="font-mono text-sm">{dev.imei}</td>
                    <td><span className="badge badge-info">{dev.protocol}</span></td>
                    <td className="text-xs text-muted">
                      {dev.last_seen_at ? new Date(dev.last_seen_at).toLocaleString() : 'Never'}
                    </td>
                    <td>
                      <button className="btn btn-ghost btn-sm" title="Signal">
                        <Signal size={13} />
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Groups */}
      {tab === 'groups' && (
        <div className="card animate-fade-in">
          <div style={{ padding: 48, textAlign: 'center', color: 'var(--color-text-muted)' }}>
            <Grid size={40} style={{ margin: '0 auto 16px', display: 'block', opacity: 0.4 }} />
            <p>Vehicle groups let you apply alert rules and view fleet subsets together.</p>
            <button className="btn btn-primary" style={{ marginTop: 16 }}>
              <Plus size={14} /> Create Group
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
