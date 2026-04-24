'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineFire, HiOutlinePlusCircle } from 'react-icons/hi2'
import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts'

const weeklyFuel = [
  { day: 'Mon', liters: 620, cost: 62000 },
  { day: 'Tue', liters: 540, cost: 54000 },
  { day: 'Wed', liters: 780, cost: 78000 },
  { day: 'Thu', liters: 450, cost: 45000 },
  { day: 'Fri', liters: 890, cost: 89000 },
  { day: 'Sat', liters: 310, cost: 31000 },
  { day: 'Sun', liters: 230, cost: 23000 },
]

const logs = [
  { vehicle: 'TRK-042', driver: 'Ravi Sharma',  liters: 85,  cost: 8500,  station: 'HP Petrol',  type: 'Diesel', date: '2026-04-24 09:15' },
  { vehicle: 'VAN-019', driver: 'Suresh Kumar', liters: 42,  cost: 4200,  station: 'IOC Bhandup', type: 'Diesel', date: '2026-04-24 11:30' },
  { vehicle: 'BUS-001', driver: 'Kiran Rao',    liters: 120, cost: 12000, station: 'BPCL Thane',  type: 'Diesel', date: '2026-04-23 16:00' },
  { vehicle: 'CAR-033', driver: 'Deepak Patel', liters: 38,  cost: 3800,  station: 'Reliance',    type: 'Petrol', date: '2026-04-23 14:22' },
]

export default function FuelPage() {
  const totalLiters = weeklyFuel.reduce((s, d) => s + d.liters, 0)
  const totalCost = weeklyFuel.reduce((s, d) => s + d.cost, 0)

  return (
    <div>
      <TopBar title="Fuel Management" subtitle="Track fuel consumption & costs" />
      <div className="p-6 space-y-5">
        {/* KPI */}
        <div className="grid grid-cols-3 gap-4">
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">This Week</p>
            <p className="text-3xl font-bold text-white mt-2">{totalLiters.toLocaleString()} <span className="text-base text-slate-400">L</span></p>
            <p className="text-xs text-green-400 mt-1">↓ 8% vs last week</p>
          </div>
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Total Cost</p>
            <p className="text-3xl font-bold text-white mt-2">₹{(totalCost / 1000).toFixed(0)}k</p>
            <p className="text-xs text-red-400 mt-1">↑ 3% vs last week</p>
          </div>
          <div className="card">
            <p className="text-xs text-slate-400 uppercase tracking-wider">Avg Efficiency</p>
            <p className="text-3xl font-bold text-white mt-2">8.2 <span className="text-base text-slate-400">km/L</span></p>
            <p className="text-xs text-slate-500 mt-1">Fleet average</p>
          </div>
        </div>

        {/* Chart */}
        <div className="card">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-white">Daily Fuel Consumption</h2>
            <button className="btn-primary text-xs flex items-center gap-1">
              <HiOutlinePlusCircle className="w-3.5 h-3.5" /> Log Fuel
            </button>
          </div>
          <ResponsiveContainer width="100%" height={200}>
            <AreaChart data={weeklyFuel}>
              <defs>
                <linearGradient id="fuelGrad" x1="0" y1="0" x2="0" y2="1">
                  <stop offset="5%"  stopColor="#f59e0b" stopOpacity={0.3} />
                  <stop offset="95%" stopColor="#f59e0b" stopOpacity={0}   />
                </linearGradient>
              </defs>
              <CartesianGrid strokeDasharray="3 3" stroke="#334155" />
              <XAxis dataKey="day" tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} />
              <YAxis tick={{ fill: '#94a3b8', fontSize: 11 }} axisLine={false} tickLine={false} />
              <Tooltip contentStyle={{ background: '#1e293b', border: '1px solid #334155', borderRadius: '8px', fontSize: 12 }} />
              <Area type="monotone" dataKey="liters" stroke="#f59e0b" strokeWidth={2} fill="url(#fuelGrad)" name="Liters" />
            </AreaChart>
          </ResponsiveContainer>
        </div>

        {/* Fuel log table */}
        <div className="card overflow-hidden p-0">
          <div className="px-5 py-3 border-b border-surface-border flex items-center gap-2">
            <HiOutlineFire className="w-4 h-4 text-amber-400" />
            <span className="text-sm font-semibold text-white">Recent Fuel Logs</span>
          </div>
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Vehicle', 'Driver', 'Liters', 'Cost (₹)', 'Station', 'Type', 'Date/Time'].map(h => (
                  <th key={h} className="px-4 py-2.5 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {logs.map((l, i) => (
                <tr key={i} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-semibold text-brand-400">{l.vehicle}</td>
                  <td className="px-4 py-3 text-slate-300">{l.driver}</td>
                  <td className="px-4 py-3 text-white font-mono">{l.liters} L</td>
                  <td className="px-4 py-3 text-amber-400 font-mono">₹{l.cost.toLocaleString()}</td>
                  <td className="px-4 py-3 text-slate-400">{l.station}</td>
                  <td className="px-4 py-3"><span className="badge bg-slate-500/20 text-slate-300">{l.type}</span></td>
                  <td className="px-4 py-3 text-slate-400 font-mono text-xs">{l.date}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
