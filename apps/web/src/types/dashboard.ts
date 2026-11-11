import type { AlertConfig, Geofence, Vehicle, Violation } from "./api"

export type DashboardData = {
  geofences: Geofence[]
  vehicles: Vehicle[]
  alerts: AlertConfig[]
  violations: Violation[]
  totalViolations: number
  apiStatus: string
}