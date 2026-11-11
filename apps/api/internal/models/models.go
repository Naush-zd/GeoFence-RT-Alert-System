package models

import "time"

type GeofenceRequest struct {
	Name        string      `json:"name" binding:"required"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates" binding:"required,min=4,dive,len=2"`
	Category    string      `json:"category" binding:"required,oneof=delivery_zone restricted_zone toll_zone customer_area"`
}

type VehicleRequest struct {
	VehicleNumber string `json:"vehicle_number" binding:"required"`
	DriverName    string `json:"driver_name" binding:"required"`
	VehicleType   string `json:"vehicle_type" binding:"required"`
	Phone         string `json:"phone" binding:"required"`
}

type Geofence struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Coordinates [][]float64 `json:"coordinates"`
	Category    string      `json:"category"`
	CreatedAt   time.Time   `json:"created_at"`
}

type Vehicle struct {
	ID            string    `json:"id"`
	VehicleNumber string    `json:"vehicle_number"`
	DriverName    string    `json:"driver_name"`
	VehicleType   string    `json:"vehicle_type"`
	Phone         string    `json:"phone"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type LocationRequest struct {
	VehicleID string    `json:"vehicle_id" binding:"required,uuid"`
	Latitude  float64   `json:"latitude" binding:"latitude"`
	Longitude float64   `json:"longitude" binding:"longitude"`
	Timestamp time.Time `json:"timestamp" binding:"required"`
}

type CurrentGeofence struct {
	GeofenceID   string `json:"geofence_id"`
	GeofenceName string `json:"geofence_name"`
	Category     string `json:"category,omitempty"`
	Status       string `json:"status,omitempty"`
}

type LocationResponse struct {
	VehicleID        string            `json:"vehicle_id"`
	VehicleNumber    string            `json:"vehicle_number,omitempty"`
	LocationUpdated  bool              `json:"location_updated,omitempty"`
	CurrentLocation  *LocationRequest  `json:"current_location,omitempty"`
	CurrentGeofences []CurrentGeofence `json:"current_geofences"`
}

type AlertConfigRequest struct {
	GeofenceID string `json:"geofence_id" binding:"required,uuid"`
	VehicleID  string `json:"vehicle_id" binding:"omitempty,uuid"`
	EventType  string `json:"event_type" binding:"required,oneof=entry exit both"`
}

type AlertConfig struct {
	AlertID       string    `json:"alert_id"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name,omitempty"`
	VehicleID     string    `json:"vehicle_id,omitempty"`
	VehicleNumber string    `json:"vehicle_number,omitempty"`
	EventType     string    `json:"event_type"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Violation struct {
	ID            string    `json:"id"`
	VehicleID     string    `json:"vehicle_id"`
	VehicleNumber string    `json:"vehicle_number"`
	GeofenceID    string    `json:"geofence_id"`
	GeofenceName  string    `json:"geofence_name"`
	EventType     string    `json:"event_type"`
	Latitude      float64   `json:"latitude"`
	Longitude     float64   `json:"longitude"`
	Timestamp     time.Time `json:"timestamp"`
}

type AlertEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
	Vehicle   struct {
		ID         string `json:"vehicle_id"`
		Number     string `json:"vehicle_number"`
		DriverName string `json:"driver_name"`
	} `json:"vehicle"`
	Geofence struct {
		ID       string `json:"geofence_id"`
		Name     string `json:"geofence_name"`
		Category string `json:"category"`
	} `json:"geofence"`
	Location struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"location"`
}
