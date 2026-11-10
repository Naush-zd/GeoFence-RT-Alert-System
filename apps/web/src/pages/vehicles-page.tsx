import { useState } from "react"
import type { Vehicle } from "../lib/api"
import { api as apiClient } from "../lib/api"
import { Button } from "../components/ui/button"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from "../components/ui/dialog"
import { Input } from "../components/ui/input"
import { Label } from "../components/ui/label"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "../components/ui/select"
import { ListCard, PageSection, VehicleList } from "../components/dashboard-widgets"

export function VehiclesPage({ vehicles, onCreated }: { vehicles: Vehicle[]; onCreated: () => void }) {
  return <PageSection title="Vehicle registry" count={`${vehicles.length} vehicles`} action="+ Register vehicle" actionContent={<RegisterVehicleDialog onCreated={onCreated} />}><ListCard><VehicleList items={vehicles} /></ListCard></PageSection>
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
