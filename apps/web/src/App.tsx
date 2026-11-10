import { useCallback, useEffect, useState } from "react"
import type { ReactNode } from "react"
import { BrowserRouter, NavLink, Route, Routes, useLocation } from "react-router-dom"
import {
  AlertTriangle,
  BellRing,
  CircleGauge,
  MapPinned,
  Play,
  Radio,
  Truck,
  Trash2,
} from "lucide-react"
import { api, type AlertConfig, type Geofence, type Vehicle, type Violation } from "./lib/api"
import { ModeToggle } from "./components/mode-toggle"
import { ThemeProvider } from "./components/theme-provider"
import { Badge } from "./components/ui/badge"
import { Button } from "./components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "./components/ui/dialog"
import { Input } from "./components/ui/input"
import { Label } from "./components/ui/label"
import { ScrollArea } from "./components/ui/scroll-area"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./components/ui/select"
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
import { OperationsMap } from "./components/operations-map"
import { api as apiClient } from "./lib/api"

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
            <Route path="/" element={<OverviewPage data={data} onRefresh={loadDashboard} />} />
            <Route path="/geofences" element={<GeofencesPage geofences={data.geofences} onCreated={() => void loadDashboard()} />} />
            <Route path="/vehicles" element={<VehiclesPage vehicles={data.vehicles} onCreated={() => void loadDashboard()} />} />
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

function OverviewPage({ data, onRefresh }: { data: DashboardData; onRefresh: () => Promise<void> }) {
  return <><section className="grid gap-4 py-6 sm:grid-cols-2 xl:grid-cols-4"><MetricCard label="Active geofences" value={data.geofences.length} detail="Configured boundaries" icon={<MapPinned />} /><MetricCard label="Registered vehicles" value={data.vehicles.length} detail="Tracked fleet" icon={<Truck />} /><MetricCard label="Alert rules" value={data.alerts.length} detail="Active configurations" icon={<BellRing />} /><MetricCard label="Violation events" value={data.totalViolations} detail="Recorded history" icon={<AlertTriangle />} /></section><SimulationCard onComplete={onRefresh} /><section className="grid gap-4 lg:grid-cols-2"><Panel title="Geofences" description="Configured monitoring boundaries" link="/geofences"><GeofenceList items={data.geofences.slice(0, 4)} /></Panel><Panel title="Fleet" description="Registered vehicles" link="/vehicles"><VehicleList items={data.vehicles.slice(0, 4)} /></Panel><Panel title="Recent violations" description={`${data.totalViolations} total recorded events`} link="/violations"><ViolationList items={data.violations} /></Panel><Panel title="Alert rules" description="Rules watching for transitions" link="/alerts"><AlertList items={data.alerts.slice(0, 4)} /></Panel></section></>
}

