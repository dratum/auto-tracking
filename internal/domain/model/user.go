package model

import "time"

type User struct {
	ID           string    `json:"id"         bson:"_id,omitempty"`
	Username     string    `json:"username"   bson:"username"`
	PasswordHash string    `json:"-"          bson:"password_hash"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
}
