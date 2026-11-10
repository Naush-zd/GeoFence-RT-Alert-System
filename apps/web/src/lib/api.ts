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

type ApiResponse = {
  time_ns: string
}

async function request<T extends ApiResponse>(path: string): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`)
  const payload = (await response.json()) as T & { error?: string }
  if (!response.ok) {
    throw new Error(payload.error ?? `Request failed with status ${response.status}`)
  }
  return payload
}

export const api = {
  health: () => request<{ status: string; time_ns: string }>("/healthz"),
  geofences: () => request<{ geofences: Geofence[]; time_ns: string }>("/geofences"),
  vehicles: () => request<{ vehicles: Vehicle[]; time_ns: string }>("/vehicles"),
  alerts: () => request<{ alerts: AlertConfig[]; time_ns: string }>("/alerts"),
  violations: () =>
    request<{ violations: Violation[]; total_count: number; time_ns: string }>(
      "/violations/history?limit=5",
    ),
}
