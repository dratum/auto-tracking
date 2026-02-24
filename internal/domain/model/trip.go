package model

import "time"

type TripStatus string

const (
	TripStatusActive    TripStatus = "active"
	TripStatusCompleted TripStatus = "completed"
)

type Trip struct {
	ID         string     `json:"id"          bson:"_id,omitempty"`
	VehicleID  string     `json:"vehicle_id"  bson:"vehicle_id"`
	StartTime  time.Time  `json:"start_time"  bson:"start_time"`
	EndTime    *time.Time `json:"end_time"    bson:"end_time,omitempty"`
	DistanceKM float64    `json:"distance_km" bson:"distance_km"`
	MaxSpeed   float64    `json:"max_speed"   bson:"max_speed"`
	AvgSpeed   float64    `json:"avg_speed"   bson:"avg_speed"`
	Status     TripStatus `json:"status"      bson:"status"`
	CreatedAt  time.Time  `json:"created_at"  bson:"created_at"`
}
