package database

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/yourusername/geofencing-platform/apps/api/internal/geospatial"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func New(ctx context.Context, connectionString string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return db, nil
}

func Migrate(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Exec(`CREATE EXTENSION IF NOT EXISTS postgis;`).Error; err != nil {
		return err
	}
	if err := db.WithContext(ctx).Exec(`
		CREATE TABLE IF NOT EXISTS geofences (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			coordinates JSONB NOT NULL,
			category TEXT NOT NULL CHECK (category IN ('delivery_zone', 'restricted_zone', 'toll_zone', 'customer_area')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		ALTER TABLE geofences
			ADD COLUMN IF NOT EXISTS boundary geometry(Polygon, 4326);

		CREATE TABLE IF NOT EXISTS vehicles (
			id UUID PRIMARY KEY,
			vehicle_number TEXT NOT NULL UNIQUE,
			driver_name TEXT NOT NULL,
			vehicle_type TEXT NOT NULL,
			phone TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS vehicle_locations (
			id UUID PRIMARY KEY,
			vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			recorded_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS alert_configurations (
			id UUID PRIMARY KEY,
			geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
			vehicle_id UUID REFERENCES vehicles(id) ON DELETE CASCADE,
			event_type TEXT NOT NULL CHECK (event_type IN ('entry', 'exit', 'both')),
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE TABLE IF NOT EXISTS violations (
			id UUID PRIMARY KEY,
			vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
			geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
			event_type TEXT NOT NULL CHECK (event_type IN ('entry', 'exit')),
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			recorded_at TIMESTAMPTZ NOT NULL
		);

		CREATE TABLE IF NOT EXISTS alert_events (
			id UUID PRIMARY KEY,
			vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
			geofence_id UUID NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
			event_type TEXT NOT NULL CHECK (event_type IN ('entry', 'exit')),
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			occurred_at TIMESTAMPTZ NOT NULL,
			payload JSONB NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

		CREATE INDEX IF NOT EXISTS vehicle_locations_vehicle_time_idx
			ON vehicle_locations (vehicle_id, recorded_at DESC);
		CREATE INDEX IF NOT EXISTS violations_vehicle_time_idx
			ON violations (vehicle_id, recorded_at DESC);
		CREATE INDEX IF NOT EXISTS alert_events_occurred_at_idx
			ON alert_events (occurred_at DESC);
		CREATE INDEX IF NOT EXISTS geofences_boundary_gist_idx
			ON geofences USING GIST (boundary);
	`).Error; err != nil {
		return err
	}

	var geofences []migrationGeofence
	if err := db.WithContext(ctx).Table("geofences").Select("id, coordinates").Where("boundary IS NULL").Find(&geofences).Error; err != nil {
		return err
	}
	for _, geofence := range geofences {
		var coordinates [][]float64
		if err := json.Unmarshal(geofence.Coordinates, &coordinates); err != nil {
			continue
		}
		wkt, err := geospatial.PolygonWKT(coordinates)
		if err != nil {
			continue
		}
		if err := db.WithContext(ctx).Exec(
			"UPDATE geofences SET boundary = ST_SetSRID(ST_GeomFromText(?), 4326) WHERE id = ? AND boundary IS NULL",
			wkt,
			geofence.ID,
		).Error; err != nil {
			return err
		}
	}
	return nil
}

type migrationGeofence struct {
	ID          uuid.UUID      `gorm:"column:id"`
	Coordinates datatypes.JSON `gorm:"column:coordinates"`
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
