package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"gorm.io/gorm"
)

type VehicleRepository struct {
	db *gorm.DB
}

type vehicleRecord struct {
	ID            uuid.UUID `gorm:"column:id;type:uuid;primaryKey"`
	VehicleNumber string    `gorm:"column:vehicle_number"`
	DriverName    string    `gorm:"column:driver_name"`
	VehicleType   string    `gorm:"column:vehicle_type"`
	Phone         string    `gorm:"column:phone"`
	Status        string    `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
}

func (vehicleRecord) TableName() string {
	return "vehicles"
}

func NewVehicleRepository(db *gorm.DB) *VehicleRepository {
	return &VehicleRepository{db: db}
}

func (r *VehicleRepository) Create(ctx context.Context, input models.VehicleRequest) (models.Vehicle, error) {
	record := vehicleRecord{
		ID:            uuid.New(),
		VehicleNumber: input.VehicleNumber,
		DriverName:    input.DriverName,
		VehicleType:   input.VehicleType,
		Phone:         input.Phone,
		Status:        "active",
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return models.Vehicle{}, err
	}
	return toVehicleModel(record), nil
}

func (r *VehicleRepository) List(ctx context.Context) ([]models.Vehicle, error) {
	var records []vehicleRecord
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, err
	}

	vehicles := make([]models.Vehicle, 0, len(records))
	for _, record := range records {
		vehicles = append(vehicles, toVehicleModel(record))
	}
	return vehicles, nil
}

func (r *VehicleRepository) Get(ctx context.Context, id uuid.UUID) (models.Vehicle, error) {
	var record vehicleRecord
	if err := r.db.WithContext(ctx).First(&record, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Vehicle{}, gorm.ErrRecordNotFound
		}
		return models.Vehicle{}, err
	}
	return toVehicleModel(record), nil
}

func toVehicleModel(record vehicleRecord) models.Vehicle {
	return models.Vehicle{
		ID:            record.ID.String(),
		VehicleNumber: record.VehicleNumber,
		DriverName:    record.DriverName,
		VehicleType:   record.VehicleType,
		Phone:         record.Phone,
		Status:        record.Status,
		CreatedAt:     record.CreatedAt,
	}
}