function SimulationCard({ onComplete }: { onComplete: () => Promise<void> }) {
  const [running, setRunning] = useState(false)
  const [simulationGeofence, setSimulationGeofence] = useState<Geofence | null>(null)
  const [simulationVehicle, setSimulationVehicle] = useState<Vehicle | null>(null)
  const [vehiclePosition, setVehiclePosition] = useState<[number, number] | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const run = async () => {
    setRunning(true)
    setMessage(null)
    setError(null)
    try {
      const delhiCoordinates = [
        [28.48, 77.05],
        [28.48, 77.38],
        [28.78, 77.38],
        [28.78, 77.05],
        [28.48, 77.05],
      ]
      const [currentGeofences, currentVehicles] = await Promise.all([apiClient.geofences(), apiClient.vehicles()])
      const existingGeofence = currentGeofences.geofences.find((item) => item.name === "Delhi delivery boundary")
      let geofence: Geofence
      if (existingGeofence) {
        geofence = existingGeofence
      } else {
        const result = await apiClient.createGeofence({
          name: "Delhi delivery boundary",
          description: "Delhi boundary for the vehicle entry and exit walkthrough.",
          category: "delivery_zone",
          coordinates: delhiCoordinates,
        })
        geofence = {
          id: result.id,
          name: result.name,
          description: "Delhi boundary for the vehicle entry and exit walkthrough.",
          category: "delivery_zone",
          coordinates: delhiCoordinates,
          created_at: new Date().toISOString(),
        }
      }
      const existingVehicle = currentVehicles.vehicles.find((item) => item.vehicle_number === "DEMO-001")
      let vehicle: Vehicle
      if (existingVehicle) {
        vehicle = existingVehicle
      } else {
        const result = await apiClient.createVehicle({
          vehicle_number: "DEMO-001",
          driver_name: "Simulation Driver",
          vehicle_type: "van",
          phone: "+0000000000",
        })
        vehicle = {
          id: result.id,
          vehicle_number: result.vehicle_number,
          driver_name: "Simulation Driver",
          vehicle_type: "van",
          phone: "+0000000000",
          status: "active",
          created_at: new Date().toISOString(),
        }
      }
      const currentAlerts = await apiClient.alerts()
      const hasDemoAlert = currentAlerts.alerts.some((alert) => alert.geofence_id === geofence.id && alert.vehicle_id === vehicle.id && alert.event_type === "both")
      if (!hasDemoAlert) await apiClient.configureAlert({ geofence_id: geofence.id, vehicle_id: vehicle.id, event_type: "both" })

      const startPosition: [number, number] = [29.35, 77.6]
      await apiClient.updateVehicleLocation({
        vehicle_id: vehicle.id,
        latitude: startPosition[0],
        longitude: startPosition[1],
        timestamp: new Date().toISOString(),
      })
      setSimulationGeofence(geofence)
      setSimulationVehicle(vehicle)
      setVehiclePosition(startPosition)
      await onComplete()
      setMessage("Delhi simulation ready. Drag the blue marker into the shaded zone, then back out.")
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to run the simulation")
    } finally {
      setRunning(false)
    }
  }

  const moveVehicle = async (point: [number, number]) => {
    if (!simulationVehicle) return
    setVehiclePosition(point)
    setError(null)
    try {
      const result = await apiClient.updateVehicleLocation({
        vehicle_id: simulationVehicle.id,
        latitude: point[0],
        longitude: point[1],
        timestamp: new Date().toISOString(),
      })
      const transition = result.current_geofences.length > 0 ? "entered" : "exited"
      setMessage(`DEMO-001 ${transition} the Delhi delivery boundary.`)
      await onComplete()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to update the simulated vehicle")
    }
  }

  return <Card><CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div><CardTitle>Vehicle alert simulation</CardTitle><CardDescription>Set up a Delhi geofence, then drag the blue vehicle marker from Uttar Pradesh into Delhi and back out to generate entry and exit alerts.</CardDescription></div><Button onClick={() => void run()} disabled={running}><Play data-icon="inline-start" />{running ? "Preparing…" : simulationGeofence ? "Reset simulation" : "Set up Delhi simulation"}</Button></CardHeader>{simulationGeofence && vehiclePosition && <CardContent className="pt-0"><div className="h-[28rem] overflow-hidden rounded-lg border"><OperationsMap geofences={[simulationGeofence]} selectedGeofenceId={simulationGeofence.id} vehiclePosition={vehiclePosition} onMoveVehicle={(point) => void moveVehicle(point)} /></div><p className="mt-3 text-xs text-muted-foreground">Drag the blue marker into the shaded Delhi boundary for entry, then drag it back to Uttar Pradesh for exit.</p></CardContent>}{message && <CardContent className="pt-0"><p className="text-sm text-primary" role="status">{message}</p></CardContent>}{error && <CardContent className="pt-0"><p className="text-sm text-destructive" role="alert">{error}</p></CardContent>}</Card>
}

