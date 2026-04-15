import { useParams, useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { ArrowLeft } from 'lucide-react'
import {
  RadarChart, Radar, PolarGrid, PolarAngleAxis, PolarRadiusAxis,
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer
} from 'recharts'
import api from '../../shared/api/client'

export function DriverDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()

  const { data: driver } = useQuery({
    queryKey: ['driver', id],
    queryFn: async () => { const r = await api.get(`/drivers/${id}`); return r.data.data },
    enabled: !!id,
  })

  const { data: score } = useQuery<{
    score: number; trips: number; duration_s: number
    harsh_accel: number; harsh_brake: number; overspeed: number
  }>({
    queryKey: ['driver-score', id],
    queryFn: async () => { const r = await api.get(`/drivers/${id}/score`); return r.data.data },
    enabled: !!id,
  })

  // 12-week score trend (demo data)
  const scoreTrend = Array.from({ length: 12 }, (_, i) => ({
    week: `W${i + 1}`,
    score: Math.max(60, Math.min(100, (score?.score ?? 85) + (Math.random() - 0.5) * 20)),
  }))

  // Radar chart breakdown
  const radarData = score ? [
    { metric: 'Smooth Braking',    value: Math.max(0, 100 - (score.harsh_brake ?? 0) * 10) },
    { metric: 'Smooth Accel',      value: Math.max(0, 100 - (score.harsh_accel ?? 0) * 10) },
    { metric: 'Speed Compliance',  value: Math.max(0, 100 - (score.overspeed ?? 0) * 5) },
    { metric: 'Cornering',         value: 85 },
    { metric: 'Seat Belt',         value: 90 },
    { metric: 'Phone Usage',       value: 95 },
  ] : []

  const scoreColor = (s: number) =>
    s >= 85 ? 'var(--color-success)' : s >= 70 ? 'var(--color-warning)' : 'var(--color-danger)'

  const driverName = driver ? driver[1] : 'Driver'

  return (
    <div className="page" style={{ padding: 'var(--space-6)', height: '100%', overflowY: 'auto' }}>
      {/* Header */}
      <div className="flex items-center gap-3" style={{ marginBottom: 'var(--space-6)' }}>
        <button className="btn btn-ghost btn-sm" onClick={() => navigate('/fleet')}>
          <ArrowLeft size={15} /> Fleet
        </button>
        <div className="user-avatar" style={{ width: 40, height: 40, fontSize: 18 }}>
          {driverName[0]?.toUpperCase() ?? 'D'}
        </div>
        <div>
          <h1 style={{ fontSize: 'var(--text-2xl)', fontWeight: 700 }}>{driverName}</h1>
          <p className="text-muted text-sm">Driver Profile</p>
        </div>
      </div>

      {/* Score badge */}
      {score && (
        <div className="flex gap-4" style={{ marginBottom: 'var(--space-6)', flexWrap: 'wrap' }}>
          <div className="stat-card" style={{ flex: '0 0 auto', minWidth: 140, textAlign: 'center' }}>
            <div className="stat-label">Safety Score</div>
            <div className="stat-value" style={{ color: scoreColor(score.score), fontSize: 48 }}>
              {score.score}
            </div>
            <div className="stat-change">out of 100</div>
          </div>
          <div className="stat-card flex-1"><div className="stat-label">Total Trips</div><div className="stat-value">{score.trips}</div></div>
          <div className="stat-card flex-1">
            <div className="stat-label">Drive Time</div>
            <div className="stat-value">{Math.round((score.duration_s ?? 0) / 3600)}h</div>
          </div>
          <div className="stat-card flex-1">
            <div className="stat-label">Harsh Events</div>
            <div className="stat-value" style={{ color: (score.harsh_accel + score.harsh_brake) > 0 ? 'var(--color-warning)' : undefined }}>
              {(score.harsh_accel ?? 0) + (score.harsh_brake ?? 0)}
            </div>
          </div>
          <div className="stat-card flex-1">
            <div className="stat-label">Overspeed</div>
            <div className="stat-value" style={{ color: (score.overspeed ?? 0) > 0 ? 'var(--color-danger)' : undefined }}>
              {score.overspeed ?? 0}
            </div>
          </div>
        </div>
      )}

      {/* Charts */}
      <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 20, marginBottom: 20 }}>
        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Behavior Breakdown</h2></div>
          <div style={{ padding: '0 16px 16px' }}>
            {radarData.length > 0 ? (
              <ResponsiveContainer width="100%" height={260}>
                <RadarChart data={radarData}>
                  <PolarGrid stroke="var(--color-border)" />
                  <PolarAngleAxis dataKey="metric" tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                  <PolarRadiusAxis domain={[0, 100]} tick={{ fontSize: 10, fill: 'var(--color-text-muted)' }} />
                  <Radar
                    name="Score"
                    dataKey="value"
                    stroke="var(--color-accent)"
                    fill="var(--color-accent)"
                    fillOpacity={0.25}
                  />
                  <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)' }} />
                </RadarChart>
              </ResponsiveContainer>
            ) : (
              <div style={{ textAlign: 'center', color: 'var(--color-text-muted)', padding: 40 }}>No data</div>
            )}
          </div>
        </div>

        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Score Trend (12 weeks)</h2></div>
          <div style={{ padding: '0 16px 16px' }}>
            <ResponsiveContainer width="100%" height={260}>
              <LineChart data={scoreTrend}>
                <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" />
                <XAxis dataKey="week" tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                <YAxis domain={[0, 100]} tick={{ fontSize: 11, fill: 'var(--color-text-muted)' }} />
                <Tooltip contentStyle={{ background: 'var(--color-surface)', border: '1px solid var(--color-border)' }} />
                <Line
                  type="monotone"
                  dataKey="score"
                  stroke="var(--color-accent)"
                  strokeWidth={2.5}
                  dot={{ r: 3, fill: 'var(--color-accent)' }}
                />
              </LineChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Profile info */}
      {driver && (
        <div className="card animate-fade-in">
          <div className="card-header"><h2 className="card-title">Profile</h2></div>
          <div style={{ padding: '0 16px 16px' }}>
            {[
              ['Name', driver[1]],
              ['License Number', driver[2] ?? '—'],
              ['RFID UID', driver[3] ?? '—'],
              ['Phone', driver[4] ?? '—'],
              ['Added', driver[5] ? new Date(driver[5]).toLocaleDateString() : '—'],
            ].map(([label, value]) => (
              <div key={label as string} className="flex justify-between" style={{ padding: '8px 0', borderBottom: '1px solid var(--color-border)' }}>
                <span className="text-sm text-muted">{label as string}</span>
                <span className="text-sm" style={{ fontWeight: 500 }}>{value as string}</span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
