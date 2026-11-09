package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/models"
	"github.com/yourusername/geofencing-platform/apps/api/internal/repositories"
)

type AlertService struct {
	repository *repositories.AlertRepository
}

type ViolationFilters struct {
	VehicleID  string
	GeofenceID string
	StartDate  string
	EndDate    string
	Limit      string
}

func NewAlertService(repository *repositories.AlertRepository) *AlertService {
	return &AlertService{repository: repository}
}

func (s *AlertService) Configure(ctx context.Context, input models.AlertConfigRequest) (models.AlertConfig, error) {
	geofenceID, err := uuid.Parse(input.GeofenceID)
	if err != nil {
		return models.AlertConfig{}, errors.New("invalid geofence_id")
	}
	if input.EventType != "entry" && input.EventType != "exit" && input.EventType != "both" {
		return models.AlertConfig{}, errors.New("event_type must be entry, exit, or both")
	}
	geofenceExists, err := s.repository.GeofenceExists(ctx, geofenceID)
	if err != nil {
		return models.AlertConfig{}, err
	}
	if !geofenceExists {
		return models.AlertConfig{}, ErrNotFound
	}

	var vehicleID *uuid.UUID
	if input.VehicleID != "" {
		parsed, err := uuid.Parse(input.VehicleID)
		if err != nil {
			return models.AlertConfig{}, errors.New("invalid vehicle_id")
		}
		exists, err := s.repository.VehicleExists(ctx, parsed)
		if err != nil {
			return models.AlertConfig{}, err
		}
		if !exists {
			return models.AlertConfig{}, ErrNotFound
		}
		vehicleID = &parsed
	}
	return s.repository.CreateConfiguration(ctx, geofenceID, vehicleID, input.EventType)
}

func (s *AlertService) List(ctx context.Context, geofenceIDValue, vehicleIDValue string) ([]models.AlertConfig, error) {
	geofenceID, err := optionalUUID(geofenceIDValue, "invalid geofence_id")
	if err != nil {
		return nil, err
	}
	vehicleID, err := optionalUUID(vehicleIDValue, "invalid vehicle_id")
	if err != nil {
		return nil, err
	}
	return s.repository.ListConfigurations(ctx, geofenceID, vehicleID)
}

func (s *AlertService) ListViolations(ctx context.Context, filters ViolationFilters) ([]models.Violation, int64, error) {
	vehicleID, err := optionalUUID(filters.VehicleID, "invalid filter id")
	if err != nil {
		return nil, 0, err
	}
	geofenceID, err := optionalUUID(filters.GeofenceID, "invalid filter id")
	if err != nil {
		return nil, 0, err
	}
	start, err := optionalTime(filters.StartDate, "invalid start_date")
	if err != nil {
		return nil, 0, err
	}
	end, err := optionalTime(filters.EndDate, "invalid end_date")
	if err != nil {
		return nil, 0, err
	}

	limit := 50
	if filters.Limit != "" {
		parsed, err := strconv.Atoi(filters.Limit)
		if err != nil || parsed < 1 {
			return nil, 0, errors.New("limit must be a positive integer")
		}
		if parsed > 500 {
			parsed = 500
		}
		limit = parsed
	}
	return s.repository.ListViolations(ctx, vehicleID, geofenceID, start, end, limit)
}

func (s *AlertService) CreateEvent(ctx context.Context, vehicleID uuid.UUID, vehicleNumber, driverName string, geofence models.Geofence, eventType string, input models.LocationRequest) ([]byte, bool, error) {
	geofenceID, err := uuid.Parse(geofence.ID)
	if err != nil {
		return nil, false, fmt.Errorf("parse geofence id: %w", err)
	}
	configured, err := s.repository.HasMatchingConfiguration(ctx, geofenceID, vehicleID, eventType)
	if err != nil || !configured {
		return nil, configured, err
	}

	event := models.AlertEvent{
		EventID:   uuid.NewString(),
		EventType: eventType,
		Timestamp: input.Timestamp,
	}
	event.Vehicle.ID = vehicleID.String()
	event.Vehicle.Number = vehicleNumber
	event.Vehicle.DriverName = driverName
	event.Geofence.ID = geofence.ID
	event.Geofence.Name = geofence.Name
	event.Geofence.Category = geofence.Category
	event.Location.Latitude = input.Latitude
	event.Location.Longitude = input.Longitude
	payload, err := json.Marshal(event)
	if err != nil {
		return nil, false, err
	}
	if err := s.repository.CreateEvent(ctx, event, payload); err != nil {
		return nil, false, err
	}
	return payload, true, nil
}

var ErrNotFound = errors.New("not found")

func optionalUUID(value, message string) (*uuid.UUID, error) {
	if value == "" {
		return nil, nil
	}
	id, err := uuid.Parse(value)
	if err != nil {
		return nil, errors.New(message)
	}
	return &id, nil
}

func optionalTime(value, message string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil, errors.New(message)
	}
	return &parsed, nil
}
