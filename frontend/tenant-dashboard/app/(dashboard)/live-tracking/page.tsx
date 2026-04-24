'use client'

import TopBar from '@/components/TopBar'
import { useEffect, useRef, useState } from 'react'
import { HiOutlineTruck, HiOutlineSignal, HiOutlineMapPin } from 'react-icons/hi2'

interface VehiclePing {
  vehicle_id: string
  lat: number
  lng: number
  speed: number
  heading: number
  ignition: boolean
  timestamp: string
}

const MOCK_VEHICLES: VehiclePing[] = [
  { vehicle_id: 'TRK-042', lat: 19.076, lng: 72.877, speed: 62, heading: 45,  ignition: true,  timestamp: '' },
  { vehicle_id: 'VAN-019', lat: 19.054, lng: 72.842, speed: 0,  heading: 0,   ignition: false, timestamp: '' },
  { vehicle_id: 'CAR-011', lat: 19.091, lng: 72.865, speed: 38, heading: 180, ignition: true,  timestamp: '' },
  { vehicle_id: 'TRK-007', lat: 19.038, lng: 72.910, speed: 85, heading: 90,  ignition: true,  timestamp: '' },
]

export default function LiveTrackingPage() {
  const mapRef = useRef<HTMLDivElement>(null)
  const [selected, setSelected] = useState<VehiclePing | null>(null)
  const [vehicles, setVehicles] = useState<VehiclePing[]>(MOCK_VEHICLES)
  const [connected, setConnected] = useState(false)

  // Simulate live updates
  useEffect(() => {
    setConnected(true)
    const interval = setInterval(() => {
      setVehicles(prev =>
        prev.map(v => ({
          ...v,
          lat: v.ignition ? v.lat + (Math.random() - 0.5) * 0.002 : v.lat,
          lng: v.ignition ? v.lng + (Math.random() - 0.5) * 0.002 : v.lng,
          speed: v.ignition ? Math.max(0, v.speed + (Math.random() - 0.5) * 10) : 0,
          timestamp: new Date().toISOString(),
        }))
      )
    }, 2000)
    return () => clearInterval(interval)
  }, [])

  const sel = selected ? vehicles.find(v => v.vehicle_id === selected.vehicle_id) : null

  return (
    <div className="h-screen flex flex-col overflow-hidden">
      <TopBar title="Live Tracking" subtitle="Real-time GPS fleet positions" />
      
      <div className="flex flex-1 overflow-hidden">
        {/* Vehicle list sidebar */}
        <div className="w-72 border-r border-surface-border bg-surface-card overflow-y-auto flex-shrink-0">
          <div className="p-4 border-b border-surface-border flex items-center justify-between">
            <span className="text-sm font-medium text-white">{vehicles.length} Vehicles</span>
            <span className={`badge ${connected ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'} animate-pulse`}>
              {connected ? '● LIVE' : '○ OFFLINE'}
            </span>
          </div>
          <div className="space-y-0 divide-y divide-surface-border">
            {vehicles.map(v => (
              <button
                key={v.vehicle_id}
                onClick={() => setSelected(v)}
                className={`w-full text-left px-4 py-3 hover:bg-surface-hover transition-colors ${selected?.vehicle_id === v.vehicle_id ? 'bg-brand-500/10 border-l-2 border-brand-500' : ''}`}
              >
                <div className="flex items-center justify-between mb-1">
                  <span className="text-sm font-semibold text-white flex items-center gap-2">
                    <HiOutlineTruck className="w-3.5 h-3.5 text-brand-400" />
                    {v.vehicle_id}
                  </span>
                  <span className={`w-2 h-2 rounded-full ${v.ignition ? 'bg-green-500 animate-pulse' : 'bg-slate-500'}`} />
                </div>
                <div className="flex items-center gap-3 text-xs text-slate-500">
                  <span>{v.speed.toFixed(0)} km/h</span>
                  <span className="text-slate-600">·</span>
                  <span>{v.lat.toFixed(4)}, {v.lng.toFixed(4)}</span>
                </div>
              </button>
            ))}
          </div>
        </div>

        {/* Map area */}
        <div className="flex-1 relative bg-surface" ref={mapRef}>
          {/* Simulated map placeholder — in production replaced by react-leaflet */}
          <div className="absolute inset-0 overflow-hidden">
            <div className="w-full h-full bg-[#1a2438] relative"
              style={{
                backgroundImage: 'radial-gradient(circle at 50% 50%, #1e3a5f 0%, #0f172a 70%)',
              }}>
              
              {/* Grid lines for map feel */}
              <svg className="absolute inset-0 w-full h-full opacity-10" xmlns="http://www.w3.org/2000/svg">
                <defs>
                  <pattern id="grid" width="60" height="60" patternUnits="userSpaceOnUse">
                    <path d="M 60 0 L 0 0 0 60" fill="none" stroke="#6366f1" strokeWidth="0.5"/>
                  </pattern>
                </defs>
                <rect width="100%" height="100%" fill="url(#grid)" />
              </svg>

              {/* Vehicle markers */}
              {vehicles.map((v, i) => {
                const x = ((v.lng - 72.83) / 0.1) * 60 + 30
                const y = ((19.10 - v.lat) / 0.07) * 60 + 20
                return (
                  <button
                    key={v.vehicle_id}
                    onClick={() => setSelected(v)}
                    className="absolute transform -translate-x-1/2 -translate-y-1/2 group"
                    style={{ left: `${Math.min(95, Math.max(5, x))}%`, top: `${Math.min(90, Math.max(10, y))}%` }}
                  >
                    <div className={`w-8 h-8 rounded-full flex items-center justify-center shadow-lg transition-transform group-hover:scale-125
                      ${v.ignition ? 'bg-green-500/90 shadow-green-500/30' : 'bg-slate-500/90 shadow-slate-500/20'}`}>
                      <HiOutlineTruck className="w-4 h-4 text-white" />
                    </div>
                    {v.ignition && (
                      <div className="absolute -inset-1 rounded-full bg-green-500/20 animate-ping" />
                    )}
                    <div className="absolute bottom-full left-1/2 -translate-x-1/2 mb-2 px-2 py-1 bg-surface-card border border-surface-border rounded text-xs text-white whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
                      {v.vehicle_id} · {v.speed.toFixed(0)} km/h
                    </div>
                  </button>
                )
              })}

              {/* Geofence demo circle */}
              <div
                className="absolute border-2 border-brand-500/40 rounded-full bg-brand-500/5"
                style={{ width: 120, height: 120, left: '45%', top: '35%', transform: 'translate(-50%,-50%)' }}
              />

              {/* Map label */}
              <div className="absolute bottom-4 right-4 glass rounded-lg px-3 py-2 text-xs text-slate-400">
                <HiOutlineMapPin className="inline w-3 h-3 mr-1" />
                Mumbai Metropolitan Region
              </div>
              <div className="absolute top-4 left-4 glass rounded-lg px-3 py-2 text-xs text-slate-400">
                ℹ️ Simulated positions — connect to live WebSocket for production
              </div>
            </div>
          </div>

          {/* Selected vehicle panel */}
          {sel && (
            <div className="absolute bottom-4 left-1/2 -translate-x-1/2 glass rounded-xl p-4 min-w-80 shadow-xl">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <HiOutlineTruck className="w-4 h-4 text-brand-400" />
                  <span className="font-semibold text-white text-sm">{sel.vehicle_id}</span>
                  <span className={`badge ${sel.ignition ? 'bg-green-500/20 text-green-400' : 'bg-slate-500/20 text-slate-400'}`}>
                    {sel.ignition ? 'Moving' : 'Parked'}
                  </span>
                </div>
                <button onClick={() => setSelected(null)} className="text-slate-500 hover:text-white text-lg leading-none">&times;</button>
              </div>
              <div className="grid grid-cols-3 gap-3 text-center">
                <div className="bg-surface/60 rounded-lg p-2">
                  <p className="text-lg font-bold text-white">{sel.speed.toFixed(0)}</p>
                  <p className="text-xs text-slate-500">km/h</p>
                </div>
                <div className="bg-surface/60 rounded-lg p-2">
                  <p className="text-xs font-mono text-white">{sel.lat.toFixed(4)}</p>
                  <p className="text-xs text-slate-500">Latitude</p>
                </div>
                <div className="bg-surface/60 rounded-lg p-2">
                  <p className="text-xs font-mono text-white">{sel.lng.toFixed(4)}</p>
                  <p className="text-xs text-slate-500">Longitude</p>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Connection status + signal strength panel */}
        <div className="w-56 border-l border-surface-border bg-surface-card p-4 overflow-y-auto flex-shrink-0">
          <p className="text-xs font-semibold text-slate-400 uppercase tracking-wider mb-3 flex items-center gap-2">
            <HiOutlineSignal className="w-3.5 h-3.5" /> Live Stats
          </p>
          <div className="space-y-3">
            {[
              { label: 'Moving',  val: vehicles.filter(v => v.ignition).length,  color: 'text-green-400' },
              { label: 'Parked',  val: vehicles.filter(v => !v.ignition).length, color: 'text-slate-400' },
              { label: 'Avg Speed', val: `${(vehicles.filter(v => v.ignition).reduce((s, v) => s + v.speed, 0) / Math.max(1, vehicles.filter(v => v.ignition).length)).toFixed(0)} km/h`, color: 'text-brand-400' },
            ].map(s => (
              <div key={s.label} className="flex justify-between items-center py-1.5 border-b border-surface-border">
                <span className="text-xs text-slate-500">{s.label}</span>
                <span className={`text-sm font-bold ${s.color}`}>{s.val}</span>
              </div>
            ))}
          </div>

          <div className="mt-4">
            <p className="text-xs text-slate-500 mb-2">Update Rate</p>
            <div className="flex items-end gap-1 h-8">
              {[4, 7, 5, 9, 6, 8, 10, 7, 9, 8].map((h, i) => (
                <div key={i} className="flex-1 bg-brand-500 rounded-sm opacity-70" style={{ height: `${h * 10}%` }} />
              ))}
            </div>
            <p className="text-xs text-slate-600 mt-1 text-center">2s interval</p>
          </div>
        </div>
      </div>
    </div>
  )
}
