import { useState } from "react"
import { AlertTriangle, BellRing, MapPinned, Play, Truck } from "lucide-react"
import type { Geofence, Vehicle } from "../lib/api"
import { api as apiClient } from "../lib/api"
import type { DashboardData } from "../lib/dashboard"
import { useVehicleLocations } from "../hooks/use-vehicle-locations"
import { OperationsMap } from "../components/operations-map"
import { AlertList, GeofenceList, MetricCard, Panel, VehicleList, ViolationList } from "../components/dashboard-widgets"
import { Badge } from "../components/ui/badge"
import { Button } from "../components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "../components/ui/card"

export function OverviewPage({ data, onRefresh }: { data: DashboardData; onRefresh: () => Promise<void> }) {
  return <><section className="grid grid-cols-2 gap-3 py-5 sm:gap-4 xl:grid-cols-4"><MetricCard label="Active geofences" value={data.geofences.length} detail="Configured boundaries" icon={<MapPinned />} /><MetricCard label="Registered vehicles" value={data.vehicles.length} detail="Tracked fleet" icon={<Truck />} /><MetricCard label="Alert rules" value={data.alerts.length} detail="Active configurations" icon={<BellRing />} /><MetricCard label="Violation events" value={data.totalViolations} detail="Recorded history" icon={<AlertTriangle />} /></section><LiveOperationsMap geofences={data.geofences} vehicles={data.vehicles} /><SimulationCard onComplete={onRefresh} /><section className="grid gap-4 py-5 lg:grid-cols-2"><Panel title="Geofences" description="Configured monitoring boundaries" link="/geofences"><GeofenceList items={data.geofences.slice(0, 4)} /></Panel><Panel title="Fleet" description="Registered vehicles" link="/vehicles"><VehicleList items={data.vehicles.slice(0, 4)} /></Panel><Panel title="Recent violations" description={`${data.totalViolations} total recorded events`} link="/violations"><ViolationList items={data.violations.slice(0, 4)} /></Panel><Panel title="Alert rules" description="Rules watching for transitions" link="/alerts"><AlertList items={data.alerts.slice(0, 4)} /></Panel></section></>
}

function LiveOperationsMap({ geofences, vehicles }: { geofences: Geofence[]; vehicles: Vehicle[] }) {
  const locationsQuery = useVehicleLocations(vehicles)
  const locations = locationsQuery.data ?? []
  const vehicleSummary = vehicles.length === 0
    ? "No vehicles registered"
    : `${locations.length} of ${vehicles.length} vehicles located`

  return (
    <Card className="overflow-hidden">
      <CardHeader className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <CardTitle>Live fleet map</CardTitle>
          <CardDescription>Vehicle positions and configured boundaries · refreshes every 5 seconds</CardDescription>
        </div>
        <Badge variant="secondary" className="w-fit">{vehicleSummary}</Badge>
      </CardHeader>
      <CardContent className="pt-0">
        <div className="relative h-80 overflow-hidden rounded-md border sm:h-120">
          <OperationsMap geofences={geofences} vehicleLocations={locations} fitAll />
          {geofences.length === 0 && locations.length === 0 && (
            <div className="pointer-events-none absolute inset-0 z-400 grid place-items-center bg-background/30 p-4">
              <p className="rounded-md border bg-background/90 px-3 py-2 text-center text-sm text-muted-foreground">
                Add geofences or report vehicle locations to populate the map.
              </p>
            </div>
          )}
        </div>
        <div className="flex flex-col gap-1 pt-3 text-xs text-muted-foreground sm:flex-row sm:items-center sm:justify-between">
          <span>Select a vehicle marker for its latest location and boundary status.</span>
          <span>{locationsQuery.dataUpdatedAt ? `Last checked ${new Date(locationsQuery.dataUpdatedAt).toLocaleTimeString()}` : "Checking vehicle locations…"}</span>
        </div>
      </CardContent>
    </Card>
  )
}

