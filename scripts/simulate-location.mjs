const apiUrl = (process.env.API_URL ?? "http://localhost:8080").replace(/\/+$/, "")
const intervalMs = Number(process.env.LOCATION_INTERVAL_MS ?? 1500)
const stepsPerLeg = 24
const route = [
  [28.45, 77.22],
  [28.82, 77.22],
  [28.82, 77.30],
  [28.45, 77.30],
]

if (!Number.isFinite(intervalMs) || intervalMs < 100) {
  console.error("LOCATION_INTERVAL_MS must be at least 100 milliseconds")
  process.exit(1)
}

async function request(path, options) {
  const response = await fetch(`${apiUrl}${path}`, options)
  const payload = await response.json()
  if (!response.ok) {
    throw new Error(payload.error ?? `Request failed with status ${response.status}`)
  }
  return payload
}

async function main() {
  const { vehicles } = await request("/vehicles")
  const vehicle = vehicles.find((item) => item.vehicle_number === "DEMO-001")
  if (!vehicle) {
    throw new Error("DEMO-001 was not found. Create it from the Overview simulation first.")
  }

  console.log(`Moving ${vehicle.vehicle_number} around the Delhi boundary every ${intervalMs} ms. Press Ctrl+C to stop.`)
  let segment = 0
  let step = 0

  while (true) {
    const start = route[segment]
    const end = route[(segment + 1) % route.length]
    const progress = step / stepsPerLeg
    const latitude = start[0] + (end[0] - start[0]) * progress
    const longitude = start[1] + (end[1] - start[1]) * progress
    const result = await request("/vehicles/location", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        vehicle_id: vehicle.id,
        latitude,
        longitude,
        timestamp: new Date().toISOString(),
      }),
    })

    const boundaries = result.current_geofences.map((geofence) => geofence.geofence_name)
    console.log(
      `${vehicle.vehicle_number} ${latitude.toFixed(5)}, ${longitude.toFixed(5)} | ${boundaries.length ? `Inside ${boundaries.join(", ")}` : "Outside configured boundaries"}`,
    )

    step += 1
    if (step > stepsPerLeg) {
      step = 0
      segment = (segment + 1) % route.length
    }
    await new Promise((resolve) => setTimeout(resolve, intervalMs))
  }
}

main().catch((error) => {
  console.error(error instanceof Error ? error.message : error)
  process.exitCode = 1
})
