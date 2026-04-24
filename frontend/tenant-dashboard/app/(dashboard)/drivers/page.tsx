'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineUserGroup, HiOutlineStar, HiOutlinePlusCircle } from 'react-icons/hi2'

const DRIVERS = [
  { id: 'd1', name: 'Ravi Sharma',   license: 'MH1234567', phone: '+91 9876543210', score: 92, trips: 284, status: 'ON_DUTY',  vehicle: 'TRK-042' },
  { id: 'd2', name: 'Suresh Kumar',  license: 'MH9876543', phone: '+91 9123456780', score: 78, trips: 198, status: 'ON_DUTY',  vehicle: 'VAN-019' },
  { id: 'd3', name: 'Amit Singh',    license: 'MH5555555', phone: '+91 9000000001', score: 65, trips: 154, status: 'OFF_DUTY', vehicle: null      },
  { id: 'd4', name: 'Deepak Patel',  license: 'MH3333333', phone: '+91 9000000002', score: 88, trips: 312, status: 'ON_DUTY',  vehicle: 'BUS-001' },
  { id: 'd5', name: 'Kiran Rao',     license: 'MH7777777', phone: '+91 9000000003', score: 55, trips: 89,  status: 'LEAVE',    vehicle: null      },
]

const statusBadge: Record<string, string> = {
  ON_DUTY:  'bg-green-500/20 text-green-400',
  OFF_DUTY: 'bg-slate-500/20 text-slate-400',
  LEAVE:    'bg-amber-500/20 text-amber-400',
}

function ScoreBar({ score }: { score: number }) {
  const color = score >= 80 ? 'bg-green-500' : score >= 65 ? 'bg-amber-500' : 'bg-red-500'
  return (
    <div className="flex items-center gap-2">
      <div className="w-20 h-1.5 bg-surface rounded-full overflow-hidden">
        <div className={`h-full ${color} rounded-full`} style={{ width: `${score}%` }} />
      </div>
      <span className={`text-xs font-bold ${score >= 80 ? 'text-green-400' : score >= 65 ? 'text-amber-400' : 'text-red-400'}`}>{score}</span>
    </div>
  )
}

export default function DriversPage() {
  return (
    <div>
      <TopBar title="Drivers" subtitle="Driver management & scoring" />
      <div className="p-6 space-y-4">
        {/* Summary */}
        <div className="grid grid-cols-3 gap-4">
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Total Drivers</p>
            <p className="text-3xl font-bold text-white mt-1">{DRIVERS.length}</p>
          </div>
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">On Duty</p>
            <p className="text-3xl font-bold text-green-400 mt-1">{DRIVERS.filter(d => d.status === 'ON_DUTY').length}</p>
          </div>
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Avg Score</p>
            <p className="text-3xl font-bold text-brand-400 mt-1 flex items-center gap-1">
              {(DRIVERS.reduce((s, d) => s + d.score, 0) / DRIVERS.length).toFixed(0)}
              <HiOutlineStar className="w-5 h-5" />
            </p>
          </div>
        </div>

        {/* Toolbar + table */}
        <div className="flex justify-between items-center">
          <h2 className="text-sm font-semibold text-white flex items-center gap-2">
            <HiOutlineUserGroup className="w-4 h-4 text-brand-400" /> All Drivers
          </h2>
          <button className="btn-primary flex items-center gap-2 text-sm">
            <HiOutlinePlusCircle className="w-4 h-4" /> Add Driver
          </button>
        </div>

        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Driver', 'License', 'Phone', 'Score', 'Trips', 'Vehicle', 'Status'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {DRIVERS.map(d => (
                <tr key={d.id} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-medium text-white">{d.name}</td>
                  <td className="px-4 py-3 text-slate-400 font-mono text-xs">{d.license}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">{d.phone}</td>
                  <td className="px-4 py-3"><ScoreBar score={d.score} /></td>
                  <td className="px-4 py-3 text-slate-300">{d.trips}</td>
                  <td className="px-4 py-3 text-brand-400 font-semibold text-xs">{d.vehicle ?? '—'}</td>
                  <td className="px-4 py-3">
                    <span className={`badge ${statusBadge[d.status]}`}>{d.status.replace('_', ' ')}</span>
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