function SimulationCard({ onComplete }: { onComplete: () => Promise<void> }) {
  const [running, setRunning] = useState(false)
  const [simulationGeofence, setSimulationGeofence] = useState<Geofence | null>(null)
  const [simulationVehicle, setSimulationVehicle] = useState<Vehicle | null>(null)
  const [vehiclePosition, setVehiclePosition] = useState<[number, number] | null>(null)
  const [message, setMessage] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const run = async () => {
    setRunning(true); setMessage(null); setError(null)
    try {
      const delhiCoordinates = [[28.48, 77.05], [28.48, 77.38], [28.78, 77.38], [28.78, 77.05], [28.48, 77.05]]
      const [currentGeofences, currentVehicles] = await Promise.all([apiClient.geofences(), apiClient.vehicles()])
      const existingGeofence = currentGeofences.geofences.find((item) => item.name === "Delhi delivery boundary")
      const geofence = existingGeofence ?? await createDemoGeofence(delhiCoordinates)
      const existingVehicle = currentVehicles.vehicles.find((item) => item.vehicle_number === "DEMO-001")
      const vehicle = existingVehicle ?? await createDemoVehicle()
      const currentAlerts = await apiClient.alerts()
      if (!currentAlerts.alerts.some((alert) => alert.geofence_id === geofence.id && alert.vehicle_id === vehicle.id && alert.event_type === "both")) {
        await apiClient.configureAlert({ geofence_id: geofence.id, vehicle_id: vehicle.id, event_type: "both" })
      }
      const startPosition: [number, number] = [29.35, 77.6]
      await apiClient.updateVehicleLocation({ vehicle_id: vehicle.id, latitude: startPosition[0], longitude: startPosition[1], timestamp: new Date().toISOString() })
      setSimulationGeofence(geofence); setSimulationVehicle(vehicle); setVehiclePosition(startPosition)
      await onComplete(); setMessage("Delhi simulation ready. Drag the blue marker into the shaded zone, then back out.")
    } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to run the simulation") }
    finally { setRunning(false) }
  }

  const moveVehicle = async (point: [number, number]) => {
    if (!simulationVehicle) return
    setVehiclePosition(point); setError(null)
    try {
      const result = await apiClient.updateVehicleLocation({ vehicle_id: simulationVehicle.id, latitude: point[0], longitude: point[1], timestamp: new Date().toISOString() })
      setMessage(`DEMO-001 ${result.current_geofences.length > 0 ? "entered" : "exited"} the Delhi delivery boundary.`)
      await onComplete()
    } catch (requestError) { setError(requestError instanceof Error ? requestError.message : "Unable to update the simulated vehicle") }
  }

  return <Card><CardHeader className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between"><div><CardTitle>Vehicle alert simulation</CardTitle><CardDescription>Set up a Delhi geofence, then drag the blue vehicle marker from Uttar Pradesh into Delhi and back out to generate entry and exit alerts.</CardDescription></div><Button onClick={() => void run()} disabled={running}><Play data-icon="inline-start" />{running ? "Preparing…" : simulationGeofence ? "Reset simulation" : "Set up Delhi simulation"}</Button></CardHeader>{simulationGeofence && vehiclePosition && <CardContent className="pt-0"><div className="h-112 overflow-hidden rounded-lg border"><OperationsMap geofences={[simulationGeofence]} selectedGeofenceId={simulationGeofence.id} vehiclePosition={vehiclePosition} onMoveVehicle={(point) => void moveVehicle(point)} /></div><p className="mt-3 text-xs text-muted-foreground">Drag the blue marker into the shaded Delhi boundary for entry, then drag it back to Uttar Pradesh for exit.</p></CardContent>}{message && <CardContent className="pt-0"><p className="text-sm text-primary" role="status">{message}</p></CardContent>}{error && <CardContent className="pt-0"><p className="text-sm text-destructive" role="alert">{error}</p></CardContent>}</Card>
}

async function createDemoGeofence(coordinates: number[][]): Promise<Geofence> {
  const result = await apiClient.createGeofence({ name: "Delhi delivery boundary", description: "Delhi boundary for the vehicle entry and exit walkthrough.", category: "delivery_zone", coordinates })
  return { id: result.id, name: result.name, description: "Delhi boundary for the vehicle entry and exit walkthrough.", category: "delivery_zone", coordinates, created_at: new Date().toISOString() }
}

async function createDemoVehicle(): Promise<Vehicle> {
  const result = await apiClient.createVehicle({ vehicle_number: "DEMO-001", driver_name: "Simulation Driver", vehicle_type: "van", phone: "+0000000000" })
  return { id: result.id, vehicle_number: result.vehicle_number, driver_name: "Simulation Driver", vehicle_type: "van", phone: "+0000000000", status: "active", created_at: new Date().toISOString() }
}
