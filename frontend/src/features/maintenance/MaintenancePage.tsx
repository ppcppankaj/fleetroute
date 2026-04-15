import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Wrench, Plus, CheckCircle, AlertTriangle, XCircle,
  FileText, Package, Clock, Gauge
} from 'lucide-react'
import api from '../../shared/api/client'

type ServiceStatus = 'ok' | 'due_soon' | 'overdue'
type DocStatus = 'valid' | 'expiring_soon' | 'expired'

interface Schedule {
  id: string
  vehicle_id: string
  registration: string
  service_type: string
  description?: string
  interval_days?: number
  interval_km?: number
  next_due_at?: string
  next_due_odometer?: number
  current_odometer_m?: number
  days_until_due?: number
  status: ServiceStatus
  enabled: boolean
}

interface VehicleDoc {
  id: string
  vehicle_id: string
  registration: string
  doc_type: string
  doc_number?: string
  expires_at?: string
  days_until_expiry?: number
  expiry_status: DocStatus
  file_url?: string
}

interface SparePart {
  id: string
  name: string
  part_number?: string
  qty_in_stock: number
  reorder_threshold: number
  unit_cost?: number
  currency: string
  supplier?: string
}

type Tab = 'schedules' | 'documents' | 'parts' | 'log'

const statusConfig: Record<ServiceStatus, { color: string; icon: React.ComponentType<{ size: number }> }> = {
  ok:        { color: 'var(--color-success)', icon: CheckCircle },
  due_soon:  { color: 'var(--color-warning)', icon: AlertTriangle },
  overdue:   { color: 'var(--color-danger)',  icon: XCircle },
}

const docStatusConfig: Record<DocStatus, { color: string; badge: string }> = {
  valid:          { color: 'var(--color-success)', badge: 'badge-success' },
  expiring_soon:  { color: 'var(--color-warning)', badge: 'badge-warning' },
  expired:        { color: 'var(--color-danger)',  badge: 'badge-danger' },
}

