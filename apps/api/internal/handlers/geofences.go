package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/services"
	"gorm.io/gorm"
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

func (h Handler) DeleteGeofence(c *gin.Context) {
	started := time.Now()
	err := h.Geofences.Delete(c.Request.Context(), c.Param("geofence_id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid geofence_id" {
			status = http.StatusBadRequest
		} else if errors.Is(err, gorm.ErrRecordNotFound) {
			status = http.StatusNotFound
		}
		writeError(c, status, err.Error(), started)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"status": "deleted", "geofence_id": c.Param("geofence_id")}, started)
}
func validateGeofence(input models.GeofenceRequest) error {
	return services.ValidateGeofence(input)
}
