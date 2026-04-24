import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import api from '../../shared/api/client'
import { LogIn, User, Car, Shield, Clock, Filter, Download } from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface AuditEvent {
  id: string
  user_id?: string
  action: string
  resource_type?: string
  resource_id?: string
  ip_address?: string
  status: 'success' | 'failed' | 'denied'
  created_at: string
}

const actionIcon: Record<string, React.ElementType> = {
  login: LogIn, logout: LogIn,
  create: Car, update: Car, delete: Car,
  export_data: Download,
}

const statusColor: Record<string, string> = {
  success: 'badge-success', failed: 'badge-danger', denied: 'badge-warning',
}

const resourceBadge: Record<string, string> = {
  vehicle: 'badge-info', driver: 'badge-success', device: 'badge-warning',
  alert: 'badge-danger', user: 'badge-muted',
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function ActivityPage() {
  const [tab, setTab] = useState<'audit' | 'logins' | 'security'>('audit')
  const [filterAction, setFilterAction] = useState('')

  const { data: events = [], isLoading } = useQuery<AuditEvent[]>({
    queryKey: ['audit', tab],
    queryFn: async () => (await api.get('/audit')).data.data ?? [],
    refetchInterval: 30_000,
  })

  const filtered = filterAction
    ? events.filter(e => e.action.includes(filterAction))
    : events

  // Simulated login history
  const loginEvents = [
    { id: '1', email: 'admin@company.com', ip: '203.0.113.42', ua: 'Chrome/Windows', ts: new Date().toISOString(), status: 'success' },
    { id: '2', email: 'manager@company.com', ip: '203.0.113.18', ua: 'Safari/macOS', ts: new Date(Date.now() - 3600000).toISOString(), status: 'success' },
    { id: '3', email: 'unknown@hack.com', ip: '185.234.218.40', ua: 'curl/7.68.0', ts: new Date(Date.now() - 7200000).toISOString(), status: 'failed' },
  ]

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Activity & Audit Logs</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>M16 · Track all user actions, logins, and security events</p>
        </div>
        <button id="audit-export" className="btn btn-secondary">
          <Download size={14} /> Export Logs
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)' }}>
        {[
          { key: 'audit', icon: User, label: 'Audit Trail' },
          { key: 'logins', icon: LogIn, label: 'Login History' },
          { key: 'security', icon: Shield, label: 'Security Events' },
        ].map(({ key, icon: Icon, label }) => (
          <button key={key} id={`activity-tab-${key}`}
            className={`btn ${tab === key ? 'btn-primary' : 'btn-ghost'}`}
            style={{ borderRadius: '6px 6px 0 0' }}
            onClick={() => setTab(key as any)}>
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* Filters */}
      {tab === 'audit' && (
        <div className="flex gap-2" style={{ marginBottom: 'var(--space-4)' }}>
          <Filter size={16} style={{ alignSelf: 'center', color: 'var(--color-text-muted)' }} />
          {['', 'login', 'create', 'update', 'delete', 'export_data'].map(action => (
            <button key={action}
              id={`audit-filter-${action || 'all'}`}
              className={`btn btn-sm ${filterAction === action ? 'btn-primary' : 'btn-secondary'}`}
              onClick={() => setFilterAction(action)}>
              {action || 'All'}
            </button>
          ))}
        </div>
      )}

      {/* ── Tab: Audit Trail ── */}
      {tab === 'audit' && (
        <div className="card animate-fade-in">
          {isLoading ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}><div className="spinner" /></div>
          ) : filtered.length === 0 ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <Shield size={40} color="var(--color-text-muted)" style={{ margin: '0 auto 12px' }} />
              <p className="text-muted">No audit events recorded yet.</p>
            </div>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Timestamp</th>
                    <th>Action</th>
                    <th>Resource</th>
                    <th>IP Address</th>
                    <th>Status</th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map(ev => {
                    const Icon = actionIcon[ev.action?.split('.')[0]] ?? User
                    return (
                      <tr key={ev.id}>
                        <td>
                          <span className="text-xs font-mono text-muted">
                            {new Date(ev.created_at).toLocaleString()}
                          </span>
                        </td>
                        <td>
                          <div className="flex items-center gap-2">
                            <Icon size={14} color="var(--color-primary)" />
                            <span className="text-sm font-medium">{ev.action}</span>
                          </div>
                        </td>
                        <td>
                          {ev.resource_type ? (
                            <span className={`badge ${resourceBadge[ev.resource_type] ?? 'badge-muted'}`}>
                              {ev.resource_type}
                            </span>
                          ) : '—'}
                        </td>
                        <td className="font-mono text-sm">{ev.ip_address || '—'}</td>
                        <td>
                          <span className={`badge ${statusColor[ev.status] ?? 'badge-muted'}`}>{ev.status}</span>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ── Tab: Login History ── */}
      {tab === 'logins' && (
        <div className="card animate-fade-in">
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>User</th>
                  <th>IP Address</th>
                  <th>Browser / OS</th>
                  <th>Time</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {loginEvents.map(ev => (
                  <tr key={ev.id}>
                    <td className="text-sm">{ev.email}</td>
                    <td className="font-mono text-sm">{ev.ip}</td>
                    <td className="text-sm text-muted">{ev.ua}</td>
                    <td>
                      <span className="text-xs font-mono text-muted">{new Date(ev.ts).toLocaleString()}</span>
                    </td>
                    <td>
                      <span className={`badge ${ev.status === 'success' ? 'badge-success' : 'badge-danger'}`}>
                        {ev.status}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── Tab: Security Events ── */}
      {tab === 'security' && (
        <div className="animate-fade-in" style={{ display: 'grid', gap: 'var(--space-3)' }}>
          {[
            { title: 'Failed Login Attempts', count: 3, color: 'var(--color-danger)', icon: LogIn, sub: 'From 185.234.218.40 in the last 24h' },
            { title: 'API Rate Limit Exceeded', count: 0, color: 'var(--color-warning)', icon: Shield, sub: 'No rate limit events this week' },
            { title: 'Admin Actions', count: 12, color: 'var(--color-info)', icon: User, sub: 'Last action 2 hours ago' },
            { title: '2FA Bypasses Attempted', count: 0, color: 'var(--color-success)', icon: Clock, sub: 'None detected' },
          ].map(({ title, count, color, icon: Icon, sub }) => (
            <div key={title} className="card" style={{ padding: 'var(--space-5)', display: 'flex', alignItems: 'center', gap: 'var(--space-4)' }}>
              <div style={{
                width: 44, height: 44, borderRadius: 10,
                background: `${color}18`, display: 'grid', placeItems: 'center', flexShrink: 0,
              }}>
                <Icon size={20} color={color} />
              </div>
              <div style={{ flex: 1 }}>
                <div className="flex items-center gap-2">
                  <span className="font-medium">{title}</span>
                  <span style={{
                    padding: '2px 8px', borderRadius: 99, fontSize: 12, fontWeight: 700,
                    background: `${color}18`, color,
                  }}>{count}</span>
                </div>
                <p className="text-sm text-muted">{sub}</p>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
