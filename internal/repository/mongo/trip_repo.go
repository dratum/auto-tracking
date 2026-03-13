package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"auto-tracking/internal/domain/model"
)

type TripRepo struct {
	col *mongo.Collection
}

func NewTripRepo(db *mongo.Database) *TripRepo {
	return &TripRepo{col: db.Collection("trips")}
}

func (r *TripRepo) Create(ctx context.Context, trip model.Trip) (string, error) {
	res, err := r.col.InsertOne(ctx, trip)
	if err != nil {
		return "", fmt.Errorf("trip_repo create: %w", err)
	}

	id, ok := res.InsertedID.(string)
	if !ok {
		return "", fmt.Errorf("trip_repo create: unexpected id type %T", res.InsertedID)
	}

	return id, nil
}

func (r *TripRepo) GetByID(ctx context.Context, id string) (*model.Trip, error) {
	var trip model.Trip

	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&trip)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("trip_repo get by id: %w", err)
	}

	return &trip, nil
}

func (r *TripRepo) GetActiveByVehicleID(ctx context.Context, vehicleID string) (*model.Trip, error) {
	var trip model.Trip

	filter := bson.M{
		"vehicle_id": vehicleID,
		"status":     model.TripStatusActive,
	}
	err := r.col.FindOne(ctx, filter).Decode(&trip)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("trip_repo get active: %w", err)
	}

	return &trip, nil
}

func (r *TripRepo) EndTrip(ctx context.Context, tripID string, endTime time.Time) error {
	filter := bson.M{"_id": tripID}
	update := bson.M{
		"$set": bson.M{
			"end_time": endTime,
			"status":   model.TripStatusCompleted,
		},
	}

	res, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("trip_repo end trip: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("trip_repo end trip: trip %s not found", tripID)
	}

	return nil
}

// UpdateStats increments distance and updates max speed atomically.
func (r *TripRepo) UpdateStats(ctx context.Context, tripID string, addDistKM float64, pointSpeed float64) error {
	filter := bson.M{"_id": tripID}
	pipeline := bson.A{
		bson.M{
			"$set": bson.M{
				"distance_km": bson.M{"$add": bson.A{"$distance_km", addDistKM}},
				"max_speed":   bson.M{"$max": bson.A{"$max_speed", pointSpeed}},
			},
		},
	}

	res, err := r.col.UpdateOne(ctx, filter, pipeline)
	if err != nil {
		return fmt.Errorf("trip_repo update stats: %w", err)
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("trip_repo update stats: trip %s not found", tripID)
	}

	return nil
}

func (r *TripRepo) SetAvgSpeed(ctx context.Context, tripID string, avgSpeed float64) error {
	filter := bson.M{"_id": tripID}
	update := bson.M{"$set": bson.M{"avg_speed": avgSpeed}}

	_, err := r.col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("trip_repo set avg speed: %w", err)
	}

	return nil
}

func (r *TripRepo) List(ctx context.Context, vehicleID string, from, to *time.Time, limit, offset int64) ([]model.Trip, error) {
	filter := bson.M{"status": model.TripStatusCompleted}
	if vehicleID != "" {
		filter["vehicle_id"] = vehicleID
	}
	if from != nil || to != nil {
		timeFilter := bson.M{}
		if from != nil {
			timeFilter["$gte"] = *from
		}
		if to != nil {
			timeFilter["$lte"] = *to
		}
		filter["start_time"] = timeFilter
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "start_time", Value: -1}}).
		SetLimit(limit).
		SetSkip(offset)

	cursor, err := r.col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("trip_repo list: %w", err)
	}
	defer cursor.Close(ctx)

	var trips []model.Trip
	if err := cursor.All(ctx, &trips); err != nil {
		return nil, fmt.Errorf("trip_repo list decode: %w", err)
	}

	return trips, nil
}

func (r *TripRepo) Count(ctx context.Context, vehicleID string, from, to *time.Time) (int64, error) {
	filter := bson.M{"status": model.TripStatusCompleted}
	if vehicleID != "" {
		filter["vehicle_id"] = vehicleID
	}
	if from != nil || to != nil {
		timeFilter := bson.M{}
		if from != nil {
			timeFilter["$gte"] = *from
		}
		if to != nil {
			timeFilter["$lte"] = *to
		}
		filter["start_time"] = timeFilter
	}

	count, err := r.col.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("trip_repo count: %w", err)
	}

	return count, nil
}

type TripStats struct {
	TotalDistanceKM float64 `bson:"total_distance_km"`
	TotalTrips      int64   `bson:"total_trips"`
	TotalDurationMs int64   `bson:"total_duration_ms"`
}

func (r *TripRepo) AggregateStats(ctx context.Context, from, to time.Time) (*TripStats, error) {
	pipeline := bson.A{
		bson.M{"$match": bson.M{
			"status":     model.TripStatusCompleted,
			"start_time": bson.M{"$gte": from, "$lte": to},
			"end_time":   bson.M{"$ne": nil},
		}},
		bson.M{"$group": bson.M{
			"_id":               nil,
			"total_distance_km": bson.M{"$sum": "$distance_km"},
			"total_trips":       bson.M{"$sum": 1},
			"total_duration_ms": bson.M{"$sum": bson.M{
				"$subtract": bson.A{"$end_time", "$start_time"},
			}},
		}},
	}

	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("trip_repo aggregate stats: %w", err)
	}
	defer cursor.Close(ctx)

	var results []TripStats
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("trip_repo aggregate stats decode: %w", err)
	}

	if len(results) == 0 {
		return &TripStats{}, nil
	}

	return &results[0], nil
}
