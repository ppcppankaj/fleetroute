import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import api from '../../shared/api/client'
import {
  CreditCard, Package, CheckCircle, Clock, Download, AlertTriangle,
  Zap, BarChart3, HardDrive, Cpu
} from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface Plan {
  id: string
  name: string
  price: number
  currency: string
  billing_cycle: 'monthly' | 'annual'
  features: {
    max_devices: number
    max_users: number
    api_rate_limit?: number
    storage_gb?: number
    video_storage?: boolean
    ai_features?: boolean
  }
}

interface Subscription {
  plan: string
  status: 'trial' | 'active' | 'past_due' | 'cancelled'
  current_period_end: string
  auto_renew: boolean
}

interface Invoice {
  id: string
  invoice_number: string
  total: number
  currency: string
  status: 'paid' | 'sent' | 'overdue' | 'draft'
  due_date: string
  created_at: string
  pdf_url?: string
}

interface Usage {
  devices_count: number
  api_calls: number
  storage_gb: number
  video_hours: number
}

const statusColor: Record<string, string> = {
  active: 'badge-success', trial: 'badge-info',
  past_due: 'badge-danger', cancelled: 'badge-muted',
}

const invoiceStatusColor: Record<string, string> = {
  paid: 'badge-success', sent: 'badge-info',
  overdue: 'badge-danger', draft: 'badge-muted',
}

const fmtCurrency = (n: number, currency = 'INR') =>
  new Intl.NumberFormat('en-IN', { style: 'currency', currency, maximumFractionDigits: 0 }).format(n)

