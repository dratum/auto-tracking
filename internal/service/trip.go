package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"auto-tracking/internal/domain/model"
	mongorepo "auto-tracking/internal/repository/mongo"
	"auto-tracking/internal/repository/timescale"
)

type TripService struct {
	tripRepo *mongorepo.TripRepo
	gpsRepo  *timescale.GPSRepo
}

func NewTripService(tripRepo *mongorepo.TripRepo, gpsRepo *timescale.GPSRepo) *TripService {
	return &TripService{
		tripRepo: tripRepo,
		gpsRepo:  gpsRepo,
	}
}

// StartTrip creates a new active trip and returns its UUID.
func (s *TripService) StartTrip(ctx context.Context, vehicleID string) (string, error) {
	existing, err := s.tripRepo.GetActiveByVehicleID(ctx, vehicleID)
	if err != nil {
		return "", fmt.Errorf("trip service start: check active: %w", err)
	}
	if existing != nil {
		return "", fmt.Errorf("trip service start: vehicle %s already has active trip %s", vehicleID, existing.ID)
	}

	tripID := uuid.New().String()
	now := time.Now().UTC()

	trip := model.Trip{
		ID:        tripID,
		VehicleID: vehicleID,
		StartTime: now,
		Status:    model.TripStatusActive,
		CreatedAt: now,
	}

	id, err := s.tripRepo.Create(ctx, trip)
	if err != nil {
		return "", fmt.Errorf("trip service start: create: %w", err)
	}

	return id, nil
}

// EndTrip finalizes the active trip for a vehicle.
func (s *TripService) EndTrip(ctx context.Context, vehicleID string) error {
	trip, err := s.tripRepo.GetActiveByVehicleID(ctx, vehicleID)
	if err != nil {
		return fmt.Errorf("trip service end: find active: %w", err)
	}
	if trip == nil {
		return fmt.Errorf("trip service end: no active trip for vehicle %s", vehicleID)
	}

	now := time.Now().UTC()

	if err := s.tripRepo.EndTrip(ctx, trip.ID, now); err != nil {
		return fmt.Errorf("trip service end: %w", err)
	}

	// Compute average speed from all GPS points.
	points, err := s.gpsRepo.FindByTripID(ctx, trip.ID)
	if err != nil {
		return fmt.Errorf("trip service end: get points: %w", err)
	}

	if len(points) > 0 {
		var totalSpeed float64
		for _, p := range points {
			totalSpeed += float64(p.Speed)
		}
		avgSpeed := totalSpeed / float64(len(points))

		if err := s.tripRepo.SetAvgSpeed(ctx, trip.ID, avgSpeed); err != nil {
			return fmt.Errorf("trip service end: set avg speed: %w", err)
		}
	}

	return nil
}

// ListTrips returns paginated completed trips with optional date filtering.
func (s *TripService) ListTrips(ctx context.Context, from, to *time.Time, limit, offset int64) ([]model.Trip, int64, error) {
	trips, err := s.tripRepo.List(ctx, "", from, to, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("trip service list: %w", err)
	}

	total, err := s.tripRepo.Count(ctx, "", from, to)
	if err != nil {
		return nil, 0, fmt.Errorf("trip service count: %w", err)
	}

	return trips, total, nil
}

// GetTrip returns a single trip by ID.
func (s *TripService) GetTrip(ctx context.Context, id string) (*model.Trip, error) {
	trip, err := s.tripRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("trip service get: %w", err)
	}

	return trip, nil
}

// GetTripPoints returns all GPS points for a trip.
func (s *TripService) GetTripPoints(ctx context.Context, tripID string) ([]model.GPSPoint, error) {
	points, err := s.gpsRepo.FindByTripID(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip service get points: %w", err)
	}

	return points, nil
}
