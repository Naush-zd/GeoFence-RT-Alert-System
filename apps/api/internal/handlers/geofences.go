package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

func (h Handler) CreateGeofence(c *gin.Context) {
	started := time.Now()
	var input models.GeofenceRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	if err := validateGeofence(input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}

	coordinates, err := json.Marshal(input.Coordinates)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "encode coordinates: "+err.Error(), started)
		return
	}

	id := uuid.New()
	_, err = h.DB.Exec(c.Request.Context(), `
		INSERT INTO geofences (id, name, description, coordinates, category)
		VALUES ($1, $2, $3, $4, $5)
	`, id, input.Name, input.Description, coordinates, input.Category)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "create geofence: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, map[string]any{
		"id":     id,
		"name":   input.Name,
		"status": "active",
	}, started)
}
func (h Handler) ListGeofences(c *gin.Context) {
	started := time.Now()
	category := c.Query("category")
	query := `SELECT id, name, description, coordinates, category, created_at FROM geofences`
	args := []any{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.DB.Query(c.Request.Context(), query, args...)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list geofences: "+err.Error(), started)
		return
	}
	defer rows.Close()

	geofences := make([]models.Geofence, 0)
	for rows.Next() {
		var item models.Geofence
		var rawCoordinates []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &rawCoordinates, &item.Category, &item.CreatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "read geofence: "+err.Error(), started)
			return
		}
		if err := json.Unmarshal(rawCoordinates, &item.Coordinates); err != nil {
			writeError(c, http.StatusInternalServerError, "decode coordinates: "+err.Error(), started)
			return
		}
		geofences = append(geofences, item)
	}
	if err := rows.Err(); err != nil {
		writeError(c, http.StatusInternalServerError, "read geofences: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusOK, map[string]any{"geofences": geofences}, started)
}
func validateGeofence(input models.GeofenceRequest) error {
	if strings.TrimSpace(input.Name) == "" || strings.TrimSpace(input.Category) == "" {
		return errors.New("name and category are required")
	}
	if len(input.Coordinates) < 4 {
		return errors.New("coordinates must contain at least 4 points")
	}
	switch input.Category {
	case "delivery_zone", "restricted_zone", "toll_zone", "customer_area":
	default:
		return errors.New("invalid category")
	}
	for _, point := range input.Coordinates {
		if len(point) != 2 || point[0] < -90 || point[0] > 90 || point[1] < -180 || point[1] > 180 {
			return errors.New("coordinates must be [latitude, longitude] within valid ranges")
		}
	}
	if input.Coordinates[0][0] != input.Coordinates[len(input.Coordinates)-1][0] ||
		input.Coordinates[0][1] != input.Coordinates[len(input.Coordinates)-1][1] {
		return errors.New("first and last coordinates must match")
	}
	return nil
}
