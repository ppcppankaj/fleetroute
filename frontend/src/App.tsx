import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './shared/store/authStore'
import AppShell from './shared/components/AppShell'
import LoginPage from './features/auth/LoginPage'
import LiveTracking from './features/tracking/LiveTracking'
import FleetPage from './features/fleet/FleetPage'
import AlertsPage from './features/alerts/AlertsPage'
import { ReportsPage } from './features/reports/ReportsPage'
import { SettingsPage } from './features/settings/SettingsPage'
import { GeofencePage } from './features/geofences/GeofencePage'
import { MaintenancePage } from './features/maintenance/MaintenancePage'
import { RoutePlayback } from './features/playback/RoutePlayback'
import { VehicleDetail } from './features/fleet/VehicleDetail'
import { DriverDetail } from './features/fleet/DriverDetail'
import FuelPage from './features/fuel/FuelPage'
import DevicesPage from './features/devices/DevicesPage'
import UsersPage from './features/users/UsersPage'
import BillingPage from './features/billing/BillingPage'
import ActivityPage from './features/activity/ActivityPage'
import VideoTelematicsPage from './features/video/VideoTelematicsPage'

export default function App() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)

  if (!isAuthenticated) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/tracking" replace />} />
        <Route path="/tracking" element={<LiveTracking />} />
        <Route path="/fleet" element={<FleetPage />} />
        <Route path="/fleet/vehicles/:id" element={<VehicleDetail />} />
        <Route path="/fleet/drivers/:id" element={<DriverDetail />} />
        <Route path="/geofences" element={<GeofencePage />} />
        <Route path="/alerts" element={<AlertsPage />} />
        <Route path="/maintenance" element={<MaintenancePage />} />
        <Route path="/fuel" element={<FuelPage />} />
        <Route path="/devices" element={<DevicesPage />} />
        <Route path="/video" element={<VideoTelematicsPage />} />
        <Route path="/users" element={<UsersPage />} />
        <Route path="/billing" element={<BillingPage />} />
        <Route path="/activity" element={<ActivityPage />} />
        <Route path="/reports" element={<ReportsPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/playback/:deviceId" element={<RoutePlayback />} />
        <Route path="*" element={<Navigate to="/tracking" replace />} />
      </Route>
    </Routes>
  )
}
