'use client'

import { HiOutlineUserGroup, HiOutlinePlusCircle, HiOutlineMagnifyingGlass, HiOutlinePauseCircle, HiOutlineCheckCircle } from 'react-icons/hi2'
import { useState } from 'react'

const TENANTS = [
  { id: 't1', name: 'Acme Transport',    slug: 'acme-transport',  plan: 'Pro',        vehicles: 97, status: 'ACTIVE',    mrr: 14999, created: '2025-08-12' },
  { id: 't2', name: 'Swift Logistics',   slug: 'swift-logistics', plan: 'Enterprise', vehicles: 240,status: 'ACTIVE',    mrr: 39999, created: '2025-06-04' },
  { id: 't3', name: 'City Cab Services', slug: 'city-cab',        plan: 'Starter',    vehicles: 32, status: 'TRIAL',     mrr: 0,     created: '2026-04-10' },
  { id: 't4', name: 'QuickFleet Ltd',    slug: 'quickfleet',      plan: 'Pro',        vehicles: 65, status: 'ACTIVE',    mrr: 14999, created: '2025-11-19' },
  { id: 't5', name: 'Old Corp',          slug: 'old-corp',        plan: 'Pro',        vehicles: 8,  status: 'SUSPENDED', mrr: 0,     created: '2024-01-01' },
]

const statusBadge: Record<string, string> = {
  ACTIVE:    'bg-green-500/20 text-green-400',
  TRIAL:     'bg-brand-500/20 text-amber-400',
  SUSPENDED: 'bg-red-500/20 text-red-400',
}

export default function TenantsPage() {
  const [search, setSearch] = useState('')
  const filtered = TENANTS.filter(t =>
    t.name.toLowerCase().includes(search.toLowerCase()) ||
    t.slug.toLowerCase().includes(search.toLowerCase()),
  )

  return (
    <div>
      <div className="h-16 border-b border-zinc-800 flex items-center px-6 bg-zinc-900/50">
        <h1 className="text-base font-semibold text-white flex items-center gap-2">
          <HiOutlineUserGroup className="w-5 h-5 text-amber-400" /> Tenant Management
        </h1>
      </div>
      <div className="p-6 space-y-4">

        {/* Summary */}
        <div className="grid grid-cols-4 gap-4">
          {[
            { label: 'Total',     count: TENANTS.length,                                    color: 'text-white'      },
            { label: 'Active',    count: TENANTS.filter(t => t.status === 'ACTIVE').length, color: 'text-green-400'  },
            { label: 'Trial',     count: TENANTS.filter(t => t.status === 'TRIAL').length,  color: 'text-amber-400'  },
            { label: 'Suspended', count: TENANTS.filter(t => t.status === 'SUSPENDED').length, color: 'text-red-400' },
          ].map(s => (
            <div key={s.label} className="card text-center">
              <p className={`text-2xl font-bold ${s.color}`}>{s.count}</p>
              <p className="text-xs text-zinc-500 mt-1">{s.label}</p>
            </div>
          ))}
        </div>

        {/* Toolbar */}
        <div className="flex items-center gap-3">
          <div className="relative flex-1 max-w-sm">
            <HiOutlineMagnifyingGlass className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
            <input className="input pl-9" placeholder="Search tenants…" value={search} onChange={e => setSearch(e.target.value)} />
          </div>
          <button className="btn-primary flex items-center gap-2 text-sm ml-auto">
            <HiOutlinePlusCircle className="w-4 h-4" /> Create Tenant
          </button>
        </div>

        {/* Table */}
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-zinc-800 bg-zinc-800/30">
                {['Tenant', 'Slug', 'Plan', 'Vehicles', 'MRR', 'Status', 'Created', 'Actions'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-zinc-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-zinc-800">
              {filtered.map(t => (
                <tr key={t.id} className="hover:bg-zinc-800/40 transition-colors group">
                  <td className="px-4 py-3 font-semibold text-white">{t.name}</td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-500">{t.slug}</td>
                  <td className="px-4 py-3"><span className="badge bg-zinc-700/50 text-zinc-300">{t.plan}</span></td>
                  <td className="px-4 py-3 text-zinc-300">{t.vehicles}</td>
                  <td className="px-4 py-3 text-amber-400 font-mono text-xs">{t.mrr > 0 ? `₹${t.mrr.toLocaleString()}` : '—'}</td>
                  <td className="px-4 py-3"><span className={`badge ${statusBadge[t.status]}`}>{t.status}</span></td>
                  <td className="px-4 py-3 text-zinc-500 text-xs">{t.created}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                      {t.status === 'ACTIVE' ? (
                        <button className="btn-ghost p-1.5 text-amber-400 rounded" title="Suspend">
                          <HiOutlinePauseCircle className="w-3.5 h-3.5" />
                        </button>
                      ) : (
                        <button className="btn-ghost p-1.5 text-green-400 rounded" title="Activate">
                          <HiOutlineCheckCircle className="w-3.5 h-3.5" />
                        </button>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
