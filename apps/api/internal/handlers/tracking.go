package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"gorm.io/gorm"
)

func (h Handler) UpdateVehicleLocation(c *gin.Context) {
	started := time.Now()
	var input models.LocationRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}

	result, err := h.Tracking.UpdateLocation(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "vehicle_id, valid coordinates") {
			status = http.StatusBadRequest
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error(), started)
		return
	}

	for _, transition := range result.Transitions {
		h.broadcastAlert(
			c,
			uuid.MustParse(result.Vehicle.ID),
			result.Vehicle.VehicleNumber,
			result.Vehicle.DriverName,
			transition.Geofence,
			transition.EventType,
			input,
		)
	}

	writeJSON(c, http.StatusOK, map[string]any{
		"vehicle_id":        input.VehicleID,
		"location_updated":  true,
		"current_geofences": result.CurrentGeofences,
	}, started)
}

func (h Handler) GetVehicleLocation(c *gin.Context) {
	started := time.Now()
	vehicleID, err := uuid.Parse(c.Param("vehicle_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid vehicle_id", started)
		return
	}

	vehicle, location, currentGeofences, err := h.Tracking.CurrentLocation(c.Request.Context(), vehicleID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error(), started)
		return
	}

	writeJSON(c, http.StatusOK, map[string]any{
		"vehicle_id":        vehicle.ID,
		"vehicle_number":    vehicle.VehicleNumber,
		"current_location":  location,
		"current_geofences": currentGeofences,
	}, started)
}
