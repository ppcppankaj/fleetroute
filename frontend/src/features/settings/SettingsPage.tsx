import { useState } from 'react'
import {
  Users, Shield, Bell, Key, Webhook, Palette, CreditCard,
  Plus, Trash2, CheckCircle, Copy, Eye, EyeOff
} from 'lucide-react'

type Tab = 'users' | 'roles' | 'notifications' | 'rules' | 'apikeys' | 'integrations' | 'whitelabel' | 'billing'

const PERMISSIONS = ['devices.view','devices.edit','vehicles.view','vehicles.edit',
  'drivers.view','drivers.edit','geofences.view','geofences.edit',
  'alerts.view','alerts.acknowledge','reports.view','reports.export',
  'maintenance.view','maintenance.edit','settings.view','settings.edit',
  'commands.send',
]

const ROLES = [
  { name: 'Tenant Admin',   perms: PERMISSIONS },
  { name: 'Fleet Manager',  perms: PERMISSIONS.filter((p) => !p.includes('settings')) },
  { name: 'Dispatcher',     perms: ['devices.view','vehicles.view','alerts.view','alerts.acknowledge'] },
  { name: 'Driver',         perms: ['devices.view','vehicles.view'] },
]

export function SettingsPage() {
  const [tab, setTab] = useState<Tab>('users')
  const [showKey, setShowKey] = useState(false)
  const [primaryColor, setPrimaryColor] = useState('#6366f1')
  const [companyName, setCompanyName] = useState('FleetOS')
  const [logoURL, setLogoURL] = useState('')

  const tabs = [
    { key: 'users',         icon: Users,      label: 'Users' },
    { key: 'roles',         icon: Shield,     label: 'Roles' },
    { key: 'notifications', icon: Bell,       label: 'Notifications' },
    { key: 'rules',         icon: Bell,       label: 'Alert Rules' },
    { key: 'apikeys',       icon: Key,        label: 'API Keys' },
    { key: 'integrations',  icon: Webhook,    label: 'Integrations' },
    { key: 'whitelabel',    icon: Palette,    label: 'White Label' },
    { key: 'billing',       icon: CreditCard, label: 'Billing' },
  ] as const

  return (
    <div className="page" style={{ padding: 'var(--space-6)', height: '100%', overflowY: 'auto' }}>
      <div style={{ marginBottom: 'var(--space-6)' }}>
        <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Settings</h1>
        <p className="text-muted text-sm" style={{ marginTop: 4 }}>Manage your organization configuration</p>
      </div>

      <div style={{ display: 'grid', gridTemplateColumns: '200px 1fr', gap: 24 }}>
        {/* Side nav */}
        <div className="card" style={{ height: 'fit-content', padding: 8 }}>
          {tabs.map(({ key, icon: Icon, label }) => (
            <button
              key={key}
              id={`settings-tab-${key}`}
              onClick={() => setTab(key as Tab)}
              className={`nav-item${tab === key ? ' active' : ''}`}
              style={{ width: '100%', marginBottom: 2 }}
            >
              <Icon size={15} className="nav-icon" />
              {label}
            </button>
          ))}
        </div>

        {/* Content */}
        <div>
          {/* Users */}
          {tab === 'users' && (
            <div className="card animate-fade-in">
              <div className="card-header">
                <h2 className="card-title">Users</h2>
                <button id="invite-user" className="btn btn-primary btn-sm"><Plus size={13} /> Invite User</button>
              </div>
              <div className="table-wrap">
                <table>
                  <thead>
                    <tr><th>User</th><th>Role</th><th>Status</th><th>Actions</th></tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>
                        <div className="flex items-center gap-2">
                          <div className="user-avatar" style={{ width: 28, height: 28, fontSize: 12 }}>A</div>
                          <div>
                            <div style={{ fontWeight: 500 }}>admin@gpsgo.com</div>
                            <div className="text-xs text-muted">Super Admin</div>
                          </div>
                        </div>
                      </td>
                      <td><span className="badge badge-info">Tenant Admin</span></td>
                      <td><span className="badge badge-success">Active</span></td>
                      <td>—</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Roles */}
          {tab === 'roles' && (
            <div className="card animate-fade-in">
              <div className="card-header">
                <h2 className="card-title">Role Permissions Matrix</h2>
                <button className="btn btn-primary btn-sm"><Plus size={13} /> Custom Role</button>
              </div>
              <div style={{ padding: '0 16px 16px', overflowX: 'auto' }}>
                <table style={{ minWidth: 700 }}>
                  <thead>
                    <tr>
                      <th style={{ minWidth: 180 }}>Permission</th>
                      {ROLES.map((r) => <th key={r.name}>{r.name}</th>)}
                    </tr>
                  </thead>
                  <tbody>
                    {PERMISSIONS.map((perm) => (
                      <tr key={perm}>
                        <td className="font-mono text-xs">{perm}</td>
                        {ROLES.map((r) => (
                          <td key={r.name} style={{ textAlign: 'center' }}>
                            {r.perms.includes(perm)
                              ? <CheckCircle size={14} style={{ color: 'var(--color-success)' }} />
                              : <span style={{ color: 'var(--color-text-muted)', fontSize: 16 }}>–</span>}
                          </td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Notifications */}
          {tab === 'notifications' && (
            <div className="animate-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              {[
                { type: 'email',   label: 'Email Notifications', placeholder: 'alerts@company.com' },
                { type: 'sms',     label: 'SMS Notifications',   placeholder: '+91 9876543210' },
                { type: 'webhook', label: 'Webhook URL',         placeholder: 'https://hooks.yourapp.com/…' },
              ].map(({ type, label, placeholder }) => (
                <div key={type} className="card">
                  <div className="card-header">
                    <h3 className="card-title">{label}</h3>
                    <button className="btn btn-primary btn-sm"><Plus size={13} /> Add</button>
                  </div>
                  <div style={{ padding: '8px 16px 16px' }}>
                    <div className="flex gap-2">
                      <input
                        type="text"
                        className="form-input"
                        placeholder={placeholder}
                        style={{ flex: 1 }}
                      />
                      <button className="btn btn-secondary btn-sm">Test</button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Alert Rules */}
          {tab === 'rules' && (
            <div className="card animate-fade-in">
              <div className="card-header">
                <h2 className="card-title">Alert Rules</h2>
                <button id="create-rule" className="btn btn-primary btn-sm"><Plus size={13} /> Create Rule</button>
              </div>
              <div style={{ padding: 24, textAlign: 'center', color: 'var(--color-text-muted)' }}>
                <p>Configure rules from built-in templates or create custom condition trees.</p>
                <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))', gap: 12, marginTop: 16 }}>
                  {[
                    'Overspeed Alert', 'Geofence Entry', 'Geofence Exit',
                    'Excessive Idling', 'Harsh Driving', 'Fuel Theft Detection',
                    'Device Tamper', 'SOS Button', 'Power Cut', 'Device Offline',
                  ].map((t) => (
                    <button key={t} className="btn btn-secondary" style={{ textAlign: 'left' }}>
                      + {t}
                    </button>
                  ))}
                </div>
              </div>
            </div>
          )}

          {/* API Keys */}
          {tab === 'apikeys' && (
            <div className="card animate-fade-in">
              <div className="card-header">
                <h2 className="card-title">API Keys</h2>
                <button id="generate-api-key" className="btn btn-primary btn-sm"><Plus size={13} /> Generate Key</button>
              </div>
              <div style={{ padding: 16 }}>
                <div className="flex items-center gap-2" style={{ padding: 12, background: 'var(--color-surface-raised)', borderRadius: 8 }}>
                  <span className="font-mono text-sm" style={{ flex: 1 }}>
                    {showKey ? 'gps_abc123def456ghi789jkl012mno345pqr678stu' : 'gps_abc123••••••••••••••••••••••••••••••••'}
                  </span>
                  <button className="btn btn-ghost btn-sm" onClick={() => setShowKey((v) => !v)}>
                    {showKey ? <EyeOff size={13} /> : <Eye size={13} />}
                  </button>
                  <button className="btn btn-ghost btn-sm" title="Copy">
                    <Copy size={13} />
                  </button>
                  <button className="btn btn-ghost btn-sm" title="Revoke">
                    <Trash2 size={13} />
                  </button>
                </div>
                <p className="text-xs text-muted" style={{ marginTop: 8 }}>
                  Rate limit: 1000 req/min · Permissions: read, write · Created: today
                </p>
              </div>
            </div>
          )}

          {/* Integrations */}
          {tab === 'integrations' && (
            <div className="card animate-fade-in">
              <div className="card-header">
                <h2 className="card-title">Webhook Endpoints</h2>
                <button className="btn btn-primary btn-sm"><Plus size={13} /> Add Endpoint</button>
              </div>
              <div style={{ padding: 16 }}>
                <div style={{ display: 'grid', gap: 16 }}>
                  {['alert', 'trip', 'geofence'].map((ev) => (
                    <label key={ev} className="flex items-center gap-2">
                      <input type="checkbox" defaultChecked />
                      <span className="text-sm">Forward <strong>{ev}</strong> events</span>
                    </label>
                  ))}
                  <input type="url" className="form-input" placeholder="https://your-endpoint.com/webhook" />
                  <input type="text" className="form-input" placeholder="Authorization header value (optional)" />
                  <button className="btn btn-secondary btn-sm" style={{ width: 'fit-content' }}>Test Delivery</button>
                </div>
              </div>
            </div>
          )}

          {/* White Label */}
          {tab === 'whitelabel' && (
            <div className="card animate-fade-in">
              <div className="card-header"><h2 className="card-title">White Label Customization</h2></div>
              <div style={{ padding: 20, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
                <label>
                  <span className="text-sm text-muted" style={{ display: 'block', marginBottom: 6 }}>Company Name</span>
                  <input
                    type="text"
                    className="form-input"
                    value={companyName}
                    onChange={(e) => setCompanyName(e.target.value)}
                  />
                </label>
                <label>
                  <span className="text-sm text-muted" style={{ display: 'block', marginBottom: 6 }}>Primary Color</span>
                  <div className="flex gap-2 items-center">
                    <input
                      type="color"
                      value={primaryColor}
                      onChange={(e) => setPrimaryColor(e.target.value)}
                      style={{ width: 40, height: 36, border: 'none', borderRadius: 6, cursor: 'pointer' }}
                    />
                    <input
                      type="text"
                      className="form-input"
                      value={primaryColor}
                      onChange={(e) => setPrimaryColor(e.target.value)}
                      style={{ flex: 1 }}
                    />
                  </div>
                </label>
                <label style={{ gridColumn: '1 / -1' }}>
                  <span className="text-sm text-muted" style={{ display: 'block', marginBottom: 6 }}>Logo URL</span>
                  <input
                    type="url"
                    className="form-input"
                    placeholder="https://your-company.com/logo.png"
                    value={logoURL}
                    onChange={(e) => setLogoURL(e.target.value)}
                  />
                </label>
                <div style={{ gridColumn: '1 / -1' }}>
                  <div style={{ padding: 20, background: 'var(--color-surface-raised)', borderRadius: 8, display: 'flex', alignItems: 'center', gap: 12 }}>
                    {logoURL && <img src={logoURL} alt="Logo preview" style={{ height: 36, objectFit: 'contain' }} />}
                    <div style={{ fontWeight: 700, color: primaryColor, fontSize: 20 }}>{companyName}</div>
                  </div>
                  <p className="text-xs text-muted" style={{ marginTop: 6 }}>Preview of branded sidebar logo</p>
                </div>
                <button id="save-whitelabel" className="btn btn-primary" style={{ gridColumn: '1 / -1', width: 'fit-content' }}>
                  Save Branding
                </button>
              </div>
            </div>
          )}

          {/* Billing */}
          {tab === 'billing' && (
            <div className="animate-fade-in" style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
              <div className="card">
                <div className="card-header"><h2 className="card-title">Current Plan</h2></div>
                <div style={{ padding: 20 }}>
                  <div className="flex items-center gap-4">
                    <div>
                      <div style={{ fontSize: 24, fontWeight: 700 }}>Professional</div>
                      <div className="text-muted text-sm">Up to 500 devices · Full feature access</div>
                    </div>
                    <div style={{ marginLeft: 'auto', textAlign: 'right' }}>
                      <div style={{ fontSize: 28, fontWeight: 700, color: 'var(--color-accent)' }}>₹12,999<span style={{ fontSize: 14, fontWeight: 400 }}>/mo</span></div>
                    </div>
                  </div>
                  <div style={{ marginTop: 16 }}>
                    <div className="flex justify-between text-sm" style={{ marginBottom: 6 }}>
                      <span>Devices used</span>
                      <span><strong>47</strong> / 500</span>
                    </div>
                    <div style={{ height: 6, background: 'var(--color-surface-raised)', borderRadius: 3 }}>
                      <div style={{ width: '9.4%', height: '100%', background: 'var(--color-accent)', borderRadius: 3 }} />
                    </div>
                  </div>
                </div>
              </div>
              <div className="card">
                <div className="card-header"><h2 className="card-title">Invoices</h2></div>
                <div className="table-wrap">
                  <table>
                    <thead><tr><th>Period</th><th>Amount</th><th>Status</th><th>Download</th></tr></thead>
                    <tbody>
                      {['April 2026', 'March 2026', 'February 2026'].map((month, i) => (
                        <tr key={month}>
                          <td>{month}</td>
                          <td>₹12,999</td>
                          <td><span className="badge badge-success">Paid</span></td>
                          <td><button className="btn btn-ghost btn-sm">PDF</button></td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
