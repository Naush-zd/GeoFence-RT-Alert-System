# Local setup

## Prerequisites

- Node.js 26+
- npm 11+
- Go 1.27+
- Docker Desktop (optional, for PostgreSQL/PostGIS)

## Install

From the repository root:

```powershell
npm install
cd apps/api
go mod download
cd ../..
```

## Start the applications

Run both the Go API and React frontend:

```powershell
npm run dev
```

The frontend is available at `http://localhost:5173` and the API at
`http://localhost:8080`. Verify the API with:

```powershell
Invoke-WebRequest http://localhost:8080/healthz
```

To run either app separately:

```powershell
npm run dev:web
npm run dev:api
```

## Validate

```powershell
npm run build
npm run test
```

## Database

If Docker Desktop is installed, start the PostGIS service with:

```powershell
docker compose up -d postgres
```

The configured local database is `geofencing` on port `5432` with user and
password `postgres`.

## Repository layout

```text
apps/
  api/   Go HTTP API
  web/   React + Vite frontend
docker-compose.yml
package.json
```
