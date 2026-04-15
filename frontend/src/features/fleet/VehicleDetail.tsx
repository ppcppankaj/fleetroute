import { useState } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft, MapPin, Navigation, Zap, Fuel, Gauge, Play } from 'lucide-react'
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar
} from 'recharts'
import api from '../../shared/api/client'

type Tab = 'overview' | 'livedata' | 'trips' | 'driver_history' | 'maintenance' | 'statistics'

export function VehicleDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [tab, setTab] = useState<Tab>('overview')

  const { data: vehicle } = useQuery({
    queryKey: ['vehicle', id],
    queryFn: async () => { const r = await api.get(`/vehicles/${id}`); return r.data.data },
    enabled: !!id,
  })

  const { data: trips = [] } = useQuery({
    queryKey: ['vehicle-trips', id],
    queryFn: async () => {
      const r = await api.get('/reports/trips', { params: { vehicle_id: id, limit: 30 } })
      return r.data.data ?? []
    },
    enabled: !!id && tab === 'trips',
  })

  // Mock statistics data
  const utilizationData = Array.from({ length: 7 }, (_, i) => ({
    day: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'][i],
    hours: Math.random() * 10 + 2,
    distance: Math.random() * 200 + 50,
  }))

  const tabs = [
    { key: 'overview',       label: 'Overview' },
    { key: 'livedata',       label: 'Live Data' },
    { key: 'trips',          label: 'Trip History' },
    { key: 'driver_history', label: 'Drivers' },
    { key: 'maintenance',    label: 'Maintenance' },
    { key: 'statistics',     label: 'Statistics' },
  ] as const

  return (
    <div className="page" style={{ padding: 'var(--space-6)', height: '100%', overflowY: 'auto' }}>
      {/* Header */}
      <div className="flex items-center gap-3" style={{ marginBottom: 'var(--space-6)' }}>
        <button className="btn btn-ghost btn-sm" onClick={() => navigate('/fleet')}>
          <ArrowLeft size={15} /> Fleet
        </button>
        <div style={{ flex: 1 }}>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>
            {vehicle ? vehicle[2] : 'Vehicle'}
          </h1>
          <p className="text-muted text-sm">Vehicle Detail</p>
        </div>
        <button
          className="btn btn-secondary btn-sm"
          onClick={() => navigate(`/playback/${vehicle?.[6] ?? id}`)}
        >
          <Play size={13} /> Route Playback
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', flexWrap: 'wrap' }}>
        {tabs.map(({ key, label }) => (
          <button
            key={key}
            id={`vehicle-tab-${key}`}
            className={`btn btn-sm ${tab === key ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setTab(key as Tab)}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Overview */}
      {tab === 'overview' && vehicle && (
        <div className="animate-fade-in" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          <div className="card">
            <div className="card-header"><h2 className="card-title">Vehicle Info</h2></div>
            <div style={{ padding: '0 16px 16px' }}>
              {[
                ['Registration', vehicle[2]],
                ['Make', vehicle[3]],
                ['Model', vehicle[4]],
                ['Year', vehicle[5]],
                ['Device ID', vehicle[6] ?? 'Not assigned'],
                ['Tenant', vehicle[1]],
              ].map(([label, value]) => (
                <div key={label as string} className="flex justify-between" style={{ padding: '8px 0', borderBottom: '1px solid var(--color-border)' }}>
                  <span className="text-sm text-muted">{label as string}</span>
                  <span className="text-sm" style={{ fontWeight: 500 }}>{value as string}</span>
                </div>
              ))}
            </div>
          </div>
          <div className="card">
            <div className="card-header"><h2 className="card-title">Quick Stats</h2></div>
            <div className="stat-grid" style={{ padding: '0 16px 16px', gridTemplateColumns: '1fr 1fr' }}>
              <div className="stat-card"><div className="stat-label">Total Trips</div><div className="stat-value">—</div></div>
              <div className="stat-card"><div className="stat-label">Total Distance</div><div className="stat-value">—</div></div>
              <div className="stat-card"><div className="stat-label">Avg Score</div><div className="stat-value">—</div></div>
              <div className="stat-card"><div className="stat-label">Fuel Used</div><div className="stat-value">—</div></div>
            </div>
          </div>
        </div>
      )}

      {/* Live Data */}
      {tab === 'livedata' && (
        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Live Telemetry</h2></div>
          <div style={{ padding: 16, display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(160px, 1fr))', gap: 12 }}>
            {[
              { label: 'Speed', value: '— km/h', icon: <Gauge size={18} /> },
              { label: 'Heading', value: '—°', icon: <Navigation size={18} /> },
              { label: 'Location', value: '—', icon: <MapPin size={18} /> },
              { label: 'Ignition', value: '—', icon: <Zap size={18} /> },
              { label: 'Fuel Level', value: '—%', icon: <Fuel size={18} /> },
            ].map((item) => (
              <div key={item.label} className="stat-card" style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 8 }}>
                <span style={{ color: 'var(--color-accent)' }}>{item.icon}</span>
                <div className="stat-label">{item.label}</div>
                <div style={{ fontWeight: 700 }}>{item.value}</div>
              </div>
            ))}
          </div>
          <p className="text-sm text-muted" style={{ padding: '0 16px 16px', textAlign: 'center' }}>
            Live data updates via WebSocket when the vehicle reports.
          </p>
        </div>
      )}

      {/* Trips */}
      {tab === 'trips' && (
        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Trip History</h2></div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr><th>Start</th><th>End</th><th>Distance</th><th>Duration</th><th>Max Speed</th><th>Driver</th></tr>
              </thead>
              <tbody>
                {trips.length === 0 && (
                  <tr><td colSpan={6} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No trips found</td></tr>
                )}
                {(trips as any[]).map((t: any, i: number) => (
                  <tr key={t.id ?? i}>
                    <td className="text-xs">{t.started_at ? new Date(t.started_at).toLocaleString() : '—'}</td>
                    <td className="text-xs">{t.ended_at ? new Date(t.ended_at).toLocaleString() : 'In progress'}</td>
                    <td>{t.distance_m ? `${(t.distance_m/1000).toFixed(1)} km` : '—'}</td>
                    <td>{t.duration_s ? `${Math.floor(t.duration_s/60)} min` : '—'}</td>
                    <td>{t.max_speed ? `${t.max_speed} km/h` : '—'}</td>
                    <td>{t.driver_id ?? '—'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Statistics */}
      {tab === 'statistics' && (
        <div className="animate-fade-in" style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
          <div className="card">
            <div className="card-header"><h2 className="card-title">Daily Distance (km)</h2></div>
            <div style={{ padding: '0 16px 16px' }}>
              <ResponsiveContainer width="100%" height={200}>
                <AreaChart data={utilizationData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                  <XAxis dataKey="day" tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                  <YAxis tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                  <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)' }} />
                  <Area type="monotone" dataKey="distance" stroke="var(--color-accent)" fill="var(--color-accent)" fillOpacity={0.2} />
                </AreaChart>
              </ResponsiveContainer>
            </div>
          </div>
          <div className="card">
            <div className="card-header"><h2 className="card-title">Daily Utilization (hours)</h2></div>
            <div style={{ padding: '0 16px 16px' }}>
              <ResponsiveContainer width="100%" height={200}>
                <BarChart data={utilizationData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                  <XAxis dataKey="day" tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                  <YAxis tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                  <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)' }} />
                  <Bar dataKey="hours" fill="var(--color-accent)" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          </div>
        </div>
      )}

      {/* Driver History / Maintenance placeholders */}
      {(tab === 'driver_history' || tab === 'maintenance') && (
        <div className="card animate-fade-in">
          <div style={{ padding: 48, textAlign: 'center', color: 'var(--color-text-muted)' }}>
            <p>Data will appear here once the vehicle has recorded activity.</p>
          </div>
        </div>
      )}
    </div>
  )
}
