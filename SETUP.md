# Local setup

## Prerequisites

- Node.js 26+
- npm 11+
- Go 1.27+
- Docker Desktop (optional, for PostgreSQL/PostGIS)

## Install

From the repository root:

```sh
npm install
go mod download -C apps/api
```

## Start the applications

Run both the Go API and React frontend:

```sh
npm run dev
```

The frontend is available at `http://localhost:5173` and the API at
`http://localhost:8080`. Verify the API with:

```sh
curl http://localhost:8080/healthz
```

To run either app separately:

```sh
npm run dev:web
npm run dev:api
```

## Validate

```sh
npm run build
npm run test
```

## Database

If Docker Desktop is installed, start the PostGIS service with:

```sh
docker compose up -d postgres
```

The configured local database is `geofencing` on port `5432` with user and
password `postgres`.

The API applies the initial `geofences` and `vehicles` tables automatically on
startup. The first implemented endpoints are:

```text
GET  /healthz
POST /geofences
GET  /geofences?category=delivery_zone
POST /vehicles
GET  /vehicles
POST /vehicles/location
GET  /vehicles/location/:vehicle_id
POST /alerts/configure
GET  /alerts
GET  /violations/history
GET  /ws/alerts
```

Example request:

```sh
curl -X POST http://localhost:8080/geofences \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Downtown Delivery Zone",
    "description": "Main delivery area",
    "coordinates": [
      [37.7749, -122.4194],
      [37.7849, -122.4194],
      [37.7849, -122.4094],
      [37.7749, -122.4094],
      [37.7749, -122.4194]
    ],
    "category": "delivery_zone"
  }'
```

## Repository layout

```text
apps/
  api/   Go HTTP API
  web/   React + Vite frontend
docker-compose.yml
package.json
```
