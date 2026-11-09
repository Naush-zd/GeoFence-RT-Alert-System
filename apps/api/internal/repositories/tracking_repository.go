package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"gorm.io/gorm"
)

type TrackingRepository struct {
	db *gorm.DB
}

type vehicleLocationRecord struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	VehicleID  uuid.UUID `gorm:"column:vehicle_id;type:uuid"`
	Latitude   float64   `gorm:"column:latitude"`
	Longitude  float64   `gorm:"column:longitude"`
	RecordedAt time.Time `gorm:"column:recorded_at"`
}

type violationRecord struct {
	ID         uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	VehicleID  uuid.UUID `gorm:"column:vehicle_id;type:uuid"`
	GeofenceID uuid.UUID `gorm:"column:geofence_id;type:uuid"`
	EventType  string    `gorm:"column:event_type"`
	Latitude   float64   `gorm:"column:latitude"`
	Longitude  float64   `gorm:"column:longitude"`
	RecordedAt time.Time `gorm:"column:recorded_at"`
}

func (vehicleLocationRecord) TableName() string {
	return "vehicle_locations"
}

func (violationRecord) TableName() string {
	return "violations"
}

func NewTrackingRepository(db *gorm.DB) *TrackingRepository {
	return &TrackingRepository{db: db}
}

func (r *TrackingRepository) RecordLocation(ctx context.Context, input models.LocationRequest) error {
	vehicleID, err := uuid.Parse(input.VehicleID)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(&vehicleLocationRecord{
		ID:         uuid.New(),
		VehicleID:  vehicleID,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		RecordedAt: input.Timestamp,
	}).Error
}

func (r *TrackingRepository) LatestLocation(ctx context.Context, vehicleID uuid.UUID) (models.LocationRequest, error) {
	var record vehicleLocationRecord
	if err := r.db.WithContext(ctx).
		Where("vehicle_id = ?", vehicleID).
		Order("recorded_at DESC").
		First(&record).Error; err != nil {
		return models.LocationRequest{}, err
	}
	return models.LocationRequest{
		VehicleID: vehicleID.String(),
		Latitude:  record.Latitude,
		Longitude: record.Longitude,
		Timestamp: record.RecordedAt,
	}, nil
}

func (r *TrackingRepository) RecordViolation(ctx context.Context, vehicleID, geofenceID uuid.UUID, eventType string, input models.LocationRequest) error {
	return r.db.WithContext(ctx).Create(&violationRecord{
		ID:         uuid.New(),
		VehicleID:  vehicleID,
		GeofenceID: geofenceID,
		EventType:  eventType,
		Latitude:   input.Latitude,
		Longitude:  input.Longitude,
		RecordedAt: input.Timestamp,
	}).Error
}
