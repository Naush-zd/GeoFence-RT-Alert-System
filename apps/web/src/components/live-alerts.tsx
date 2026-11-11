import { useEffect, useState } from "react"
import { AlertTriangle, X } from "lucide-react"
import type { AlertEvent } from "../lib/api"
import { alertsWebSocketUrl } from "../lib/api"
import { Button } from "./ui/button"

export function LiveAlerts({ onEvent }: { onEvent: (event: AlertEvent) => void }) {
  const [events, setEvents] = useState<AlertEvent[]>([])
  useEffect(() => {
    let active = true
    let socket: WebSocket | undefined
    let reconnectTimer: number | undefined
    const connect = () => {
      socket = new WebSocket(alertsWebSocketUrl())
      socket.onmessage = (message) => {
        try {
          const event = JSON.parse(message.data as string) as AlertEvent
          setEvents((current) => [event, ...current].slice(0, 3))
          onEvent(event)
        } catch {
          // Ignore malformed messages from the alert stream.
        }
      }
      socket.onerror = () => socket?.close()
      socket.onclose = () => {
        if (active) reconnectTimer = window.setTimeout(connect, 1000)
      }
    }
    connect()
    return () => {
      active = false
      if (reconnectTimer !== undefined) window.clearTimeout(reconnectTimer)
      socket?.close()
    }
  }, [onEvent])
  if (events.length === 0) return null
  return <div className="fixed right-4 top-4 z-1100 grid w-[min(24rem,calc(100vw-2rem))] gap-2">{events.map((event) => <div className="rounded-lg border border-destructive/30 bg-background p-3 shadow-lg" key={event.event_id}><div className="flex items-start gap-3"><AlertTriangle className="mt-0.5 shrink-0 text-destructive" size={18} /><div className="min-w-0 flex-1"><p className="text-sm font-medium">{event.vehicle.vehicle_number} {event.event_type === "entry" ? "entered" : "exited"} {event.geofence.geofence_name}</p><p className="text-xs text-muted-foreground">{new Date(event.timestamp).toLocaleString()} · {event.location.latitude.toFixed(4)}, {event.location.longitude.toFixed(4)}</p></div><Button variant="ghost" size="icon-sm" aria-label="Dismiss alert" onClick={() => setEvents((current) => current.filter((item) => item.event_id !== event.event_id))}><X /></Button></div></div>)}</div>
}
