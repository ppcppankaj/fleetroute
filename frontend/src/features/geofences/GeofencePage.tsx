import { useState, useEffect, useRef, useCallback } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Shield, Plus, Eye, EyeOff, Edit2, Trash2, Circle, Square, Minus } from 'lucide-react'
import api from '../../shared/api/client'
import { OpenLayersAdapter } from '../../shared/map/OpenLayersAdapter'
import {
  AreaChart, Area, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer
} from 'recharts'

interface Geofence {
  id: string
  name: string
  shape_type: 'polygon' | 'circle' | 'corridor'
  geometry: string
  created_at: string
  vehicle_groups?: string[]
}

type DrawMode = 'none' | 'polygon' | 'circle' | 'corridor'

export function GeofencePage() {
  const qc = useQueryClient()
  const mapRef = useRef<HTMLDivElement>(null)
  const adapterRef = useRef<OpenLayersAdapter | null>(null)
  const [selected, setSelected] = useState<Geofence | null>(null)
  const [visible, setVisible] = useState<Record<string, boolean>>({})
  const [drawMode, setDrawMode] = useState<DrawMode>('none')
  const [showCreateForm, setShowCreateForm] = useState(false)
  const [newName, setNewName] = useState('')
  const [drawnGeoJSON, setDrawnGeoJSON] = useState<string | null>(null)

  const { data: geofences = [] } = useQuery<Geofence[]>({
    queryKey: ['geofences'],
    queryFn: async () => {
      const r = await api.get('/geofences')
      return r.data.data ?? []
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/geofences/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['geofences'] }),
  })

  const createMutation = useMutation({
    mutationFn: (body: { name: string; shape_type: string; geojson: string }) =>
      api.post('/geofences', body),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['geofences'] })
      setShowCreateForm(false)
      setNewName('')
      setDrawnGeoJSON(null)
      setDrawMode('none')
    },
  })

  // Mock dwell time data for the selected geofence
  const dwellData = [
    { vehicle: 'MH01AB1234', avg: 12, max: 35 },
    { vehicle: 'MH02CD5678', avg: 8, max: 22 },
    { vehicle: 'DL3EF9012', avg: 25, max: 60 },
  ]

  // Initialize map
  useEffect(() => {
    if (!mapRef.current) return
    adapterRef.current = new OpenLayersAdapter()
    adapterRef.current.init(mapRef.current)
    return () => { adapterRef.current?.destroy(); adapterRef.current = null }
  }, [])

  const toggleVisibility = useCallback((id: string) => {
    setVisible((v) => ({ ...v, [id]: !v[id] }))
  }, [])

  const startDraw = (mode: DrawMode) => {
    setDrawMode(mode)
    setShowCreateForm(mode !== 'none')
    // In a real implementation: adapterRef.current?.enableDrawMode(mode)
  }

  const handleCreate = () => {
    if (!newName.trim()) return
    // Use drawn geometry or a default for demo
    const geoJSON = drawnGeoJSON ?? JSON.stringify({
      type: 'Polygon',
      coordinates: [[[72.8, 18.9], [72.9, 18.9], [72.9, 19.0], [72.8, 19.0], [72.8, 18.9]]]
    })
    createMutation.mutate({ name: newName, shape_type: drawMode === 'none' ? 'polygon' : drawMode, geojson: geoJSON })
  }

  const shapeIcon = (type: string) => {
    if (type === 'circle') return <Circle size={13} />
    if (type === 'corridor') return <Minus size={13} />
    return <Square size={13} />
  }

  return (
    <div className="page flex" style={{ flexDirection: 'row', height: '100%' }}>
      {/* Left Panel */}
      <div
        className="card animate-fade-in"
        style={{ width: 320, minWidth: 320, height: '100%', display: 'flex', flexDirection: 'column', borderRadius: 0, overflowY: 'auto' }}
      >
        <div className="card-header" style={{ flexShrink: 0 }}>
          <div className="flex items-center gap-2">
            <Shield size={18} style={{ color: 'var(--color-accent)' }} />
            <h2 className="card-title">Geofences</h2>
          </div>
          <span className="badge badge-info">{geofences.length}</span>
        </div>

        {/* Draw controls */}
        <div className="flex gap-2" style={{ padding: '12px 16px', borderBottom: '1px solid var(--color-border)', flexShrink: 0 }}>
          <button
            id="draw-polygon"
            className={`btn btn-sm ${drawMode === 'polygon' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => startDraw('polygon')}
            title="Draw Polygon"
          >
            <Square size={13} /> Polygon
          </button>
          <button
            id="draw-circle"
            className={`btn btn-sm ${drawMode === 'circle' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => startDraw('circle')}
            title="Draw Circle"
          >
            <Circle size={13} /> Circle
          </button>
          <button
            id="draw-corridor"
            className={`btn btn-sm ${drawMode === 'corridor' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => startDraw('corridor')}
            title="Draw Corridor"
          >
            <Minus size={13} /> Corridor
          </button>
        </div>

        {/* Create form */}
        {showCreateForm && (
          <div style={{ padding: '12px 16px', background: 'var(--color-surface-raised)', borderBottom: '1px solid var(--color-border)', flexShrink: 0 }}>
            <p className="text-sm text-muted" style={{ marginBottom: 8 }}>
              {drawMode === 'polygon' && 'Click to add vertices on the map, double-click to close.'}
              {drawMode === 'circle' && 'Click center point, then drag to set radius.'}
              {drawMode === 'corridor' && 'Draw centerline, then set buffer width.'}
            </p>
            <input
              type="text"
              className="form-input"
              placeholder="Geofence name…"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              style={{ marginBottom: 8, width: '100%' }}
            />
            <div className="flex gap-2">
              <button className="btn btn-primary btn-sm flex-1" onClick={handleCreate} disabled={!newName.trim()}>
                <Plus size={13} /> Save
              </button>
              <button className="btn btn-secondary btn-sm" onClick={() => { setShowCreateForm(false); setDrawMode('none') }}>
                Cancel
              </button>
            </div>
          </div>
        )}

        {/* Geofence list */}
        <div style={{ flex: 1, overflowY: 'auto' }}>
          {geofences.length === 0 && (
            <div style={{ textAlign: 'center', padding: 32, color: 'var(--color-text-muted)' }}>
              No geofences yet
            </div>
          )}
          {geofences.map((gf) => (
            <div
              key={gf.id}
              id={`geofence-item-${gf.id}`}
              className={`flex items-center gap-2`}
              style={{
                padding: '10px 16px',
                borderBottom: '1px solid var(--color-border)',
                cursor: 'pointer',
                background: selected?.id === gf.id ? 'var(--color-surface-raised)' : 'transparent',
                transition: 'background 0.15s',
              }}
              onClick={() => setSelected(gf)}
            >
              <span style={{ color: 'var(--color-accent)' }}>{shapeIcon(gf.shape_type)}</span>
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{ fontWeight: 500, fontSize: 'var(--text-sm)', whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                  {gf.name}
                </div>
                <div className="text-xs text-muted">{gf.shape_type}</div>
              </div>
              <div className="flex gap-1">
                <button
                  className="btn btn-ghost btn-sm"
                  onClick={(e) => { e.stopPropagation(); toggleVisibility(gf.id) }}
                  title={visible[gf.id] === false ? 'Show' : 'Hide'}
                >
                  {visible[gf.id] === false ? <EyeOff size={13} /> : <Eye size={13} />}
                </button>
                <button
                  className="btn btn-ghost btn-sm"
                  onClick={(e) => { e.stopPropagation(); deleteMutation.mutate(gf.id) }}
                  title="Delete"
                >
                  <Trash2 size={13} />
                </button>
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Map + Detail */}
      <div style={{ flex: 1, display: 'flex', flexDirection: 'column', position: 'relative' }}>
        <div ref={mapRef} style={{ flex: 1 }} />

        {/* Geofence detail panel overlay */}
        {selected && (
          <div
            className="card animate-slide-up"
            style={{
              position: 'absolute', bottom: 0, left: 0, right: 0, height: 280,
              borderRadius: '12px 12px 0 0', padding: 20, overflowY: 'auto',
            }}
          >
            <div className="flex items-center justify-between" style={{ marginBottom: 16 }}>
              <div>
                <h3 style={{ fontSize: 'var(--text-lg)', fontWeight: 700 }}>{selected.name}</h3>
                <span className="badge badge-info">{selected.shape_type}</span>
              </div>
              <button className="btn btn-ghost btn-sm" onClick={() => setSelected(null)}>✕</button>
            </div>
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20 }}>
              <div>
                <p className="text-sm text-muted" style={{ marginBottom: 8 }}>Dwell Time (last 30 days)</p>
                <ResponsiveContainer width="100%" height={160}>
                  <BarChart data={dwellData} margin={{ top: 0, right: 0, bottom: 0, left: -20 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                    <XAxis dataKey="vehicle" tick={{ fontSize: 10, fill: 'var(--color-text-muted)' }} />
                    <YAxis tick={{ fontSize: 10, fill: 'var(--color-text-muted)' }} />
                    <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)', fontSize: 11 }} />
                    <Bar dataKey="avg" name="Avg (min)" fill="var(--color-accent)" radius={[3, 3, 0, 0]} />
                    <Bar dataKey="max" name="Max (min)" fill="var(--color-warning)" radius={[3, 3, 0, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              </div>
              <div>
                <p className="text-sm text-muted" style={{ marginBottom: 8 }}>Recent Events</p>
                {['MH01AB1234 entered', 'MH02CD5678 exited', 'DL3EF9012 entered'].map((ev, i) => (
                  <div key={i} className="flex items-center gap-2" style={{ marginBottom: 6 }}>
                    <div className={`dot ${i % 2 === 0 ? 'dot-online' : 'dot-warning'}`} />
                    <span className="text-sm">{ev}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
