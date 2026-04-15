import { create } from 'zustand'

export interface Alert {
  id: string
  device_id: string
  vehicle_id?: string
  rule_id?: string
  alert_type: string
  severity: 'info' | 'warning' | 'critical'
  message: string
  triggered_at: string
  acknowledged_at?: string
  lat?: number
  lng?: number
}

interface AlertStore {
  alerts: Alert[]
  unreadCount: number
  setAlerts: (alerts: Alert[]) => void
  addAlert: (alert: Alert) => void
  acknowledge: (id: string) => void
  clearUnread: () => void
}

export const useAlertStore = create<AlertStore>((set) => ({
  alerts: [],
  unreadCount: 0,

  setAlerts: (alerts) =>
    set({
      alerts,
      unreadCount: alerts.filter((a) => !a.acknowledged_at).length,
    }),

  addAlert: (alert) =>
    set((s) => ({
      alerts: [alert, ...s.alerts].slice(0, 500),
      unreadCount: s.unreadCount + (alert.acknowledged_at ? 0 : 1),
    })),

  acknowledge: (id) =>
    set((s) => ({
      alerts: s.alerts.map((a) =>
        a.id === id ? { ...a, acknowledged_at: new Date().toISOString() } : a
      ),
      unreadCount: Math.max(0, s.unreadCount - 1),
    })),

  clearUnread: () => set({ unreadCount: 0 }),
}))
