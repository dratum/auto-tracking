package handler

import (
	"context"
	"net/http"

	"auto-tracking/internal/service"
)

type statsService interface {
	GetStats(ctx context.Context, period string) (*service.StatsResult, error)
}

type StatsHandler struct {
	stats statsService
}

func NewStatsHandler(stats statsService) *StatsHandler {
	return &StatsHandler{stats: stats}
}

type statsResponse struct {
	Period           string  `json:"period"`
	TotalDistanceKM  float64 `json:"total_distance_km"`
	TotalTrips       int64   `json:"total_trips"`
	TotalDurationMin float64 `json:"total_duration_min"`
	AvgTripDistKM    float64 `json:"avg_trip_distance_km"`
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}

	validPeriods := map[string]bool{"day": true, "week": true, "month": true, "year": true}
	if !validPeriods[period] {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid period, must be one of: day, week, month, year"})
		return
	}

	result, err := h.stats.GetStats(r.Context(), period)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to compute stats"})
		return
	}

	writeJSON(w, http.StatusOK, statsResponse{
		Period:           result.Period,
		TotalDistanceKM:  result.TotalDistanceKM,
		TotalTrips:       result.TotalTrips,
		TotalDurationMin: result.TotalDurationMin,
		AvgTripDistKM:    result.AvgTripDistKM,
	})
}
