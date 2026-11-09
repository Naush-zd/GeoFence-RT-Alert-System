package services

import (
	"context"
	"errors"
	"strings"

	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/repositories"
)

type VehicleService struct {
	repository *repositories.VehicleRepository
}

func NewVehicleService(repository *repositories.VehicleRepository) *VehicleService {
	return &VehicleService{repository: repository}
}

func (s *VehicleService) Create(ctx context.Context, input models.VehicleRequest) (models.Vehicle, error) {
	if err := ValidateVehicle(input); err != nil {
		return models.Vehicle{}, err
	}
	return s.repository.Create(ctx, input)
}

func (s *VehicleService) List(ctx context.Context) ([]models.Vehicle, error) {
	return s.repository.List(ctx)
}

func ValidateVehicle(input models.VehicleRequest) error {
	if strings.TrimSpace(input.VehicleNumber) == "" ||
		strings.TrimSpace(input.DriverName) == "" ||
		strings.TrimSpace(input.VehicleType) == "" ||
		strings.TrimSpace(input.Phone) == "" {
		return errors.New("vehicle_number, driver_name, vehicle_type, and phone are required")
	}
	return nil
}
