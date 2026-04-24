import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'TrackOra — Super Admin',
  description: 'Platform Administration — Trackora Technologies Pvt. Ltd.',
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body className="bg-surface text-zinc-100 font-sans antialiased">
        {children}
      </body>
    </html>
  )
}
