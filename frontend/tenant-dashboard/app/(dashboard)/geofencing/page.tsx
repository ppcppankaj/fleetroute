'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineGlobeAlt, HiOutlinePlusCircle, HiOutlineTrash } from 'react-icons/hi2'
import { useState } from 'react'

const ZONES = [
  { id: 'z1', name: 'Mumbai Depot',    type: 'POLYGON', vehicles: 6, status: 'ACTIVE',   updated: '2026-04-20' },
  { id: 'z2', name: 'Airport Pickup',  type: 'CIRCLE',  vehicles: 2, status: 'ACTIVE',   updated: '2026-04-18' },
  { id: 'z3', name: 'Industrial Zone', type: 'POLYGON', vehicles: 0, status: 'INACTIVE', updated: '2026-04-10' },
  { id: 'z4', name: 'Restricted Area', type: 'POLYGON', vehicles: 0, status: 'ACTIVE',   updated: '2026-04-22' },
]

export default function GeofencingPage() {
  const [showModal, setShowModal] = useState(false)

  return (
    <div>
      <TopBar title="Geofencing" subtitle="Manage geographic zones & boundaries" />
      <div className="p-6 space-y-5">
        {/* Map preview */}
        <div className="card p-0 overflow-hidden">
          <div className="h-64 bg-[#1a2438] relative flex items-center justify-center"
            style={{ backgroundImage: 'radial-gradient(ellipse at center, #1e3a5f 0%, #0f172a 80%)' }}>
            <svg className="absolute inset-0 w-full h-full opacity-10" xmlns="http://www.w3.org/2000/svg">
              <defs><pattern id="g2" width="40" height="40" patternUnits="userSpaceOnUse">
                <path d="M 40 0 L 0 0 0 40" fill="none" stroke="#6366f1" strokeWidth="0.5"/>
              </pattern></defs>
              <rect width="100%" height="100%" fill="url(#g2)" />
            </svg>

            {/* Geofence polygons */}
            <div className="absolute w-48 h-36 border-2 border-green-500/60 bg-green-500/10 rounded-lg top-8 left-24 flex items-center justify-center">
              <span className="text-xs text-green-400 font-medium">Mumbai Depot</span>
            </div>
            <div className="absolute w-24 h-24 border-2 border-brand-500/60 bg-brand-500/10 rounded-full bottom-8 right-32 flex items-center justify-center">
              <span className="text-xs text-brand-400 font-medium text-center">Airport</span>
            </div>
            <div className="absolute w-32 h-20 border-2 border-red-500/60 bg-red-500/10 rounded top-12 right-16 flex items-center justify-center">
              <span className="text-xs text-red-400 font-medium">Restricted</span>
            </div>

            <div className="absolute bottom-3 right-3 glass rounded px-2 py-1 text-xs text-slate-400">
              <HiOutlineGlobeAlt className="inline w-3 h-3 mr-1" /> Mumbai Region
            </div>
          </div>
        </div>

        {/* Toolbar */}
        <div className="flex justify-between items-center">
          <h2 className="text-sm font-semibold text-white">Defined Zones ({ZONES.length})</h2>
          <button onClick={() => setShowModal(true)} className="btn-primary flex items-center gap-2 text-sm">
            <HiOutlinePlusCircle className="w-4 h-4" /> Create Zone
          </button>
        </div>

        {/* Zone cards */}
        <div className="grid grid-cols-2 gap-4">
          {ZONES.map(z => (
            <div key={z.id} className="card flex items-start gap-4 hover:border-brand-500/30 transition-all group">
              <div className={`w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0 ${z.type === 'CIRCLE' ? 'bg-brand-500/20' : 'bg-green-500/20'}`}>
                <HiOutlineGlobeAlt className={`w-5 h-5 ${z.type === 'CIRCLE' ? 'text-brand-400' : 'text-green-400'}`} />
              </div>
              <div className="flex-1">
                <div className="flex items-center gap-2">
                  <span className="text-sm font-semibold text-white">{z.name}</span>
                  <span className={`badge ${z.status === 'ACTIVE' ? 'bg-green-500/20 text-green-400' : 'bg-slate-500/20 text-slate-400'}`}>{z.status}</span>
                </div>
                <div className="flex items-center gap-3 mt-1 text-xs text-slate-500">
                  <span>{z.type}</span>
                  <span>{z.vehicles} vehicles inside</span>
                  <span>Updated {z.updated}</span>
                </div>
              </div>
              <button className="btn-ghost p-2 opacity-0 group-hover:opacity-100 transition-opacity text-red-400 hover:text-red-300">
                <HiOutlineTrash className="w-4 h-4" />
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
