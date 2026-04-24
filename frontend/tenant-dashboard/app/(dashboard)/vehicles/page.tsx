'use client'

import TopBar from '@/components/TopBar'
import { useState } from 'react'
import { HiOutlineTruck, HiOutlinePlusCircle, HiOutlineMagnifyingGlass, HiOutlinePencilSquare, HiOutlineSignalSlash } from 'react-icons/hi2'

const MOCK_VEHICLES = [
  { id: 'v1', plate: 'MH12AB1234', type: 'Truck',  make: 'Tata',   model: 'Signa 4825', year: 2022, status: 'ACTIVE',   odometer: 84200, driver: 'Ravi Sharma' },
  { id: 'v2', plate: 'MH03CD5678', type: 'Van',    make: 'Force',  model: 'Traveller',  year: 2021, status: 'ACTIVE',   odometer: 51400, driver: 'Suresh Kumar' },
  { id: 'v3', plate: 'MH14EF9012', type: 'Car',    make: 'Maruti', model: 'Eeco',       year: 2023, status: 'INACTIVE', odometer: 12300, driver: 'Unassigned' },
  { id: 'v4', plate: 'MH01GH3456', type: 'Truck',  make: 'Ashok',  model: 'Leyland 1917', year: 2020, status: 'MAINTENANCE', odometer: 145000, driver: 'Amit Singh' },
  { id: 'v5', plate: 'MH04IJ7890', type: 'Bus',    make: 'Volvo',  model: 'B9R',        year: 2019, status: 'ACTIVE',   odometer: 320000, driver: 'Deepak Patel' },
]

const statusColors: Record<string, string> = {
  ACTIVE:      'bg-green-500/20 text-green-400',
  INACTIVE:    'bg-slate-500/20 text-slate-400',
  MAINTENANCE: 'bg-amber-500/20 text-amber-400',
}

export default function VehiclesPage() {
  const [search, setSearch] = useState('')
  const filtered = MOCK_VEHICLES.filter(v =>
    v.plate.toLowerCase().includes(search.toLowerCase()) ||
    v.driver.toLowerCase().includes(search.toLowerCase()) ||
    v.make.toLowerCase().includes(search.toLowerCase()),
  )

  return (
    <div>
      <TopBar title="Vehicles" subtitle="Manage your fleet" />
      <div className="p-6 space-y-4">
        {/* Toolbar */}
        <div className="flex items-center gap-3">
          <div className="relative flex-1 max-w-sm">
            <HiOutlineMagnifyingGlass className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-slate-500" />
            <input
              className="input pl-9"
              placeholder="Search vehicles, plates, drivers…"
              value={search}
              onChange={e => setSearch(e.target.value)}
            />
          </div>
          <select className="input w-auto text-xs">
            <option>All Types</option>
            <option>Truck</option><option>Van</option><option>Car</option><option>Bus</option>
          </select>
          <select className="input w-auto text-xs">
            <option>All Statuses</option>
            <option>Active</option><option>Inactive</option><option>Maintenance</option>
          </select>
          <button className="btn-primary flex items-center gap-2 text-sm ml-auto">
            <HiOutlinePlusCircle className="w-4 h-4" /> Add Vehicle
          </button>
        </div>

        {/* Table */}
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/30">
                {['Plate No.', 'Type', 'Make / Model', 'Year', 'Odometer', 'Driver', 'Status', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider whitespace-nowrap">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {filtered.map(v => (
                <tr key={v.id} className="hover:bg-surface-hover/40 transition-colors group">
                  <td className="px-4 py-3 font-mono font-semibold text-white text-xs tracking-wide">{v.plate}</td>
                  <td className="px-4 py-3">
                    <span className="flex items-center gap-1.5 text-slate-300">
                      <HiOutlineTruck className="w-3.5 h-3.5 text-brand-400" />
                      {v.type}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-300">{v.make} {v.model}</td>
                  <td className="px-4 py-3 text-slate-400">{v.year}</td>
                  <td className="px-4 py-3 text-slate-300 font-mono">{v.odometer.toLocaleString()} km</td>
                  <td className="px-4 py-3 text-slate-300">{v.driver}</td>
                  <td className="px-4 py-3">
                    <span className={`badge ${statusColors[v.status]}`}>{v.status}</span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      <button className="btn-ghost p-1.5 rounded"><HiOutlinePencilSquare className="w-3.5 h-3.5" /></button>
                      <button className="btn-ghost p-1.5 rounded text-red-400 hover:text-red-300"><HiOutlineSignalSlash className="w-3.5 h-3.5" /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {filtered.length === 0 && (
            <div className="text-center py-12 text-slate-500">No vehicles match your search.</div>
          )}
        </div>

        <p className="text-xs text-slate-600 text-right">{filtered.length} of {MOCK_VEHICLES.length} vehicles</p>
      </div>
    </div>
  )
}
