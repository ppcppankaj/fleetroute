import { useState, useEffect, useRef, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import {
  Play, Pause, SkipBack, SkipForward, ArrowLeft, Gauge,
  Navigation, Zap, Fuel, Thermometer
} from 'lucide-react'
import api from '../../shared/api/client'
import { OpenLayersAdapter } from '../../shared/map/OpenLayersAdapter'

interface PositionRecord {
  timestamp: string
  lat: number
  lng: number
  speed: number
  heading: number
  ignition: boolean
  fuel_level_pct?: number
  temperature_1_c?: number
  external_voltage_v?: number
  sos_event?: boolean
}

const SPEED_MULTIPLIERS = [1, 2, 5, 10, 30, 60]

function getSpeedColor(speed: number): string {
  if (speed > 80) return '#ef4444'  // red — overspeed
  if (speed > 50) return '#f59e0b'  // amber — approaching limit
  return '#10b981'                   // green — normal
}

export function RoutePlayback() {
  const { deviceId } = useParams<{ deviceId: string }>()
  const navigate = useNavigate()
  const mapRef = useRef<HTMLDivElement>(null)
  const adapterRef = useRef<OpenLayersAdapter | null>(null)
  const playTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)

  const [date, setDate] = useState(() => new Date().toISOString().slice(0, 10))
  const [currentIdx, setCurrentIdx] = useState(0)
  const [playing, setPlaying] = useState(false)
  const [speedIdx, setSpeedIdx] = useState(0)   // index into SPEED_MULTIPLIERS

  // Fetch trip history for the selected date
  const { data: positions = [], isLoading } = useQuery<PositionRecord[]>({
    queryKey: ['playback', deviceId, date],
    queryFn: async () => {
      const from = `${date}T00:00:00Z`
      const to   = `${date}T23:59:59Z`
      const r = await api.get(`/devices/${deviceId}/history`, { params: { from, to, limit: 2000 } })
      return (r.data.data ?? []) as PositionRecord[]
    },
    enabled: !!deviceId,
  })

  const current = positions[currentIdx]
  const total   = positions.length

  // Init map
  useEffect(() => {
    if (!mapRef.current) return
    adapterRef.current = new OpenLayersAdapter()
    adapterRef.current.init(mapRef.current)
    return () => { adapterRef.current?.destroy(); adapterRef.current = null }
  }, [])

  // Draw route whenever positions load
  useEffect(() => {
    const adapter = adapterRef.current
    if (!adapter || positions.length < 2) return

    const coords = positions.map((p) => ({ lat: p.lat, lng: p.lng }))
    adapter.clearAll?.()
    adapter.addPolyline?.('route', coords, '#10b981', 3)

    // Fit to route
    if (coords.length > 0) {
      adapter.fitBounds?.(coords)
    }
    setCurrentIdx(0)
    setPlaying(false)
  }, [positions])

  // Update marker position during playback
  useEffect(() => {
    if (!current || !adapterRef.current) return
    adapterRef.current.updateMarker(deviceId!, {
      lat: current.lat,
      lng: current.lng,
      heading: current.heading,
      speed: current.speed,
      ignition: current.ignition,
      online: true,
      label: '',
    })
    adapterRef.current.flyTo?.(current.lat, current.lng, 14)
  }, [currentIdx, current, deviceId])

  // Playback tick
  useEffect(() => {
    if (playing && playTimerRef.current === null) {
      const mult = SPEED_MULTIPLIERS[speedIdx]
      // Assume 30-second GPS interval  →  tick interval = 30000 / mult ms
      const interval = Math.max(50, 30000 / mult)
      playTimerRef.current = setInterval(() => {
        setCurrentIdx((idx) => {
          if (idx >= total - 1) {
            setPlaying(false)
            return idx
          }
          return idx + 1
        })
      }, interval)
    }
    return () => {
      if (playTimerRef.current) {
        clearInterval(playTimerRef.current)
        playTimerRef.current = null
      }
    }
  }, [playing, speedIdx, total])

  const togglePlay = useCallback(() => setPlaying((p) => !p), [])
  const stepBack  = useCallback(() => setCurrentIdx((i) => Math.max(0, i - 1)), [])
  const stepFwd   = useCallback(() => setCurrentIdx((i) => Math.min(total - 1, i + 1)), [total])
  const cycleSpeed = useCallback(() => setSpeedIdx((i) => (i + 1) % SPEED_MULTIPLIERS.length), [])

  const progressPct = total > 1 ? (currentIdx / (total - 1)) * 100 : 0

  return (
    <div className="page" style={{ position: 'relative', height: '100%', display: 'flex', flexDirection: 'column' }}>
      {/* Top bar */}
      <div className="flex items-center gap-3" style={{ padding: '12px 20px', background: 'var(--color-surface)', borderBottom: '1px solid var(--color-border)', flexShrink: 0 }}>
        <button className="btn btn-ghost btn-sm" onClick={() => navigate(-1)}>
          <ArrowLeft size={16} /> Back
        </button>
        <div style={{ flex: 1 }}>
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>Route Playback</h2>
          <p className="text-sm text-muted">Device: <span className="font-mono">{deviceId}</span></p>
        </div>
        <input
          type="date"
          value={date}
          max={new Date().toISOString().slice(0, 10)}
          onChange={(e) => { setDate(e.target.value); setPlaying(false) }}
          style={{ background: 'var(--color-surface-raised)', border: '1px solid var(--color-border)', borderRadius: 8, padding: '6px 12px', color: 'var(--color-text)', fontSize: 14 }}
        />
      </div>

      {/* Map + telemetry */}
      <div style={{ flex: 1, display: 'flex', position: 'relative', overflow: 'hidden' }}>
        <div ref={mapRef} style={{ flex: 1 }} />

        {/* Telemetry panel */}
        {current && (
          <div
            className="card animate-fade-in"
            style={{
              position: 'absolute', top: 16, right: 16, width: 220,
              padding: 16, display: 'flex', flexDirection: 'column', gap: 12,
            }}
          >
            <div className="text-xs text-muted font-mono">
              {new Date(current.timestamp).toLocaleTimeString()}
            </div>
            <TelemetryItem icon={<Gauge size={15} />} label="Speed" value={`${current.speed} km/h`}
              valueStyle={{ color: getSpeedColor(current.speed), fontWeight: 700 }} />
            <TelemetryItem icon={<Navigation size={15} />} label="Heading" value={`${current.heading}°`} />
            <TelemetryItem icon={<Zap size={15} />} label="Ignition"
              value={current.ignition ? 'ON' : 'OFF'}
              valueStyle={{ color: current.ignition ? 'var(--color-success)' : 'var(--color-text-muted)' }} />
            {current.fuel_level_pct !== undefined && (
              <TelemetryItem icon={<Fuel size={15} />} label="Fuel" value={`${current.fuel_level_pct}%`} />
            )}
            {current.temperature_1_c !== undefined && (
              <TelemetryItem icon={<Thermometer size={15} />} label="Temp" value={`${current.temperature_1_c}°C`} />
            )}
            <div style={{ fontSize: 11, color: 'var(--color-text-muted)' }}>
              {current.lat.toFixed(6)}, {current.lng.toFixed(6)}
            </div>
            {current.sos_event && (
              <div className="badge badge-danger" style={{ justifyContent: 'center' }}>⚠ SOS ACTIVE</div>
            )}
          </div>
        )}

        {/* Loading */}
        {isLoading && (
          <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.4)' }}>
            <div style={{ color: '#fff', fontSize: 18 }}>Loading route…</div>
          </div>
        )}

        {!isLoading && total === 0 && (
          <div style={{ position: 'absolute', inset: 0, display: 'flex', alignItems: 'center', justifyContent: 'center', background: 'rgba(0,0,0,0.4)' }}>
            <div className="card" style={{ padding: 32, textAlign: 'center' }}>
              <p style={{ color: 'var(--color-text-muted)' }}>No GPS data for this date</p>
            </div>
          </div>
        )}
      </div>

      {/* Timeline + controls */}
      <div
        style={{
          padding: '16px 24px', background: 'var(--color-surface)', borderTop: '1px solid var(--color-border)',
          flexShrink: 0,
        }}
      >
        {/* Timeline slider */}
        <div style={{ marginBottom: 12 }}>
          <input
            type="range"
            min={0}
            max={Math.max(0, total - 1)}
            value={currentIdx}
            onChange={(e) => { setCurrentIdx(+e.target.value); setPlaying(false) }}
            style={{ width: '100%', accentColor: 'var(--color-accent)', cursor: 'pointer' }}
          />
          <div className="flex justify-between text-xs text-muted" style={{ marginTop: 4 }}>
            <span>{total > 0 ? new Date(positions[0].timestamp).toLocaleTimeString() : '--'}</span>
            <span className="badge badge-info">{currentIdx + 1} / {total} positions</span>
            <span>{total > 0 ? new Date(positions[total - 1].timestamp).toLocaleTimeString() : '--'}</span>
          </div>
        </div>

        {/* Controls */}
        <div className="flex items-center gap-3" style={{ justifyContent: 'center' }}>
          <button id="playback-step-back" className="btn btn-secondary btn-sm" onClick={stepBack} disabled={currentIdx === 0}>
            <SkipBack size={15} />
          </button>
          <button
            id="playback-play-pause"
            className={`btn btn-sm ${playing ? 'btn-warning' : 'btn-primary'}`}
            onClick={togglePlay}
            disabled={total === 0}
            style={{ minWidth: 100 }}
          >
            {playing ? <><Pause size={15} /> Pause</> : <><Play size={15} /> Play</>}
          </button>
          <button id="playback-step-fwd" className="btn btn-secondary btn-sm" onClick={stepFwd} disabled={currentIdx >= total - 1}>
            <SkipForward size={15} />
          </button>
          <button
            id="playback-speed"
            className="btn btn-ghost btn-sm"
            onClick={cycleSpeed}
            title="Cycle playback speed"
            style={{ minWidth: 56, fontWeight: 700 }}
          >
            {SPEED_MULTIPLIERS[speedIdx]}×
          </button>
          <div style={{ flex: 1, maxWidth: 300 }}>
            <div style={{ height: 4, background: 'var(--color-surface-raised)', borderRadius: 2 }}>
              <div
                style={{
                  width: `${progressPct}%`,
                  height: '100%',
                  background: 'var(--color-accent)',
                  borderRadius: 2,
                  transition: 'width 0.1s linear',
                }}
              />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

function TelemetryItem({ icon, label, value, valueStyle }: {
  icon: React.ReactNode
  label: string
  value: string
  valueStyle?: React.CSSProperties
}) {
  return (
    <div className="flex items-center gap-2">
      <span style={{ color: 'var(--color-text-muted)' }}>{icon}</span>
      <span className="text-xs text-muted" style={{ flex: 1 }}>{label}</span>
      <span className="text-sm" style={{ fontWeight: 600, ...valueStyle }}>{value}</span>
    </div>
  )
}
