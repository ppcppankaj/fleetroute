interface Props {
  title: string
  value: string | number
  subtitle?: string
  trend?: { value: number; label: string }
  icon: React.ReactNode
  accent?: string
}

export default function StatCard({ title, value, subtitle, trend, icon, accent = 'brand' }: Props) {
  const accentMap: Record<string, string> = {
    brand:  'from-brand-500/20 border-brand-500/30 text-brand-400',
    green:  'from-green-500/20 border-green-500/30 text-green-400',
    amber:  'from-amber-500/20 border-amber-500/30 text-amber-400',
    red:    'from-red-500/20 border-red-500/30 text-red-400',
    purple: 'from-purple-500/20 border-purple-500/30 text-purple-400',
    cyan:   'from-cyan-500/20 border-cyan-500/30 text-cyan-400',
  }
  const cls = accentMap[accent] || accentMap.brand

  return (
    <div className={`card bg-gradient-to-br ${cls} relative overflow-hidden`}>
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-slate-400 uppercase tracking-wider">{title}</p>
          <p className="mt-2 text-3xl font-bold text-white">{value}</p>
          {subtitle && <p className="mt-1 text-xs text-slate-500">{subtitle}</p>}
          {trend && (
            <p className={`mt-2 text-xs font-medium ${trend.value >= 0 ? 'text-green-400' : 'text-red-400'}`}>
              {trend.value >= 0 ? '↑' : '↓'} {Math.abs(trend.value)}% {trend.label}
            </p>
          )}
        </div>
        <div className={`p-3 rounded-xl bg-gradient-to-br ${cls} opacity-80`}>
          {icon}
        </div>
      </div>
    </div>
  )
}
