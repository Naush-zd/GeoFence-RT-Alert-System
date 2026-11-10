const apiBaseUrl = import.meta.env.VITE_API_URL ?? "/api"

export type Geofence = {
  id: string
  name: string
  description: string
  category: string
  coordinates: number[][]
  created_at: string
}

export type Vehicle = {
  id: string
  vehicle_number: string
  driver_name: string
  vehicle_type: string
  phone: string
  status: string
  created_at: string
}

export type AlertConfig = {
  alert_id: string
  geofence_id: string
  geofence_name: string
  vehicle_id?: string
  vehicle_number?: string
  event_type: string
  status: string
  created_at: string
}

export type Violation = {
  id: string
  vehicle_id: string
  vehicle_number: string
  geofence_id: string
  geofence_name: string
  event_type: string
  latitude: number
  longitude: number
  timestamp: string
}

export type ViolationFilters = {
  vehicle_id?: string
  geofence_id?: string
  start_date?: string
  end_date?: string
  limit?: number
}

export type CurrentGeofence = {
  geofence_id: string
  geofence_name: string
  category?: string
  status?: string
}

export type VehicleLocation = {
  vehicle_id: string
  vehicle_number: string
  current_location: {
    latitude: number
    longitude: number
    timestamp: string
  }
  current_geofences: CurrentGeofence[]
}

export type AlertEvent = {
  event_id: string
  event_type: "entry" | "exit"
  timestamp: string
  vehicle: { vehicle_id: string; vehicle_number: string; driver_name: string }
  geofence: { geofence_id: string; geofence_name: string; category: string }
  location: { latitude: number; longitude: number }
}

type ApiResponse = {
  time_ns: string
}

type RequestOptions = {
  method?: "GET" | "POST" | "DELETE"
  body?: unknown
}

async function request<T extends ApiResponse>(path: string, options: RequestOptions = {}): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`, {
    method: options.method ?? "GET",
    headers: options.body ? { "Content-Type": "application/json" } : undefined,
    body: options.body ? JSON.stringify(options.body) : undefined,
  })
  const payload = (await response.json()) as T & { error?: string }
  if (!response.ok) {
    throw new Error(payload.error ?? `Request failed with status ${response.status}`)
  }
  return payload
}

export const api = {
  health: () => request<{ status: string; time_ns: string }>("/healthz"),
  geofences: () => request<{ geofences: Geofence[]; time_ns: string }>("/geofences"),
  createGeofence: (input: {
    name: string
    description: string
    category: string
    coordinates: number[][]
  }) => request<{ id: string; name: string; status: string; time_ns: string }>("/geofences", { method: "POST", body: input }),
  deleteGeofence: (geofenceId: string) => request<{ status: string; geofence_id: string; time_ns: string }>(`/geofences/${geofenceId}`, { method: "DELETE" }),
  vehicles: () => request<{ vehicles: Vehicle[]; time_ns: string }>("/vehicles"),
  vehicleLocation: (vehicleId: string) => request<VehicleLocation & { time_ns: string }>(`/vehicles/location/${vehicleId}`),
  createVehicle: (input: {
    vehicle_number: string
    driver_name: string
    vehicle_type: string
    phone: string
  }) => request<{ id: string; vehicle_number: string; status: string; time_ns: string }>("/vehicles", { method: "POST", body: input }),
  updateVehicleLocation: (input: {
    vehicle_id: string
    latitude: number
    longitude: number
    timestamp: string
  }) => request<{ vehicle_id: string; location_updated: boolean; current_geofences: CurrentGeofence[]; time_ns: string }>("/vehicles/location", { method: "POST", body: input }),
  configureAlert: (input: {
    geofence_id: string
    vehicle_id: string
    event_type: "entry" | "exit" | "both"
  }) => request<{ alert_id: string; geofence_id: string; vehicle_id: string; event_type: string; status: string; time_ns: string }>("/alerts/configure", { method: "POST", body: input }),
  alerts: () => request<{ alerts: AlertConfig[]; time_ns: string }>("/alerts"),
  violations: (filters: ViolationFilters = {}) =>
    request<{ violations: Violation[]; total_count: number; time_ns: string }>(
      `/violations/history?${new URLSearchParams(Object.entries({ ...filters, limit: filters.limit ?? 50 }).filter(([, value]) => value !== undefined && value !== "").map(([key, value]) => [key, String(value)]))}`,
    ),
}

export function alertsWebSocketUrl(): string {
  const configuredUrl = import.meta.env.VITE_API_URL as string | undefined
  if (configuredUrl?.startsWith("http")) {
    return configuredUrl.replace(/^http/, "ws") + "/ws/alerts"
  }
  return `${window.location.protocol === "https:" ? "wss:" : "ws:"}//${window.location.host}/api/ws/alerts`
}
