package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"auto-tracking/internal/domain/model"
)

type tripQueryService interface {
	ListTrips(ctx context.Context, from, to *time.Time, limit, offset int64) ([]model.Trip, int64, error)
	GetTrip(ctx context.Context, id string) (*model.Trip, error)
	GetTripPoints(ctx context.Context, tripID string) ([]model.GPSPoint, error)
}

type TripHandler struct {
	trips tripQueryService
}

func NewTripHandler(trips tripQueryService) *TripHandler {
	return &TripHandler{trips: trips}
}

type tripListResponse struct {
	Trips []tripResponse `json:"trips"`
	Total int64          `json:"total"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

type tripResponse struct {
	ID          string     `json:"id"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	DistanceKM  float64    `json:"distance_km"`
	DurationMin float64    `json:"duration_min"`
	MaxSpeed    float64    `json:"max_speed"`
	AvgSpeed    float64    `json:"avg_speed"`
}

type pointsResponse struct {
	Points []pointDTO `json:"points"`
}

type pointDTO struct {
	Lat   float64   `json:"lat"`
	Lon   float64   `json:"lon"`
	Speed float32   `json:"speed"`
	Time  time.Time `json:"time"`
}

func tripToResponse(t model.Trip) tripResponse {
	var durationMin float64
	if t.EndTime != nil {
		durationMin = t.EndTime.Sub(t.StartTime).Minutes()
	}

	return tripResponse{
		ID:          t.ID,
		StartTime:   t.StartTime,
		EndTime:     t.EndTime,
		DistanceKM:  t.DistanceKM,
		DurationMin: durationMin,
		MaxSpeed:    t.MaxSpeed,
		AvgSpeed:    t.AvgSpeed,
	}
}

func (h *TripHandler) ListTrips(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := int64((page - 1) * limit)

	var from, to *time.Time
	if fromStr := r.URL.Query().Get("from"); fromStr != "" {
		t, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid 'from' date format, expected YYYY-MM-DD"})
			return
		}
		from = &t
	}

	if toStr := r.URL.Query().Get("to"); toStr != "" {
		t, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid 'to' date format, expected YYYY-MM-DD"})
			return
		}
		endOfDay := t.Add(24*time.Hour - time.Nanosecond)
		to = &endOfDay
	}

	trips, total, err := h.trips.ListTrips(r.Context(), from, to, int64(limit), offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list trips"})
		return
	}

	resp := tripListResponse{
		Trips: make([]tripResponse, 0, len(trips)),
		Total: total,
		Page:  page,
		Limit: limit,
	}

	for _, t := range trips {
		resp.Trips = append(resp.Trips, tripToResponse(t))
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *TripHandler) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "trip id is required"})
		return
	}

	trip, err := h.trips.GetTrip(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get trip"})
		return
	}
	if trip == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "trip not found"})
		return
	}

	writeJSON(w, http.StatusOK, tripToResponse(*trip))
}

func (h *TripHandler) GetTripPoints(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "trip id is required"})
		return
	}

	points, err := h.trips.GetTripPoints(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get trip points"})
		return
	}

	dtos := make([]pointDTO, 0, len(points))
	for _, p := range points {
		dtos = append(dtos, pointDTO{
			Lat:   p.Lat,
			Lon:   p.Lon,
			Speed: p.Speed,
			Time:  p.Time,
		})
	}

	writeJSON(w, http.StatusOK, pointsResponse{Points: dtos})
}
