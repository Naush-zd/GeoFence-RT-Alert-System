import { useCallback, useEffect, useState } from "react"
import type { ReactNode } from "react"
import { api, type AlertConfig, type Geofence, type Vehicle, type Violation } from "./lib/api"
import "./App.css"

type DashboardData = {
  geofences: Geofence[]
  vehicles: Vehicle[]
  alerts: AlertConfig[]
  violations: Violation[]
  totalViolations: number
  apiStatus: string
}

const initialData: DashboardData = {
  geofences: [],
  vehicles: [],
  alerts: [],
  violations: [],
  totalViolations: 0,
  apiStatus: "checking",
}

function App() {
  const [data, setData] = useState(initialData)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const loadDashboard = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const [health, geofences, vehicles, alerts, violations] = await Promise.all([
        api.health(),
        api.geofences(),
        api.vehicles(),
        api.alerts(),
        api.violations(),
      ])
      setData({
        geofences: geofences.geofences,
        vehicles: vehicles.vehicles,
        alerts: alerts.alerts,
        violations: violations.violations,
        totalViolations: violations.total_count,
        apiStatus: health.status,
      })
    } catch (requestError) {
      setData((current) => ({ ...current, apiStatus: "unavailable" }))
      setError(requestError instanceof Error ? requestError.message : "Unable to load dashboard data")
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const timer = window.setTimeout(() => {
      void loadDashboard()
    }, 0)
    return () => window.clearTimeout(timer)
  }, [loadDashboard])

  return (
    <main className="dashboard">
      <header className="topbar">
        <div>
          <p className="eyebrow">Operations center</p>
          <h1>GeoFence command</h1>
          <p className="subtitle">Monitor geofences, vehicles, and movement events from one place.</p>
        </div>
        <button className="refresh-button" type="button" onClick={() => void loadDashboard()} disabled={loading}>
          {loading ? "Refreshing…" : "Refresh data"}
        </button>
      </header>

      {error && <div className="error-banner" role="alert">{error}</div>}

      <section className="status-row" aria-label="System status">
        <div className={`status-pill status-${data.apiStatus}`}>
          <span className="status-dot" />
          API {data.apiStatus}
        </div>
        <span className="updated-label">{loading ? "Loading live data…" : "Live data loaded"}</span>
      </section>

      <section className="metric-grid" aria-label="System overview">
        <MetricCard label="Active geofences" value={data.geofences.length} detail="Configured boundaries" />
        <MetricCard label="Registered vehicles" value={data.vehicles.length} detail="Tracked fleet" />
        <MetricCard label="Alert rules" value={data.alerts.length} detail="Active configurations" />
        <MetricCard label="Violation events" value={data.totalViolations} detail="Recorded history" />
      </section>

      <section className="content-grid">
        <Panel title="Geofences" description="Configured monitoring boundaries">
          {data.geofences.length === 0 ? (
            <EmptyState text="No geofences configured yet." />
          ) : (
            <div className="record-list">
              {data.geofences.map((geofence) => (
                <div className="record" key={geofence.id}>
                  <div>
                    <strong>{geofence.name}</strong>
                    <span>{geofence.description || "No description"}</span>
                  </div>
                  <Badge text={geofence.category.replace("_", " ")} />
                </div>
              ))}
            </div>
          )}
        </Panel>

        <Panel title="Fleet" description="Registered vehicles">
          {data.vehicles.length === 0 ? (
            <EmptyState text="No vehicles registered yet." />
          ) : (
            <div className="record-list">
              {data.vehicles.map((vehicle) => (
                <div className="record" key={vehicle.id}>
                  <div>
                    <strong>{vehicle.vehicle_number}</strong>
                    <span>{vehicle.driver_name} · {vehicle.vehicle_type}</span>
                  </div>
                  <Badge text={vehicle.status} />
                </div>
              ))}
            </div>
          )}
        </Panel>

        <Panel title="Recent violations" description={`${data.totalViolations} total recorded events`}>
          {data.violations.length === 0 ? (
            <EmptyState text="No violation events recorded." />
          ) : (
            <div className="record-list">
              {data.violations.map((violation) => (
                <div className="record" key={violation.id}>
                  <div>
                    <strong>{violation.geofence_name}</strong>
                    <span>{violation.vehicle_number} · {formatDate(violation.timestamp)}</span>
                  </div>
                  <Badge text={violation.event_type} danger />
                </div>
              ))}
            </div>
          )}
        </Panel>

        <Panel title="Alert rules" description="Rules currently watching for transitions">
          {data.alerts.length === 0 ? (
            <EmptyState text="No alert rules configured." />
          ) : (
            <div className="record-list">
              {data.alerts.map((alert) => (
                <div className="record" key={alert.alert_id}>
                  <div>
                    <strong>{alert.geofence_name}</strong>
                    <span>{alert.vehicle_number || "All vehicles"} · {alert.event_type}</span>
                  </div>
                  <Badge text={alert.status} />
                </div>
              ))}
            </div>
          )}
        </Panel>
      </section>
    </main>
  )
}

function MetricCard({ label, value, detail }: { label: string; value: number; detail: string }) {
  return (
    <article className="metric-card">
      <span>{label}</span>
      <strong>{value}</strong>
      <small>{detail}</small>
    </article>
  )
}

function Panel({ title, description, children }: { title: string; description: string; children: ReactNode }) {
  return (
    <article className="panel">
      <div className="panel-header">
        <div>
          <h2>{title}</h2>
          <p>{description}</p>
        </div>
      </div>
      {children}
    </article>
  )
}

function Badge({ text, danger = false }: { text: string; danger?: boolean }) {
  return <span className={`badge ${danger ? "badge-danger" : ""}`}>{text}</span>
}

function EmptyState({ text }: { text: string }) {
  return <p className="empty-state">{text}</p>
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value))
}

export default App