function GeofencesPage({ geofences, onCreated }: { geofences: Geofence[]; onCreated: () => void }) {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [draftCoordinates, setDraftCoordinates] = useState<number[][]>([])
  const [selectedGeofenceId, setSelectedGeofenceId] = useState<string | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Geofence | null>(null)

  useEffect(() => {
    if (selectedGeofenceId && !geofences.some((geofence) => geofence.id === selectedGeofenceId)) {
      setSelectedGeofenceId(null)
    }
  }, [geofences, selectedGeofenceId])

  return <PageSection title="Geofence registry" count={`${geofences.length} zones`} action="+ New geofence" actionContent={<><CreateGeofenceDialog open={dialogOpen} onOpenChange={setDialogOpen} coordinates={draftCoordinates} onCoordinatesChange={setDraftCoordinates} onCreated={onCreated} /><DeleteGeofenceDialog target={deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }} onDeleted={() => { setDeleteTarget(null); setSelectedGeofenceId(null); onCreated() }} /></>}><div className="grid gap-4 xl:grid-cols-[minmax(0,1.3fr)_minmax(20rem,0.7fr)]"><Card className="overflow-hidden"><CardHeader><CardTitle>Live boundary map</CardTitle><CardDescription>Select a boundary to focus and highlight it on the map.</CardDescription></CardHeader><CardContent className="p-0"><div className="h-[28rem]"><OperationsMap geofences={geofences} selectedGeofenceId={selectedGeofenceId} /></div></CardContent></Card><Card><CardHeader><CardTitle>Configured boundaries</CardTitle><CardDescription>Choose a geofence to inspect or remove it from monitoring.</CardDescription></CardHeader><CardContent className="pt-0"><GeofenceList items={geofences} selectedId={selectedGeofenceId} onSelect={setSelectedGeofenceId} onDelete={setDeleteTarget} /></CardContent></Card></div></PageSection>
}
function VehiclesPage({ vehicles, onCreated }: { vehicles: Vehicle[]; onCreated: () => void }) { return <PageSection title="Vehicle registry" count={`${vehicles.length} vehicles`} action="+ Register vehicle" actionContent={<RegisterVehicleDialog onCreated={onCreated} />}><ListCard><VehicleList items={vehicles} /></ListCard></PageSection> }
function AlertsPage({ alerts }: { alerts: AlertConfig[] }) { return <PageSection title="Alert rules" count={`${alerts.length} rules`} action="+ Configure alert"><ListCard><AlertList items={alerts} /></ListCard></PageSection> }
function ViolationsPage({ violations, total }: { violations: Violation[]; total: number }) { return <PageSection title="Violation history" count={`${total} total events`} action="Export history" variant="outline"><ListCard><ViolationList items={violations} /></ListCard></PageSection> }

