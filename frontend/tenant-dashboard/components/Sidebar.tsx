'use client'

import Link from 'next/link'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import { clsx } from 'clsx'
import {
  HiOutlineMap,
  HiOutlineTruck,
  HiOutlineUserGroup,
  HiOutlineShieldCheck,
  HiOutlineBell,
  HiOutlineDocumentChartBar,
  HiOutlineWrenchScrewdriver,
  HiOutlineFire,
  HiOutlineGlobeAlt,
  HiOutlineCalendarDays,
  HiOutlineCpuChip,
  HiOutlineUsers,
  HiOutlineReceiptPercent,
  HiOutlineChartBarSquare,
  HiOutlineHome,
  HiOutlineSparkles,
  HiOutlineCog6Tooth,
} from 'react-icons/hi2'

const navItems = [
  { href: '/dashboard',      label: 'Overview',        icon: HiOutlineHome },
  { href: '/live-tracking',  label: 'Live Tracking',   icon: HiOutlineMap,        badge: 'LIVE' },
  { href: '/vehicles',       label: 'Vehicles',        icon: HiOutlineTruck },
  { href: '/drivers',        label: 'Drivers',         icon: HiOutlineUserGroup },
  { href: '/routes',         label: 'Routes & Trips',  icon: HiOutlineCalendarDays },
  { href: '/geofencing',     label: 'Geofencing',      icon: HiOutlineGlobeAlt },
  { href: '/alerts',         label: 'Alerts',          icon: HiOutlineBell },
  { href: '/maintenance',    label: 'Maintenance',     icon: HiOutlineWrenchScrewdriver },
  { href: '/fuel',           label: 'Fuel',            icon: HiOutlineFire },
  { href: '/reports',        label: 'Reports',         icon: HiOutlineDocumentChartBar },
  { href: '/devices',        label: 'Devices',         icon: HiOutlineCpuChip },
  { href: '/users',          label: 'Users',           icon: HiOutlineUsers },
  { href: '/security',       label: 'Security',        icon: HiOutlineShieldCheck },
  { href: '/billing',        label: 'Billing',         icon: HiOutlineReceiptPercent },
  { href: '/activity',       label: 'Activity',        icon: HiOutlineChartBarSquare },
  { href: '/roadmap',        label: 'Roadmap',         icon: HiOutlineSparkles },
  { href: '/settings',       label: 'Settings',        icon: HiOutlineCog6Tooth },
]

export default function Sidebar() {
  const pathname = usePathname()

  return (
    <aside className="fixed left-0 top-0 h-screen w-60 bg-surface-card border-r border-surface-border flex flex-col z-50">
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 py-4 border-b border-surface-border">
        <div className="w-9 h-9 rounded-lg overflow-hidden flex items-center justify-center bg-surface flex-shrink-0">
          <Image
            src="/logo.png"
            alt="TrackOra Logo"
            width={36}
            height={36}
            className="object-contain"
            priority
          />
        </div>
        <div>
          <span className="font-bold text-white text-sm tracking-wide">TrackOra</span>
          <p className="text-xs text-slate-500 leading-tight">Enterprise GPS</p>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-0.5">
        {navItems.map(({ href, label, icon: Icon, badge }) => {
          const active = pathname === href || (pathname ?? '').startsWith(href + '/')
          return (
            <Link
              key={href}
              href={href}
              className={clsx(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-150 group',
                active
                  ? 'bg-brand-500/20 text-brand-400 font-medium'
                  : 'text-slate-400 hover:text-slate-100 hover:bg-surface-hover',
              )}
            >
              <Icon className={clsx('w-4 h-4 flex-shrink-0', active ? 'text-brand-400' : 'text-slate-500 group-hover:text-slate-300')} />
              <span className="flex-1 truncate">{label}</span>
              {badge && (
                <span className="badge bg-green-500/20 text-green-400 animate-pulse">
                  {badge}
                </span>
              )}
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-4 border-t border-surface-border">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 rounded-full bg-gradient-brand flex items-center justify-center text-xs font-bold text-white flex-shrink-0">
            TT
          </div>
          <div className="flex-1 min-w-0">
            <p className="text-xs font-medium text-slate-200 truncate">Trackora Technologies</p>
            <p className="text-xs text-slate-500 truncate">Pro Plan</p>
          </div>
        </div>
      </div>
    </aside>
  )
}
