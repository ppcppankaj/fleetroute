'use client'

import TopBar from '@/components/TopBar'
import { HiOutlineSparkles, HiOutlineHandThumbUp, HiOutlineCheck } from 'react-icons/hi2'

const FEATURES = [
  { id: 'f1', title: 'Trip Replay with Heatmap',     category: 'Tracking',     status: 'PLANNED',     votes: 84  },
  { id: 'f2', title: 'AI Driver Coaching Dashboard', category: 'Driver',       status: 'IN_PROGRESS', votes: 67  },
  { id: 'f3', title: 'Multi-fleet Management',       category: 'Platform',     status: 'PLANNED',     votes: 52  },
  { id: 'f4', title: 'WhatsApp Alert Notifications', category: 'Alerts',       status: 'DONE',        votes: 148 },
  { id: 'f5', title: 'Mobile App for Drivers',       category: 'Mobile',       status: 'PLANNED',     votes: 201 },
  { id: 'f6', title: 'Fuel Theft Detection',         category: 'Fuel',         status: 'IN_PROGRESS', votes: 93  },
]

const statusStyle: Record<string, string> = {
  PLANNED:     'bg-brand-500/20 text-brand-400',
  IN_PROGRESS: 'bg-amber-500/20 text-amber-400',
  DONE:        'bg-green-500/20 text-green-400',
  CANCELLED:   'bg-red-500/20 text-red-400',
}

const statusIcon: Record<string, React.ReactNode> = {
  IN_PROGRESS: <span className="text-amber-400">⚙</span>,
  DONE:        <HiOutlineCheck className="text-green-400 w-3.5 h-3.5" />,
  PLANNED:     <span className="text-brand-400">○</span>,
}

export default function RoadmapPage() {
  return (
    <div>
      <TopBar title="Product Roadmap" subtitle="Vote on features & track what's coming" />
      <div className="p-6 space-y-5">
        <div className="flex items-start gap-4 card bg-gradient-to-r from-brand-500/10 to-purple-500/10 border-brand-500/20">
          <HiOutlineSparkles className="w-8 h-8 text-brand-400 flex-shrink-0 mt-1" />
          <div>
            <h2 className="font-semibold text-white">Help shape the future of FleetRoute</h2>
            <p className="text-sm text-slate-400 mt-1">Vote on features you'd like to see. Our team reviews votes every sprint and plans accordingly.</p>
          </div>
        </div>

        {/* Category pills */}
        <div className="flex flex-wrap gap-2">
          {['All', 'Tracking', 'Driver', 'Alerts', 'Fuel', 'Mobile', 'Platform'].map(c => (
            <button key={c} className={`text-xs px-3 py-1.5 rounded-full font-medium transition-all ${c === 'All' ? 'bg-brand-500 text-white' : 'bg-surface-hover text-slate-400 hover:text-white'}`}>
              {c}
            </button>
          ))}
        </div>

        {/* Feature cards */}
        <div className="grid grid-cols-1 gap-3">
          {FEATURES.sort((a, b) => b.votes - a.votes).map(f => (
            <div key={f.id} className="card flex items-center gap-4 hover:border-brand-500/30 transition-all group">
              {/* Vote button */}
              <button className="flex flex-col items-center gap-1 px-3 py-2 rounded-lg bg-surface hover:bg-surface-hover transition-colors min-w-[52px]">
                <HiOutlineHandThumbUp className="w-4 h-4 text-slate-400 group-hover:text-brand-400 transition-colors" />
                <span className="text-sm font-bold text-white">{f.votes}</span>
              </button>

              {/* Content */}
              <div className="flex-1">
                <div className="flex items-center gap-2 flex-wrap">
                  <span className="text-sm font-semibold text-white">{f.title}</span>
                  <span className={`badge ${statusStyle[f.status]}`}>
                    <span className="mr-0.5">{statusIcon[f.status]}</span>
                    {f.status.replace('_', ' ')}
                  </span>
                </div>
                <span className="text-xs text-slate-500 mt-0.5 inline-block">{f.category}</span>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
