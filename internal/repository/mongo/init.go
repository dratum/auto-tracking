package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func InitSchema(ctx context.Context, db *mongo.Database) error {
	tripsCol := db.Collection("trips")
	tripsIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "vehicle_id", Value: 1}, {Key: "start_time", Value: -1}},
			Options: options.Index().SetName("idx_trips_vehicle_time"),
		},
		{
			Keys:    bson.D{{Key: "status", Value: 1}},
			Options: options.Index().SetName("idx_trips_status"),
		},
	}
	if _, err := tripsCol.Indexes().CreateMany(ctx, tripsIndexes); err != nil {
		return fmt.Errorf("mongo init trips indexes: %w", err)
	}

	usersCol := db.Collection("users")
	usersIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetName("idx_users_username").SetUnique(true),
		},
	}
	if _, err := usersCol.Indexes().CreateMany(ctx, usersIndexes); err != nil {
		return fmt.Errorf("mongo init users indexes: %w", err)
	}

	return nil
}
