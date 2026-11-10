import { MapContainer, Marker, Polygon, TileLayer, useMap, useMapEvents } from "react-leaflet"
import { divIcon } from "leaflet"
import type { LatLngExpression } from "leaflet"
import type { Geofence } from "../lib/api"
import { useEffect } from "react"
import "leaflet/dist/leaflet.css"

const defaultCenter: LatLngExpression = [20, 0]
const draftPointIcon = divIcon({
  className: "border-0 bg-transparent",
  html: '<span style="display:block;width:14px;height:14px;border:3px solid white;border-radius:9999px;background:var(--primary);box-shadow:0 1px 5px rgba(0,0,0,.45);"></span>',
  iconSize: [14, 14],
  iconAnchor: [7, 7],
})
const vehicleIcon = divIcon({
  className: "border-0 bg-transparent",
  html: '<span style="display:flex;width:30px;height:30px;align-items:center;justify-content:center;border:3px solid white;border-radius:9999px;background:#2563eb;color:white;box-shadow:0 2px 8px rgba(0,0,0,.45);font-size:16px;">●</span>',
  iconSize: [30, 30],
  iconAnchor: [15, 15],
})

function MapClickCapture({ onAddPoint }: { onAddPoint?: (point: [number, number]) => void }) {
  useMapEvents({
    click: (event) => onAddPoint?.([event.latlng.lat, event.latlng.lng]),
  })
  return null
}

function MapSelection({ geofence, vehiclePosition }: { geofence?: Geofence; vehiclePosition?: [number, number] }) {
  const map = useMap()

  useEffect(() => {
    if (!geofence || geofence.coordinates.length < 2) return
    const points = vehiclePosition ? [...geofence.coordinates, vehiclePosition] : geofence.coordinates
    map.fitBounds(points as [number, number][], { padding: [32, 32], maxZoom: 10 })
  }, [geofence, map, vehiclePosition])

  return null
}

function OperationsMap({
  geofences,
  selectedGeofenceId,
  draftCoordinates = [],
  onAddPoint,
  onMovePoint,
  vehiclePosition,
  onMoveVehicle,
}: {
  geofences: Geofence[]
  selectedGeofenceId?: string | null
  draftCoordinates?: number[][]
  onAddPoint?: (point: [number, number]) => void
  onMovePoint?: (index: number, point: [number, number]) => void
  vehiclePosition?: [number, number]
  onMoveVehicle?: (point: [number, number]) => void
}) {
  return (
    <MapContainer center={defaultCenter} zoom={2} scrollWheelZoom doubleClickZoom={false} className="relative z-0 isolate h-full min-h-96 w-full">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <MapClickCapture onAddPoint={onAddPoint} />
      <MapSelection geofence={geofences.find((geofence) => geofence.id === selectedGeofenceId)} vehiclePosition={vehiclePosition} />
      {geofences.map((geofence) => (
        <Polygon
          key={geofence.id}
          positions={geofence.coordinates as LatLngExpression[]}
          pathOptions={geofence.id === selectedGeofenceId
            ? { color: "var(--ring)", fillColor: "var(--primary)", fillOpacity: 0.34, weight: 4 }
            : { color: "var(--primary)", fillColor: "var(--primary)", fillOpacity: 0.18 }}
        />
      ))}
      {draftCoordinates.length > 2 && (
        <Polygon
          positions={draftCoordinates as LatLngExpression[]}
          pathOptions={{ color: "var(--destructive)", fillColor: "var(--destructive)", fillOpacity: 0.16, dashArray: "6 6" }}
        />
      )}
      {draftCoordinates.map((point, index) => (
        <Marker
          key={`${index}-${point[0]}-${point[1]}`}
          position={point as LatLngExpression}
          icon={draftPointIcon}
          draggable
          eventHandlers={{
            dragend: (event) => {
              const position = event.target.getLatLng()
              onMovePoint?.(index, [position.lat, position.lng])
            },
          }}
        />
      ))}
      {vehiclePosition && (
        <Marker
          position={vehiclePosition}
          icon={vehicleIcon}
          draggable
          eventHandlers={{
            dragend: (event) => {
              const position = event.target.getLatLng()
              onMoveVehicle?.([position.lat, position.lng])
            },
          }}
        />
      )}
    </MapContainer>
  )
}

export { OperationsMap }
