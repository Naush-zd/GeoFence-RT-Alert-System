package services

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/repositories"
)

type GeofenceService struct {
	repository *repositories.GeofenceRepository
}

var ErrInvalidGeofence = errors.New("invalid geofence")

type geofenceValidationError string

func (err geofenceValidationError) Error() string {
	return string(err)
}

func (geofenceValidationError) Is(target error) bool {
	return target == ErrInvalidGeofence
}

func NewGeofenceService(repository *repositories.GeofenceRepository) *GeofenceService {
	return &GeofenceService{repository: repository}
}

func (s *GeofenceService) Create(ctx context.Context, input models.GeofenceRequest) (models.Geofence, error) {
	if err := ValidateGeofence(input); err != nil {
		return models.Geofence{}, err
	}
	return s.repository.Create(ctx, input)
}

func (s *GeofenceService) List(ctx context.Context, category string) ([]models.Geofence, error) {
	return s.repository.List(ctx, category)
}

func (s *GeofenceService) Delete(ctx context.Context, value string) error {
	id, err := uuid.Parse(value)
	if err != nil {
		return errors.New("invalid geofence_id")
	}
	return s.repository.Delete(ctx, id)
}

func ValidateGeofence(input models.GeofenceRequest) error {
	if len(input.Coordinates) < 4 {
		return geofenceValidationError("coordinates must contain at least 4 points")
	}
	for _, point := range input.Coordinates {
		if len(point) != 2 || point[0] < -90 || point[0] > 90 || point[1] < -180 || point[1] > 180 {
			return geofenceValidationError("coordinates must be [latitude, longitude] within valid ranges")
		}
	}
	if input.Coordinates[0][0] != input.Coordinates[len(input.Coordinates)-1][0] ||
		input.Coordinates[0][1] != input.Coordinates[len(input.Coordinates)-1][1] {
		return geofenceValidationError("first and last coordinates must match")
	}
	return nil
}
