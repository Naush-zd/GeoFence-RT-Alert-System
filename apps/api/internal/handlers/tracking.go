package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/yourusername/geofencing-platform/apps/api/internal/geospatial"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

func (h Handler) UpdateVehicleLocation(c *gin.Context) {
	started := time.Now()
	var input models.LocationRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	vehicleID, err := uuid.Parse(input.VehicleID)
	if err != nil || input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 || input.Timestamp.IsZero() {
		writeError(c, http.StatusBadRequest, "vehicle_id, valid coordinates, and timestamp are required", started)
		return
	}

	var vehicleNumber string
	var driverName string
	if err := h.DB.QueryRow(c.Request.Context(), `
		SELECT vehicle_number, driver_name FROM vehicles WHERE id = $1
	`, vehicleID).Scan(&vehicleNumber, &driverName); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(c, http.StatusNotFound, "vehicle not found", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "find vehicle: "+err.Error(), started)
		return
	}

	geofences, err := h.loadGeofences(c)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error(), started)
		return
	}

	previousLatitude, previousLongitude, hasPrevious, err := h.previousLocation(c, vehicleID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "find previous location: "+err.Error(), started)
		return
	}

	if _, err := h.DB.Exec(c.Request.Context(), `
		INSERT INTO vehicle_locations (id, vehicle_id, latitude, longitude, recorded_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), vehicleID, input.Latitude, input.Longitude, input.Timestamp); err != nil {
		writeError(c, http.StatusInternalServerError, "store vehicle location: "+err.Error(), started)
		return
	}

	currentGeofences := make([]models.CurrentGeofence, 0)
	for _, geofence := range geofences {
		currentInside := geospatial.PointInPolygon(input.Latitude, input.Longitude, geofence.Coordinates)
		if currentInside {
			currentGeofences = append(currentGeofences, models.CurrentGeofence{
				GeofenceID:   geofence.ID,
				GeofenceName: geofence.Name,
				Category:     geofence.Category,
				Status:       "inside",
			})
		}

		previousInside := false
		if hasPrevious {
			previousInside = geospatial.PointInPolygon(previousLatitude, previousLongitude, geofence.Coordinates)
		}
		if currentInside == previousInside {
			continue
		}

		eventType := "entry"
		if !currentInside {
			eventType = "exit"
		}
		if err := h.recordViolation(c, vehicleID, geofence.ID, eventType, input); err != nil {
			writeError(c, http.StatusInternalServerError, "record violation: "+err.Error(), started)
			return
		}
		if err := h.broadcastAlert(c, vehicleID, vehicleNumber, driverName, geofence, eventType, input); err != nil {
			writeError(c, http.StatusInternalServerError, "record alert: "+err.Error(), started)
			return
		}
	}

	writeJSON(c, http.StatusOK, map[string]any{
		"vehicle_id":        input.VehicleID,
		"location_updated":  true,
		"current_geofences": currentGeofences,
	}, started)
}
func (h Handler) GetVehicleLocation(c *gin.Context) {
	started := time.Now()
	vehicleID, err := uuid.Parse(c.Param("vehicle_id"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid vehicle_id", started)
		return
	}

	var vehicleNumber string
	if err := h.DB.QueryRow(c.Request.Context(), `SELECT vehicle_number FROM vehicles WHERE id = $1`, vehicleID).Scan(&vehicleNumber); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(c, http.StatusNotFound, "vehicle not found", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "find vehicle: "+err.Error(), started)
		return
	}

	var location models.LocationRequest
	if err := h.DB.QueryRow(c.Request.Context(), `
		SELECT latitude, longitude, recorded_at
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`, vehicleID).Scan(&location.Latitude, &location.Longitude, &location.Timestamp); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(c, http.StatusNotFound, "vehicle has no recorded location", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "get vehicle location: "+err.Error(), started)
		return
	}
	location.VehicleID = vehicleID.String()

	geofences, err := h.loadGeofences(c)
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error(), started)
		return
	}
	currentGeofences := make([]models.CurrentGeofence, 0)
	for _, geofence := range geofences {
		if geospatial.PointInPolygon(location.Latitude, location.Longitude, geofence.Coordinates) {
			currentGeofences = append(currentGeofences, models.CurrentGeofence{
				GeofenceID:   geofence.ID,
				GeofenceName: geofence.Name,
				Category:     geofence.Category,
			})
		}
	}

	writeJSON(c, http.StatusOK, map[string]any{
		"vehicle_id":        vehicleID.String(),
		"vehicle_number":    vehicleNumber,
		"current_location":  location,
		"current_geofences": currentGeofences,
	}, started)
}
func (h Handler) loadGeofences(c *gin.Context) ([]models.Geofence, error) {
	rows, err := h.DB.Query(c.Request.Context(), `SELECT id, name, coordinates, category FROM geofences`)
	if err != nil {
		return nil, fmt.Errorf("list geofences: %w", err)
	}
	defer rows.Close()

	geofences := make([]models.Geofence, 0)
	for rows.Next() {
		var item models.Geofence
		var rawCoordinates []byte
		if err := rows.Scan(&item.ID, &item.Name, &rawCoordinates, &item.Category); err != nil {
			return nil, fmt.Errorf("read geofence: %w", err)
		}
		if err := json.Unmarshal(rawCoordinates, &item.Coordinates); err != nil {
			return nil, fmt.Errorf("decode coordinates: %w", err)
		}
		geofences = append(geofences, item)
	}
	return geofences, rows.Err()
}
func (h Handler) previousLocation(c *gin.Context, vehicleID uuid.UUID) (float64, float64, bool, error) {
	var latitude, longitude float64
	err := h.DB.QueryRow(c.Request.Context(), `
		SELECT latitude, longitude
		FROM vehicle_locations
		WHERE vehicle_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`, vehicleID).Scan(&latitude, &longitude)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, 0, false, nil
	}
	return latitude, longitude, err == nil, err
}
func (h Handler) recordViolation(c *gin.Context, vehicleID uuid.UUID, geofenceID string, eventType string, input models.LocationRequest) error {
	_, err := h.DB.Exec(c.Request.Context(), `
		INSERT INTO violations (id, vehicle_id, geofence_id, event_type, latitude, longitude, recorded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New(), vehicleID, geofenceID, eventType, input.Latitude, input.Longitude, input.Timestamp)
	return err
}
