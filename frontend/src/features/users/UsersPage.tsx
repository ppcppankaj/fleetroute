import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../shared/api/client'
import { Users, Shield, Key, Plus, Trash2, Edit2, CheckCircle, Clock } from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface User {
  id: string
  email: string
  first_name: string
  last_name: string
  role: 'admin' | 'manager' | 'dispatcher' | 'driver' | 'viewer'
  status: 'active' | 'inactive' | 'suspended'
  last_login_at?: string
  created_at: string
}

interface Role {
  name: string
  permissions: string[]
}

const roleColor: Record<string, string> = {
  admin: 'badge-danger', manager: 'badge-warning',
  dispatcher: 'badge-info', driver: 'badge-success', viewer: 'badge-muted',
}

const statusColor: Record<string, string> = {
  active: 'badge-success', inactive: 'badge-muted', suspended: 'badge-danger',
}

/* ─── Invite User Modal ──────────────────────────────────────────────────── */
function InviteModal({ roles, onClose }: { roles: Role[]; onClose: () => void }) {
  const qc = useQueryClient()
  const [form, setForm] = useState({ email: '', first_name: '', last_name: '', role: 'viewer', password: '' })
  const set = (k: string, v: string) => setForm(f => ({ ...f, [k]: v }))

  const mutation = useMutation({
    mutationFn: (data: typeof form) => api.post('/users', data),
    onSuccess: () => { qc.invalidateQueries({ queryKey: ['users'] }); onClose() },
  })

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal" style={{ width: 460 }} onClick={e => e.stopPropagation()}>
        <div className="modal-header">
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>Invite User</h2>
          <button className="btn btn-ghost" onClick={onClose}>✕</button>
        </div>
        <div className="modal-body" style={{ display: 'grid', gap: 'var(--space-4)' }}>
          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">First Name</label>
              <input className="input" placeholder="Raj" value={form.first_name} onChange={e => set('first_name', e.target.value)} />
            </div>
            <div style={{ display: 'grid', gap: 6 }}>
              <label className="text-sm font-medium">Last Name</label>
              <input className="input" placeholder="Sharma" value={form.last_name} onChange={e => set('last_name', e.target.value)} />
            </div>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            <label className="text-sm font-medium">Email *</label>
            <input className="input" type="email" placeholder="raj.sharma@company.com"
              value={form.email} onChange={e => set('email', e.target.value)} />
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            <label className="text-sm font-medium">Role *</label>
            <select className="input" value={form.role} onChange={e => set('role', e.target.value)}>
              {roles.map(r => <option key={r.name} value={r.name} style={{ textTransform: 'capitalize' }}>{r.name}</option>)}
            </select>
          </div>
          <div style={{ display: 'grid', gap: 6 }}>
            <label className="text-sm font-medium">Temporary Password *</label>
            <input className="input" type="password" placeholder="Min 8 characters"
              value={form.password} onChange={e => set('password', e.target.value)} />
          </div>
          {form.role && (
            <div className="card" style={{ background: 'var(--color-surface-raised)', padding: 'var(--space-3)' }}>
              <p className="text-sm text-muted font-medium" style={{ marginBottom: 6 }}>Permissions for <strong>{form.role}</strong>:</p>
              <div className="flex flex-wrap gap-1">
                {(roles.find(r => r.name === form.role)?.permissions ?? []).slice(0, 8).map(p => (
                  <span key={p} className="badge badge-info" style={{ fontSize: 10 }}>{p}</span>
                ))}
              </div>
            </div>
          )}
        </div>
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose}>Cancel</button>
          <button id="user-invite-save" className="btn btn-primary"
            disabled={!form.email || !form.password || mutation.isPending}
            onClick={() => mutation.mutate(form)}>
            {mutation.isPending ? 'Inviting…' : 'Send Invite'}
          </button>
        </div>
      </div>
    </div>
  )
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function UsersPage() {
  const [tab, setTab] = useState<'users' | 'roles' | 'api-keys'>('users')
  const [showInvite, setShowInvite] = useState(false)
  const qc = useQueryClient()

  const { data: users = [], isLoading } = useQuery<User[]>({
    queryKey: ['users'],
    queryFn: async () => (await api.get('/users')).data.data ?? [],
  })

  const { data: roles = [] } = useQuery<Role[]>({
    queryKey: ['roles'],
    queryFn: async () => (await api.get('/roles')).data.data ?? [],
  })

  const deleteUser = useMutation({
    mutationFn: (id: string) => api.delete(`/users/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
  })

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {showInvite && <InviteModal roles={roles} onClose={() => setShowInvite(false)} />}

      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Users & Access</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>M11 · Manage users, roles, and API keys</p>
        </div>
        <button id="users-invite" className="btn btn-primary" onClick={() => setShowInvite(true)}>
          <Plus size={14} /> Invite User
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)' }}>
        {[
          { key: 'users', icon: Users, label: `Users (${users.length})` },
          { key: 'roles', icon: Shield, label: 'Roles & Permissions' },
          { key: 'api-keys', icon: Key, label: 'API Keys' },
        ].map(({ key, icon: Icon, label }) => (
          <button key={key} id={`users-tab-${key}`}
            className={`btn ${tab === key ? 'btn-primary' : 'btn-ghost'}`}
            style={{ borderRadius: '6px 6px 0 0' }}
            onClick={() => setTab(key as any)}>
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* ── Tab: Users ── */}
      {tab === 'users' && (
        <div className="card animate-fade-in">
          {isLoading ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}><div className="spinner" /></div>
          ) : users.length === 0 ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <Users size={40} color="var(--color-text-muted)" style={{ margin: '0 auto 12px' }} />
              <p className="text-muted">No users yet. Invite your first team member.</p>
            </div>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>User</th>
                    <th>Email</th>
                    <th>Role</th>
                    <th>Status</th>
                    <th>Last Login</th>
                    <th>Joined</th>
                    <th>Actions</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map(u => (
                    <tr key={u.id}>
                      <td>
                        <div className="flex items-center gap-2">
                          <div style={{
                            width: 32, height: 32, borderRadius: '50%',
                            background: 'var(--color-primary)', color: '#fff',
                            display: 'grid', placeItems: 'center', fontSize: 12, fontWeight: 700,
                            flexShrink: 0,
                          }}>
                            {(u.first_name?.[0] ?? u.email[0]).toUpperCase()}
                          </div>
                          <span className="font-medium">{u.first_name} {u.last_name}</span>
                        </div>
                      </td>
                      <td className="text-sm">{u.email}</td>
                      <td>
                        <span className={`badge ${roleColor[u.role] ?? 'badge-muted'}`} style={{ textTransform: 'capitalize' }}>
                          {u.role}
                        </span>
                      </td>
                      <td>
                        <span className={`badge ${statusColor[u.status] ?? 'badge-muted'}`}>
                          {u.status === 'active' ? <CheckCircle size={10} /> : <Clock size={10} />} {u.status}
                        </span>
                      </td>
                      <td className="text-xs text-muted">
                        {u.last_login_at ? new Date(u.last_login_at).toLocaleString() : 'Never'}
                      </td>
                      <td className="text-xs text-muted">
                        {new Date(u.created_at).toLocaleDateString()}
                      </td>
                      <td>
                        <div className="flex gap-1">
                          <button id={`user-edit-${u.id}`} className="btn btn-secondary btn-sm" title="Edit">
                            <Edit2 size={12} />
                          </button>
                          <button id={`user-delete-${u.id}`} className="btn btn-danger btn-sm" title="Remove"
                            onClick={() => {
                              if (confirm(`Remove ${u.email}?`)) deleteUser.mutate(u.id)
                            }}>
                            <Trash2 size={12} />
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
      )}

      {/* ── Tab: Roles ── */}
      {tab === 'roles' && (
        <div style={{ display: 'grid', gap: 'var(--space-4)' }} className="animate-fade-in">
          {roles.map(role => (
            <div key={role.name} className="card" style={{ padding: 'var(--space-5)' }}>
              <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-3)' }}>
                <div className="flex items-center gap-3">
                  <Shield size={18} color="var(--color-primary)" />
                  <span className="font-semibold" style={{ textTransform: 'capitalize' }}>{role.name}</span>
                  <span className="badge badge-muted">{role.permissions.length} permissions</span>
                </div>
              </div>
              <div className="flex flex-wrap gap-1">
                {role.permissions.map(p => (
                  <span key={p} className="badge badge-info" style={{ fontSize: 11 }}>{p}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── Tab: API Keys ── */}
      {tab === 'api-keys' && (
        <div className="card animate-fade-in" style={{ padding: 'var(--space-8)', textAlign: 'center' }}>
          <Key size={40} color="var(--color-primary)" style={{ margin: '0 auto 16px' }} />
          <p className="font-medium" style={{ marginBottom: 8 }}>API Key Management</p>
          <p className="text-sm text-muted" style={{ marginBottom: 'var(--space-4)' }}>
            Generate API keys for programmatic access. Keys are hashed and only shown once.
          </p>
          <button id="apikey-generate" className="btn btn-primary">
            <Plus size={14} /> Generate API Key
          </button>
        </div>
      )}
    </div>
  )
}
