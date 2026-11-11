import { useState } from "react"
import { Download } from "lucide-react"
import type { AlertConfig, AlertEvent, Geofence, Vehicle, Violation, ViolationFilters } from "../types/api"
import { api as apiClient } from "../lib/api"
import { AlertList, ListCard, PageSection, ViolationList } from "../components/dashboard-widgets"
import { Button, buttonVariants } from "../components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "../components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"

export function AlertsPage({ alerts, events, geofences, vehicles, onCreated }: { alerts: AlertConfig[]; events: AlertEvent[]; geofences: Geofence[]; vehicles: Vehicle[]; onCreated: () => void }) {
  return <div className="grid gap-5"><PageSection title="Configured rules" count={`${alerts.length} ${alerts.length === 1 ? "rule" : "rules"}`} action="+ Configure alert" actionContent={<ConfigureAlertDialog geofences={geofences} vehicles={vehicles} onCreated={onCreated} />}><ListCard><AlertList items={alerts} /></ListCard></PageSection><section className="grid gap-3"><div><h2 className="text-lg font-semibold">Recent live events</h2><p className="text-sm text-muted-foreground">{events.length} {events.length === 1 ? "event" : "events"} this session</p></div><ListCard><RecentAlertEvents events={events} /></ListCard></section></div>
}

function RecentAlertEvents({ events }: { events: AlertEvent[] }) {
  if (events.length === 0) return <p className="py-5 text-center text-sm text-muted-foreground">Waiting for alert events.</p>
  return <div className="grid gap-2">{events.map((event) => <div className="flex items-center justify-between gap-3 rounded-lg bg-muted/50 p-3" key={event.event_id}><div className="min-w-0"><p className="truncate text-sm font-medium">{event.vehicle.vehicle_number} {event.event_type === "entry" ? "entered" : "exited"} {event.geofence.geofence_name}</p><p className="truncate text-xs text-muted-foreground">{new Date(event.timestamp).toLocaleString()} · {event.location.latitude.toFixed(4)}, {event.location.longitude.toFixed(4)}</p></div></div>)}</div>
}

