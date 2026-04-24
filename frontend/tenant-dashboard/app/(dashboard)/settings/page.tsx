'use client'
import TopBar from '@/components/TopBar'
import { HiOutlineCog6Tooth } from 'react-icons/hi2'

export default function SettingsPage() {
  return (
    <div>
      <TopBar title="Settings" subtitle="Platform configuration" />
      <div className="p-6 space-y-5">
        <div className="grid grid-cols-2 gap-4">
          {[
            { section: 'Company', fields: [{ label: 'Company Name', val: 'Acme Transport' }, { label: 'Timezone', val: 'Asia/Kolkata (IST)' }, { label: 'Units', val: 'Metric (km)' }] },
            { section: 'Notifications', fields: [{ label: 'Email Alerts', val: 'Enabled' }, { label: 'SMS Alerts', val: 'Disabled' }, { label: 'Webhook URL', val: 'https://hooks.acme.com/fleet' }] },
          ].map(s => (
            <div key={s.section} className="card space-y-4">
              <h2 className="font-semibold text-white text-sm flex items-center gap-2">
                <HiOutlineCog6Tooth className="w-4 h-4 text-brand-400" /> {s.section}
              </h2>
              {s.fields.map(f => (
                <div key={f.label}>
                  <label className="text-xs text-slate-500 block mb-1">{f.label}</label>
                  <input className="input" defaultValue={f.val} />
                </div>
              ))}
              <button className="btn-primary text-sm w-full mt-2">Save {s.section}</button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
