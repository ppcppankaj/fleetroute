'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineShieldCheck } from 'react-icons/hi2'

const LOGS = [
  { action: 'LOGIN_SUCCESS', user: 'admin@acme.com', resource: 'auth',     ip: '103.43.12.8',  time: '2026-04-24 19:45' },
  { action: 'VEHICLE_UPDATE', user: 'ops@acme.com', resource: 'vehicles', ip: '103.43.12.9',  time: '2026-04-24 18:32' },
  { action: 'ALERT_RESOLVED', user: 'ops@acme.com', resource: 'alerts',   ip: '103.43.12.9',  time: '2026-04-24 18:10' },
  { action: 'LOGIN_FAILED',  user: 'unknown',       resource: 'auth',     ip: '45.32.99.201', time: '2026-04-24 17:55' },
  { action: 'REPORT_RUN',    user: 'admin@acme.com', resource: 'reports', ip: '103.43.12.8',  time: '2026-04-24 17:00' },
]

const actionColor: Record<string, string> = {
  LOGIN_SUCCESS:  'bg-green-500/20 text-green-400',
  LOGIN_FAILED:   'bg-red-500/20 text-red-400',
  VEHICLE_UPDATE: 'bg-brand-500/20 text-brand-400',
  ALERT_RESOLVED: 'bg-amber-500/20 text-amber-400',
  REPORT_RUN:     'bg-purple-500/20 text-purple-400',
}

export default function SecurityPage() {
  return (
    <div>
      <TopBar title="Security & Audit" subtitle="Login events & actionaudit trail" />
      <div className="p-6 space-y-4">
        <div className="flex items-center gap-3 card bg-green-500/5 border-green-500/20">
          <HiOutlineShieldCheck className="w-8 h-8 text-green-400" />
          <div>
            <p className="font-semibold text-white text-sm">Security status: Good</p>
            <p className="text-xs text-slate-400">No critical incidents in the past 30 days. 1 failed login attempt flagged.</p>
          </div>
        </div>
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Action', 'User', 'Resource', 'IP Address', 'Time'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {LOGS.map((l, i) => (
                <tr key={i} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3"><span className={`badge ${actionColor[l.action] || 'bg-slate-500/20 text-slate-400'}`}>{l.action}</span></td>
                  <td className="px-4 py-3 text-slate-300 text-xs">{l.user}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">{l.resource}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">{l.ip}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-500">{l.time}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
