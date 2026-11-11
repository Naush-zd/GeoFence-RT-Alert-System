export type HealthResponse = ApiResponse & {
  status: string
}

export type GeofencesResponse = ApiResponse & {
  geofences: Geofence[]
}

export type CreateGeofenceRequest = {
  name: string
  description: string
  category: string
  coordinates: number[][]
}

export type CreateGeofenceResponse = ApiResponse & {
  id: string
  name: string
  status: string
}

export type DeleteGeofenceResponse = ApiResponse & {
  status: string
  geofence_id: string
}

export type VehiclesResponse = ApiResponse & {
  vehicles: Vehicle[]
}

export type VehicleLocationResponse = VehicleLocation & ApiResponse

export type CreateVehicleRequest = {
  vehicle_number: string
  driver_name: string
  vehicle_type: string
  phone: string
}

export type CreateVehicleResponse = ApiResponse & {
  id: string
  vehicle_number: string
  status: string
}

export type UpdateVehicleLocationRequest = {
  vehicle_id: string
  latitude: number
  longitude: number
  timestamp: string
}

export type UpdateVehicleLocationResponse = ApiResponse & {
  vehicle_id: string
  location_updated: boolean
  current_geofences: CurrentGeofence[]
}

export type ConfigureAlertRequest = {
  geofence_id: string
  vehicle_id: string
  event_type: "entry" | "exit" | "both"
}

export type ConfigureAlertResponse = ApiResponse & {
  alert_id: string
  geofence_id: string
  vehicle_id: string
  event_type: string
  status: string
}

export type AlertsResponse = ApiResponse & {
  alerts: AlertConfig[]
}

export type ViolationsResponse = ApiResponse & {
  violations: Violation[]
  total_count: number
}
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

export type ApiResponse = {
  time_ns: string
}

export type RequestOptions = {
  method?: "GET" | "POST" | "DELETE"
  body?: unknown
}