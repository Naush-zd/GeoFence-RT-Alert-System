package services

import (
	"context"

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
	return s.repository.Create(ctx, input)
}

func (s *VehicleService) List(ctx context.Context) ([]models.Vehicle, error) {
	return s.repository.List(ctx)
}
