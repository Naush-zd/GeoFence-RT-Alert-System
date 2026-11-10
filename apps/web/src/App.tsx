import { useCallback, useEffect, useState } from "react"
import type { ReactNode } from "react"
import { BrowserRouter, NavLink, Route, Routes, useLocation } from "react-router-dom"
import {
  AlertTriangle,
  BellRing,
  CircleGauge,
  MapPinned,
  Radio,
  Truck,
} from "lucide-react"
import { api, type AlertConfig, type Geofence, type Vehicle, type Violation } from "./lib/api"
import { ModeToggle } from "./components/mode-toggle"
import { ThemeProvider } from "./components/theme-provider"
import { Badge } from "./components/ui/badge"
import { Button } from "./components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./components/ui/card"
import { ScrollArea } from "./components/ui/scroll-area"
import { Separator } from "./components/ui/separator"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "./components/ui/sidebar"
import { TooltipProvider } from "./components/ui/tooltip"

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

const navItems = [
  { to: "/", label: "Overview", icon: CircleGauge, end: true },
  { to: "/geofences", label: "Geofences", icon: MapPinned },
  { to: "/vehicles", label: "Vehicles", icon: Truck },
  { to: "/alerts", label: "Alert rules", icon: BellRing },
  { to: "/violations", label: "Violations", icon: AlertTriangle },
]

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
    const timer = window.setTimeout(() => void loadDashboard(), 0)
    return () => window.clearTimeout(timer)
  }, [loadDashboard])

  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <BrowserRouter>
        <TooltipProvider>
          <SidebarProvider defaultOpen={false}>
          <Sidebar collapsible="icon" variant="sidebar">
            <SidebarHeader className="p-3">
              <div className="flex items-center gap-2 px-1">
                <div className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground"><Radio size={18} /></div>
                <div className="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden">
                  <span className="truncate font-semibold text-sidebar-foreground">GeoFence</span>
                  <span className="truncate text-xs text-sidebar-foreground/60">Control center</span>
                </div>
              </div>
            </SidebarHeader>
            <SidebarSeparator />
            <SidebarContent className="overflow-hidden">
              <ScrollArea className="h-full">
                <SidebarGroup>
                  <SidebarGroupLabel>Workspace</SidebarGroupLabel>
                  <SidebarGroupContent>
                    <SidebarMenu>
                      {navItems.map(({ to, label, icon: Icon, end }) => (
                        <SidebarMenuItem key={to}>
                          <NavLink to={to} end={end}>
                            {({ isActive }) => (
                              <SidebarMenuButton render={<span />} isActive={isActive} tooltip={label}>
                                <Icon />
                                <span>{label}</span>
                              </SidebarMenuButton>
                            )}
                          </NavLink>
                        </SidebarMenuItem>
                      ))}
                    </SidebarMenu>
                  </SidebarGroupContent>
                </SidebarGroup>
              </ScrollArea>
            </SidebarContent>
            <SidebarFooter className="gap-2 p-3">
              <Separator className="bg-sidebar-border" />
              <div className="flex items-center gap-2 text-xs">
                <span className={`size-2 rounded-full ${data.apiStatus === "ok" ? "bg-emerald-400" : data.apiStatus === "checking" ? "bg-amber-400" : "bg-red-400"}`} />
                <span className="truncate group-data-[collapsible=icon]:hidden">{data.apiStatus === "ok" ? "All systems operational" : "Check API connection"}</span>
              </div>
              <span className="text-[11px] text-sidebar-foreground/50 group-data-[collapsible=icon]:hidden">v0.1.0 · Local workspace</span>
            </SidebarFooter>
          </Sidebar>
          <SidebarInset>
            <header className="flex h-14 items-center gap-3 border-b bg-background/95 px-4 backdrop-blur md:px-6">
              <SidebarTrigger aria-label="Toggle navigation" />
              <Separator orientation="vertical" className="h-5" />
              <span className="text-sm font-semibold">GeoFence</span>
              <span className="ml-auto flex items-center gap-3 text-xs text-muted-foreground"><span className={`flex items-center gap-2`}><span className={`size-2 rounded-full ${data.apiStatus === "ok" ? "bg-emerald-500" : "bg-amber-500"}`} /> API {data.apiStatus}</span><ModeToggle /></span>
            </header>
            <main className="min-w-0 flex-1 px-4 pb-12 md:px-8 lg:px-12">
          <PageHeader data={data} loading={loading} onRefresh={() => void loadDashboard()} />
          {error && <div className="mt-4 rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive" role="alert">{error}</div>}
          <Routes>
            <Route path="/" element={<OverviewPage data={data} />} />
            <Route path="/geofences" element={<GeofencesPage geofences={data.geofences} />} />
            <Route path="/vehicles" element={<VehiclesPage vehicles={data.vehicles} />} />
            <Route path="/alerts" element={<AlertsPage alerts={data.alerts} />} />
            <Route path="/violations" element={<ViolationsPage violations={data.violations} total={data.totalViolations} />} />
          </Routes>
            </main>
          </SidebarInset>
          </SidebarProvider>
        </TooltipProvider>
      </BrowserRouter>
    </ThemeProvider>
  )
}

