package routes

import (
	"net/http"

	"github.com/yourusername/geofencing-platform/apps/api/internal/handlers"
)

func New(api handlers.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", api.Health)
	mux.HandleFunc("POST /geofences", api.CreateGeofence)
	mux.HandleFunc("GET /geofences", api.ListGeofences)
	mux.HandleFunc("POST /vehicles", api.CreateVehicle)
	mux.HandleFunc("GET /vehicles", api.ListVehicles)
	return mux
}