function ConfigureAlertDialog({ geofences, vehicles, onCreated }: { geofences: Geofence[]; vehicles: Vehicle[]; onCreated: () => void }) {
  const [open, setOpen] = useState(false); const [geofenceId, setGeofenceId] = useState(""); const [vehicleId, setVehicleId] = useState("all"); const [eventType, setEventType] = useState<"entry" | "exit" | "both">("both"); const [error, setError] = useState<string | null>(null); const [saving, setSaving] = useState(false)
  const submit = async () => {
    if (!geofenceId) { setError("Select a geofence."); return }
    setSaving(true); setError(null)
    try { await apiClient.configureAlert({ geofence_id: geofenceId, vehicle_id: vehicleId === "all" ? "" : vehicleId, event_type: eventType }); onCreated(); setOpen(false) }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to configure alert") }
    finally { setSaving(false) }
  }
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger render={<Button className="w-full sm:w-auto" />}>+ Configure alert</DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Configure alert rule</DialogTitle>
          <DialogDescription>Choose the boundary, vehicle scope, and transition events to monitor.</DialogDescription>
        </DialogHeader>
        <div className="grid gap-4">
          <Select value={geofenceId} onValueChange={(value) => setGeofenceId(value ?? "")}>
            <SelectTrigger className="w-full"><SelectValue>{geofences.find((geofence) => geofence.id === geofenceId)?.name ?? "Select geofence"}</SelectValue></SelectTrigger>
            <SelectContent>{geofences.map((geofence) => <SelectItem value={geofence.id} key={geofence.id}>{geofence.name}</SelectItem>)}</SelectContent>
          </Select>
          <Select value={vehicleId} onValueChange={(value) => setVehicleId(value ?? "all")}>
            <SelectTrigger className="w-full"><SelectValue>{vehicleId === "all" ? "All vehicles" : vehicles.find((vehicle) => vehicle.id === vehicleId)?.vehicle_number ?? "Select vehicle"}</SelectValue></SelectTrigger>
            <SelectContent><SelectItem value="all">All vehicles</SelectItem>{vehicles.map((vehicle) => <SelectItem value={vehicle.id} key={vehicle.id}>{vehicle.vehicle_number}</SelectItem>)}</SelectContent>
          </Select>
          <Select value={eventType} onValueChange={(value) => setEventType((value ?? "both") as "entry" | "exit" | "both")}>
            <SelectTrigger className="w-full"><SelectValue>{eventType === "both" ? "Entry and exit" : eventType === "entry" ? "Entry" : "Exit"}</SelectValue></SelectTrigger>
            <SelectContent><SelectItem value="entry">Entry</SelectItem><SelectItem value="exit">Exit</SelectItem><SelectItem value="both">Entry and exit</SelectItem></SelectContent>
          </Select>
          {error && <p className="text-sm text-destructive" role="alert">{error}</p>}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
          <Button onClick={() => void submit()} disabled={saving}>{saving ? "Saving…" : "Save alert rule"}</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export function ViolationsPage({ violations, total, vehicles, geofences, onFilter }: { violations: Violation[]; total: number; vehicles: Vehicle[]; geofences: Geofence[]; onFilter: (filters: ViolationFilters) => Promise<void> }) {
  const [vehicleId, setVehicleId] = useState(""); const [geofenceId, setGeofenceId] = useState(""); const [startDate, setStartDate] = useState(""); const [endDate, setEndDate] = useState("")
  const rows = [["Vehicle", "Geofence", "Event", "Timestamp", "Latitude", "Longitude"], ...violations.map((item) => [item.vehicle_number, item.geofence_name, item.event_type, item.timestamp, String(item.latitude), String(item.longitude)])]
  const csv = rows.map((row) => row.map((value) => `"${value.replaceAll('"', '""')}"`).join(",")).join("\r\n")
  const csvHref = `data:text/csv;charset=utf-8,${encodeURIComponent(csv)}`
  const applyFilters = () => onFilter({ vehicle_id: vehicleId, geofence_id: geofenceId, start_date: startDate ? `${startDate}T00:00:00Z` : undefined, end_date: endDate ? `${endDate}T23:59:59Z` : undefined, limit: 50 })
  return <PageSection title="Violation history" count={`${total} ${total === 1 ? "event" : "events"}`} action="Export visible CSV" actionContent={violations.length > 0 ? <a className={buttonVariants({ variant: "outline", className: "w-full sm:w-auto" })} href={csvHref} download="violations.csv"><Download data-icon="inline-start" />Export visible CSV</a> : <Button variant="outline" className="w-full sm:w-auto" disabled><Download data-icon="inline-start" />Export visible CSV</Button>}><div className="grid gap-4"><div className="grid grid-cols-1 gap-2 rounded-lg border p-3 sm:grid-cols-2 xl:grid-cols-4"><Select value={vehicleId} onValueChange={(value) => setVehicleId(value ?? "")}><SelectTrigger className="w-full"><SelectValue>{vehicles.find((vehicle) => vehicle.id === vehicleId)?.vehicle_number ?? "All vehicles"}</SelectValue></SelectTrigger><SelectContent><SelectItem value="">All vehicles</SelectItem>{vehicles.map((vehicle) => <SelectItem value={vehicle.id} key={vehicle.id}>{vehicle.vehicle_number}</SelectItem>)}</SelectContent></Select><Select value={geofenceId} onValueChange={(value) => setGeofenceId(value ?? "")}><SelectTrigger className="w-full"><SelectValue>{geofences.find((geofence) => geofence.id === geofenceId)?.name ?? "All geofences"}</SelectValue></SelectTrigger><SelectContent><SelectItem value="">All geofences</SelectItem>{geofences.map((geofence) => <SelectItem value={geofence.id} key={geofence.id}>{geofence.name}</SelectItem>)}</SelectContent></Select><input className="h-9 min-w-0 rounded-md border bg-background px-3 text-sm" type="date" value={startDate} onChange={(event) => setStartDate(event.target.value)} aria-label="Start date" /><input className="h-9 min-w-0 rounded-md border bg-background px-3 text-sm" type="date" value={endDate} onChange={(event) => setEndDate(event.target.value)} aria-label="End date" /><Button className="w-full sm:col-span-2 xl:col-span-4" onClick={applyFilters}>Apply filters</Button></div><ListCard><ViolationList items={violations} /></ListCard></div></PageSection>
}
