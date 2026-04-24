'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineBell, HiOutlineCheckCircle, HiOutlineFunnel } from 'react-icons/hi2'
import { useState } from 'react'

const MOCK_ALERTS = [
  { id: '1', type: 'SPEEDING',    vehicle: 'TRK-042', driver: 'Ravi Sharma',  message: 'Sustained speed of 118 km/h', severity: 'HIGH',   status: 'TRIGGERED',    time: '2m ago'  },
  { id: '2', type: 'GEOFENCE',    vehicle: 'VAN-019', driver: 'Suresh Kumar', message: 'Exited Mumbai Depot zone',    severity: 'MEDIUM', status: 'TRIGGERED',    time: '8m ago'  },
  { id: '3', type: 'MAINTENANCE', vehicle: 'TRK-007', driver: 'Amit Singh',   message: 'Oil change overdue 200km',   severity: 'MEDIUM', status: 'ACKNOWLEDGED', time: '32m ago' },
  { id: '4', type: 'OFFLINE',     vehicle: 'CAR-033', driver: 'Deepak Patel', message: 'Device offline 35 minutes',  severity: 'LOW',    status: 'TRIGGERED',    time: '37m ago' },
  { id: '5', type: 'SPEEDING',    vehicle: 'BUS-001', driver: 'Kiran Rao',    message: 'Exceeded 90 km/h speed limit', severity: 'HIGH', status: 'RESOLVED',     time: '1h ago'  },
]

const sevColor: Record<string, string> = {
  HIGH:   'bg-red-500/20 text-red-400 border-red-500/30',
  MEDIUM: 'bg-amber-500/20 text-amber-400 border-amber-500/30',
  LOW:    'bg-slate-500/20 text-slate-400 border-slate-500/30',
}
const statusIcon: Record<string, React.ReactNode> = {
  TRIGGERED:    <span className="w-2 h-2 rounded-full bg-red-500 animate-pulse" />,
  ACKNOWLEDGED: <span className="w-2 h-2 rounded-full bg-amber-400" />,
  RESOLVED:     <HiOutlineCheckCircle className="w-3.5 h-3.5 text-green-400" />,
}

export default function AlertsPage() {
  const [filter, setFilter] = useState('ALL')
  const filtered = filter === 'ALL' ? MOCK_ALERTS : MOCK_ALERTS.filter(a => a.status === filter)

  return (
    <div>
      <TopBar title="Alerts" subtitle="Active fleet alerts & rules" />
      <div className="p-6 space-y-4">

        {/* Summary cards */}
        <div className="grid grid-cols-3 gap-4">
          {[
            { label: 'Triggered',    count: MOCK_ALERTS.filter(a => a.status === 'TRIGGERED').length,    color: 'border-red-500/40 text-red-400'   },
            { label: 'Acknowledged', count: MOCK_ALERTS.filter(a => a.status === 'ACKNOWLEDGED').length, color: 'border-amber-500/40 text-amber-400' },
            { label: 'Resolved Today', count: MOCK_ALERTS.filter(a => a.status === 'RESOLVED').length,   color: 'border-green-500/40 text-green-400' },
          ].map(s => (
            <div key={s.label} className={`card border ${s.color} text-center`}>
              <p className={`text-3xl font-bold ${s.color.split(' ')[1]}`}>{s.count}</p>
              <p className="text-xs text-slate-500 mt-1">{s.label}</p>
            </div>
          ))}
        </div>

        {/* Filter bar */}
        <div className="flex items-center gap-2">
          <HiOutlineFunnel className="w-4 h-4 text-slate-500" />
          {['ALL', 'TRIGGERED', 'ACKNOWLEDGED', 'RESOLVED'].map(f => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`text-xs px-3 py-1.5 rounded-full font-medium transition-all ${filter === f ? 'bg-brand-500 text-white' : 'bg-surface-hover text-slate-400 hover:text-white'}`}
            >
              {f}
            </button>
          ))}
        </div>

        {/* Alert list */}
        <div className="space-y-2">
          {filtered.map(a => (
            <div key={a.id} className="card flex items-start gap-4 hover:border-brand-500/30 transition-all">
              <div className={`badge border ${sevColor[a.severity]} mt-0.5 flex-shrink-0`}>{a.severity}</div>
              <div className="flex items-center gap-1.5 mt-0.5 flex-shrink-0">
                {statusIcon[a.status]}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-xs font-semibold text-brand-400">{a.type}</span>
                  <span className="text-xs text-slate-400">·</span>
                  <span className="text-xs font-medium text-white">{a.vehicle}</span>
                  <span className="text-xs text-slate-500">({a.driver})</span>
                </div>
                <p className="text-sm text-slate-300 mt-0.5">{a.message}</p>
                <p className="text-xs text-slate-500 mt-1">{a.time}</p>
              </div>
              {a.status === 'TRIGGERED' && (
                <button className="btn-primary text-xs px-3 py-1.5 flex-shrink-0">Resolve</button>
              )}
            </div>
          ))}
          {filtered.length === 0 && (
            <div className="card text-center py-12">
              <HiOutlineBell className="w-8 h-8 text-slate-600 mx-auto mb-2" />
              <p className="text-slate-500">No alerts in this category</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
