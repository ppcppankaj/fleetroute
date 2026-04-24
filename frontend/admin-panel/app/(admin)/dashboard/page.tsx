'use client'

import { HiOutlineUserGroup, HiOutlineCurrencyDollar, HiOutlineServerStack, HiOutlineTicket, HiOutlineArrowTrendingUp } from 'react-icons/hi2'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts'

const revenueData = [
  { month: 'Nov', mrr: 142000, tenants: 38 },
  { month: 'Dec', mrr: 156000, tenants: 42 },
  { month: 'Jan', mrr: 178000, tenants: 48 },
  { month: 'Feb', mrr: 195000, tenants: 53 },
  { month: 'Mar', mrr: 218000, tenants: 59 },
  { month: 'Apr', mrr: 241000, tenants: 64 },
]

const serviceHealth = [
  { name: 'M01 Live Tracking',   status: 'UP',   latency: 12  },
  { name: 'M02 Routes & Trips',  status: 'UP',   latency: 8   },
  { name: 'M03 Geofencing',      status: 'UP',   latency: 15  },
  { name: 'M04 Alerts',          status: 'UP',   latency: 9   },
  { name: 'M05 Reports',         status: 'WARN', latency: 234 },
  { name: 'M06 Vehicles',        status: 'UP',   latency: 11  },
  { name: 'M07 Drivers',         status: 'UP',   latency: 10  },
  { name: 'M08 Maintenance',     status: 'UP',   latency: 7   },
  { name: 'M09 Fuel',            status: 'UP',   latency: 8   },
  { name: 'M10 Multi-Tenant',    status: 'UP',   latency: 6   },
]

export default function AdminDashboard() {
  const fmtMRR = (n: number) => `₹${(n / 1000).toFixed(0)}k`

  return (
    <div>
      {/* Header */}
      <div className="h-16 border-b border-zinc-800 flex items-center px-6 justify-between bg-zinc-900/50">
        <div>
          <h1 className="text-base font-semibold text-white">Platform Overview</h1>
          <p className="text-xs text-zinc-500">TrackOra Platform Admin</p>
        </div>
        <span className="badge bg-green-500/20 text-green-400 animate-pulse">● All Systems Operational</span>
      </div>

      <div className="p-6 space-y-6">
        {/* KPI row */}
        <div className="grid grid-cols-4 gap-4">
          {[
            { label: 'Active Tenants', value: '64',      sub: '+5 this month', color: 'text-amber-400',  icon: HiOutlineUserGroup      },
            { label: 'MRR',            value: '₹2.41L',  sub: '↑ 10.5% MoM',  color: 'text-green-400',  icon: HiOutlineCurrencyDollar  },
            { label: 'Services Up',    value: '16/17',   sub: '1 degraded',    color: 'text-amber-400',  icon: HiOutlineServerStack     },
            { label: 'Open Tickets',   value: '8',       sub: '2 critical',    color: 'text-red-400',    icon: HiOutlineTicket          },
          ].map(k => (
            <div key={k.label} className="card">
              <div className="flex items-start justify-between">
                <div>
                  <p className="text-xs text-zinc-400 uppercase tracking-wider">{k.label}</p>
                  <p className={`text-3xl font-bold mt-2 ${k.color}`}>{k.value}</p>
                  <p className="text-xs text-zinc-500 mt-1">{k.sub}</p>
                </div>
                <k.icon className={`w-5 h-5 ${k.color} mt-1`} />
              </div>
            </div>
          ))}
        </div>

        {/* Charts */}
        <div className="grid grid-cols-2 gap-4">
          {/* MRR trend */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-sm font-semibold text-white flex items-center gap-2">
                <HiOutlineArrowTrendingUp className="w-4 h-4 text-amber-400" /> Monthly Recurring Revenue — TrackOra
              </h2>
            </div>
            <ResponsiveContainer width="100%" height={200}>
              <AreaChart data={revenueData}>
                <defs>
                  <linearGradient id="mrrGrad" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%"  stopColor="#d97706" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#d97706" stopOpacity={0}   />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis dataKey="month" tick={{ fill: '#71717a', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tickFormatter={fmtMRR} tick={{ fill: '#71717a', fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip
                  formatter={(v: number) => [`₹${v.toLocaleString()}`, 'MRR']}
                  contentStyle={{ background: '#18181b', border: '1px solid #27272a', borderRadius: '8px', fontSize: 12 }}
                />
                <Area type="monotone" dataKey="mrr" stroke="#d97706" strokeWidth={2} fill="url(#mrrGrad)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Tenant growth */}
          <div className="card">
            <h2 className="text-sm font-semibold text-white mb-4">Tenant Growth</h2>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={revenueData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                <XAxis dataKey="month" tick={{ fill: '#71717a', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: '#71717a', fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip contentStyle={{ background: '#18181b', border: '1px solid #27272a', borderRadius: '8px', fontSize: 12 }} />
                <Bar dataKey="tenants" fill="#d97706" radius={[4, 4, 0, 0]} name="Tenants" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Service health */}
        <div className="card">
          <h2 className="text-sm font-semibold text-white mb-4 flex items-center gap-2">
            <HiOutlineServerStack className="w-4 h-4 text-amber-400" /> TrackOra Service Health (17 Microservices)
          </h2>
          <div className="grid grid-cols-2 gap-2">
            {serviceHealth.map(s => (
              <div key={s.name} className="flex items-center justify-between py-2 px-3 bg-zinc-800/50 rounded-lg">
                <div className="flex items-center gap-2">
                  <span className={`w-2 h-2 rounded-full ${s.status === 'UP' ? 'bg-green-500' : s.status === 'WARN' ? 'bg-amber-500 animate-pulse' : 'bg-red-500'}`} />
                  <span className="text-xs text-zinc-300">{s.name}</span>
                </div>
                <span className={`text-xs font-mono ${s.latency > 100 ? 'text-amber-400' : 'text-zinc-500'}`}>{s.latency}ms</span>
              </div>
            ))}
          </div>
        </div>

      </div>
    </div>
  )
}
