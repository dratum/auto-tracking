package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/crypto/bcrypt"

	"auto-tracking/internal/api"
	"auto-tracking/internal/api/handler"
	"auto-tracking/internal/config"
	"auto-tracking/internal/domain/model"
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
	pgPool, err := pgxpool.New(ctx, cfg.Timescale.DSN())
	if err != nil {
		return fmt.Errorf("open TimescaleDB: %w", err)
	}
	defer pgPool.Close()

	if err := pgPool.Ping(ctx); err != nil {
		return fmt.Errorf("ping TimescaleDB: %w", err)
	}
	log.Println("connected to TimescaleDB")

	if err := timescale.InitSchema(ctx, pgPool); err != nil {
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
	gpsRepo := timescale.NewGPSRepo(pgPool)
	tripRepo := mongorepo.NewTripRepo(mongoDB)
	userRepo := mongorepo.NewUserRepo(mongoDB)

	trackingService := service.NewTrackingService(gpsRepo, tripRepo)
	tripService := service.NewTripService(tripRepo, gpsRepo)
	statsService := service.NewStatsService(tripRepo)

	const defaultVehicleID = "1"
	deviceHandler := handler.NewDeviceHandler(trackingService, tripService, defaultVehicleID)
	authHandler := handler.NewAuthHandler(userRepo, cfg.Auth.JWTSecret, cfg.Auth.JWTExpiry)
	tripHandler := handler.NewTripHandler(tripService)
	statsHandler := handler.NewStatsHandler(statsService)

	// Seed default admin user
	if err := seedAdminUser(ctx, userRepo, cfg.Admin); err != nil {
		return fmt.Errorf("seed admin: %w", err)
	}

	// Static files (SPA)
	var webFS fs.FS
	if info, err := os.Stat("web/build"); err == nil && info.IsDir() {
		webFS = os.DirFS("web/build")
		log.Println("serving SPA from web/build")
	}

	// HTTP server
	router := api.NewRouter(
		deviceHandler,
		authHandler,
		tripHandler,
		statsHandler,
		cfg.Auth.APIKey,
		cfg.Auth.JWTSecret,
		webFS,
	)

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

func seedAdminUser(ctx context.Context, userRepo *mongorepo.UserRepo, admin config.AdminConfig) error {
	existing, err := userRepo.GetByUsername(ctx, admin.Username)
	if err != nil {
		return fmt.Errorf("check admin user: %w", err)
	}
	if existing != nil {
		log.Printf("admin user %q already exists, skipping seed", admin.Username)
		return nil
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	user := model.User{
		ID:           uuid.New().String(),
		Username:     admin.Username,
		PasswordHash: string(hash),
		CreatedAt:    time.Now().UTC(),
	}

	if err := userRepo.Create(ctx, user); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	log.Printf("admin user %q created", admin.Username)
	return nil
}
