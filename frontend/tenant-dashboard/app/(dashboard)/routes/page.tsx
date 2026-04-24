'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineCalendarDays, HiOutlineMapPin } from 'react-icons/hi2'

const TRIPS = [
  { id: 't1', vehicle: 'TRK-042', driver: 'Ravi Sharma',  start: 'Mumbai Depot',       end: 'Pune Central',    km: 148.2, duration: '2h 34m', fuel: 18.2, date: '2026-04-24 07:15', status: 'COMPLETED' },
  { id: 't2', vehicle: 'VAN-019', driver: 'Suresh Kumar', start: 'Bandra Office',       end: 'Airport Terminal', km: 22.8,  duration: '48m',    fuel: 3.1,  date: '2026-04-24 09:00', status: 'COMPLETED' },
  { id: 't3', vehicle: 'BUS-001', driver: 'Kiran Rao',    start: 'Andheri West',        end: 'Thane Station',   km: 34.5,  duration: '1h 12m', fuel: 5.8,  date: '2026-04-24 11:30', status: 'IN_PROGRESS' },
  { id: 't4', vehicle: 'TRK-007', driver: 'Amit Singh',   start: 'Navi Mumbai Depot',   end: 'Nashik Factory',  km: 210.0, duration: '—',      fuel: 0,    date: '2026-04-24 14:00', status: 'IN_PROGRESS' },
]

const statusBadge: Record<string, string> = {
  COMPLETED:   'bg-green-500/20 text-green-400',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400 animate-pulse',
  CANCELLED:   'bg-red-500/20 text-red-400',
}

export default function RoutesPage() {
  return (
    <div>
      <TopBar title="Routes & Trips" subtitle="Trip history, playback & route planning" />
      <div className="p-6 space-y-4">
        <div className="grid grid-cols-3 gap-4">
          <div className="card"><p className="text-xs text-slate-400 uppercase tracking-wider">Today's Trips</p><p className="text-3xl font-bold text-white mt-2">24</p></div>
          <div className="card"><p className="text-xs text-slate-400 uppercase tracking-wider">Total KM Today</p><p className="text-3xl font-bold text-brand-400 mt-2">1,842</p></div>
          <div className="card"><p className="text-xs text-slate-400 uppercase tracking-wider">In Progress</p><p className="text-3xl font-bold text-amber-400 mt-2">6</p></div>
        </div>
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Vehicle', 'Driver', 'From → To', 'Distance', 'Duration', 'Fuel', 'Date', 'Status'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider whitespace-nowrap">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {TRIPS.map(t => (
                <tr key={t.id} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-semibold text-brand-400 text-xs">{t.vehicle}</td>
                  <td className="px-4 py-3 text-slate-300 text-xs">{t.driver}</td>
                  <td className="px-4 py-3">
                    <span className="flex items-center gap-1 text-xs text-slate-300">
                      <HiOutlineMapPin className="w-3 h-3 text-slate-500 flex-shrink-0" />{t.start}
                      <span className="text-slate-600 mx-1">→</span>
                      <HiOutlineMapPin className="w-3 h-3 text-green-500 flex-shrink-0" />{t.end}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-300 font-mono text-xs">{t.km} km</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">{t.duration}</td>
                  <td className="px-4 py-3 text-amber-400 font-mono text-xs">{t.fuel > 0 ? `${t.fuel} L` : '—'}</td>
                  <td className="px-4 py-3 text-slate-500 font-mono text-xs">{t.date}</td>
                  <td className="px-4 py-3"><span className={`badge ${statusBadge[t.status]}`}>{t.status.replace('_',' ')}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
