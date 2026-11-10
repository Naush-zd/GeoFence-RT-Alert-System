import { useCallback, useEffect, useState } from "react"
import { BrowserRouter, Route, Routes } from "react-router-dom"
import { api } from "./lib/api"
import { initialDashboardData, type DashboardData } from "./lib/dashboard"
import { ThemeProvider } from "./components/theme-provider"
import { TooltipProvider } from "./components/ui/tooltip"
import { AppLayout } from "./components/app-layout"
import { OverviewPage } from "./pages/overview-page"
import { GeofencesPage } from "./pages/geofences-page"
import { VehiclesPage } from "./pages/vehicles-page"
import { AlertsPage, ViolationsPage } from "./pages/secondary-pages"

function App() {
  const [data, setData] = useState<DashboardData>(initialDashboardData)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const loadDashboard = useCallback(async () => {
    setLoading(true); setError(null)
    try {
      const [health, geofences, vehicles, alerts, violations] = await Promise.all([api.health(), api.geofences(), api.vehicles(), api.alerts(), api.violations()])
      setData({ geofences: geofences.geofences, vehicles: vehicles.vehicles, alerts: alerts.alerts, violations: violations.violations, totalViolations: violations.total_count, apiStatus: health.status })
    } catch (requestError) { setData((current) => ({ ...current, apiStatus: "unavailable" })); setError(requestError instanceof Error ? requestError.message : "Unable to load dashboard data") }
    finally { setLoading(false) }
  }, [])
  useEffect(() => { const timer = window.setTimeout(() => void loadDashboard(), 0); return () => window.clearTimeout(timer) }, [loadDashboard])
  return <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme"><BrowserRouter><TooltipProvider><AppLayout data={data} loading={loading} error={error} onRefresh={() => void loadDashboard()}><Routes><Route path="/" element={<OverviewPage data={data} onRefresh={loadDashboard} />} /><Route path="/geofences" element={<GeofencesPage geofences={data.geofences} onCreated={() => void loadDashboard()} />} /><Route path="/vehicles" element={<VehiclesPage vehicles={data.vehicles} onCreated={() => void loadDashboard()} />} /><Route path="/alerts" element={<AlertsPage alerts={data.alerts} />} /><Route path="/violations" element={<ViolationsPage violations={data.violations} total={data.totalViolations} />} /></Routes></AppLayout></TooltipProvider></BrowserRouter></ThemeProvider>
}

export default App
