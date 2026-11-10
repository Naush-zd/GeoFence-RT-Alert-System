import { useCallback, useEffect, useRef, useState } from "react"
import { BrowserRouter, Route, Routes } from "react-router-dom"
import { api, type AlertEvent, type Violation, type ViolationFilters } from "./lib/api"
import { initialDashboardData, type DashboardData } from "./lib/dashboard"
import { ThemeProvider } from "./components/theme-provider"
import { TooltipProvider } from "./components/ui/tooltip"
import { AppLayout } from "./components/app-layout"
import { OverviewPage } from "./pages/overview-page"
import { GeofencesPage } from "./pages/geofences-page"
import { VehiclesPage } from "./pages/vehicles-page"
import { AlertsPage, ViolationsPage } from "./pages/secondary-pages"

function violationFromAlert(event: AlertEvent): Violation {
  return {
    id: event.event_id,
    vehicle_id: event.vehicle.vehicle_id,
    vehicle_number: event.vehicle.vehicle_number,
    geofence_id: event.geofence.geofence_id,
    geofence_name: event.geofence.geofence_name,
    event_type: event.event_type,
    latitude: event.location.latitude,
    longitude: event.location.longitude,
    timestamp: event.timestamp,
  }
}

function matchesViolationFilters(violation: Violation, filters: ViolationFilters) {
  if (filters.vehicle_id && violation.vehicle_id !== filters.vehicle_id) return false
  if (filters.geofence_id && violation.geofence_id !== filters.geofence_id) return false
  const timestamp = Date.parse(violation.timestamp)
  if (filters.start_date && timestamp < Date.parse(filters.start_date)) return false
  if (filters.end_date && timestamp > Date.parse(filters.end_date)) return false
  return true
}

function mergeLiveViolations(violations: Violation[], total: number, filters: ViolationFilters, liveViolations: Map<string, Violation>) {
  const persistedIds = new Set(violations.map((violation) => violation.id))
  const unpersisted = [...liveViolations.values()].filter((violation) => !persistedIds.has(violation.id) && matchesViolationFilters(violation, filters))
  return {
    violations: [...unpersisted, ...violations].slice(0, filters.limit ?? 50),
    totalViolations: total + unpersisted.length,
  }
}

function App() {
  const [data, setData] = useState<DashboardData>(initialDashboardData)
  const [liveEvents, setLiveEvents] = useState<AlertEvent[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const activeViolationFilters = useRef<ViolationFilters>({})
  const receivedViolations = useRef(new Map<string, Violation>())
  const loadDashboard = useCallback(async () => {
    setLoading(true); setError(null)
    try {
      const [health, geofences, vehicles, alerts, violations] = await Promise.all([api.health(), api.geofences(), api.vehicles(), api.alerts(), api.violations(activeViolationFilters.current)])
      const mergedViolations = mergeLiveViolations(violations.violations, violations.total_count, activeViolationFilters.current, receivedViolations.current)
      setData({ geofences: geofences.geofences, vehicles: vehicles.vehicles, alerts: alerts.alerts, ...mergedViolations, apiStatus: health.status })
    } catch (requestError) { setData((current) => ({ ...current, apiStatus: "unavailable" })); setError(requestError instanceof Error ? requestError.message : "Unable to load dashboard data") }
    finally { setLoading(false) }
  }, [])
  const filterViolations = useCallback(async (filters: ViolationFilters) => {
    activeViolationFilters.current = filters
    try {
      const result = await api.violations(filters)
      const mergedViolations = mergeLiveViolations(result.violations, result.total_count, filters, receivedViolations.current)
      setData((current) => ({ ...current, ...mergedViolations }))
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to load violation history")
    }
  }, [])
  const handleAlertEvent = useCallback((event: AlertEvent) => {
    setLiveEvents((current) => [event, ...current.filter((item) => item.event_id !== event.event_id)].slice(0, 50))
    const violation = violationFromAlert(event)
    receivedViolations.current.set(violation.id, violation)
    setData((current) => {
      const filters = activeViolationFilters.current
      if (current.violations.some((item) => item.id === violation.id) || !matchesViolationFilters(violation, filters)) return current
      return {
        ...current,
        violations: [violation, ...current.violations].slice(0, filters.limit ?? 50),
        totalViolations: current.totalViolations + 1,
      }
    })
  }, [])
  useEffect(() => { const timer = window.setTimeout(() => void loadDashboard(), 0); return () => window.clearTimeout(timer) }, [loadDashboard])
  return <ThemeProvider defaultTheme="light" storageKey="vite-ui-theme"><BrowserRouter><TooltipProvider><AppLayout data={data} loading={loading} error={error} onRefresh={() => void loadDashboard()} onAlertEvent={handleAlertEvent}><Routes><Route path="/" element={<OverviewPage data={data} onRefresh={loadDashboard} />} /><Route path="/geofences" element={<GeofencesPage geofences={data.geofences} onCreated={() => void loadDashboard()} />} /><Route path="/vehicles" element={<VehiclesPage vehicles={data.vehicles} onCreated={() => void loadDashboard()} />} /><Route path="/alerts" element={<AlertsPage alerts={data.alerts} events={liveEvents} geofences={data.geofences} vehicles={data.vehicles} onCreated={() => void loadDashboard()} />} /><Route path="/violations" element={<ViolationsPage violations={data.violations} total={data.totalViolations} vehicles={data.vehicles} geofences={data.geofences} onFilter={filterViolations} />} /></Routes></AppLayout></TooltipProvider></BrowserRouter></ThemeProvider>
}

export default App
