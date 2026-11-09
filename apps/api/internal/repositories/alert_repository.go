package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AlertRepository struct {
	db *gorm.DB
}

type alertConfigurationRecord struct {
	ID         uuid.UUID  `gorm:"column:id;type:uuid;primaryKey"`
	GeofenceID uuid.UUID  `gorm:"column:geofence_id;type:uuid"`
	VehicleID  *uuid.UUID `gorm:"column:vehicle_id;type:uuid"`
	EventType  string     `gorm:"column:event_type"`
	Status     string     `gorm:"column:status"`
	CreatedAt  time.Time  `gorm:"column:created_at"`
}

type alertConfigurationRow struct {
	ID            uuid.UUID
	GeofenceID    uuid.UUID
	GeofenceName  string
	VehicleID     *uuid.UUID
	VehicleNumber string
	EventType     string
	Status        string
	CreatedAt     time.Time
}

type alertEventRecord struct {
	ID         uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	VehicleID  uuid.UUID      `gorm:"column:vehicle_id;type:uuid"`
	GeofenceID uuid.UUID      `gorm:"column:geofence_id;type:uuid"`
	EventType  string         `gorm:"column:event_type"`
	Latitude   float64        `gorm:"column:latitude"`
	Longitude  float64        `gorm:"column:longitude"`
	OccurredAt time.Time      `gorm:"column:occurred_at"`
	Payload    datatypes.JSON `gorm:"column:payload;type:jsonb"`
}

type violationRow struct {
	ID            uuid.UUID
	VehicleID     uuid.UUID
	VehicleNumber string
	GeofenceID    uuid.UUID
	GeofenceName  string
	EventType     string
	Latitude      float64
	Longitude     float64
	Timestamp     time.Time `gorm:"column:recorded_at"`
}

func (alertConfigurationRecord) TableName() string {
	return "alert_configurations"
}

func (alertEventRecord) TableName() string {
	return "alert_events"
}

func NewAlertRepository(db *gorm.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) GeofenceExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("geofences").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *AlertRepository) VehicleExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("vehicles").Where("id = ?", id).Count(&count).Error
	return count > 0, err
}

func (r *AlertRepository) CreateConfiguration(ctx context.Context, geofenceID uuid.UUID, vehicleID *uuid.UUID, eventType string) (models.AlertConfig, error) {
	record := alertConfigurationRecord{
		ID:         uuid.New(),
		GeofenceID: geofenceID,
		VehicleID:  vehicleID,
		EventType:  eventType,
		Status:     "active",
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return models.AlertConfig{}, err
	}
	var geofence struct {
		Name string
	}
	if err := r.db.WithContext(ctx).Table("geofences").Select("name").First(&geofence, "id = ?", geofenceID).Error; err != nil {
		return models.AlertConfig{}, err
	}
	return models.AlertConfig{
		AlertID:      record.ID.String(),
		GeofenceID:   geofenceID.String(),
		GeofenceName: geofence.Name,
		VehicleID:    uuidString(vehicleID),
		EventType:    record.EventType,
		Status:       record.Status,
		CreatedAt:    record.CreatedAt,
	}, nil
}

func (r *AlertRepository) ListConfigurations(ctx context.Context, geofenceID, vehicleID *uuid.UUID) ([]models.AlertConfig, error) {
	query := r.db.WithContext(ctx).
		Table("alert_configurations AS a").
		Select("a.id, a.geofence_id, g.name AS geofence_name, a.vehicle_id, COALESCE(v.vehicle_number, '') AS vehicle_number, a.event_type, a.status, a.created_at").
		Joins("JOIN geofences AS g ON g.id = a.geofence_id").
		Joins("LEFT JOIN vehicles AS v ON v.id = a.vehicle_id").
		Order("a.created_at DESC")
	if geofenceID != nil {
		query = query.Where("a.geofence_id = ?", *geofenceID)
	}
	if vehicleID != nil {
		query = query.Where("a.vehicle_id = ?", *vehicleID)
	}

	var rows []alertConfigurationRow
	if err := query.Scan(&rows).Error; err != nil {
		return nil, err
	}
	alerts := make([]models.AlertConfig, 0, len(rows))
	for _, row := range rows {
		alerts = append(alerts, models.AlertConfig{
			AlertID:       row.ID.String(),
			GeofenceID:    row.GeofenceID.String(),
			GeofenceName:  row.GeofenceName,
			VehicleID:     uuidString(row.VehicleID),
			VehicleNumber: row.VehicleNumber,
			EventType:     row.EventType,
			Status:        row.Status,
			CreatedAt:     row.CreatedAt,
		})
	}
	return alerts, nil
}

func (r *AlertRepository) HasMatchingConfiguration(ctx context.Context, geofenceID, vehicleID uuid.UUID, eventType string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&alertConfigurationRecord{}).
		Where("geofence_id = ? AND status = 'active' AND (vehicle_id IS NULL OR vehicle_id = ?) AND (event_type = ? OR event_type = 'both')", geofenceID, vehicleID, eventType).
		Count(&count).Error
	return count > 0, err
}

func (r *AlertRepository) CreateEvent(ctx context.Context, event models.AlertEvent, payload []byte) error {
	vehicleID, err := uuid.Parse(event.Vehicle.ID)
	if err != nil {
		return err
	}
	geofenceID, err := uuid.Parse(event.Geofence.ID)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(&alertEventRecord{
		ID:         uuid.MustParse(event.EventID),
		VehicleID:  vehicleID,
		GeofenceID: geofenceID,
		EventType:  event.EventType,
		Latitude:   event.Location.Latitude,
		Longitude:  event.Location.Longitude,
		OccurredAt: event.Timestamp,
		Payload:    datatypes.JSON(payload),
	}).Error
}

func (r *AlertRepository) ListViolations(ctx context.Context, vehicleID, geofenceID *uuid.UUID, start, end *time.Time, limit int) ([]models.Violation, int64, error) {
	query := r.db.WithContext(ctx).
		Table("violations AS violation").
		Joins("JOIN vehicles AS vehicle ON vehicle.id = violation.vehicle_id").
		Joins("JOIN geofences AS geofence ON geofence.id = violation.geofence_id").
		Where("1 = 1")
	if vehicleID != nil {
		query = query.Where("violation.vehicle_id = ?", *vehicleID)
	}
	if geofenceID != nil {
		query = query.Where("violation.geofence_id = ?", *geofenceID)
	}
	if start != nil {
		query = query.Where("violation.recorded_at >= ?", *start)
	}
	if end != nil {
		query = query.Where("violation.recorded_at <= ?", *end)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []violationRow
	if err := query.Select("violation.id, violation.vehicle_id, vehicle.vehicle_number, violation.geofence_id, geofence.name AS geofence_name, violation.event_type, violation.latitude, violation.longitude, violation.recorded_at").
		Order("violation.recorded_at DESC").
		Limit(limit).
		Scan(&rows).Error; err != nil {
		return nil, 0, err
	}
	violations := make([]models.Violation, 0, len(rows))
	for _, row := range rows {
		violations = append(violations, models.Violation{
			ID:            row.ID.String(),
			VehicleID:     row.VehicleID.String(),
			VehicleNumber: row.VehicleNumber,
			GeofenceID:    row.GeofenceID.String(),
			GeofenceName:  row.GeofenceName,
			EventType:     row.EventType,
			Latitude:      row.Latitude,
			Longitude:     row.Longitude,
			Timestamp:     row.Timestamp,
		})
	}
	return violations, total, nil
}

func uuidString(value *uuid.UUID) string {
	if value == nil {
		return ""
	}
	return value.String()
}
