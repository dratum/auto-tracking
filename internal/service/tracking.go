package service

import (
	"context"
	"fmt"

	"auto-tracking/internal/domain/geo"
	"auto-tracking/internal/domain/model"
	mongorepo "auto-tracking/internal/repository/mongo"
	"auto-tracking/internal/repository/timescale"
)

type TrackingService struct {
	gpsRepo  *timescale.GPSRepo
	tripRepo *mongorepo.TripRepo
}

func NewTrackingService(gpsRepo *timescale.GPSRepo, tripRepo *mongorepo.TripRepo) *TrackingService {
	return &TrackingService{
		gpsRepo:  gpsRepo,
		tripRepo: tripRepo,
	}
}

// SavePoint stores a GPS point and incrementally updates trip distance and max speed.
func (s *TrackingService) SavePoint(ctx context.Context, point model.GPSPoint) error {
	lastPoint, err := s.gpsRepo.LastByTripID(ctx, point.TripID)
	if err != nil {
		return fmt.Errorf("tracking: get last point: %w", err)
	}

	if err := s.gpsRepo.Insert(ctx, point); err != nil {
		return fmt.Errorf("tracking: insert point: %w", err)
	}

	if lastPoint != nil {
		distKM := geo.HaversineDistance(lastPoint.Lat, lastPoint.Lon, point.Lat, point.Lon)
		if err := s.tripRepo.UpdateStats(ctx, point.TripID, distKM, float64(point.Speed)); err != nil {
			return fmt.Errorf("tracking: update stats: %w", err)
		}
	}

	return nil
}
