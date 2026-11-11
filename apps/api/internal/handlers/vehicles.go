package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

func (h Handler) CreateVehicle(c *gin.Context) {
	started := time.Now()
	var input models.VehicleRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	vehicle, err := h.Vehicles.Create(c.Request.Context(), input)
	if err != nil {
		if strings.Contains(err.Error(), "vehicles_vehicle_number_key") {
			writeError(c, http.StatusConflict, "vehicle_number already exists", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "create vehicle: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, gin.H{
		"id":             vehicle.ID,
		"vehicle_number": vehicle.VehicleNumber,
		"status":         "active",
	}, started)
}
func (h Handler) ListVehicles(c *gin.Context) {
	started := time.Now()
	vehicles, err := h.Vehicles.List(c.Request.Context())
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list vehicles: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusOK, gin.H{"vehicles": vehicles}, started)
}
