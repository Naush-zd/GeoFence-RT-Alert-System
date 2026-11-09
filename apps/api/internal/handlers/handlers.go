package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
)

type Handler struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) Handler {
	return Handler{DB: db}
}

func (h Handler) Health(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	if err := h.DB.Ping(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "database_unavailable",
			"error":  "database is unavailable",
		}, started)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok"}, started)
}

func (h Handler) CreateGeofence(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	var input models.GeofenceRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), started)
		return
	}
	if err := validateGeofence(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), started)
		return
	}

	coordinates, err := json.Marshal(input.Coordinates)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode coordinates: "+err.Error(), started)
		return
	}

	id := uuid.New()
	_, err = h.DB.Exec(r.Context(), `
		INSERT INTO geofences (id, name, description, coordinates, category)
		VALUES ($1, $2, $3, $4, $5)
	`, id, input.Name, input.Description, coordinates, input.Category)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "create geofence: "+err.Error(), started)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":     id,
		"name":   input.Name,
		"status": "active",
	}, started)
}

func (h Handler) ListGeofences(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	category := r.URL.Query().Get("category")
	query := `SELECT id, name, description, coordinates, category, created_at FROM geofences`
	args := []any{}
	if category != "" {
		query += " WHERE category = $1"
		args = append(args, category)
	}
	query += " ORDER BY created_at DESC"

	rows, err := h.DB.Query(r.Context(), query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list geofences: "+err.Error(), started)
		return
	}
	defer rows.Close()

	geofences := make([]models.Geofence, 0)
	for rows.Next() {
		var item models.Geofence
		var rawCoordinates []byte
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &rawCoordinates, &item.Category, &item.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "read geofence: "+err.Error(), started)
			return
		}
		if err := json.Unmarshal(rawCoordinates, &item.Coordinates); err != nil {
			writeError(w, http.StatusInternalServerError, "decode coordinates: "+err.Error(), started)
			return
		}
		geofences = append(geofences, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "read geofences: "+err.Error(), started)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"geofences": geofences}, started)
}

func (h Handler) CreateVehicle(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	var input models.VehicleRequest
	if err := decodeJSON(r, &input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), started)
		return
	}
	if err := validateVehicle(input); err != nil {
		writeError(w, http.StatusBadRequest, err.Error(), started)
		return
	}

	id := uuid.New()
	if err := h.DB.QueryRow(r.Context(), `
		INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`, id, input.VehicleNumber, input.DriverName, input.VehicleType, input.Phone).Scan(new(time.Time)); err != nil {
		if strings.Contains(err.Error(), "vehicles_vehicle_number_key") {
			writeError(w, http.StatusConflict, "vehicle_number already exists", started)
			return
		}
		writeError(w, http.StatusInternalServerError, "create vehicle: "+err.Error(), started)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":             id,
		"vehicle_number": input.VehicleNumber,
		"status":         "active",
	}, started)
}

func (h Handler) ListVehicles(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	rows, err := h.DB.Query(r.Context(), `
		SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
		FROM vehicles
		ORDER BY created_at DESC
	`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "list vehicles: "+err.Error(), started)
		return
	}
	defer rows.Close()

	vehicles := make([]models.Vehicle, 0)
	for rows.Next() {
		var item models.Vehicle
		if err := rows.Scan(&item.ID, &item.VehicleNumber, &item.DriverName, &item.VehicleType, &item.Phone, &item.Status, &item.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "read vehicle: "+err.Error(), started)
			return
		}
		vehicles = append(vehicles, item)
	}
	if err := rows.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "read vehicles: "+err.Error(), started)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"vehicles": vehicles}, started)
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

func validateVehicle(input models.VehicleRequest) error {
	if strings.TrimSpace(input.VehicleNumber) == "" ||
		strings.TrimSpace(input.DriverName) == "" ||
		strings.TrimSpace(input.VehicleType) == "" ||
		strings.TrimSpace(input.Phone) == "" {
		return errors.New("vehicle_number, driver_name, vehicle_type, and phone are required")
	}
	return nil
}

func decodeJSON(r *http.Request, destination any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	return nil
}

func writeError(w http.ResponseWriter, status int, message string, started time.Time) {
	writeJSON(w, status, map[string]any{"error": message}, started)
}

func writeJSON(w http.ResponseWriter, status int, payload map[string]any, started time.Time) {
	payload["time_ns"] = fmt.Sprintf("%d", time.Since(started).Nanoseconds())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
