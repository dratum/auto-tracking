package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"auto-tracking/internal/api"
	"auto-tracking/internal/api/handler"
	"auto-tracking/internal/config"
	"auto-tracking/internal/service"
	mongorepo "auto-tracking/internal/repository/mongo"
	"auto-tracking/internal/repository/timescale"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("fatal: %v", err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect to TimescaleDB
	pgDB, err := sql.Open("postgres", cfg.Timescale.DSN())
	if err != nil {
		return fmt.Errorf("open TimescaleDB: %w", err)
	}
	defer pgDB.Close()

	if err := pgDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping TimescaleDB: %w", err)
	}
	log.Println("connected to TimescaleDB")

	if err := timescale.InitSchema(ctx, pgDB); err != nil {
		return fmt.Errorf("init TimescaleDB: %w", err)
	}
	log.Println("TimescaleDB schema initialized")

	// Connect to MongoDB
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return fmt.Errorf("connect MongoDB: %w", err)
	}
	defer func() {
		disconnectCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		mongoClient.Disconnect(disconnectCtx)
	}()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping MongoDB: %w", err)
	}
	log.Println("connected to MongoDB")

	mongoDB := mongoClient.Database(cfg.Mongo.DB)
	if err := mongorepo.InitSchema(ctx, mongoDB); err != nil {
		return fmt.Errorf("init mongodb: %w", err)
	}
	log.Println("MongoDB indexes initialized")

	// Build dependency graph
	gpsRepo := timescale.NewGPSRepo(pgDB)
	tripRepo := mongorepo.NewTripRepo(mongoDB)

	trackingService := service.NewTrackingService(gpsRepo, tripRepo)
	tripService := service.NewTripService(tripRepo, gpsRepo)

	const defaultVehicleID = "default"
	deviceHandler := handler.NewDeviceHandler(trackingService, tripService, defaultVehicleID)

	// HTTP server
	router := api.NewRouter(deviceHandler, cfg.Auth.APIKey)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("server listening on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("received signal %v, shutting down...", sig)
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	log.Println("server stopped gracefully")
	return nil
}
