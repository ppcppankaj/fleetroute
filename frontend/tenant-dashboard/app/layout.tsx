import type { Metadata } from 'next'
import './globals.css'

export const metadata: Metadata = {
  title: 'TrackOra — Dashboard',
  description: 'Enterprise GPS Fleet Tracking & Management by Trackora Technologies Pvt. Ltd.',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" className="dark">
      <body className="bg-surface text-slate-100 font-sans antialiased">
        {children}
      </body>
    </html>
  )
}
