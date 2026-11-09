package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(ctx context.Context, connectionString string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, connectionString)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS geofences (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			coordinates JSONB NOT NULL,
			category TEXT NOT NULL CHECK (category IN ('delivery_zone', 'restricted_zone', 'toll_zone', 'customer_area')),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);

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
	`)
	return err
}
