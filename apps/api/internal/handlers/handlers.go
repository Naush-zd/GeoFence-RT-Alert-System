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

	"github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourusername/geofencing-platform/apps/api/internal/geospatial"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/realtime"
)

type Handler struct {
	DB  *pgxpool.Pool
	Hub *realtime.Hub
}

func New(db *pgxpool.Pool, hub *realtime.Hub) Handler {
	return Handler{DB: db, Hub: hub}
}

func (h Handler) Health(c *gin.Context) {
	started := time.Now()
	if err := h.DB.Ping(c.Request.Context()); err != nil {
		writeJSON(c, http.StatusServiceUnavailable, map[string]any{
			"status": "database_unavailable",
			"error":  "database is unavailable",
		}, started)
		return
	}
	writeJSON(c, http.StatusOK, map[string]any{"status": "ok"}, started)
}

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
		h.broadcastAlert(c, vehicleID, vehicleNumber, driverName, geofence, eventType, input)
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

func (h Handler) AlertsWebSocket(c *gin.Context) {
	conn, err := websocket.Accept(c.Writer, c.Request, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	h.Hub.Add(conn)
	defer h.Hub.Remove(conn)
	defer conn.Close(websocket.StatusNormalClosure, "connection closed")

	for {
		if _, _, err := conn.Read(c.Request.Context()); err != nil {
			return
		}
	}
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

func (h Handler) broadcastAlert(c *gin.Context, vehicleID uuid.UUID, vehicleNumber, driverName string, geofence models.Geofence, eventType string, input models.LocationRequest) {
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT event_type
		FROM alert_configurations
		WHERE geofence_id = $1
		  AND status = 'active'
		  AND (vehicle_id IS NULL OR vehicle_id = $2)
		  AND (event_type = $3 OR event_type = 'both')
	`, geofence.ID, vehicleID, eventType)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var configuredEvent string
		if rows.Scan(&configuredEvent) != nil {
			continue
		}
		var event models.AlertEvent
		event.EventID = uuid.NewString()
		event.EventType = eventType
		event.Timestamp = input.Timestamp
		event.Vehicle.ID = vehicleID.String()
		event.Vehicle.Number = vehicleNumber
		event.Vehicle.DriverName = driverName
		event.Geofence.ID = geofence.ID
		event.Geofence.Name = geofence.Name
		event.Geofence.Category = geofence.Category
		event.Location.Latitude = input.Latitude
		event.Location.Longitude = input.Longitude
		payload, err := json.Marshal(event)
		if err == nil {
			go h.Hub.Broadcast(context.Background(), payload)
		}
	}
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

func decodeJSON(c *gin.Context, destination any) error {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
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
