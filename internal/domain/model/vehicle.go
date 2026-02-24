package model

import "time"

type Vehicle struct {
	ID          string    `json:"id"           bson:"_id,omitempty"`
	Name        string    `json:"name"         bson:"name"`
	PlateNumber string    `json:"plate_number" bson:"plate_number"`
	CreatedAt   time.Time `json:"created_at"   bson:"created_at"`
}
