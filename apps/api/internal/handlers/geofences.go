package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/services"
)

func (h Handler) CreateGeofence(c *gin.Context) {
	started := time.Now()
	var input models.GeofenceRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	if err := services.ValidateGeofence(input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}

	geofence, err := h.Geofences.Create(c.Request.Context(), input)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "create geofence: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, map[string]any{
		"id":     geofence.ID,
		"name":   geofence.Name,
		"status": "active",
	}, started)
}
func (h Handler) ListGeofences(c *gin.Context) {
	started := time.Now()
	geofences, err := h.Geofences.List(c.Request.Context(), c.Query("category"))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list geofences: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusOK, map[string]any{"geofences": geofences}, started)
}
func validateGeofence(input models.GeofenceRequest) error {
	return services.ValidateGeofence(input)
}
