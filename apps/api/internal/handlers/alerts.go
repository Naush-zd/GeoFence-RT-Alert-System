package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

func (h Handler) ConfigureAlert(c *gin.Context) {
	started := time.Now()
	var input models.AlertConfigRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	geofenceID, err := uuid.Parse(input.GeofenceID)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid geofence_id", started)
		return
	}
	var vehicleID *uuid.UUID
	if input.VehicleID != "" {
		parsed, parseErr := uuid.Parse(input.VehicleID)
		if parseErr != nil {
			writeError(c, http.StatusBadRequest, "invalid vehicle_id", started)
			return
		}
		vehicleID = &parsed
	}
	if input.EventType != "entry" && input.EventType != "exit" && input.EventType != "both" {
		writeError(c, http.StatusBadRequest, "event_type must be entry, exit, or both", started)
		return
	}

	var geofenceName string
	if err := h.DB.QueryRow(c.Request.Context(), `SELECT name FROM geofences WHERE id = $1`, geofenceID).Scan(&geofenceName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(c, http.StatusNotFound, "geofence not found", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "find geofence: "+err.Error(), started)
		return
	}
	if vehicleID != nil {
		var exists bool
		if err := h.DB.QueryRow(c.Request.Context(), `SELECT EXISTS (SELECT 1 FROM vehicles WHERE id = $1)`, *vehicleID).Scan(&exists); err != nil || !exists {
			writeError(c, http.StatusNotFound, "vehicle not found", started)
			return
		}
	}

	alertID := uuid.New()
	if _, err := h.DB.Exec(c.Request.Context(), `
		INSERT INTO alert_configurations (id, geofence_id, vehicle_id, event_type)
		VALUES ($1, $2, $3, $4)
	`, alertID, geofenceID, vehicleID, input.EventType); err != nil {
		writeError(c, http.StatusInternalServerError, "configure alert: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, map[string]any{
		"alert_id":      alertID,
		"geofence_id":   geofenceID,
		"vehicle_id":    input.VehicleID,
		"event_type":    input.EventType,
		"status":        "active",
		"geofence_name": geofenceName,
	}, started)
}
func (h Handler) ListAlerts(c *gin.Context) {
	started := time.Now()
	query := `
		SELECT a.id, a.geofence_id, g.name, a.vehicle_id, COALESCE(v.vehicle_number, ''),
		       a.event_type, a.status, a.created_at
		FROM alert_configurations a
		JOIN geofences g ON g.id = a.geofence_id
		LEFT JOIN vehicles v ON v.id = a.vehicle_id
		WHERE 1 = 1`
	args := make([]any, 0, 2)
	argIndex := 1
	if value := c.Query("geofence_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid geofence_id", started)
			return
		}
		query += fmt.Sprintf(" AND a.geofence_id = $%d", argIndex)
		args = append(args, id)
		argIndex++
	}
	if value := c.Query("vehicle_id"); value != "" {
		id, err := uuid.Parse(value)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid vehicle_id", started)
			return
		}
		query += fmt.Sprintf(" AND a.vehicle_id = $%d", argIndex)
		args = append(args, id)
	}
	query += " ORDER BY a.created_at DESC"

	rows, err := h.DB.Query(c.Request.Context(), query, args...)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list alerts: "+err.Error(), started)
		return
	}
	defer rows.Close()

	alerts := make([]models.AlertConfig, 0)
	for rows.Next() {
		var item models.AlertConfig
		var vehicleID *uuid.UUID
		if err := rows.Scan(&item.AlertID, &item.GeofenceID, &item.GeofenceName, &vehicleID, &item.VehicleNumber, &item.EventType, &item.Status, &item.CreatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "read alert: "+err.Error(), started)
			return
		}
		if vehicleID != nil {
			item.VehicleID = vehicleID.String()
		}
		alerts = append(alerts, item)
	}
	writeJSON(c, http.StatusOK, map[string]any{"alerts": alerts}, started)
}
func (h Handler) ListViolations(c *gin.Context) {
	started := time.Now()
	limit := 50
	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			writeError(c, http.StatusBadRequest, "limit must be a positive integer", started)
			return
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}

	filters := []string{"1 = 1"}
	args := make([]any, 0, 5)
	argIndex := 1
	for _, filter := range []struct {
		query string
		value string
	}{
		{"v.vehicle_id = $%d", c.Query("vehicle_id")},
		{"v.geofence_id = $%d", c.Query("geofence_id")},
	} {
		if filter.value == "" {
			continue
		}
		id, err := uuid.Parse(filter.value)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid filter id", started)
			return
		}
		filters = append(filters, fmt.Sprintf(filter.query, argIndex))
		args = append(args, id)
		argIndex++
	}

	query := `
		SELECT v.id, v.vehicle_id, vh.vehicle_number, v.geofence_id, g.name,
		       v.event_type, v.latitude, v.longitude, v.recorded_at
		FROM violations v
		JOIN vehicles vh ON vh.id = v.vehicle_id
		JOIN geofences g ON g.id = v.geofence_id
		WHERE ` + strings.Join(filters, " AND ")
	countQuery := "SELECT COUNT(*) FROM violations v WHERE " + strings.Join(filters, " AND ")
	if value := c.Query("start_date"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid start_date", started)
			return
		}
		query += fmt.Sprintf(" AND v.recorded_at >= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND v.recorded_at >= $%d", argIndex)
		args = append(args, parsed)
		argIndex++
	}
	if value := c.Query("end_date"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid end_date", started)
			return
		}
		query += fmt.Sprintf(" AND v.recorded_at <= $%d", argIndex)
		countQuery += fmt.Sprintf(" AND v.recorded_at <= $%d", argIndex)
		args = append(args, parsed)
		argIndex++
	}
	countArgs := append([]any(nil), args...)
	query += fmt.Sprintf(" ORDER BY v.recorded_at DESC LIMIT $%d", argIndex)
	args = append(args, limit)

	var total int
	if err := h.DB.QueryRow(c.Request.Context(), countQuery, countArgs...).Scan(&total); err != nil {
		writeError(c, http.StatusInternalServerError, "count violations: "+err.Error(), started)
		return
	}
	rows, err := h.DB.Query(c.Request.Context(), query, args...)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list violations: "+err.Error(), started)
		return
	}
	defer rows.Close()

	violations := make([]models.Violation, 0)
	for rows.Next() {
		var item models.Violation
		if err := rows.Scan(&item.ID, &item.VehicleID, &item.VehicleNumber, &item.GeofenceID, &item.GeofenceName, &item.EventType, &item.Latitude, &item.Longitude, &item.Timestamp); err != nil {
			writeError(c, http.StatusInternalServerError, "read violation: "+err.Error(), started)
			return
		}
		violations = append(violations, item)
	}
	writeJSON(c, http.StatusOK, map[string]any{"violations": violations, "total_count": total}, started)
}
func (h Handler) broadcastAlert(c *gin.Context, vehicleID uuid.UUID, vehicleNumber, driverName string, geofence models.Geofence, eventType string, input models.LocationRequest) error {
	geofenceID, err := uuid.Parse(geofence.ID)
	if err != nil {
		return fmt.Errorf("parse geofence id: %w", err)
	}

	var configured bool
	if err := h.DB.QueryRow(c.Request.Context(), `
		SELECT EXISTS (
			SELECT 1
			FROM alert_configurations
			WHERE geofence_id = $1
			  AND status = 'active'
			  AND (vehicle_id IS NULL OR vehicle_id = $2)
			  AND (event_type = $3 OR event_type = 'both')
		)
	`, geofenceID, vehicleID, eventType).Scan(&configured); err != nil {
		return err
	}
	if !configured {
		return nil
	}

	event := models.AlertEvent{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Timestamp: input.Timestamp,
	}
	event.Vehicle.ID = vehicleID.String()
	event.Vehicle.Number = vehicleNumber
	event.Vehicle.DriverName = driverName
	event.Geofence.ID = geofence.ID
	event.Geofence.Name = geofence.Name
	event.Geofence.Category = geofence.Category
	event.Location.Latitude = input.Latitude
	event.Location.Longitude = input.Longitude

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := h.DB.Exec(c.Request.Context(), `
		INSERT INTO alert_events (
			id, vehicle_id, geofence_id, event_type, latitude, longitude, occurred_at, payload
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, event.EventID, vehicleID, geofenceID, eventType, input.Latitude, input.Longitude, input.Timestamp, payload); err != nil {
		return err
	}

	go h.Hub.Broadcast(context.Background(), payload)
	return nil
}
