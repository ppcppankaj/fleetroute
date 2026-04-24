'use client'

import TopBar from '@/components/TopBar'
import { HiOutlinePlusCircle, HiOutlineWrenchScrewdriver, HiOutlineCalendarDays } from 'react-icons/hi2'

const TASKS = [
  { id: '1', vehicle: 'TRK-042', type: 'OIL_CHANGE',   title: 'Engine Oil Change',       status: 'OVERDUE',   dueAt: '2026-04-15', cost: null, vendor: null  },
  { id: '2', vehicle: 'BUS-001', type: 'TYRE_ROTATION', title: 'Tyre Rotation',           status: 'SCHEDULED', dueAt: '2026-04-30', cost: null, vendor: null  },
  { id: '3', vehicle: 'VAN-019', type: 'BRAKE_CHECK',   title: 'Brake System Inspection', status: 'COMPLETED', dueAt: '2026-04-20', cost: 4500, vendor: 'AutoCare' },
  { id: '4', vehicle: 'TRK-007', type: 'FILTER_AIR',   title: 'Air Filter Replacement',  status: 'SCHEDULED', dueAt: '2026-05-10', cost: null, vendor: null  },
  { id: '5', vehicle: 'CAR-033', type: 'BATTERY',       title: 'Battery Check & Replace', status: 'IN_PROGRESS', dueAt: '2026-04-25', cost: null, vendor: 'MotorWorks' },
]

const statusStyle: Record<string, string> = {
  SCHEDULED:   'bg-brand-500/20 text-brand-400',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400',
  COMPLETED:   'bg-green-500/20 text-green-400',
  OVERDUE:     'bg-red-500/20 text-red-400',
}

export default function MaintenancePage() {
  const summary = {
    overdue: TASKS.filter(t => t.status === 'OVERDUE').length,
    scheduled: TASKS.filter(t => t.status === 'SCHEDULED').length,
    inProgress: TASKS.filter(t => t.status === 'IN_PROGRESS').length,
    completed: TASKS.filter(t => t.status === 'COMPLETED').length,
  }

  return (
    <div>
      <TopBar title="Maintenance" subtitle="Fleet maintenance scheduling & tracking" />
      <div className="p-6 space-y-5">
        {/* Summary */}
        <div className="grid grid-cols-4 gap-4">
          {[
            { label: 'Overdue',      count: summary.overdue,     color: 'text-red-400'   },
            { label: 'In Progress',  count: summary.inProgress,  color: 'text-amber-400' },
            { label: 'Scheduled',    count: summary.scheduled,   color: 'text-brand-400' },
            { label: 'Completed',    count: summary.completed,   color: 'text-green-400' },
          ].map(s => (
            <div key={s.label} className="card text-center">
              <p className={`text-2xl font-bold ${s.color}`}>{s.count}</p>
              <p className="text-xs text-slate-500 mt-1">{s.label}</p>
            </div>
          ))}
        </div>

        {/* Toolbar */}
        <div className="flex justify-between">
          <h2 className="text-sm font-semibold text-white flex items-center gap-2">
            <HiOutlineWrenchScrewdriver className="w-4 h-4 text-brand-400" />
            All Maintenance Tasks
          </h2>
          <button className="btn-primary flex items-center gap-2 text-sm">
            <HiOutlinePlusCircle className="w-4 h-4" /> Schedule Task
          </button>
        </div>

        {/* Task list */}
        <div className="space-y-2">
          {TASKS.map(t => (
            <div key={t.id} className="card flex items-center gap-4 hover:border-surface-hover transition-all">
              <div className="w-10 h-10 rounded-lg bg-surface flex items-center justify-center flex-shrink-0">
                <HiOutlineWrenchScrewdriver className="w-5 h-5 text-brand-400" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-semibold text-white">{t.title}</span>
                  <span className={`badge ${statusStyle[t.status]}`}>{t.status}</span>
                </div>
                <div className="flex items-center gap-3 mt-0.5 text-xs text-slate-500">
                  <span className="flex items-center gap-1">
                    <HiOutlineCalendarDays className="w-3 h-3" /> Due: {t.dueAt}
                  </span>
                  <span>Vehicle: <span className="text-slate-300">{t.vehicle}</span></span>
                  {t.vendor && <span>Vendor: <span className="text-slate-300">{t.vendor}</span></span>}
                  {t.cost && <span>Cost: <span className="text-green-400">₹{t.cost.toLocaleString()}</span></span>}
                </div>
              </div>
              {t.status !== 'COMPLETED' && (
                <button className="btn-ghost text-xs px-3">Mark Done</button>
              )}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
