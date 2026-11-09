package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

func (h Handler) CreateVehicle(c *gin.Context) {
	started := time.Now()
	var input models.VehicleRequest
	if err := decodeJSON(c, &input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}
	if err := validateVehicle(input); err != nil {
		writeError(c, http.StatusBadRequest, err.Error(), started)
		return
	}

	id := uuid.New()
	if err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, id, input.VehicleNumber, input.DriverName, input.VehicleType, input.Phone).Scan(new(time.Time)); err != nil {
		if strings.Contains(err.Error(), "vehicles_vehicle_number_key") {
			writeError(c, http.StatusConflict, "vehicle_number already exists", started)
			return
		}
		writeError(c, http.StatusInternalServerError, "create vehicle: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusCreated, map[string]any{
		"id":             id,
		"vehicle_number": input.VehicleNumber,
		"status":         "active",
	}, started)
}
func (h Handler) ListVehicles(c *gin.Context) {
	started := time.Now()
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
		FROM vehicles
		ORDER BY created_at DESC
	`)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "list vehicles: "+err.Error(), started)
		return
	}
	defer rows.Close()

	vehicles := make([]models.Vehicle, 0)
	for rows.Next() {
		var item models.Vehicle
		if err := rows.Scan(&item.ID, &item.VehicleNumber, &item.DriverName, &item.VehicleType, &item.Phone, &item.Status, &item.CreatedAt); err != nil {
			writeError(c, http.StatusInternalServerError, "read vehicle: "+err.Error(), started)
			return
		}
		vehicles = append(vehicles, item)
	}
	if err := rows.Err(); err != nil {
		writeError(c, http.StatusInternalServerError, "read vehicles: "+err.Error(), started)
		return
	}

	writeJSON(c, http.StatusOK, map[string]any{"vehicles": vehicles}, started)
}
func validateVehicle(input models.VehicleRequest) error {
	if strings.TrimSpace(input.VehicleNumber) == "" ||
		strings.TrimSpace(input.DriverName) == "" ||
		strings.TrimSpace(input.VehicleType) == "" ||
		strings.TrimSpace(input.Phone) == "" {
		return errors.New("vehicle_number, driver_name, vehicle_type, and phone are required")
	}
	return nil
}
