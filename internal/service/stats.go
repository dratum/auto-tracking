package service

import (
	"context"
	"fmt"
	"time"

	mongorepo "auto-tracking/internal/repository/mongo"
)

type StatsService struct {
	tripRepo *mongorepo.TripRepo
}

func NewStatsService(tripRepo *mongorepo.TripRepo) *StatsService {
	return &StatsService{tripRepo: tripRepo}
}

type StatsResult struct {
	Period           string
	TotalDistanceKM  float64
	TotalTrips       int64
	TotalDurationMin float64
	AvgTripDistKM    float64
}

func (s *StatsService) GetStats(ctx context.Context, period string) (*StatsResult, error) {
	now := time.Now().UTC()
	var from time.Time

	switch period {
	case "day":
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case "week":
		weekday := now.Weekday()
		if weekday == time.Sunday {
			weekday = 7
		}
		from = time.Date(now.Year(), now.Month(), now.Day()-int(weekday-time.Monday), 0, 0, 0, 0, time.UTC)
	case "month":
		from = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	case "year":
		from = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
	default:
		return nil, fmt.Errorf("stats service: unsupported period %q", period)
	}

	stats, err := s.tripRepo.AggregateStats(ctx, from, now)
	if err != nil {
		return nil, fmt.Errorf("stats service: %w", err)
	}

	totalDurationMin := float64(stats.TotalDurationMs) / 60000.0
	var avgDist float64
	if stats.TotalTrips > 0 {
		avgDist = stats.TotalDistanceKM / float64(stats.TotalTrips)
	}

	return &StatsResult{
		Period:           period,
		TotalDistanceKM:  stats.TotalDistanceKM,
		TotalTrips:       stats.TotalTrips,
		TotalDurationMin: totalDurationMin,
		AvgTripDistKM:    avgDist,
	}, nil
}
