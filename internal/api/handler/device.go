package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"auto-tracking/internal/domain/model"
)

type trackingService interface {
	SavePoint(ctx context.Context, point model.GPSPoint) error
}

type tripService interface {
	EndTrip(ctx context.Context, vehicleID string) error
	StartTrip(ctx context.Context, vehicleID string) (string, error)
}

type DeviceHandler struct {
	tracking  trackingService
	trip      tripService
	vehicleID string
}

func NewDeviceHandler(
	tracking trackingService, trip tripService, vehicleID string,
) *DeviceHandler {
	return &DeviceHandler{
		tracking:  tracking,
		trip:      trip,
		vehicleID: vehicleID,
	}
}

type LocationRequest struct {
	Lat        float64 `json:"lat"`
	Lon        float64 `json:"lon"`
	Speed      float32 `json:"speed"`
	Heading    float32 `json:"heading"`
	Satellites int16   `json:"satellites"`
	Timestamp  string  `json:"timestamp"`
	TripID     string  `json:"trip_id"`
}

func (h *DeviceHandler) PostLocation(w http.ResponseWriter, r *http.Request) {
	var req LocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.TripID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "trip_id is required"})
		return
	}

	ts, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid timestamp format, expected RFC3339"})
		return
	}

	point := model.GPSPoint{
		Time:       ts,
		TripID:     req.TripID,
		Lat:        req.Lat,
		Lon:        req.Lon,
		Speed:      req.Speed,
		Heading:    req.Heading,
		Satellites: req.Satellites,
	}

	if err := h.tracking.SavePoint(r.Context(), point); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to save location"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *DeviceHandler) PostTripStart(w http.ResponseWriter, r *http.Request) {
	tripID, err := h.trip.StartTrip(r.Context(), h.vehicleID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to start trip"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"trip_id": tripID})
}

func (h *DeviceHandler) PostTripEnd(w http.ResponseWriter, r *http.Request) {
	if err := h.trip.EndTrip(r.Context(), h.vehicleID); err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to end trip"})
		return
	}

	w.WriteHeader(http.StatusOK)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
