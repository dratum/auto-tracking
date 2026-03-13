package timescale

import (
	"context"
	"database/sql"
	"fmt"

	"auto-tracking/internal/domain/model"
)

type GPSRepo struct {
	db *sql.DB
}

func NewGPSRepo(db *sql.DB) *GPSRepo {
	return &GPSRepo{db: db}
}

func (r *GPSRepo) Insert(ctx context.Context, p model.GPSPoint) error {
	const query = `
		insert into gps_points (
								time
							  , trip_id
							  , lat
							  , lon
							  , speed
							  , heading
							  , satellites
							  )
						values (
								@time
							  , @trip_id
							  , @lat
							  , @lon
							  , @speed
							  , @heading
							  , @satellites
							  )
	`

	_, err := r.db.ExecContext(ctx, query,
		sql.Named("time", p.Time),
		sql.Named("trip_id", p.TripID),
		sql.Named("lat", p.Lat),
		sql.Named("lon", p.Lon),
		sql.Named("speed", p.Speed),
		sql.Named("heading", p.Heading),
		sql.Named("satellites", p.Satellites),
	)
	if err != nil {
		return fmt.Errorf("gps_repo insert: %w", err)
	}

	return nil
}

func (r *GPSRepo) FindByTripID(ctx context.Context, tripID string) ([]model.GPSPoint, error) {
	const query = `
		select
			  time
			, trip_id
			, lat
			, lon
			, speed
			, heading
			, satellites
		from gps_points
		where trip_id = $1
		order by time asc`

	rows, err := r.db.QueryContext(ctx, query, tripID)
	if err != nil {
		return nil, fmt.Errorf("gps_repo find by trip: %w", err)
	}
	defer rows.Close()

	var points []model.GPSPoint
	for rows.Next() {
		var p model.GPSPoint
		if err := rows.Scan(&p.Time, &p.TripID, &p.Lat, &p.Lon, &p.Speed, &p.Heading, &p.Satellites); err != nil {
			return nil, fmt.Errorf("gps_repo scan: %w", err)
		}
		points = append(points, p)
	}

	return points, rows.Err()
}

// LastByTripID returns the most recent GPS point for a trip, or nil if none exist.
func (r *GPSRepo) LastByTripID(ctx context.Context, tripID string) (*model.GPSPoint, error) {
	const query = `
		select
			  time
			, trip_id
			, lat
			, lon
			, speed
			, heading
			, satellites
		from gps_points
		where trip_id = $1
		order by time desc
		limit 1`

	var p model.GPSPoint
	err := r.db.QueryRowContext(ctx, query, tripID).Scan(
		&p.Time, &p.TripID, &p.Lat, &p.Lon, &p.Speed, &p.Heading, &p.Satellites,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("gps_repo last by trip: %w", err)
	}

	return &p, nil
}
