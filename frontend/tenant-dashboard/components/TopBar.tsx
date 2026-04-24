'use client'

import { HiOutlineBell, HiOutlineCommandLine } from 'react-icons/hi2'

interface Props {
  title: string
  subtitle?: string
}

export default function TopBar({ title, subtitle }: Props) {
  return (
    <header className="h-16 border-b border-surface-border flex items-center justify-between px-6 bg-surface-card/50 backdrop-blur-sm sticky top-0 z-40">
      <div>
        <h1 className="text-base font-semibold text-white">{title}</h1>
        {subtitle && <p className="text-xs text-slate-500">{subtitle}</p>}
      </div>
      <div className="flex items-center gap-3">
        <button className="btn-ghost relative p-2 rounded-full">
          <HiOutlineBell className="w-5 h-5" />
          <span className="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full animate-pulse" />
        </button>
        <button className="btn-ghost p-2 rounded-full">
          <HiOutlineCommandLine className="w-5 h-5" />
        </button>
      </div>
    </header>
  )
}
