package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/geofencing?sslmode=disable"

type geofenceRequest struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
}

type vehicleRequest struct {
	VehicleNumber string `json:"vehicle_number"`
	DriverName    string `json:"driver_name"`
	VehicleType   string `json:"vehicle_type"`
	Phone         string `json:"phone"`
}

type geofence struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
	CreatedAt   time.Time   `json:"created_at"`
}

type vehicle struct {
	ID            string    `json:"id"`
	VehicleNumber string    `json:"vehicle_number"`
	DriverName    string    `json:"driver_name"`
	VehicleType   string    `json:"vehicle_type"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	db, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	if err := migrate(ctx, db); err != nil {
		log.Fatalf("run database migrations: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", healthHandler(db))
	mux.HandleFunc("POST /geofences", createGeofenceHandler(db))
	mux.HandleFunc("GET /geofences", listGeofencesHandler(db))
	mux.HandleFunc("POST /vehicles", createVehicleHandler(db))
	mux.HandleFunc("GET /vehicles", listVehiclesHandler(db))

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8080"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           withJSONContentType(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS geofences (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			coordinates JSONB NOT NULL,
			category TEXT NOT NULL CHECK (category IN ('delivery_zone', 'restricted_zone', 'toll_zone', 'customer_area')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS vehicles (
			id UUID PRIMARY KEY,
			vehicle_number TEXT NOT NULL UNIQUE,
			driver_name TEXT NOT NULL,
			vehicle_type TEXT NOT NULL,
			phone TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	return err
}

func healthHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		if err := db.Ping(r.Context()); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "database_unavailable", "error": "database is unavailable"}, started)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"}, started)
	}
}

func createGeofenceHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		var input geofenceRequest
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
		_, err = db.Exec(r.Context(), `
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
}

func listGeofencesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		category := r.URL.Query().Get("category")
		query := `SELECT id, name, description, coordinates, category, created_at FROM geofences`
		args := []any{}
		if category != "" {
			query += " WHERE category = $1"
			args = append(args, category)
		}
		query += " ORDER BY created_at DESC"

		rows, err := db.Query(r.Context(), query, args...)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list geofences: "+err.Error(), started)
			return
		}
		defer rows.Close()

		geofences := make([]geofence, 0)
		for rows.Next() {
			var item geofence
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
}

func createVehicleHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		var input vehicleRequest
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), started)
			return
		}
		if err := validateVehicle(input); err != nil {
			writeError(w, http.StatusBadRequest, err.Error(), started)
			return
		}

		id := uuid.New()
		err := db.QueryRow(r.Context(), `
			INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING created_at
		`, id, input.VehicleNumber, input.DriverName, input.VehicleType, input.Phone).Scan(new(time.Time))
		if err != nil {
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
}

func listVehiclesHandler(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		rows, err := db.Query(r.Context(), `
			SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at
			FROM vehicles
			ORDER BY created_at DESC
		`)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "list vehicles: "+err.Error(), started)
			return
		}
		defer rows.Close()

		vehicles := make([]vehicle, 0)
		for rows.Next() {
			var item vehicle
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
}

func validateGeofence(input geofenceRequest) error {
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

func validateVehicle(input vehicleRequest) error {
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

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func writeError(w http.ResponseWriter, status int, message string, started time.Time) {
	writeJSON(w, status, map[string]any{"error": message}, started)
}

func writeJSON(w http.ResponseWriter, status int, payload any, started time.Time) {
	body, ok := payload.(map[string]any)
	if !ok {
		body = map[string]any{"data": payload}
	}
	body["time_ns"] = fmt.Sprintf("%d", time.Since(started).Nanoseconds())
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
