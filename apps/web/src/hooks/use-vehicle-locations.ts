import { useQuery } from "@tanstack/react-query"
import type { Vehicle, VehicleLocation } from "../lib/api"
import { api } from "../lib/api"

export function useVehicleLocations(vehicles: Vehicle[]) {
  const vehicleIds = vehicles.map((vehicle) => vehicle.id).sort()

  return useQuery({
    queryKey: ["vehicle-locations", vehicleIds],
    enabled: vehicleIds.length > 0,
    queryFn: async (): Promise<VehicleLocation[]> => {
      const results = await Promise.all(vehicleIds.map(async (vehicleId) => {
        try {
          return await api.vehicleLocation(vehicleId)
        } catch {
          return null
        }
      }))

      return results.filter((location): location is VehicleLocation & { time_ns: string } => location !== null)
    },
    refetchInterval: 5000,
  })
}