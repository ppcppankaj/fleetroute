'use client'

import TopBar from '@/components/TopBar'
import StatCard from '@/components/StatCard'
import { HiOutlineTruck, HiOutlineUserGroup, HiOutlineBell, HiOutlineCpuChip, HiOutlineMap, HiOutlineFire } from 'react-icons/hi2'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar, Legend } from 'recharts'

const tripData = [
  { day: 'Mon', trips: 34, km: 1240 },
  { day: 'Tue', trips: 41, km: 1580 },
  { day: 'Wed', trips: 28, km: 980  },
  { day: 'Thu', trips: 56, km: 2100 },
  { day: 'Fri', trips: 62, km: 2450 },
  { day: 'Sat', trips: 38, km: 1320 },
  { day: 'Sun', trips: 22, km: 820  },
]

const alertData = [
  { name: 'Speeding',   value: 14 },
  { name: 'Geofence',   value: 8  },
  { name: 'Maintenance',value: 5  },
  { name: 'Offline',    value: 3  },
]

const vehicleStatuses = [
  { label: 'Moving',   count: 42, color: 'bg-green-500'  },
  { label: 'Idle',     count: 18, color: 'bg-amber-500'  },
  { label: 'Parked',   count: 31, color: 'bg-slate-500'  },
  { label: 'Offline',  count: 6,  color: 'bg-red-500'    },
]

export default function DashboardPage() {
  return (
    <div>
      <TopBar title="Fleet Overview" subtitle="TrackOra — Real-time platform monitoring" />
      <div className="p-6 space-y-6">

        {/* Stats row */}
        <div className="grid grid-cols-2 xl:grid-cols-3 gap-4">
          <StatCard title="Total Vehicles" value="97" trend={{ value: 4, label: 'vs last month' }}
            icon={<HiOutlineTruck className="w-5 h-5" />} accent="brand" />
          <StatCard title="Active Drivers" value="73" trend={{ value: 2, label: 'vs last month' }}
            icon={<HiOutlineUserGroup className="w-5 h-5" />} accent="green" />
          <StatCard title="Active Alerts" value="14" subtitle="3 critical"
            icon={<HiOutlineBell className="w-5 h-5" />} accent="red" />
          <StatCard title="Devices Online" value="91/97" trend={{ value: -3, label: 'vs yesterday' }}
            icon={<HiOutlineCpuChip className="w-5 h-5" />} accent="cyan" />
          <StatCard title="Trips Today" value="62" trend={{ value: 12, label: 'vs yesterday' }}
            icon={<HiOutlineMap className="w-5 h-5" />} accent="purple" />
          <StatCard title="Fuel This Week" value="4,820 L" trend={{ value: -8, label: 'efficiency' }}
            icon={<HiOutlineFire className="w-5 h-5" />} accent="amber" />
        </div>

        {/* Charts row */}
        <div className="grid grid-cols-1 xl:grid-cols-3 gap-4">
          {/* Trip activity */}
          <div className="card xl:col-span-2">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold text-white text-sm">Trip Activity</h2>
              <span className="badge bg-brand-500/20 text-brand-400">This Week</span>
            </div>
            <ResponsiveContainer width="100%" height={220}>
              <AreaChart data={tripData}>
                <defs>
                  <linearGradient id="gradTrips" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%"  stopColor="#6366f1" stopOpacity={0.3} />
                    <stop offset="95%" stopColor="#6366f1" stopOpacity={0}   />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
                <XAxis dataKey="day" tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} />
                <Tooltip
                  contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', fontSize: 12 }}
                  labelStyle={{ color: '#f1f5f9' }}
                />
                <Area type="monotone" dataKey="trips" stroke="#6366f1" strokeWidth={2} fill="url(#gradTrips)" name="Trips" />
                <Area type="monotone" dataKey="km" stroke="#8b5cf6" strokeWidth={2} fill="none" name="KM" />
              </AreaChart>
            </ResponsiveContainer>
          </div>

          {/* Alert types */}
          <div className="card">
            <h2 className="font-semibold text-white text-sm mb-4">Alert Breakdown</h2>
            <ResponsiveContainer width="100%" height={220}>
              <BarChart data={alertData} layout="vertical">
                <XAxis type="number" tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} />
                <YAxis type="category" dataKey="name" tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} width={70} />
                <Tooltip
                  contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', fontSize: 12 }}
                />
                <Bar dataKey="value" fill="#6366f1" radius={[0, 4, 4, 0]} name="Alerts" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Vehicle status + recent alerts */}
        <div className="grid grid-cols-1 xl:grid-cols-2 gap-4">
          {/* Fleet status */}
          <div className="card">
            <h2 className="font-semibold text-white text-sm mb-4">Fleet Status</h2>
            <div className="space-y-3">
              {vehicleStatuses.map(s => (
                <div key={s.label} className="flex items-center gap-3">
                  <span className="text-xs text-slate-400 w-16">{s.label}</span>
                  <div className="flex-1 bg-surface rounded-full h-2 overflow-hidden">
                    <div
                      className={`h-full ${s.color} rounded-full transition-all duration-700`}
                      style={{ width: `${(s.count / 97) * 100}%` }}
                    />
                  </div>
                  <span className="text-xs text-slate-300 w-6 text-right font-medium">{s.count}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Recent alerts */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h2 className="font-semibold text-white text-sm">Recent Alerts</h2>
              <a href="/alerts" className="text-xs text-brand-400 hover:underline">View all</a>
            </div>
            <div className="space-y-2">
              {[
                { type: 'SPEEDING',     vehicle: 'TRK-042', msg: 'Speed 118 km/h — exceeded limit', time: '2m ago',  sev: 'high'   },
                { type: 'GEOFENCE',     vehicle: 'VAN-019', msg: 'Exited Mumbai Depot zone',        time: '8m ago',  sev: 'medium' },
                { type: 'MAINTENANCE',  vehicle: 'TRK-007', msg: 'Oil change overdue by 200km',     time: '15m ago', sev: 'medium' },
                { type: 'OFFLINE',      vehicle: 'CAR-033', msg: 'Device offline for 35 minutes',   time: '37m ago', sev: 'low'    },
              ].map((a, i) => (
                <div key={i} className="flex items-start gap-3 py-2 border-b border-surface-border last:border-0">
                  <span className={`badge mt-0.5 ${
                    a.sev === 'high'   ? 'bg-red-500/20 text-red-400' :
                    a.sev === 'medium' ? 'bg-amber-500/20 text-amber-400' :
                                         'bg-slate-500/20 text-slate-400'
                  }`}>{a.type}</span>
                  <div className="flex-1 min-w-0">
                    <p className="text-xs font-medium text-slate-200">{a.vehicle} — {a.msg}</p>
                    <p className="text-xs text-slate-500 mt-0.5">{a.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>

      </div>
    </div>
  )
}
