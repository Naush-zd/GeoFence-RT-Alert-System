import { useEffect, useState } from "react"
import type { Vehicle } from "../lib/api"
import { api as apiClient } from "../lib/api"
import { Badge } from "../components/ui/badge"
import { Button } from "../components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "../components/ui/dialog"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"
import { ListCard, PageSection, VehicleList } from "../components/dashboard-widgets"
import type { CurrentGeofence } from "../lib/api"

export function VehiclesPage({ vehicles, onCreated }: { vehicles: Vehicle[]; onCreated: () => void }) {
  const [locations, setLocations] = useState<Record<string, { latitude: number; longitude: number; geofences: CurrentGeofence[] }>>({})
  useEffect(() => {
    let active = true
    void Promise.all(vehicles.map(async (vehicle) => {
      try {
        const location = await apiClient.vehicleLocation(vehicle.id)
        return [vehicle.id, { latitude: location.current_location.latitude, longitude: location.current_location.longitude, geofences: location.current_geofences }] as const
      } catch {
        return null
      }
    })).then((entries) => {
      if (active) setLocations(Object.fromEntries(entries.filter((entry): entry is NonNullable<typeof entry> => entry !== null)))
    })
    return () => { active = false }
  }, [vehicles])
  return <PageSection title="Vehicle registry" count={`${vehicles.length} vehicles`} action="+ Register vehicle" actionContent={<RegisterVehicleDialog onCreated={onCreated} />}><div className="grid gap-4"><ListCard><VehicleList items={vehicles} /></ListCard><div className="grid gap-2">{vehicles.map((vehicle) => { const location = locations[vehicle.id]; return <div className="rounded-lg border bg-muted/30 p-3" key={vehicle.id}><div className="flex items-center justify-between gap-3"><p className="font-medium">{vehicle.vehicle_number}</p><Badge variant="outline">{location ? "Location available" : "No location"}</Badge></div><p className="mt-1 text-xs text-muted-foreground">{location ? `${location.latitude.toFixed(5)}, ${location.longitude.toFixed(5)}` : "Update this vehicle location to see its current position."}</p><p className="mt-1 text-xs text-muted-foreground">{location?.geofences.length ? `Inside: ${location.geofences.map((geofence) => geofence.geofence_name).join(", ")}` : "Outside all configured geofences"}</p></div> })}</div></div></PageSection>
}

function RegisterVehicleDialog({ onCreated }: { onCreated: () => void }) {
  const [open, setOpen] = useState(false); const [vehicleNumber, setVehicleNumber] = useState(""); const [driverName, setDriverName] = useState(""); const [vehicleType, setVehicleType] = useState("truck"); const [phone, setPhone] = useState(""); const [error, setError] = useState<string | null>(null); const [submitting, setSubmitting] = useState(false)
  const submit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault(); setSubmitting(true); setError(null)
    try { await apiClient.createVehicle({ vehicle_number: vehicleNumber, driver_name: driverName, vehicle_type: vehicleType, phone }); onCreated(); setOpen(false); setVehicleNumber(""); setDriverName(""); setVehicleType("truck"); setPhone("") }
    catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to register vehicle") }
    finally { setSubmitting(false) }
  }
  return <Dialog open={open} onOpenChange={setOpen}><DialogTrigger render={<Button />}>+ Register vehicle</DialogTrigger><DialogContent className="sm:max-w-lg"><DialogHeader><DialogTitle>Register vehicle</DialogTitle><DialogDescription>Add a vehicle to the tracked fleet.</DialogDescription></DialogHeader><form className="flex flex-col gap-4" onSubmit={(event) => void submit(event)}><div className="flex flex-col gap-2"><Label htmlFor="vehicle-number">Vehicle number</Label><Input id="vehicle-number" value={vehicleNumber} onChange={(event) => setVehicleNumber(event.target.value)} required placeholder="KA-01-AB-1234" /></div><div className="flex flex-col gap-2"><Label htmlFor="driver-name">Driver name</Label><Input id="driver-name" value={driverName} onChange={(event) => setDriverName(event.target.value)} required placeholder="John Doe" /></div><div className="flex flex-col gap-2"><Label>Vehicle type</Label><Select value={vehicleType} onValueChange={(value) => setVehicleType(value ?? "truck")}><SelectTrigger className="w-full"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="truck">Truck</SelectItem><SelectItem value="van">Van</SelectItem><SelectItem value="car">Car</SelectItem><SelectItem value="motorcycle">Motorcycle</SelectItem></SelectContent></Select></div><div className="flex flex-col gap-2"><Label htmlFor="driver-phone">Driver phone</Label><Input id="driver-phone" value={phone} onChange={(event) => setPhone(event.target.value)} required placeholder="+1234567890" /></div>{error && <p className="text-sm text-destructive" role="alert">{error}</p>}<DialogFooter><Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button><Button type="submit" disabled={submitting}>{submitting ? "Registering…" : "Register vehicle"}</Button></DialogFooter></form></DialogContent></Dialog>
}
