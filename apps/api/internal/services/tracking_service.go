package services

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/repositories"
	"gorm.io/gorm"
)

type TrackingService struct {
	vehicles  *repositories.VehicleRepository
	geofences *repositories.GeofenceRepository
	tracking  *repositories.TrackingRepository
}

type GeofenceTransition struct {
	Geofence  models.Geofence
	EventType string
}

type LocationUpdateResult struct {
	Vehicle          models.Vehicle
	CurrentGeofences []models.CurrentGeofence
	Transitions      []GeofenceTransition
}

func NewTrackingService(
	vehicles *repositories.VehicleRepository,
	geofences *repositories.GeofenceRepository,
	tracking *repositories.TrackingRepository,
) *TrackingService {
	return &TrackingService{vehicles: vehicles, geofences: geofences, tracking: tracking}
}

func (s *TrackingService) UpdateLocation(ctx context.Context, input models.LocationRequest) (LocationUpdateResult, error) {
	vehicleID, err := uuid.Parse(input.VehicleID)
	if err != nil || input.Latitude < -90 || input.Latitude > 90 || input.Longitude < -180 || input.Longitude > 180 || input.Timestamp.IsZero() {
		return LocationUpdateResult{}, errors.New("vehicle_id, valid coordinates, and timestamp are required")
	}

	vehicle, err := s.vehicles.Get(ctx, vehicleID)
	if err != nil {
		return LocationUpdateResult{}, err
	}
	currentGeofences, err := s.geofences.FindContaining(ctx, input.Latitude, input.Longitude)
	if err != nil {
		return LocationUpdateResult{}, fmt.Errorf("find current geofences: %w", err)
	}
	previous, previousErr := s.tracking.LatestLocation(ctx, vehicleID)
	if previousErr != nil && !errors.Is(previousErr, gorm.ErrRecordNotFound) {
		return LocationUpdateResult{}, fmt.Errorf("find previous location: %w", previousErr)
	}
	if err := s.tracking.RecordLocation(ctx, input); err != nil {
		return LocationUpdateResult{}, fmt.Errorf("store vehicle location: %w", err)
	}
	var previousGeofences []models.Geofence
	if previousErr == nil {
		previousGeofences, err = s.geofences.FindContaining(ctx, previous.Latitude, previous.Longitude)
		if err != nil {
			return LocationUpdateResult{}, fmt.Errorf("find previous geofences: %w", err)
		}
	}

	result := LocationUpdateResult{
		Vehicle:          vehicle,
		CurrentGeofences: make([]models.CurrentGeofence, 0),
		Transitions:      make([]GeofenceTransition, 0),
	}
	currentByID := make(map[string]models.Geofence, len(currentGeofences))
	for _, geofence := range currentGeofences {
		currentByID[geofence.ID] = geofence
		result.CurrentGeofences = append(result.CurrentGeofences, models.CurrentGeofence{
			GeofenceID:   geofence.ID,
			GeofenceName: geofence.Name,
			Category:     geofence.Category,
			Status:       "inside",
		})
	}
	previousByID := make(map[string]models.Geofence, len(previousGeofences))
	for _, geofence := range previousGeofences {
		previousByID[geofence.ID] = geofence
	}
	for id, geofence := range currentByID {
		if _, wasInside := previousByID[id]; wasInside {
			continue
		}
		geofenceID, err := uuid.Parse(geofence.ID)
		if err != nil {
			return LocationUpdateResult{}, fmt.Errorf("parse geofence id: %w", err)
		}
		if err := s.tracking.RecordViolation(ctx, vehicleID, geofenceID, "entry", input); err != nil {
			return LocationUpdateResult{}, fmt.Errorf("record violation: %w", err)
		}
		result.Transitions = append(result.Transitions, GeofenceTransition{
			Geofence:  geofence,
			EventType: "entry",
		})
	}
	for id, geofence := range previousByID {
		if _, stillInside := currentByID[id]; stillInside {
			continue
		}
		geofenceID, err := uuid.Parse(geofence.ID)
		if err != nil {
			return LocationUpdateResult{}, fmt.Errorf("parse geofence id: %w", err)
		}
		if err := s.tracking.RecordViolation(ctx, vehicleID, geofenceID, "exit", input); err != nil {
			return LocationUpdateResult{}, fmt.Errorf("record violation: %w", err)
		}
		result.Transitions = append(result.Transitions, GeofenceTransition{
			Geofence:  geofence,
			EventType: "exit",
		})
	}
	return result, nil
}

func (s *TrackingService) CurrentLocation(ctx context.Context, vehicleID uuid.UUID) (models.Vehicle, models.LocationRequest, []models.CurrentGeofence, error) {
	vehicle, err := s.vehicles.Get(ctx, vehicleID)
	if err != nil {
		return models.Vehicle{}, models.LocationRequest{}, nil, err
	}
	location, err := s.tracking.LatestLocation(ctx, vehicleID)
	if err != nil {
		return models.Vehicle{}, models.LocationRequest{}, nil, err
	}
	geofences, err := s.geofences.FindContaining(ctx, location.Latitude, location.Longitude)
	if err != nil {
		return models.Vehicle{}, models.LocationRequest{}, nil, err
	}

	current := make([]models.CurrentGeofence, 0)
	for _, geofence := range geofences {
		current = append(current, models.CurrentGeofence{
			GeofenceID:   geofence.ID,
			GeofenceName: geofence.Name,
			Category:     geofence.Category,
		})
	}
	return vehicle, location, current, nil
}
