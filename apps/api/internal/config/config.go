package config

import "os"

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/geofencing?sslmode=disable"

type Config struct {
	APIPort     string
	DatabaseURL string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = os.Getenv("API_PORT")
	}
	if port == "" {
		port = "8080"
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = defaultDatabaseURL
	}

	return Config{
		APIPort:     port,
		DatabaseURL: databaseURL,
	}
}