/* ─── Plan Card ──────────────────────────────────────────────────────────── */
function PlanCard({
  plan, current, onSelect,
}: { plan: Plan; current: string; onSelect: (id: string) => void }) {
  const isActive = plan.id === current
  return (
    <div className="card" style={{
      padding: 'var(--space-6)',
      border: isActive ? '2px solid var(--color-primary)' : '1px solid var(--color-border)',
      position: 'relative',
    }}>
      {isActive && (
        <div style={{
          position: 'absolute', top: -12, left: '50%', transform: 'translateX(-50%)',
          background: 'var(--color-primary)', color: '#fff', borderRadius: 99, padding: '2px 12px',
          fontSize: 12, fontWeight: 600,
        }}>
          Current Plan
        </div>
      )}
      <h3 style={{ fontSize: 'var(--text-xl)', fontWeight: 700, marginBottom: 4 }}>{plan.name}</h3>
      <div style={{ marginBottom: 'var(--space-4)' }}>
        <span style={{ fontSize: 'var(--text-3xl)', fontWeight: 800 }}>
          {plan.price === 0 ? 'Custom' : fmtCurrency(plan.price)}
        </span>
        {plan.price > 0 && (
          <span className="text-muted text-sm">/{plan.billing_cycle === 'monthly' ? 'mo' : 'yr'}</span>
        )}
      </div>
      <ul style={{ listStyle: 'none', padding: 0, margin: '0 0 var(--space-6)', display: 'grid', gap: 8 }}>
        {[
          { label: `${plan.features.max_devices === -1 ? 'Unlimited' : plan.features.max_devices} Devices`, icon: Cpu },
          { label: `${plan.features.max_users === -1 ? 'Unlimited' : plan.features.max_users} Users`, icon: BarChart3 },
          { label: `${plan.features.storage_gb ?? 100} GB Storage`, icon: HardDrive },
          ...(plan.features.video_storage ? [{ label: 'Video Telematics', icon: Zap }] : []),
        ].map(({ label, icon: Icon }) => (
          <li key={label} className="flex items-center gap-2">
            <CheckCircle size={14} color="var(--color-success)" />
            <span className="text-sm">{label}</span>
          </li>
        ))}
      </ul>
      <button
        id={`plan-select-${plan.id}`}
        className={`btn w-full ${isActive ? 'btn-secondary' : 'btn-primary'}`}
        onClick={() => !isActive && onSelect(plan.id)}
        disabled={isActive}
      >
        {isActive ? 'Current Plan' : plan.price === 0 ? 'Contact Sales' : 'Upgrade'}
      </button>
    </div>
  )
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function BillingPage() {
  const [tab, setTab] = useState<'overview' | 'plans' | 'invoices'>('overview')
  const qc = useQueryClient()

  const { data: plans = [] } = useQuery<Plan[]>({
    queryKey: ['billing-plans'],
    queryFn: async () => (await api.get('/billing/plans')).data.data ?? [],
  })

  const { data: subscription } = useQuery<Subscription>({
    queryKey: ['billing-subscription'],
    queryFn: async () => (await api.get('/billing/subscription')).data.data,
  })

  const { data: invoices = [] } = useQuery<Invoice[]>({
    queryKey: ['billing-invoices'],
    queryFn: async () => (await api.get('/billing/invoices')).data.data ?? [],
  })

  const { data: usage } = useQuery<Usage>({
    queryKey: ['billing-usage'],
    queryFn: async () => (await api.get('/billing/usage')).data.data,
    refetchInterval: 60_000,
  })

  const upgradePlan = useMutation({
    mutationFn: (planId: string) => api.post('/billing/subscribe', { plan_id: planId }),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['billing-subscription'] }),
  })

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {/* Header */}
      <div style={{ marginBottom: 'var(--space-6)' }}>
        <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Billing & Subscription</h1>
        <p className="text-muted text-sm" style={{ marginTop: 4 }}>M14 · Manage your plan, usage, and invoices</p>
      </div>

      {/* Subscription status banner */}
      {subscription && (
        <div className="card" style={{ padding: 'var(--space-5)', marginBottom: 'var(--space-6)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <div className="flex items-center gap-4">
            <div style={{ width: 44, height: 44, borderRadius: 10, background: 'var(--color-primary)18', display: 'grid', placeItems: 'center' }}>
              <CreditCard size={22} color="var(--color-primary)" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <span className="font-semibold" style={{ textTransform: 'capitalize' }}>{subscription.plan} Plan</span>
                <span className={`badge ${statusColor[subscription.status] ?? 'badge-muted'}`}>
                  {subscription.status}
                </span>
              </div>
              <p className="text-sm text-muted">
                {subscription.status === 'trial' ? 'Trial ends' : 'Renews'}: {' '}
                {new Date(subscription.current_period_end).toLocaleDateString('en-IN')}
                {subscription.auto_renew ? ' · Auto-renew ON' : ' · Auto-renew OFF'}
              </p>
            </div>
          </div>
          <button id="billing-manage" className="btn btn-secondary" onClick={() => setTab('plans')}>
            Change Plan
          </button>
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)' }}>
        {[
          { key: 'overview', icon: BarChart3, label: 'Usage Overview' },
          { key: 'plans', icon: Package, label: 'Plans' },
          { key: 'invoices', icon: CreditCard, label: `Invoices${invoices.length ? ` (${invoices.length})` : ''}` },
        ].map(({ key, icon: Icon, label }) => (
          <button key={key} id={`billing-tab-${key}`}
            className={`btn ${tab === key ? 'btn-primary' : 'btn-ghost'}`}
            style={{ borderRadius: '6px 6px 0 0' }}
            onClick={() => setTab(key as any)}>
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* ── Tab: Overview ── */}
      {tab === 'overview' && usage && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(2, 1fr)', gap: 'var(--space-4)' }} className="animate-fade-in">
          {[
            {
              label: 'Devices', used: usage.devices_count,
              limit: plans.find(p => p.id === subscription?.plan)?.features.max_devices ?? 100,
              icon: Cpu, color: 'var(--color-primary)',
            },
            {
              label: 'API Calls (this month)', used: usage.api_calls,
              limit: plans.find(p => p.id === subscription?.plan)?.features.api_rate_limit ?? 100000,
              icon: Zap, color: 'var(--color-info)',
            },
            {
              label: 'Storage Used', used: Math.round(usage.storage_gb),
              limit: plans.find(p => p.id === subscription?.plan)?.features.storage_gb ?? 100,
              icon: HardDrive, color: 'var(--color-warning)',
              unit: 'GB',
            },
            {
              label: 'Video Hours', used: usage.video_hours,
              limit: 0, icon: BarChart3, color: 'var(--color-success)',
            },
          ].map(({ label, used, limit, icon: Icon, color, unit }) => {
            const pct = limit > 0 ? Math.min(100, (used / limit) * 100) : 0
            return (
              <div key={label} className="card" style={{ padding: 'var(--space-5)' }}>
                <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-3)' }}>
                  <span className="text-sm font-medium">{label}</span>
                  <Icon size={16} color={color} />
                </div>
                <div style={{ fontSize: 'var(--text-2xl)', fontWeight: 700, marginBottom: 'var(--space-3)' }}>
                  {used.toLocaleString()}{unit ? ` ${unit}` : ''} {limit > 0 && <span className="text-muted text-sm font-normal">/ {limit.toLocaleString()}{unit ? ` ${unit}` : ''}</span>}
                </div>
                {limit > 0 && (
                  <div style={{ height: 8, borderRadius: 4, background: 'var(--color-surface-raised)', overflow: 'hidden' }}>
                    <div style={{
                      width: `${pct}%`, height: '100%', borderRadius: 4,
                      background: pct > 90 ? 'var(--color-danger)' : pct > 70 ? 'var(--color-warning)' : color,
                      transition: 'width 0.5s ease',
                    }} />
                  </div>
                )}
                {limit > 0 && pct > 80 && (
                  <p className="text-xs" style={{ color: 'var(--color-warning)', marginTop: 6 }}>
                    <AlertTriangle size={11} style={{ display: 'inline', marginRight: 4 }} />
                    {Math.round(pct)}% used — consider upgrading
                  </p>
                )}
              </div>
            )
          })}
        </div>
      )}

      {/* ── Tab: Plans ── */}
      {tab === 'plans' && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(3, 1fr)', gap: 'var(--space-6)' }} className="animate-fade-in">
          {plans.map(plan => (
            <PlanCard
              key={plan.id} plan={plan}
              current={subscription?.plan ?? ''}
              onSelect={id => upgradePlan.mutate(id)}
            />
          ))}
        </div>
      )}

      {/* ── Tab: Invoices ── */}
      {tab === 'invoices' && (
        <div className="card animate-fade-in">
          {invoices.length === 0 ? (
            <div style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
              <CreditCard size={40} color="var(--color-text-muted)" style={{ margin: '0 auto 12px' }} />
              <p className="text-muted">No invoices yet.</p>
            </div>
          ) : (
            <div className="table-wrap">
              <table>
                <thead>
                  <tr>
                    <th>Invoice #</th>
                    <th>Period</th>
                    <th>Amount</th>
                    <th>Status</th>
                    <th>Due Date</th>
                    <th>Action</th>
                  </tr>
                </thead>
                <tbody>
                  {invoices.map(inv => (
                    <tr key={inv.id}>
                      <td className="font-mono text-sm">{inv.invoice_number}</td>
                      <td className="text-sm text-muted">{new Date(inv.created_at).toLocaleDateString('en-IN')}</td>
                      <td className="font-medium">{fmtCurrency(inv.total, inv.currency)}</td>
                      <td><span className={`badge ${invoiceStatusColor[inv.status] ?? 'badge-muted'}`}>{inv.status}</span></td>
                      <td className="text-sm text-muted">{inv.due_date ? new Date(inv.due_date).toLocaleDateString('en-IN') : '—'}</td>
                      <td>
                        {inv.pdf_url ? (
                          <a href={inv.pdf_url} target="_blank" rel="noreferrer"
                            id={`invoice-dl-${inv.id}`} className="btn btn-secondary btn-sm">
                            <Download size={12} /> PDF
                          </a>
                        ) : (
                          <span className="text-muted text-xs">—</span>
                        )}
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
