'use client'

import { HiOutlineServerStack } from 'react-icons/hi2'

const SERVICES = [
  { id: 'm01', name: 'M01 Live Tracking',  port: 4001, status: 'UP',   latency: 12,  uptime: '99.98%' },
  { id: 'm02', name: 'M02 Routes & Trips', port: 4002, status: 'UP',   latency: 8,   uptime: '99.95%' },
  { id: 'm03', name: 'M03 Geofencing',     port: 4003, status: 'UP',   latency: 15,  uptime: '99.99%' },
  { id: 'm04', name: 'M04 Alerts',         port: 4004, status: 'UP',   latency: 9,   uptime: '99.97%' },
  { id: 'm05', name: 'M05 Reports',        port: 4005, status: 'WARN', latency: 234, uptime: '99.12%' },
  { id: 'm06', name: 'M06 Vehicles',       port: 4006, status: 'UP',   latency: 11,  uptime: '100%'   },
  { id: 'm07', name: 'M07 Drivers',        port: 4007, status: 'UP',   latency: 10,  uptime: '99.99%' },
  { id: 'm08', name: 'M08 Maintenance',    port: 4008, status: 'UP',   latency: 7,   uptime: '99.98%' },
  { id: 'm09', name: 'M09 Fuel',           port: 4009, status: 'UP',   latency: 8,   uptime: '99.96%' },
  { id: 'm10', name: 'M10 Multi-Tenant',   port: 4010, status: 'UP',   latency: 6,   uptime: '100%'   },
  { id: 'm11', name: 'M11 Users & Access', port: 4011, status: 'UP',   latency: 14,  uptime: '99.99%' },
  { id: 'm12', name: 'M12 Devices',        port: 4012, status: 'UP',   latency: 9,   uptime: '99.98%' },
  { id: 'm13', name: 'M13 Security',       port: 4013, status: 'UP',   latency: 11,  uptime: '99.99%' },
  { id: 'm14', name: 'M14 Billing',        port: 4014, status: 'UP',   latency: 18,  uptime: '99.95%' },
  { id: 'm15', name: 'M15 Admin Panel',    port: 4015, status: 'UP',   latency: 7,   uptime: '100%'   },
  { id: 'm16', name: 'M16 Activity Log',   port: 4016, status: 'UP',   latency: 6,   uptime: '100%'   },
  { id: 'm17', name: 'M17 Roadmap',        port: 4017, status: 'UP',   latency: 5,   uptime: '100%'   },
]

const statusStyle: Record<string, string> = {
  UP:   'bg-green-500/20 text-green-400',
  WARN: 'bg-amber-500/20 text-amber-400',
  DOWN: 'bg-red-500/20 text-red-400',
}

export default function ServicesPage() {
  const up = SERVICES.filter(s => s.status === 'UP').length
  return (
    <div>
      <div className="h-16 border-b border-zinc-800 flex items-center justify-between px-6 bg-zinc-900/50">
        <h1 className="text-base font-semibold text-white flex items-center gap-2">
          <HiOutlineServerStack className="w-5 h-5 text-amber-400" /> Service Health
        </h1>
        <span className="badge bg-green-500/20 text-green-400">{up}/{SERVICES.length} Healthy</span>
      </div>
      <div className="p-6">
        <div className="grid grid-cols-1 gap-2">
          {SERVICES.map(s => (
            <div key={s.id} className="card flex items-center gap-4 py-3">
              <span className={`w-2.5 h-2.5 rounded-full flex-shrink-0 ${s.status === 'UP' ? 'bg-green-500' : s.status === 'WARN' ? 'bg-amber-500 animate-pulse' : 'bg-red-500'}`} />
              <span className="text-sm font-medium text-white w-44 flex-shrink-0">{s.name}</span>
              <span className="text-xs text-zinc-500 font-mono flex-shrink-0">:{s.port}</span>
              <div className="flex-1" />
              <span className={`badge ${statusStyle[s.status]} flex-shrink-0`}>{s.status}</span>
              <span className={`text-xs font-mono w-16 text-right flex-shrink-0 ${s.latency > 100 ? 'text-amber-400' : 'text-zinc-500'}`}>{s.latency}ms</span>
              <span className="text-xs font-mono w-16 text-right text-green-400 flex-shrink-0">{s.uptime}</span>
              <a href={`http://localhost:${s.port}/health`} target="_blank" rel="noreferrer"
                className="btn-ghost text-xs text-amber-400 hover:text-amber-300">Health →</a>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
