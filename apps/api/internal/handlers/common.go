package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/yourusername/geofencing-platform/apps/api/internal/realtime"
	"github.com/yourusername/geofencing-platform/apps/api/internal/services"
	"gorm.io/gorm"
)

type Handler struct {
	DB        *gorm.DB
	Hub       *realtime.Hub
	Geofences *services.GeofenceService
	Vehicles  *services.VehicleService
	Tracking  *services.TrackingService
	Alerts    *services.AlertService
}

func New(db *gorm.DB, hub *realtime.Hub, geofences *services.GeofenceService, vehicles *services.VehicleService, tracking *services.TrackingService, alerts *services.AlertService) Handler {
	return Handler{DB: db, Hub: hub, Geofences: geofences, Vehicles: vehicles, Tracking: tracking, Alerts: alerts}
}
func (h Handler) Health(c *gin.Context) {
	started := time.Now()
	if err := h.DB.WithContext(c.Request.Context()).Exec("SELECT 1").Error; err != nil {
		writeJSON(c, http.StatusServiceUnavailable, map[string]any{
			"status": "database_unavailable",
			"error":  "database is unavailable",
		}, started)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"status": "ok"}, started)
}
func decodeJSON(c *gin.Context, destination any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if err := binding.Validator.ValidateStruct(destination); err != nil {
		return fmt.Errorf("invalid request: %w", err)
	}
	return nil
}
func writeError(c *gin.Context, status int, message string, started time.Time) {
	writeJSON(c, status, map[string]any{"error": message}, started)
}
func writeJSON(c *gin.Context, status int, payload map[string]any, started time.Time) {
	payload["time_ns"] = fmt.Sprintf("%d", time.Since(started).Nanoseconds())
	c.JSON(status, payload)
}
