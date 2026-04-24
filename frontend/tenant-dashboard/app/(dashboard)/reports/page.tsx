'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineDocumentChartBar, HiOutlinePlusCircle, HiOutlineArrowDownTray, HiOutlinePlay } from 'react-icons/hi2'

const DEFINITIONS = [
  { id: 'r1', name: 'Weekly Trip Summary',   type: 'TRIP_SUMMARY',  schedule: 'WEEKLY',  status: 'ACTIVE' },
  { id: 'r2', name: 'Monthly Fuel Report',   type: 'FUEL',          schedule: 'MONTHLY', status: 'ACTIVE' },
  { id: 'r3', name: 'Driver Score Report',   type: 'DRIVER_SCORE',  schedule: null,      status: 'ACTIVE' },
  { id: 'r4', name: 'Maintenance Log',       type: 'MAINTENANCE',   schedule: 'MONTHLY', status: 'INACTIVE' },
]

const RUNS = [
  { defName: 'Weekly Trip Summary', status: 'DONE',    size: '2.3 MB', time: '2026-04-21 08:00' },
  { defName: 'Monthly Fuel Report', status: 'DONE',    size: '1.1 MB', time: '2026-04-01 00:00' },
  { defName: 'Driver Score Report', status: 'RUNNING', size: null,     time: '2026-04-24 19:55' },
]

const runStatusStyle: Record<string, string> = {
  DONE:    'bg-green-500/20 text-green-400',
  RUNNING: 'bg-brand-500/20 text-brand-400 animate-pulse',
  FAILED:  'bg-red-500/20 text-red-400',
  PENDING: 'bg-slate-500/20 text-slate-400',
}

export default function ReportsPage() {
  return (
    <div>
      <TopBar title="Reports" subtitle="Scheduled & on-demand fleet analytics" />
      <div className="p-6 space-y-5">
        {/* Definitions */}
        <div className="flex justify-between items-center">
          <h2 className="text-sm font-semibold text-white flex items-center gap-2">
            <HiOutlineDocumentChartBar className="w-4 h-4 text-brand-400" /> Report Definitions
          </h2>
          <button className="btn-primary flex items-center gap-2 text-sm">
            <HiOutlinePlusCircle className="w-4 h-4" /> New Definition
          </button>
        </div>

        <div className="grid grid-cols-2 gap-4">
          {DEFINITIONS.map(d => (
            <div key={d.id} className="card flex items-center gap-4">
              <div className="w-10 h-10 rounded-lg bg-brand-500/20 flex items-center justify-center">
                <HiOutlineDocumentChartBar className="w-5 h-5 text-brand-400" />
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold text-white">{d.name}</span>
                  <span className={`badge ${d.status === 'ACTIVE' ? 'bg-green-500/20 text-green-400' : 'bg-slate-500/20 text-slate-400'}`}>{d.status}</span>
                </div>
                <div className="flex gap-3 mt-0.5 text-xs text-slate-500">
                  <span>{d.type}</span>
                  {d.schedule && <span>· {d.schedule}</span>}
                </div>
              </div>
              <button className="btn-ghost p-2 text-green-400 hover:text-green-300" title="Run now">
                <HiOutlinePlay className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>

        {/* Run history */}
        <h2 className="text-sm font-semibold text-white flex items-center gap-2 pt-2">
          Recent Runs
        </h2>
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Report', 'Status', 'Size', 'Executed At', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {RUNS.map((r, i) => (
                <tr key={i} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 text-white font-medium">{r.defName}</td>
                  <td className="px-4 py-3"><span className={`badge ${runStatusStyle[r.status]}`}>{r.status}</span></td>
                  <td className="px-4 py-3 text-slate-400 font-mono text-xs">{r.size ?? '—'}</td>
                  <td className="px-4 py-3 text-slate-400 font-mono text-xs">{r.time}</td>
                  <td className="px-4 py-3">
                    {r.status === 'DONE' && (
                      <button className="btn-ghost flex items-center gap-1 text-xs text-brand-400">
                        <HiOutlineArrowDownTray className="w-3.5 h-3.5" /> Download
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