function PageHeader({ data, loading, onRefresh }: { data: DashboardData; loading: boolean; onRefresh: () => void }) {
  const location = useLocation()
  const current = navItems.find((item) => item.to === location.pathname)
  return (
    <header className="flex flex-col gap-6 border-b py-8 md:flex-row md:items-end md:justify-between">
      <div><p className="mb-3 text-xs text-muted-foreground">Workspace / {current?.label ?? "Overview"}</p><h1 className="text-3xl font-semibold tracking-tight">{current?.label ?? "Overview"}</h1><p className="mt-2 text-sm text-muted-foreground">{pageDescriptions[location.pathname] ?? pageDescriptions["/"]}</p></div>
      <div className="flex items-center gap-3"><Badge variant={data.apiStatus === "ok" ? "secondary" : data.apiStatus === "checking" ? "outline" : "destructive"}>API {data.apiStatus}</Badge><Button variant="outline" onClick={onRefresh} disabled={loading}>{loading ? "Refreshing…" : "Refresh"}</Button></div>
    </header>
  )
}

const pageDescriptions: Record<string, string> = {
  "/": "A live snapshot of your geofencing operations.",
  "/geofences": "Define and monitor the boundaries that matter to your fleet.",
  "/vehicles": "Keep track of every registered vehicle in your network.",
  "/alerts": "Configure the events that need your attention.",
  "/violations": "Review historical entries and exits across your zones.",
}

function OverviewPage({ data }: { data: DashboardData }) {
  return <><section className="grid gap-4 py-6 sm:grid-cols-2 xl:grid-cols-4"><MetricCard label="Active geofences" value={data.geofences.length} detail="Configured boundaries" icon={<MapPinned />} /><MetricCard label="Registered vehicles" value={data.vehicles.length} detail="Tracked fleet" icon={<Truck />} /><MetricCard label="Alert rules" value={data.alerts.length} detail="Active configurations" icon={<BellRing />} /><MetricCard label="Violation events" value={data.totalViolations} detail="Recorded history" icon={<AlertTriangle />} /></section><section className="grid gap-4 lg:grid-cols-2"><Panel title="Geofences" description="Configured monitoring boundaries" link="/geofences"><GeofenceList items={data.geofences.slice(0, 4)} /></Panel><Panel title="Fleet" description="Registered vehicles" link="/vehicles"><VehicleList items={data.vehicles.slice(0, 4)} /></Panel><Panel title="Recent violations" description={`${data.totalViolations} total recorded events`} link="/violations"><ViolationList items={data.violations} /></Panel><Panel title="Alert rules" description="Rules watching for transitions" link="/alerts"><AlertList items={data.alerts.slice(0, 4)} /></Panel></section></>
}

