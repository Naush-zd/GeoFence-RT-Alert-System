import { useState } from "react"
import type { AlertConfig, Geofence, Vehicle, Violation } from "../lib/api"
import { api as apiClient } from "../lib/api"
import { AlertList, ListCard, PageSection, ViolationList } from "../components/dashboard-widgets"
import { Button } from "../components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "../components/ui/dialog"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"

export function AlertsPage({ alerts, geofences, vehicles, onCreated }: { alerts: AlertConfig[]; geofences: Geofence[]; vehicles: Vehicle[]; onCreated: () => void }) {
  return <PageSection title="Alert rules" count={`${alerts.length} rules`} action="+ Configure alert" actionContent={<ConfigureAlertDialog geofences={geofences} vehicles={vehicles} onCreated={onCreated} />}><ListCard><AlertList items={alerts} /></ListCard></PageSection>
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
  return <Dialog open={open} onOpenChange={setOpen}><DialogTrigger render={<Button />}>+ Configure alert</DialogTrigger><DialogContent><DialogHeader><DialogTitle>Configure alert rule</DialogTitle><DialogDescription>Choose the boundary, vehicle scope, and transition events to monitor.</DialogDescription></DialogHeader><div className="grid gap-4"><Select value={geofenceId} onValueChange={(value) => setGeofenceId(value ?? "")}><SelectTrigger><SelectValue placeholder="Select geofence" /></SelectTrigger><SelectContent>{geofences.map((geofence) => <SelectItem value={geofence.id} key={geofence.id}>{geofence.name}</SelectItem>)}</SelectContent></Select><Select value={vehicleId} onValueChange={(value) => setVehicleId(value ?? "all")}><SelectTrigger><SelectValue placeholder="Vehicle scope" /></SelectTrigger><SelectContent><SelectItem value="all">All vehicles</SelectItem>{vehicles.map((vehicle) => <SelectItem value={vehicle.id} key={vehicle.id}>{vehicle.vehicle_number}</SelectItem>)}</SelectContent></Select><Select value={eventType} onValueChange={(value) => setEventType((value ?? "both") as "entry" | "exit" | "both")}><SelectTrigger><SelectValue /></SelectTrigger><SelectContent><SelectItem value="entry">Entry</SelectItem><SelectItem value="exit">Exit</SelectItem><SelectItem value="both">Entry and exit</SelectItem></SelectContent></Select>{error && <p className="text-sm text-destructive" role="alert">{error}</p>}</div><DialogFooter><Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button><Button onClick={() => void submit()} disabled={saving}>{saving ? "Saving…" : "Save alert rule"}</Button></DialogFooter></DialogContent></Dialog>
}

export function ViolationsPage({ violations, total, vehicles, geofences, onFilter }: { violations: Violation[]; total: number; vehicles: Vehicle[]; geofences: Geofence[]; onFilter: (filters: { vehicle_id?: string; geofence_id?: string; start_date?: string; end_date?: string; limit?: number }) => Promise<void> }) {
  const [vehicleId, setVehicleId] = useState(""); const [geofenceId, setGeofenceId] = useState(""); const [startDate, setStartDate] = useState(""); const [endDate, setEndDate] = useState("")
  const applyFilters = () => onFilter({ vehicle_id: vehicleId, geofence_id: geofenceId, start_date: startDate ? `${startDate}T00:00:00Z` : undefined, end_date: endDate ? `${endDate}T23:59:59Z` : undefined, limit: 50 })
  return <PageSection title="Violation history" count={`${total} total events`} action="Export history"><div className="grid gap-4"><div className="grid gap-2 rounded-lg border p-3 md:grid-cols-4"><Select value={vehicleId} onValueChange={(value) => setVehicleId(value ?? "")}><SelectTrigger><SelectValue placeholder="All vehicles" /></SelectTrigger><SelectContent><SelectItem value="">All vehicles</SelectItem>{vehicles.map((vehicle) => <SelectItem value={vehicle.id} key={vehicle.id}>{vehicle.vehicle_number}</SelectItem>)}</SelectContent></Select><Select value={geofenceId} onValueChange={(value) => setGeofenceId(value ?? "")}><SelectTrigger><SelectValue placeholder="All geofences" /></SelectTrigger><SelectContent><SelectItem value="">All geofences</SelectItem>{geofences.map((geofence) => <SelectItem value={geofence.id} key={geofence.id}>{geofence.name}</SelectItem>)}</SelectContent></Select><input className="rounded-md border bg-background px-3 text-sm" type="date" value={startDate} onChange={(event) => setStartDate(event.target.value)} aria-label="Start date" /><input className="rounded-md border bg-background px-3 text-sm" type="date" value={endDate} onChange={(event) => setEndDate(event.target.value)} aria-label="End date" /><Button className="md:col-span-4" onClick={applyFilters}>Apply filters</Button></div><ListCard><ViolationList items={violations} /></ListCard></div></PageSection>
}
