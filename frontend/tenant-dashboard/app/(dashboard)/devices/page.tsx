'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineCpuChip } from 'react-icons/hi2'

const DEVICES = [
  { id: 'd1', imei: '352099001761481', vehicle: 'TRK-042', firmware: 'v2.4.1', lastSeen: '2 seconds ago',   status: 'ONLINE',  signal: 90 },
  { id: 'd2', imei: '352099001761482', vehicle: 'VAN-019', firmware: 'v2.4.0', lastSeen: '5 seconds ago',   status: 'ONLINE',  signal: 75 },
  { id: 'd3', imei: '352099001761483', vehicle: 'CAR-033', firmware: 'v2.3.8', lastSeen: '38 minutes ago',  status: 'OFFLINE', signal: 0  },
  { id: 'd4', imei: '352099001761484', vehicle: 'Unassigned', firmware: 'v2.4.1', lastSeen: '—',           status: 'PROVISIONED', signal: 0 },
]

const statusColor: Record<string, string> = {
  ONLINE:      'bg-green-500/20 text-green-400',
  OFFLINE:     'bg-red-500/20 text-red-400',
  PROVISIONED: 'bg-brand-500/20 text-brand-400',
}

export default function DevicesPage() {
  return (
    <div>
      <TopBar title="Devices" subtitle="GPS device fleet & provisioning" />
      <div className="p-6 space-y-4">
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['IMEI', 'Vehicle', 'Firmware', 'Signal', 'Last Seen', 'Status'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {DEVICES.map(d => (
                <tr key={d.id} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-slate-300 flex items-center gap-2">
                    <HiOutlineCpuChip className="w-3.5 h-3.5 text-brand-400 flex-shrink-0" />{d.imei}
                  </td>
                  <td className="px-4 py-3 text-white font-medium text-xs">{d.vehicle}</td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-400">{d.firmware}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <div className="w-16 h-1.5 bg-surface rounded-full overflow-hidden">
                        <div className={`h-full ${d.signal > 60 ? 'bg-green-500' : d.signal > 30 ? 'bg-amber-500' : 'bg-red-500'} rounded-full`} style={{ width: `${d.signal}%` }} />
                      </div>
                      <span className="text-xs text-slate-400">{d.signal}%</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{d.lastSeen}</td>
                  <td className="px-4 py-3"><span className={`badge ${statusColor[d.status]}`}>{d.status}</span></td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
