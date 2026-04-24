'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineUsers, HiOutlinePlusCircle } from 'react-icons/hi2'

const USERS = [
  { name: 'Admin User',    email: 'admin@acme.com',  role: 'ADMIN',    status: 'ACTIVE', last: '2 min ago' },
  { name: 'Ops Manager',   email: 'ops@acme.com',    role: 'MANAGER',  status: 'ACTIVE', last: '1h ago'    },
  { name: 'Field Agent 1', email: 'field1@acme.com', role: 'VIEWER',   status: 'ACTIVE', last: '3h ago'    },
  { name: 'Field Agent 2', email: 'field2@acme.com', role: 'VIEWER',   status: 'INACTIVE', last: '5d ago'  },
]

const roleColor: Record<string, string> = {
  ADMIN:   'bg-red-500/20 text-red-400',
  MANAGER: 'bg-brand-500/20 text-brand-400',
  VIEWER:  'bg-slate-500/20 text-slate-400',
}

export default function UsersPage() {
  return (
    <div>
      <TopBar title="Users & Access" subtitle="Team members & permissions" />
      <div className="p-6 space-y-4">
        <div className="flex justify-between">
          <span className="text-sm font-semibold text-white flex items-center gap-2">
            <HiOutlineUsers className="w-4 h-4 text-brand-400" /> {USERS.length} Members
          </span>
          <button className="btn-primary flex items-center gap-2 text-sm">
            <HiOutlinePlusCircle className="w-4 h-4" /> Invite User
          </button>
        </div>
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Name', 'Email', 'Role', 'Status', 'Last Active'].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {USERS.map((u, i) => (
                <tr key={i} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-medium text-white">{u.name}</td>
                  <td className="px-4 py-3 text-slate-400 text-xs">{u.email}</td>
                  <td className="px-4 py-3"><span className={`badge ${roleColor[u.role]}`}>{u.role}</span></td>
                  <td className="px-4 py-3">
                    <span className={`badge ${u.status === 'ACTIVE' ? 'bg-green-500/20 text-green-400' : 'bg-slate-500/20 text-slate-400'}`}>{u.status}</span>
                  </td>
                  <td className="px-4 py-3 text-slate-500 text-xs">{u.last}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