export function MaintenancePage() {
  const qc = useQueryClient()
  const [tab, setTab] = useState<Tab>('schedules')
  const [showCompleteModal, setShowCompleteModal] = useState<string | null>(null)
  const [completeForm, setCompleteForm] = useState({
    serviced_at: new Date().toISOString().slice(0, 16),
    technician: '', service_center: '', cost: '', notes: '',
  })

  const { data: schedules = [] } = useQuery<Schedule[]>({
    queryKey: ['maintenance-schedules'],
    queryFn: async () => {
      const r = await api.get('/maintenance/schedules')
      return r.data.data ?? []
    },
  })

  const { data: documents = [] } = useQuery<VehicleDoc[]>({
    queryKey: ['vehicle-documents'],
    queryFn: async () => {
      const r = await api.get('/maintenance/documents')
      return r.data.data ?? []
    },
  })

  const { data: parts = [] } = useQuery<SparePart[]>({
    queryKey: ['spare-parts'],
    queryFn: async () => {
      const r = await api.get('/maintenance/parts')
      return r.data.data ?? []
    },
  })

  const completeMutation = useMutation({
    mutationFn: ({ id, body }: { id: string; body: any }) =>
      api.post(`/maintenance/schedules/${id}/complete`, body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['maintenance-schedules'] })
      setShowCompleteModal(null)
    },
  })

  const overdueCount = schedules.filter((s) => s.status === 'overdue').length
  const dueSoonCount = schedules.filter((s) => s.status === 'due_soon').length
  const expiredDocs  = documents.filter((d) => d.expiry_status === 'expired').length
  const lowStockParts = parts.filter((p) => p.qty_in_stock <= p.reorder_threshold).length

  return (
    <div className="page" style={{ padding: 'var(--space-6)', overflowY: 'auto', height: '100%' }}>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Maintenance</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>Service schedules, documents & parts</p>
        </div>
        <button id="add-schedule" className="btn btn-primary">
          <Plus size={15} /> Add Schedule
        </button>
      </div>

      {/* Summary stats */}
      <div className="stat-grid" style={{ marginBottom: 'var(--space-6)' }}>
        <div className="stat-card">
          <div className="stat-label">Overdue Services</div>
          <div className="stat-value" style={{ color: overdueCount > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>
            {overdueCount}
          </div>
          <div className="stat-change" style={{ color: overdueCount > 0 ? 'var(--color-danger)' : undefined }}>
            {overdueCount > 0 ? 'Immediate attention needed' : 'All caught up'}
          </div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Due Soon</div>
          <div className="stat-value" style={{ color: dueSoonCount > 0 ? 'var(--color-warning)' : undefined }}>
            {dueSoonCount}
          </div>
          <div className="stat-change">within 7 days</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Expired Documents</div>
          <div className="stat-value" style={{ color: expiredDocs > 0 ? 'var(--color-danger)' : 'var(--color-success)' }}>
            {expiredDocs}
          </div>
          <div className="stat-change">certificates & permits</div>
        </div>
        <div className="stat-card">
          <div className="stat-label">Low Stock Parts</div>
          <div className="stat-value" style={{ color: lowStockParts > 0 ? 'var(--color-warning)' : undefined }}>
            {lowStockParts}
          </div>
          <div className="stat-change">below reorder threshold</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)', paddingBottom: 0 }}>
        {([
          { key: 'schedules', icon: Wrench,      label: 'Service Schedules' },
          { key: 'documents', icon: FileText,     label: 'Documents' },
          { key: 'parts',     icon: Package,      label: 'Spare Parts' },
          { key: 'log',       icon: Clock,        label: 'Service Log' },
        ] as const).map(({ key, icon: Icon, label }) => (
          <button
            key={key}
            id={`tab-${key}`}
            className={`btn btn-sm ${tab === key ? 'btn-primary' : 'btn-ghost'}`}
            onClick={() => setTab(key as Tab)}
            style={{ borderRadius: '8px 8px 0 0', marginBottom: -1 }}
          >
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* Schedules Tab */}
      {tab === 'schedules' && (
        <div className="card animate-fade-in">
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Status</th>
                  <th>Vehicle</th>
                  <th>Service Type</th>
                  <th>Interval</th>
                  <th>Next Due</th>
                  <th>Days Until Due</th>
                  <th>Odometer</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {schedules.length === 0 && (
                  <tr><td colSpan={8} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No schedules configured</td></tr>
                )}
                {schedules.map((s) => {
                  const cfg = statusConfig[s.status]
                  const StatusIcon = cfg.icon
                  return (
                    <tr key={s.id} id={`schedule-row-${s.id}`}>
                      <td>
                        <StatusIcon size={16} style={{ color: cfg.color }} />
                      </td>
                      <td><strong>{s.registration}</strong></td>
                      <td>{s.service_type}</td>
                      <td className="text-sm text-muted">
                        {s.interval_days && `${s.interval_days}d`}
                        {s.interval_days && s.interval_km && ' / '}
                        {s.interval_km && `${(s.interval_km / 1000).toFixed(0)}k km`}
                      </td>
                      <td>
                        <span className="text-sm">
                          {s.next_due_at ? new Date(s.next_due_at).toLocaleDateString() : '—'}
                        </span>
                      </td>
                      <td>
                        <span style={{ fontWeight: 600, color: cfg.color }}>
                          {s.days_until_due !== undefined
                            ? s.days_until_due < 0
                              ? `${Math.abs(s.days_until_due)}d overdue`
                              : `${s.days_until_due}d`
                            : '—'}
                        </span>
                      </td>
                      <td>
                        <div className="flex items-center gap-1">
                          <Gauge size={12} style={{ color: 'var(--color-text-muted)' }} />
                          <span className="text-sm">
                            {s.current_odometer_m
                              ? `${(s.current_odometer_m / 1000).toFixed(0)} km`
                              : '—'}
                          </span>
                        </div>
                      </td>
                      <td>
                        <button
                          id={`complete-${s.id}`}
                          className="btn btn-success btn-sm"
                          onClick={() => setShowCompleteModal(s.id)}
                        >
                          <CheckCircle size={12} /> Mark Done
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Documents Tab */}
      {tab === 'documents' && (
        <div className="card animate-fade-in">
          <div className="card-header">
            <h2 className="card-title">Vehicle Documents</h2>
            <button className="btn btn-primary btn-sm"><Plus size={13} /> Add Document</button>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Vehicle</th>
                  <th>Document Type</th>
                  <th>Doc Number</th>
                  <th>Expires</th>
                  <th>Days Left</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {documents.length === 0 && (
                  <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No documents uploaded</td></tr>
                )}
                {documents.map((doc) => {
                  const cfg = docStatusConfig[doc.expiry_status]
                  return (
                    <tr key={doc.id} id={`doc-row-${doc.id}`}>
                      <td><strong>{doc.registration}</strong></td>
                      <td>{doc.doc_type.replace(/_/g, ' ')}</td>
                      <td className="font-mono text-sm">{doc.doc_number ?? '—'}</td>
                      <td>{doc.expires_at ? new Date(doc.expires_at).toLocaleDateString() : '—'}</td>
                      <td>
                        <span style={{ fontWeight: 600, color: cfg.color }}>
                          {doc.days_until_expiry !== undefined
                            ? doc.days_until_expiry < 0
                              ? `Expired ${Math.abs(doc.days_until_expiry)}d ago`
                              : `${doc.days_until_expiry} days`
                            : '—'}
                        </span>
                      </td>
                      <td><span className={`badge ${cfg.badge}`}>{doc.expiry_status.replace('_', ' ')}</span></td>
                      <td>
                        <div className="flex gap-1">
                          {doc.file_url && (
                            <a href={doc.file_url} target="_blank" rel="noreferrer" className="btn btn-ghost btn-sm">
                              <FileText size={12} /> View
                            </a>
                          )}
                        </div>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Spare Parts Tab */}
      {tab === 'parts' && (
        <div className="card animate-fade-in">
          <div className="card-header">
            <h2 className="card-title">Spare Parts Inventory</h2>
            <button className="btn btn-primary btn-sm"><Plus size={13} /> Add Part</button>
          </div>
          <div className="table-wrap">
            <table>
              <thead>
                <tr>
                  <th>Part Name</th>
                  <th>Part Number</th>
                  <th>Stock</th>
                  <th>Reorder At</th>
                  <th>Unit Cost</th>
                  <th>Supplier</th>
                  <th>Status</th>
                </tr>
              </thead>
              <tbody>
                {parts.length === 0 && (
                  <tr><td colSpan={7} style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 32 }}>No parts in inventory</td></tr>
                )}
                {parts.map((p) => {
                  const low = p.qty_in_stock <= p.reorder_threshold
                  return (
                    <tr key={p.id} id={`part-row-${p.id}`}>
                      <td><strong>{p.name}</strong></td>
                      <td className="font-mono text-sm">{p.part_number ?? '—'}</td>
                      <td>
                        <span style={{ fontWeight: 700, color: low ? 'var(--color-danger)' : 'var(--color-success)' }}>
                          {p.qty_in_stock}
                        </span>
                      </td>
                      <td>{p.reorder_threshold}</td>
                      <td>{p.unit_cost ? `${p.currency} ${p.unit_cost}` : '—'}</td>
                      <td>{p.supplier ?? '—'}</td>
                      <td>
                        <span className={`badge ${low ? 'badge-danger' : 'badge-success'}`}>
                          {low ? 'Low Stock' : 'OK'}
                        </span>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Service Log Tab */}
      {tab === 'log' && (
        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Service History</h2></div>
          <div style={{ padding: 32, textAlign: 'center', color: 'var(--color-text-muted)' }}>
            <Clock size={32} style={{ margin: '0 auto 12px', display: 'block' }} />
            <p>Service log will appear here after services are marked complete.</p>
          </div>
        </div>
      )}

      {/* Complete Service Modal */}
      {showCompleteModal && (
        <div style={{
          position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.6)',
          display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 1000,
        }}>
          <div className="card" style={{ width: 480, padding: 32 }}>
            <h3 style={{ fontSize: 'var(--text-lg)', fontWeight: 700, marginBottom: 20 }}>Mark Service Complete</h3>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 12 }}>
              <label style={{ gridColumn: '1/-1' }}>
                <span className="text-sm text-muted">Service Date & Time</span>
                <input type="datetime-local" className="form-input"
                  value={completeForm.serviced_at}
                  onChange={(e) => setCompleteForm((f) => ({ ...f, serviced_at: e.target.value }))} />
              </label>
              <label>
                <span className="text-sm text-muted">Technician</span>
                <input type="text" className="form-input" placeholder="Name"
                  value={completeForm.technician}
                  onChange={(e) => setCompleteForm((f) => ({ ...f, technician: e.target.value }))} />
              </label>
              <label>
                <span className="text-sm text-muted">Service Center</span>
                <input type="text" className="form-input" placeholder="Location"
                  value={completeForm.service_center}
                  onChange={(e) => setCompleteForm((f) => ({ ...f, service_center: e.target.value }))} />
              </label>
              <label>
                <span className="text-sm text-muted">Cost (INR)</span>
                <input type="number" className="form-input" placeholder="0"
                  value={completeForm.cost}
                  onChange={(e) => setCompleteForm((f) => ({ ...f, cost: e.target.value }))} />
              </label>
              <label style={{ gridColumn: '1/-1' }}>
                <span className="text-sm text-muted">Notes</span>
                <textarea className="form-input" rows={3} placeholder="Optional notes…"
                  value={completeForm.notes}
                  onChange={(e) => setCompleteForm((f) => ({ ...f, notes: e.target.value }))} />
              </label>
            </div>
            <div className="flex gap-3" style={{ marginTop: 20, justifyContent: 'flex-end' }}>
              <button className="btn btn-secondary" onClick={() => setShowCompleteModal(null)}>Cancel</button>
              <button
                id="confirm-complete"
                className="btn btn-success"
                onClick={() => completeMutation.mutate({
                  id: showCompleteModal,
                  body: {
                    serviced_at: new Date(completeForm.serviced_at).toISOString(),
                    technician: completeForm.technician,
                    service_center: completeForm.service_center,
                    cost: completeForm.cost ? parseFloat(completeForm.cost) : undefined,
                    notes: completeForm.notes,
                  },
                })}
              >
                <CheckCircle size={15} /> Confirm Complete
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
