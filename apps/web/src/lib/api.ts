import axios from "axios"
import type {
  AlertsResponse,
  ApiResponse,
  ConfigureAlertRequest,
  ConfigureAlertResponse,
  CreateGeofenceRequest,
  CreateGeofenceResponse,
  CreateVehicleRequest,
  CreateVehicleResponse,
  DeleteGeofenceResponse,
  GeofencesResponse,
  HealthResponse,
  RequestOptions,
  UpdateVehicleLocationRequest,
  UpdateVehicleLocationResponse,
  VehicleLocationResponse,
  VehiclesResponse,
  ViolationFilters,
  ViolationsResponse,
} from "../types/api"

const configuredApiUrl = import.meta.env.VITE_API_URL as string | undefined
const apiBaseUrl = import.meta.env.PROD ? "/api" : configuredApiUrl ?? "/api"
const http = axios.create({ baseURL: apiBaseUrl })

async function request<T extends ApiResponse>(path: string, options: RequestOptions = {}): Promise<T> {
  try {
    const response = await http.request<T>({
      url: path,
      method: options.method ?? "GET",
      data: options.body,
      headers: options.body ? { "Content-Type": "application/json" } : undefined,
    })
    return response.data
  } catch (error) {
    if (axios.isAxiosError<{ error?: string }>(error)) {
      const message = error.response?.data?.error
      if (message) throw new Error(message)
      if (error.response) throw new Error(`Request failed with status ${error.response.status}`)
    }
    throw error
  }
}

function violationQuery(filters: ViolationFilters): string {
  const params = new URLSearchParams()
  const values = { ...filters, limit: filters.limit ?? 50 }
  for (const [key, value] of Object.entries(values)) {
    if (value !== undefined && value !== "") {
      params.set(key, String(value))
    }
  }
  return params.toString()
}

export const api = {
  health: () => request<HealthResponse>("/healthz"),

  geofences: () => request<GeofencesResponse>("/geofences"),

  createGeofence: (input: CreateGeofenceRequest) =>
    request<CreateGeofenceResponse>("/geofences", {
      method: "POST",
      body: input,
    }),

  deleteGeofence: (geofenceId: string) =>
    request<DeleteGeofenceResponse>(`/geofences/${geofenceId}`, {
      method: "DELETE",
    }),

  vehicles: () => request<VehiclesResponse>("/vehicles"),

  vehicleLocation: (vehicleId: string) => request<VehicleLocationResponse>(`/vehicles/location/${vehicleId}`),

  createVehicle: (input: CreateVehicleRequest) =>
    request<CreateVehicleResponse>("/vehicles", {
      method: "POST",
      body: input,
    }),

  updateVehicleLocation: (input: UpdateVehicleLocationRequest) =>
    request<UpdateVehicleLocationResponse>("/vehicles/location", {
      method: "POST",
      body: input,
    }),

  configureAlert: (input: ConfigureAlertRequest) =>
    request<ConfigureAlertResponse>("/alerts/configure", {
      method: "POST",
      body: input,
    }),

  alerts: () => request<AlertsResponse>("/alerts"),

  violations: (filters: ViolationFilters = {}) =>
    request<ViolationsResponse>(`/violations/history?${violationQuery(filters)}`),
}

export function alertsWebSocketUrl(): string {
  if (!import.meta.env.PROD && configuredApiUrl?.startsWith("http")) {
    return configuredApiUrl.replace(/^http/, "ws") + "/ws/alerts"
  }

  const websocketProtocol = window.location.protocol === "https:" ? "wss:" : "ws:"
  return `${websocketProtocol}//${window.location.host}${apiBaseUrl}/ws/alerts`
}