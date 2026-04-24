import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import api from '../../shared/api/client'
import { Video, Camera, Clock, AlertTriangle, Download, Settings, Play } from 'lucide-react'

/* ─── Types ──────────────────────────────────────────────────────────────── */
interface VideoEvent {
  id: string
  device_id: string
  vehicle_registration: string
  event_type: 'harsh_brake' | 'harsh_accel' | 'collision' | 'manual'
  timestamp: string
  thumbnail_url?: string
  video_url?: string
  duration_s: number
}

/* ─── Main Page ──────────────────────────────────────────────────────────── */
export default function VideoTelematicsPage() {
  const [tab, setTab] = useState<'events' | 'livestream' | 'settings'>('events')
  
  // Simulated events data (stub backend)
  const events: VideoEvent[] = [
    {
      id: 'evt_1', device_id: 'd1', vehicle_registration: 'MH 12 AB 1234',
      event_type: 'harsh_brake', timestamp: new Date().toISOString(), duration_s: 15,
      thumbnail_url: 'https://images.unsplash.com/photo-1469854523086-cc02fe5d8800?w=600&q=80',
    },
    {
      id: 'evt_2', device_id: 'd2', vehicle_registration: 'KA 01 XY 9999',
      event_type: 'collision', timestamp: new Date(Date.now() - 3600000).toISOString(), duration_s: 30,
      thumbnail_url: 'https://images.unsplash.com/photo-1549317661-bd32c8ce0db2?w=600&q=80',
    }
  ]

  return (
    <div className="page" style={{ padding: 'var(--space-6)' }}>
      {/* Header */}
      <div className="flex items-center justify-between" style={{ marginBottom: 'var(--space-6)' }}>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>Video Telematics</h1>
          <p className="text-muted text-sm" style={{ marginTop: 4 }}>M15 · Dashcam event recordings and live view</p>
        </div>
        <button id="video-request" className="btn btn-primary">
          <Camera size={14} /> Request Snapshot
        </button>
      </div>

      {/* Tabs */}
      <div className="flex gap-1" style={{ marginBottom: 'var(--space-4)', borderBottom: '1px solid var(--color-border)' }}>
        {[
          { key: 'events', icon: Video, label: 'Event Recordings' },
          { key: 'livestream', icon: Camera, label: 'Live View (Beta)' },
          { key: 'settings', icon: Settings, label: 'Camera Settings' },
        ].map(({ key, icon: Icon, label }) => (
          <button key={key} id={`video-tab-${key}`}
            className={`btn ${tab === key ? 'btn-primary' : 'btn-ghost'}`}
            style={{ borderRadius: '6px 6px 0 0' }}
            onClick={() => setTab(key as any)}>
            <Icon size={13} /> {label}
          </button>
        ))}
      </div>

      {/* ── Tab: Events ── */}
      {tab === 'events' && (
        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 'var(--space-4)' }} className="animate-fade-in">
          {events.map((ev) => (
            <div key={ev.id} className="card overflow-hidden">
              <div style={{ position: 'relative', height: 180, background: '#111' }}>
                {ev.thumbnail_url ? (
                  <img src={ev.thumbnail_url} alt="Video thumbnail" style={{ width: '100%', height: '100%', objectFit: 'cover', opacity: 0.8 }} />
                ) : (
                  <div className="flex items-center justify-center h-full"><Video size={48} color="#333" /></div>
                )}
                <div style={{ position: 'absolute', top: 8, right: 8, background: 'rgba(0,0,0,0.6)', padding: '2px 8px', borderRadius: 4, color: '#fff', fontSize: 12, fontWeight: 600 }}>
                  {ev.duration_s}s
                </div>
                <div style={{ position: 'absolute', top: 8, left: 8 }}>
                  <span className={`badge ${ev.event_type === 'collision' ? 'badge-danger' : 'badge-warning'}`}>
                    {ev.event_type.replace('_', ' ').toUpperCase()}
                  </span>
                </div>
                <button className="btn btn-primary" style={{ position: 'absolute', top: '50%', left: '50%', transform: 'translate(-50%, -50%)', borderRadius: '50%', width: 44, height: 44, padding: 0 }}>
                  <Play size={20} style={{ marginLeft: 2 }} />
                </button>
              </div>
              <div style={{ padding: 'var(--space-4)' }}>
                <div className="flex justify-between items-center mb-2">
                  <span className="font-bold">{ev.vehicle_registration}</span>
                  <span className="text-xs text-muted flex items-center gap-1"><Clock size={12} /> {new Date(ev.timestamp).toLocaleTimeString([], {hour: '2-digit', minute:'2-digit'})}</span>
                </div>
                <div className="flex justify-between items-center mt-4">
                  <span className="text-sm text-muted">{new Date(ev.timestamp).toLocaleDateString()}</span>
                  <button className="btn btn-ghost btn-sm text-muted" title="Download Video">
                    <Download size={14} />
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* ── Tab: Livestream ── */}
      {tab === 'livestream' && (
        <div className="card animate-fade-in" style={{ padding: 'var(--space-12)', textAlign: 'center' }}>
          <div style={{ display: 'inline-block', padding: 24, borderRadius: '50%', background: 'var(--color-surface-raised)', marginBottom: 16 }}>
            <Camera size={48} color="var(--color-primary)" />
          </div>
          <h2 style={{ fontSize: 'var(--text-lg)', fontWeight: 600, marginBottom: 8 }}>Live View Not Connected</h2>
          <p className="text-muted max-w-md mx-auto">
            Select an online dashcam-enabled vehicle from the fleet to start a WebRTC live video stream.
          </p>
        </div>
      )}

      {/* ── Tab: Settings ── */}
      {tab === 'settings' && (
        <div className="card animate-fade-in" style={{ padding: 'var(--space-6)' }}>
          <h3 className="font-semibold mb-4">Dashcam Configuration (Global)</h3>
          <div style={{ display: 'grid', gap: 16, maxWidth: 600 }}>
            <div className="flex items-center justify-between p-4 border rounded">
              <div>
                <div className="font-medium">Video Quality</div>
                <div className="text-sm text-muted">Resolution for event clips</div>
              </div>
              <select className="input" defaultValue="720p">
                <option value="480p">480p</option>
                <option value="720p">720p (HD)</option>
                <option value="1080p">1080p (FHD)</option>
              </select>
            </div>
            <div className="flex items-center justify-between p-4 border rounded">
              <div>
                <div className="font-medium">Pre/Post Event Recording</div>
                <div className="text-sm text-muted">Seconds to record before and after an event</div>
              </div>
              <select className="input" defaultValue="15">
                <option value="5">5s / 5s</option>
                <option value="10">10s / 10s</option>
                <option value="15">15s / 15s</option>
              </select>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