function GeofencesPage({ geofences }: { geofences: Geofence[] }) { return <PageSection title="Geofence registry" count={`${geofences.length} zones`} action="+ New geofence"><ListCard><GeofenceList items={geofences} /></ListCard></PageSection> }
function VehiclesPage({ vehicles }: { vehicles: Vehicle[] }) { return <PageSection title="Vehicle registry" count={`${vehicles.length} vehicles`} action="+ Register vehicle"><ListCard><VehicleList items={vehicles} /></ListCard></PageSection> }
function AlertsPage({ alerts }: { alerts: AlertConfig[] }) { return <PageSection title="Alert rules" count={`${alerts.length} rules`} action="+ Configure alert"><ListCard><AlertList items={alerts} /></ListCard></PageSection> }
function ViolationsPage({ violations, total }: { violations: Violation[]; total: number }) { return <PageSection title="Violation history" count={`${total} total events`} action="Export history" variant="outline"><ListCard><ViolationList items={violations} /></ListCard></PageSection> }

function PageSection({ title, count, action, variant = "default", children }: { title: string; count: string; action: string; variant?: "default" | "outline"; children: ReactNode }) { return <section className="space-y-4 py-6"><div className="flex items-end justify-between gap-4"><div><h2 className="text-lg font-semibold">{title}</h2><p className="text-sm text-muted-foreground">{count}</p></div><Button variant={variant}>{action}</Button></div>{children}</section> }
function ListCard({ children }: { children: ReactNode }) { return <Card><CardContent className="p-2">{children}</CardContent></Card> }
function MetricCard({ label, value, detail, icon }: { label: string; value: number; detail: string; icon: ReactNode }) { return <Card><CardContent className="p-5"><div className="mb-4 flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">{icon}</div><p className="text-sm text-muted-foreground">{label}</p><p className="mt-2 text-3xl font-semibold tracking-tight">{value}</p><p className="mt-1 text-xs text-muted-foreground">{detail}</p></CardContent></Card> }
function Panel({ title, description, link, children }: { title: string; description: string; link: string; children: ReactNode }) { return <Card><CardHeader className="flex flex-row items-start justify-between gap-4"><div><CardTitle>{title}</CardTitle><CardDescription>{description}</CardDescription></div><Button variant="ghost" size="sm" render={<NavLink to={link} />}>View all</Button></CardHeader><CardContent>{children}</CardContent></Card> }
function GeofenceList({ items }: { items: Geofence[] }) { return items.length === 0 ? <EmptyState text="No geofences configured yet." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary"><MapPinned size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.name}</p><p className="truncate text-xs text-muted-foreground">{item.description || "No description"}</p></div></div><Badge variant="secondary">{item.category.replaceAll("_", " ")}</Badge></div>)}</div> }
function VehicleList({ items }: { items: Vehicle[] }) { return items.length === 0 ? <EmptyState text="No vehicles registered yet." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary"><Truck size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.vehicle_number}</p><p className="truncate text-xs text-muted-foreground">{item.driver_name} · {item.vehicle_type}</p></div></div><Badge>{item.status}</Badge></div>)}</div> }
function AlertList({ items }: { items: AlertConfig[] }) { return items.length === 0 ? <EmptyState text="No alert rules configured." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.alert_id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary"><BellRing size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.geofence_name}</p><p className="truncate text-xs text-muted-foreground">{item.vehicle_number || "All vehicles"} · {item.event_type}</p></div></div><Badge>{item.status}</Badge></div>)}</div> }
function ViolationList({ items }: { items: Violation[] }) { return items.length === 0 ? <EmptyState text="No violation events recorded." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-destructive/15 text-destructive"><AlertTriangle size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.geofence_name}</p><p className="truncate text-xs text-muted-foreground">{item.vehicle_number} · {formatDate(item.timestamp)}</p></div></div><Badge variant="destructive">{item.event_type}</Badge></div>)}</div> }
function EmptyState({ text }: { text: string }) { return <p className="py-8 text-center text-sm text-muted-foreground">{text}</p> }
function formatDate(value: string) { return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) }

export default App
