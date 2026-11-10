package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/geospatial"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type GeofenceRepository struct {
	db *gorm.DB
}

type geofenceRecord struct {
	ID          uuid.UUID      `gorm:"column:id;type:uuid;primaryKey"`
	Name        string         `gorm:"column:name"`
	Description string         `gorm:"column:description"`
	Coordinates datatypes.JSON `gorm:"column:coordinates;type:jsonb"`
	Category    string         `gorm:"column:category"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
}

func (geofenceRecord) TableName() string {
	return "geofences"
}

func NewGeofenceRepository(db *gorm.DB) *GeofenceRepository {
	return &GeofenceRepository{db: db}
}

func (r *GeofenceRepository) Create(ctx context.Context, input models.GeofenceRequest) (models.Geofence, error) {
	coordinates, err := json.Marshal(input.Coordinates)
	if err != nil {
		return models.Geofence{}, err
	}

	record := geofenceRecord{
		ID:          uuid.New(),
		Name:        input.Name,
		Description: input.Description,
		Coordinates: datatypes.JSON(coordinates),
		Category:    input.Category,
	}
	if err := r.db.WithContext(ctx).Create(&record).Error; err != nil {
		return models.Geofence{}, err
	}
	wkt, err := geospatial.PolygonWKT(input.Coordinates)
	if err != nil {
		return models.Geofence{}, err
	}
	if err := r.db.WithContext(ctx).Exec(
		"UPDATE geofences SET boundary = ST_SetSRID(ST_GeomFromText(?), 4326) WHERE id = ?",
		wkt,
		record.ID,
	).Error; err != nil {
		return models.Geofence{}, err
	}

	return toGeofenceModel(record)
}

func (r *GeofenceRepository) List(ctx context.Context, category string) ([]models.Geofence, error) {
	query := r.db.WithContext(ctx).Select("id, name, description, coordinates, category, created_at").Order("created_at DESC")
	if category != "" {
		query = query.Where("category = ?", category)
	}

	var records []geofenceRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, err
	}

	geofences := make([]models.Geofence, 0, len(records))
	for _, record := range records {
		geofence, err := toGeofenceModel(record)
		if err != nil {
			return nil, err
		}
		geofences = append(geofences, geofence)
	}
	return geofences, nil
}

func (r *GeofenceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Delete(&geofenceRecord{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GeofenceRepository) FindContaining(ctx context.Context, latitude, longitude float64) ([]models.Geofence, error) {
	var records []geofenceRecord
	if err := r.db.WithContext(ctx).
		Select("id, name, description, coordinates, category, created_at").
		Where("boundary IS NOT NULL").
		Where("ST_Covers(boundary, ST_SetSRID(ST_Point(?, ?), 4326))", longitude, latitude).
		Order("created_at DESC").
		Find(&records).Error; err != nil {
		return nil, err
	}

	geofences := make([]models.Geofence, 0, len(records))
	for _, record := range records {
		geofence, err := toGeofenceModel(record)
		if err != nil {
			return nil, err
		}
		geofences = append(geofences, geofence)
	}
	return geofences, nil
}

func toGeofenceModel(record geofenceRecord) (models.Geofence, error) {
	var coordinates [][]float64
	if err := json.Unmarshal(record.Coordinates, &coordinates); err != nil {
		return models.Geofence{}, err
	}
	return models.Geofence{
		ID:          record.ID.String(),
		Name:        record.Name,
		Description: record.Description,
		Coordinates: coordinates,
		Category:    record.Category,
		CreatedAt:   record.CreatedAt,
	}, nil
}