function PageSection({ title, count, action, actionContent, children }: { title: string; count: string; action: string; variant?: "default" | "outline"; actionContent?: ReactNode; children: ReactNode }) { return <section className="flex flex-col gap-4 py-6"><div className="flex items-end justify-between gap-4"><div><h2 className="text-lg font-semibold">{title}</h2><p className="text-sm text-muted-foreground">{count}</p></div>{actionContent ?? <Button>{action}</Button>}</div>{children}</section> }
function ListCard({ children }: { children: ReactNode }) { return <Card><CardContent className="p-2">{children}</CardContent></Card> }
function CreateGeofenceDialog({
  open,
  onOpenChange,
  coordinates,
  onCoordinatesChange,
  onCreated,
}: {
  open: boolean
  onOpenChange: (open: boolean) => void
  coordinates: number[][]
  onCoordinatesChange: (coordinates: number[][]) => void
  onCreated: () => void
}) {
  const [name, setName] = useState("")
  const [description, setDescription] = useState("")
  const [category, setCategory] = useState("delivery_zone")
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const reset = () => {
    setName("")
    setDescription("")
    setCategory("delivery_zone")
    setError(null)
    onCoordinatesChange([])
  }

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (coordinates.length < 3) {
      setError("Add at least three points on the map.")
      return
    }
    setSubmitting(true)
    setError(null)
    const closedCoordinates = [...coordinates, coordinates[0]]
    try {
      await apiClient.createGeofence({ name, description, category, coordinates: closedCoordinates })
      onCreated()
      reset()
      onOpenChange(false)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to create geofence")
    } finally {
      setSubmitting(false)
    }
  }

  const handleOpenChange = (nextOpen: boolean) => {
    onOpenChange(nextOpen)
    if (!nextOpen) reset()
  }

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger render={<Button />}>+ New geofence</DialogTrigger>
      <DialogContent className="max-h-[92vh] overflow-y-auto sm:max-w-4xl">
        <DialogHeader>
          <DialogTitle>Create geofence</DialogTitle>
          <DialogDescription>Click three or more points, then drag the handles to resize the polygon.</DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={(event) => void submit(event)}>
          <div className="overflow-hidden rounded-lg border">
            <div className="h-72">
              <OperationsMap
                geofences={[]}
                draftCoordinates={coordinates}
                onAddPoint={(point) => onCoordinatesChange([...coordinates, point])}
                onMovePoint={(index, point) => onCoordinatesChange(coordinates.map((currentPoint, currentIndex) => currentIndex === index ? point : currentPoint))}
              />
            </div>
          </div>
          <div className="flex items-center justify-between gap-3">
            <p className="text-xs text-muted-foreground">{coordinates.length} points selected. A closing point will be added automatically.</p>
            <Button type="button" variant="outline" size="sm" onClick={() => onCoordinatesChange([])} disabled={coordinates.length === 0}>Reset points</Button>
          </div>
          <div className="flex flex-col gap-2"><Label htmlFor="geofence-name">Name</Label><Input id="geofence-name" value={name} onChange={(event) => setName(event.target.value)} required placeholder="Downtown delivery zone" /></div>
          <div className="flex flex-col gap-2"><Label htmlFor="geofence-description">Description</Label><Input id="geofence-description" value={description} onChange={(event) => setDescription(event.target.value)} placeholder="Optional context" /></div>
          <div className="flex flex-col gap-2"><Label>Category</Label><Select value={category} onValueChange={(value) => setCategory(value ?? "delivery_zone")}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="delivery_zone">Delivery zone</SelectItem><SelectItem value="restricted_zone">Restricted zone</SelectItem><SelectItem value="toll_zone">Toll zone</SelectItem><SelectItem value="customer_area">Customer area</SelectItem></SelectContent></Select></div>
          {error && <p className="text-sm text-destructive" role="alert">{error}</p>}
          <DialogFooter><Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button><Button type="submit" disabled={submitting}>{submitting ? "Creating…" : "Create geofence"}</Button></DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function RegisterVehicleDialog({ onCreated }: { onCreated: () => void }) {
  const [open, setOpen] = useState(false)
  const [vehicleNumber, setVehicleNumber] = useState("")
  const [driverName, setDriverName] = useState("")
  const [vehicleType, setVehicleType] = useState("truck")
  const [phone, setPhone] = useState("")
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      await apiClient.createVehicle({ vehicle_number: vehicleNumber, driver_name: driverName, vehicle_type: vehicleType, phone })
      onCreated()
      setOpen(false)
      setVehicleNumber("")
      setDriverName("")
      setVehicleType("truck")
      setPhone("")
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to register vehicle")
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button />}>+ Register vehicle</DialogTrigger>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Register vehicle</DialogTitle>
          <DialogDescription>Add a vehicle to the tracked fleet.</DialogDescription>
        </DialogHeader>
        <form className="flex flex-col gap-4" onSubmit={(event) => void submit(event)}>
          <div className="flex flex-col gap-2"><Label htmlFor="vehicle-number">Vehicle number</Label><Input id="vehicle-number" value={vehicleNumber} onChange={(event) => setVehicleNumber(event.target.value)} required placeholder="KA-01-AB-1234" /></div>
          <div className="flex flex-col gap-2"><Label htmlFor="driver-name">Driver name</Label><Input id="driver-name" value={driverName} onChange={(event) => setDriverName(event.target.value)} required placeholder="John Doe" /></div>
          <div className="flex flex-col gap-2"><Label>Vehicle type</Label><Select value={vehicleType} onValueChange={(value) => setVehicleType(value ?? "truck")}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="truck">Truck</SelectItem><SelectItem value="van">Van</SelectItem><SelectItem value="car">Car</SelectItem><SelectItem value="motorcycle">Motorcycle</SelectItem></SelectContent></Select></div>
          <div className="flex flex-col gap-2"><Label htmlFor="driver-phone">Driver phone</Label><Input id="driver-phone" value={phone} onChange={(event) => setPhone(event.target.value)} required placeholder="+1234567890" /></div>
          {error && <p className="text-sm text-destructive" role="alert">{error}</p>}
          <DialogFooter><Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button><Button type="submit" disabled={submitting}>{submitting ? "Registering…" : "Register vehicle"}</Button></DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  )
}

