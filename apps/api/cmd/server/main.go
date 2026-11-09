package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/geofencing-platform/apps/api/internal/config"
	"github.com/yourusername/geofencing-platform/apps/api/internal/database"
	"github.com/yourusername/geofencing-platform/apps/api/internal/handlers"
	"github.com/yourusername/geofencing-platform/apps/api/internal/routes"
)

func main() {
	ctx := context.Background()
	cfg := config.Load()

	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		log.Fatalf("run database migrations: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           withJSONContentType(routes.New(handlers.New(db))),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", cfg.APIPort)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func withJSONContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
