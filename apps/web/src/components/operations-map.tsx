import { MapContainer, Marker, Polygon, TileLayer, useMapEvents } from "react-leaflet"
import { divIcon } from "leaflet"
import type { LatLngExpression } from "leaflet"
import type { Geofence } from "../lib/api"
import "leaflet/dist/leaflet.css"

const defaultCenter: LatLngExpression = [20, 0]
const draftPointIcon = divIcon({
  className: "border-0 bg-transparent",
  html: '<span style="display:block;width:14px;height:14px;border:3px solid white;border-radius:9999px;background:var(--primary);box-shadow:0 1px 5px rgba(0,0,0,.45);"></span>',
  iconSize: [14, 14],
  iconAnchor: [7, 7],
})

function MapClickCapture({ onAddPoint }: { onAddPoint?: (point: [number, number]) => void }) {
  useMapEvents({
    click: (event) => onAddPoint?.([event.latlng.lat, event.latlng.lng]),
  })
  return null
}

function OperationsMap({
  geofences,
  draftCoordinates = [],
  onAddPoint,
  onMovePoint,
}: {
  geofences: Geofence[]
  draftCoordinates?: number[][]
  onAddPoint?: (point: [number, number]) => void
  onMovePoint?: (index: number, point: [number, number]) => void
}) {
  return (
    <MapContainer center={defaultCenter} zoom={2} scrollWheelZoom doubleClickZoom={false} className="relative z-0 isolate h-full min-h-96 w-full">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <MapClickCapture onAddPoint={onAddPoint} />
      {geofences.map((geofence) => (
        <Polygon
          key={geofence.id}
          positions={geofence.coordinates as LatLngExpression[]}
          pathOptions={{ color: "var(--primary)", fillColor: "var(--primary)", fillOpacity: 0.18 }}
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
    </MapContainer>
  )
}

export { OperationsMap }
