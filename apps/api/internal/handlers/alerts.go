package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/services"
)

func (h Handler) ConfigureAlert(c *gin.Context) {
	started := time.Now()
	var input models.AlertConfigRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}

	alert, err := h.Alerts.Configure(c.Request.Context(), input)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, services.ErrNotFound) {
			status = http.StatusNotFound
		} else if err.Error() == "invalid geofence_id" || err.Error() == "invalid vehicle_id" ||
			err.Error() == "event_type must be entry, exit, or both" {
			status = http.StatusBadRequest
		}
		writeError(c, status, err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, gin.H{
		"alert_id":      alert.AlertID,
		"geofence_id":   alert.GeofenceID,
		"vehicle_id":    alert.VehicleID,
		"event_type":    alert.EventType,
		"status":        alert.Status,
		"geofence_name": alert.GeofenceName,
	}, started)
}

func (h Handler) ListAlerts(c *gin.Context) {
	started := time.Now()
	alerts, err := h.Alerts.List(c.Request.Context(), c.Query("geofence_id"), c.Query("vehicle_id"))
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid geofence_id" || err.Error() == "invalid vehicle_id" {
			status = http.StatusBadRequest
		}
		writeError(c, status, err.Error(), started)
		return
	}
	writeJSON(c, http.StatusOK, gin.H{"alerts": alerts}, started)
}

func (h Handler) ListViolations(c *gin.Context) {
	started := time.Now()
	violations, total, err := h.Alerts.ListViolations(c.Request.Context(), services.ViolationFilters{
		VehicleID:  c.Query("vehicle_id"),
		GeofenceID: c.Query("geofence_id"),
		StartDate:  c.Query("start_date"),
		EndDate:    c.Query("end_date"),
		Limit:      c.Query("limit"),
	})
	if err != nil {
		status := http.StatusInternalServerError
		switch err.Error() {
		case "invalid filter id", "invalid start_date", "invalid end_date", "limit must be a positive integer":
			status = http.StatusBadRequest
		}
		writeError(c, status, err.Error(), started)
		return
	}
	writeJSON(c, http.StatusOK, gin.H{"violations": violations, "total_count": total}, started)
}

func (h Handler) broadcastAlert(c *gin.Context, vehicleID uuid.UUID, vehicleNumber, driverName string, geofence models.Geofence, eventType string, input models.LocationRequest) error {
	payload, configured, err := h.Alerts.CreateEvent(c.Request.Context(), vehicleID, vehicleNumber, driverName, geofence, eventType, input)
	if err != nil || !configured {
		return err
	}
	go h.Hub.Broadcast(context.Background(), payload)
	return nil
}