function MetricCard({ label, value, detail, icon }: { label: string; value: number; detail: string; icon: ReactNode }) { return <Card><CardContent className="p-5"><div className="mb-4 flex size-9 items-center justify-center rounded-lg bg-primary/10 text-primary">{icon}</div><p className="text-sm text-muted-foreground">{label}</p><p className="mt-2 text-3xl font-semibold tracking-tight">{value}</p><p className="mt-1 text-xs text-muted-foreground">{detail}</p></CardContent></Card> }
function Panel({ title, description, link, children }: { title: string; description: string; link: string; children: ReactNode }) { return <Card><CardHeader className="flex flex-row items-start justify-between gap-4"><div><CardTitle>{title}</CardTitle><CardDescription>{description}</CardDescription></div><Button variant="ghost" size="sm" render={<NavLink to={link} />}>View all</Button></CardHeader><CardContent>{children}</CardContent></Card> }
function GeofenceList({ items, selectedId, onSelect, onDelete }: { items: Geofence[]; selectedId?: string | null; onSelect?: (id: string) => void; onDelete?: (geofence: Geofence) => void }) { return items.length === 0 ? <EmptyState text="No geofences configured yet." /> : <div className="grid gap-2 overflow-hidden">{items.map((item) => { const selected = item.id === selectedId; return <div className={`grid min-w-0 grid-cols-[minmax(0,1fr)_auto] items-center gap-2 rounded-lg border p-2 transition-colors ${selected ? "border-primary bg-primary/10" : "border-transparent bg-muted/50"}`} key={item.id}><Button type="button" variant="ghost" className="w-full min-w-0 justify-start gap-3 overflow-hidden px-2 text-left" aria-pressed={selected} onClick={() => onSelect?.(item.id)}><span className={`flex size-8 shrink-0 items-center justify-center rounded-md ${selected ? "bg-primary text-primary-foreground" : "bg-primary/15 text-primary"}`}><MapPinned /></span><span className="min-w-0 flex-1 overflow-hidden"><span className="block truncate text-sm font-medium" title={item.name}>{item.name}</span><span className="block truncate text-xs text-muted-foreground" title={item.description || "No description"}>{item.description || "No description"}</span></span></Button><div className="flex shrink-0 items-center gap-1"><Badge variant="secondary" className="hidden md:inline-flex">{item.category.replaceAll("_", " ")}</Badge><Button type="button" variant="ghost" size="icon" aria-label={`Remove ${item.name}`} onClick={() => onDelete?.(item)}><Trash2 data-icon="inline-start" /></Button></div></div> })}</div> }

function DeleteGeofenceDialog({ target, onOpenChange, onDeleted }: { target: Geofence | null; onOpenChange: (open: boolean) => void; onDeleted: () => void }) {
  const [error, setError] = useState<string | null>(null)
  const [deleting, setDeleting] = useState(false)

  const remove = async () => {
    if (!target) return
    setDeleting(true)
    setError(null)
    try {
      await apiClient.deleteGeofence(target.id)
      onDeleted()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Unable to remove geofence")
    } finally {
      setDeleting(false)
    }
  }

  return <Dialog open={Boolean(target)} onOpenChange={onOpenChange}><DialogContent className="sm:max-w-md"><DialogHeader><DialogTitle>Remove geofence?</DialogTitle><DialogDescription>This removes {target?.name ?? "this boundary"} and its related alert configurations and violation history.</DialogDescription></DialogHeader>{error && <p className="text-sm text-destructive" role="alert">{error}</p>}<DialogFooter><Button type="button" variant="outline" onClick={() => onOpenChange(false)}>Cancel</Button><Button type="button" variant="destructive" onClick={() => void remove()} disabled={deleting}>{deleting ? "Removing…" : "Remove geofence"}</Button></DialogFooter></DialogContent></Dialog>
}
function VehicleList({ items }: { items: Vehicle[] }) { return items.length === 0 ? <EmptyState text="No vehicles registered yet." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary"><Truck size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.vehicle_number}</p><p className="truncate text-xs text-muted-foreground">{item.driver_name} · {item.vehicle_type}</p></div></div><Badge>{item.status}</Badge></div>)}</div> }
function AlertList({ items }: { items: AlertConfig[] }) { return items.length === 0 ? <EmptyState text="No alert rules configured." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.alert_id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-primary/15 text-primary"><BellRing size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.geofence_name}</p><p className="truncate text-xs text-muted-foreground">{item.vehicle_number || "All vehicles"} · {item.event_type}</p></div></div><Badge>{item.status}</Badge></div>)}</div> }
function ViolationList({ items }: { items: Violation[] }) { return items.length === 0 ? <EmptyState text="No violation events recorded." /> : <div className="grid gap-2">{items.map((item) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={item.id}><div className="flex min-w-0 items-center gap-3"><span className="flex size-8 shrink-0 items-center justify-center rounded-md bg-destructive/15 text-destructive"><AlertTriangle size={16} /></span><div className="min-w-0"><p className="truncate text-sm font-medium">{item.geofence_name}</p><p className="truncate text-xs text-muted-foreground">{item.vehicle_number} · {formatDate(item.timestamp)}</p></div></div><Badge variant="destructive">{item.event_type}</Badge></div>)}</div> }
function EmptyState({ text }: { text: string }) { return <p className="py-8 text-center text-sm text-muted-foreground">{text}</p> }
function formatDate(value: string) { return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(new Date(value)) }

export default App
