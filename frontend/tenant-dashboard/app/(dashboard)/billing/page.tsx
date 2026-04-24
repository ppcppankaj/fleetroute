'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineReceiptPercent, HiOutlineCheckBadge } from 'react-icons/hi2'

const INVOICES = [
  { id: 'inv-001', period: 'April 2026',  amount: 14999, status: 'PAID',   date: '2026-04-01', pdf: true  },
  { id: 'inv-002', period: 'March 2026',  amount: 14999, status: 'PAID',   date: '2026-03-01', pdf: true  },
  { id: 'inv-003', period: 'May 2026',    amount: 14999, status: 'OPEN',   date: '2026-05-01', pdf: false },
]

export default function BillingPage() {
  return (
    <div>
      <TopBar title="Billing" subtitle="Subscription & invoice management" />
      <div className="p-6 space-y-5">
        {/* Plan card */}
        <div className="card bg-gradient-to-r from-brand-500/10 to-purple-500/10 border-brand-500/20 flex items-center gap-6">
          <div className="w-14 h-14 rounded-2xl bg-gradient-brand flex items-center justify-center flex-shrink-0 shadow-lg shadow-brand-500/30">
            <HiOutlineCheckBadge className="w-7 h-7 text-white" />
          </div>
          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="text-lg font-bold text-white">Pro Plan</span>
              <span className="badge bg-green-500/20 text-green-400">ACTIVE</span>
            </div>
            <p className="text-sm text-slate-400 mt-0.5">Up to 100 vehicles · 20 users · Priority support</p>
            <p className="text-xs text-slate-500 mt-1">Renews on <strong className="text-slate-300">May 1, 2026</strong> · ₹14,999/month</p>
          </div>
          <button className="btn-ghost text-sm">Upgrade Plan</button>
        </div>

        {/* Invoices */}
        <h2 className="text-sm font-semibold text-white flex items-center gap-2">
          <HiOutlineReceiptPercent className="w-4 h-4 text-brand-400" /> Invoice History
        </h2>
        <div className="card overflow-hidden p-0">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-surface-border bg-surface-hover/20">
                {['Invoice', 'Period', 'Amount', 'Status', 'Date', ''].map(h => (
                  <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-400 uppercase tracking-wider">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-surface-border">
              {INVOICES.map(inv => (
                <tr key={inv.id} className="hover:bg-surface-hover/30 transition-colors">
                  <td className="px-4 py-3 font-mono text-xs text-slate-300">{inv.id}</td>
                  <td className="px-4 py-3 text-white">{inv.period}</td>
                  <td className="px-4 py-3 text-white font-bold">₹{inv.amount.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <span className={`badge ${inv.status === 'PAID' ? 'bg-green-500/20 text-green-400' : 'bg-amber-500/20 text-amber-400'}`}>{inv.status}</span>
                  </td>
                  <td className="px-4 py-3 text-slate-400 font-mono text-xs">{inv.date}</td>
                  <td className="px-4 py-3">
                    {inv.pdf && <button className="btn-ghost text-xs text-brand-400">Download PDF</button>}
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
