'use client'

import Link from 'next/link'
import Image from 'next/image'
import { usePathname } from 'next/navigation'
import { clsx } from 'clsx'
import {
  HiOutlineHome,
  HiOutlineUserGroup,
  HiOutlineTicket,
  HiOutlineCurrencyDollar,
  HiOutlineServerStack,
  HiOutlineChartBar,
  HiOutlineShieldCheck,
  HiOutlineMegaphone,
  HiOutlineSparkles,
  HiOutlineUserCircle,
  HiOutlineCog6Tooth,
} from 'react-icons/hi2'

const nav = [
  { href: '/dashboard',       label: 'Platform Overview', icon: HiOutlineHome           },
  { href: '/tenants',         label: 'Tenants',           icon: HiOutlineUserGroup      },
  { href: '/tickets',         label: 'Support Tickets',   icon: HiOutlineTicket         },
  { href: '/billing',         label: 'Platform Billing',  icon: HiOutlineCurrencyDollar },
  { href: '/services',        label: 'Service Health',    icon: HiOutlineServerStack    },
  { href: '/activity',        label: 'Platform Activity', icon: HiOutlineChartBar       },
  { href: '/security',        label: 'Security',          icon: HiOutlineShieldCheck    },
  { href: '/announcements',   label: 'Announcements',     icon: HiOutlineMegaphone      },
  { href: '/roadmap',         label: 'Roadmap Admin',     icon: HiOutlineSparkles       },
  { href: '/admins',          label: 'Admin Users',       icon: HiOutlineUserCircle     },
  { href: '/settings',        label: 'Settings',          icon: HiOutlineCog6Tooth      },
]

export default function AdminSidebar() {
  const pathname = usePathname()

  return (
    <aside className="fixed left-0 top-0 h-screen w-60 bg-zinc-900 border-r border-zinc-800 flex flex-col z-50">
      {/* Logo */}
      <div className="flex items-center gap-3 px-4 py-4 border-b border-zinc-800">
        <div className="w-9 h-9 rounded-lg overflow-hidden flex items-center justify-center bg-zinc-800 flex-shrink-0">
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
          <p className="text-xs text-zinc-500 leading-tight">Super Admin</p>
        </div>
      </div>

      {/* Nav */}
      <nav className="flex-1 overflow-y-auto py-4 px-3 space-y-0.5">
        {nav.map(({ href, label, icon: Icon }) => {
          const active = pathname === href || (pathname ?? '').startsWith(href + '/')
          return (
            <Link
              key={href}
              href={href}
              className={clsx(
                'flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-all duration-150 group',
                active
                  ? 'bg-amber-500/15 text-amber-400 font-medium'
                  : 'text-zinc-500 hover:text-zinc-100 hover:bg-zinc-800',
              )}
            >
              <Icon className={clsx('w-4 h-4 flex-shrink-0', active ? 'text-amber-400' : 'text-zinc-600 group-hover:text-zinc-400')} />
              <span className="truncate">{label}</span>
            </Link>
          )
        })}
      </nav>

      {/* Footer */}
      <div className="px-4 py-4 border-t border-zinc-800">
        <div className="flex items-center gap-3">
          <div className="w-9 h-9 rounded-lg overflow-hidden flex-shrink-0 bg-zinc-800">
            <Image src="/logo.png" alt="TrackOra" width={36} height={36} className="object-contain" />
          </div>
          <div className="min-w-0">
            <p className="text-xs font-semibold text-zinc-200 truncate">Trackora Technologies</p>
            <p className="text-xs text-zinc-600 truncate">Pvt. Ltd.</p>
          </div>
        </div>
      </div>
    </aside>
  )
}
