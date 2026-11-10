import type { AlertConfig, Geofence, Vehicle, Violation } from "./api"

export type DashboardData = {
  geofences: Geofence[]
  vehicles: Vehicle[]
  alerts: AlertConfig[]
  violations: Violation[]
  totalViolations: number
  apiStatus: string
}

export const initialDashboardData: DashboardData = {
  geofences: [],
  vehicles: [],
  alerts: [],
  violations: [],
  totalViolations: 0,
  apiStatus: "checking",
}

export const pageDescriptions: Record<string, string> = {
  "/": "A live snapshot of your geofencing operations.",
  "/geofences": "Define and monitor the boundaries that matter to your fleet.",
  "/vehicles": "Keep track of every registered vehicle in your network.",
  "/alerts": "Configure the events that need your attention.",
  "/violations": "Review historical entries and exits across your zones.",
}
