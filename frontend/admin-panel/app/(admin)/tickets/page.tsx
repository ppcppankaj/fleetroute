'use client'

import { HiOutlineTicket } from 'react-icons/hi2'

const TICKETS = [
  { id: 'TKT-001', tenant: 'Acme Transport',    subject: 'Live tracking delay > 30s',    status: 'OPEN',        priority: 'HIGH',   created: '2026-04-24 16:00' },
  { id: 'TKT-002', tenant: 'Swift Logistics',   subject: 'Bulk vehicle import failed',   status: 'IN_PROGRESS', priority: 'MEDIUM', created: '2026-04-24 14:30' },
  { id: 'TKT-003', tenant: 'City Cab Services', subject: 'Need custom report template',  status: 'OPEN',        priority: 'LOW',    created: '2026-04-23 11:00' },
  { id: 'TKT-004', tenant: 'QuickFleet Ltd',    subject: 'Invoice PDF not generating',   status: 'RESOLVED',    priority: 'HIGH',   created: '2026-04-22 09:15' },
]

const statusBadge: Record<string, string> = {
  OPEN:        'bg-red-500/20 text-red-400',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400',
  RESOLVED:    'bg-green-500/20 text-green-400',
}
const priorityBadge: Record<string, string> = {
  HIGH:   'bg-red-500/20 text-red-300',
  MEDIUM: 'bg-amber-500/20 text-amber-300',
  LOW:    'bg-zinc-500/20 text-zinc-400',
}

export default function TicketsPage() {
  return (
    <div>
      <div className="h-16 border-b border-zinc-800 flex items-center px-6 bg-zinc-900/50">
        <h1 className="text-base font-semibold text-white flex items-center gap-2">
          <HiOutlineTicket className="w-5 h-5 text-amber-400" /> Support Tickets
        </h1>
      </div>
      <div className="p-6 space-y-4">
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 bg-zinc-800/30">
                {['ID', 'Tenant', 'Subject', 'Priority', 'Status', 'Created', 'Actions'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800">
              {TICKETS.map(t => (
                <tr key={t.id} className="hover:bg-zinc-800/40 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-amber-400">{t.id}</td>
                  <td className="px-4 py-3 text-white font-medium">{t.tenant}</td>
                  <td className="px-4 py-3 text-zinc-300">{t.subject}</td>
                  <td className="px-4 py-3"><span className={`badge ${priorityBadge[t.priority]}`}>{t.priority}</span></td>
                  <td className="px-4 py-3"><span className={`badge ${statusBadge[t.status]}`}>{t.status.replace('_', ' ')}</span></td>
                  <td className="px-4 py-3 text-zinc-500 font-mono text-xs">{t.created}</td>
                  <td className="px-4 py-3">
                    <button className="btn-ghost text-xs text-amber-400">Open</button>
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
