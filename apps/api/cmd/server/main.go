package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourusername/geofencing-platform/apps/api/internal/config"
	"github.com/yourusername/geofencing-platform/apps/api/internal/database"
	"github.com/yourusername/geofencing-platform/apps/api/internal/handlers"
	"github.com/yourusername/geofencing-platform/apps/api/internal/realtime"
	"github.com/yourusername/geofencing-platform/apps/api/internal/routes"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()

	db, err := database.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Printf("close database: %v", err)
		}
	}()

	if err := database.Migrate(context.Background(), db); err != nil {
		log.Fatalf("run database migrations: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.APIPort,
		Handler:           routes.New(handlers.New(db, realtime.NewHub())),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("API listening on http://localhost:%s", cfg.APIPort)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			log.Printf("shutdown API: %v", err)
		}
	}
}
