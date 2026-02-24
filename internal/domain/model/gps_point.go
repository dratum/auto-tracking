package model

import "time"

type GPSPoint struct {
	Time       time.Time `json:"time"       db:"time"`
	TripID     string    `json:"trip_id"    db:"trip_id"`
	Lat        float64   `json:"lat"        db:"lat"`
	Lon        float64   `json:"lon"        db:"lon"`
	Speed      float32   `json:"speed"      db:"speed"`
	Heading    float32   `json:"heading"    db:"heading"`
	Satellites int16     `json:"satellites" db:"satellites"`
}
