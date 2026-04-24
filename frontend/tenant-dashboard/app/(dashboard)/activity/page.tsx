'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineChartBarSquare } from 'react-icons/hi2'

const EVENTS = [
  { type: 'TRIP_STARTED',    title: 'TRK-042 started a trip',           time: '2 min ago',  vehicle: 'TRK-042', driver: 'Ravi Sharma'  },
  { type: 'ALERT_TRIGGERED', title: 'Speeding alert triggered for TRK-042', time: '2 min ago',  vehicle: 'TRK-042', driver: 'Ravi Sharma'  },
  { type: 'TRIP_COMPLETED',  title: 'VAN-019 completed trip (22.8 km)',  time: '14 min ago', vehicle: 'VAN-019', driver: 'Suresh Kumar' },
  { type: 'GEOFENCE_BREACH', title: 'VAN-019 exited Mumbai Depot zone',  time: '8 min ago',  vehicle: 'VAN-019', driver: 'Suresh Kumar' },
  { type: 'MAINTENANCE_DUE', title: 'Oil change overdue for TRK-007',   time: '1 hr ago',   vehicle: 'TRK-007', driver: null          },
  { type: 'TRIP_STARTED',    title: 'BUS-001 started a trip',           time: '30 min ago', vehicle: 'BUS-001', driver: 'Kiran Rao'   },
]

const typeColor: Record<string, string> = {
  TRIP_STARTED:    'bg-green-500/20 text-green-400',
  TRIP_COMPLETED:  'bg-brand-500/20 text-brand-400',
  ALERT_TRIGGERED: 'bg-red-500/20 text-red-400',
  GEOFENCE_BREACH: 'bg-amber-500/20 text-amber-400',
  MAINTENANCE_DUE: 'bg-orange-500/20 text-orange-400',
}

export default function ActivityPage() {
  return (
    <div>
      <TopBar title="Activity Log" subtitle="Comprehensive fleet activity timeline" />
      <div className="p-6 space-y-4">
        <div className="flex items-center gap-2 mb-4">
          <HiOutlineChartBarSquare className="w-4 h-4 text-brand-400" />
          <span className="text-sm font-semibold text-white">Recent Activity</span>
          <span className="badge bg-green-500/20 text-green-400 ml-2 animate-pulse">● LIVE</span>
        </div>
        <div className="relative">
          {/* Timeline line */}
          <div className="absolute left-5 top-3 bottom-3 w-0.5 bg-surface-border" />
          <div className="space-y-4">
            {EVENTS.map((e, i) => (
              <div key={i} className="flex gap-4 relative">
                <div className={`w-10 h-10 rounded-full flex items-center justify-center flex-shrink-0 z-10 ${typeColor[e.type]?.split(' ')[0] || 'bg-slate-500/20'}`}>
                  <span className="text-xs font-bold">{e.type[0]}</span>
                </div>
                <div className="card flex-1 py-3">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-white">{e.title}</span>
                    <span className="text-xs text-slate-500 ml-4 flex-shrink-0">{e.time}</span>
                  </div>
                  <div className="flex gap-3 mt-1 text-xs text-slate-500">
                    <span className={`badge ${typeColor[e.type]}`}>{e.type.replace(/_/g, ' ')}</span>
                    {e.driver && <span>{e.driver}</span>}
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
