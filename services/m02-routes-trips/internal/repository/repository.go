package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Trip struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	VehicleID   string     `json:"vehicle_id"`
	DriverID    *string    `json:"driver_id"`
	Status      string     `json:"status"`
	StartLat    *float64   `json:"start_lat"`
	StartLng    *float64   `json:"start_lng"`
	EndLat      *float64   `json:"end_lat"`
	EndLng      *float64   `json:"end_lng"`
	DistanceKM  float64    `json:"distance_km"`
	FuelUsed    float64    `json:"fuel_used"`
	DurationSec int        `json:"duration_sec"`
	StartedAt   time.Time  `json:"started_at"`
	EndedAt     *time.Time `json:"ended_at"`
}

type Route struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Waypoints  any       `json:"waypoints"`
	DistanceKM float64   `json:"distance_km"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) StartTrip(ctx context.Context, t Trip) (Trip, error) {
	const q = `
INSERT INTO trips (tenant_id, vehicle_id, driver_id, start_lat, start_lng, status)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, 'IN_PROGRESS')
RETURNING id::text, started_at`
	err := r.db.QueryRow(ctx, q, t.TenantID, t.VehicleID, nullable(t.DriverID), t.StartLat, t.StartLng).
		Scan(&t.ID, &t.StartedAt)
	return t, err
}

func (r *Repository) EndTrip(ctx context.Context, tripID string, endLat, endLng, distKM, fuelUsed float64, durationSec int) error {
	const q = `
UPDATE trips SET
  status = 'COMPLETED', end_lat = $1, end_lng = $2,
  distance_km = $3, fuel_used = $4, duration_sec = $5,
  ended_at = NOW()
WHERE id = $6::uuid`
	_, err := r.db.Exec(ctx, q, endLat, endLng, distKM, fuelUsed, durationSec, tripID)
	return err
}

func (r *Repository) GetActiveTrip(ctx context.Context, vehicleID, tenantID string) (Trip, error) {
	const q = `
SELECT id::text, tenant_id::text, vehicle_id::text, driver_id::text, status,
       start_lat, start_lng, end_lat, end_lng, distance_km, fuel_used, duration_sec, started_at, ended_at
FROM trips WHERE vehicle_id = $1::uuid AND tenant_id = $2::uuid AND status = 'IN_PROGRESS'
LIMIT 1`
	var t Trip
	err := r.db.QueryRow(ctx, q, vehicleID, tenantID).Scan(
		&t.ID, &t.TenantID, &t.VehicleID, &t.DriverID, &t.Status,
		&t.StartLat, &t.StartLng, &t.EndLat, &t.EndLng,
		&t.DistanceKM, &t.FuelUsed, &t.DurationSec, &t.StartedAt, &t.EndedAt,
	)
	return t, err
}

func (r *Repository) ListTrips(ctx context.Context, tenantID string, limit int) ([]Trip, error) {
	const q = `
SELECT id::text, tenant_id::text, vehicle_id::text, driver_id::text, status,
       start_lat, start_lng, end_lat, end_lng, distance_km, fuel_used, duration_sec, started_at, ended_at
FROM trips WHERE tenant_id = $1::uuid ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ts []Trip
	for rows.Next() {
		var t Trip
		if err := rows.Scan(&t.ID, &t.TenantID, &t.VehicleID, &t.DriverID, &t.Status,
			&t.StartLat, &t.StartLng, &t.EndLat, &t.EndLng,
			&t.DistanceKM, &t.FuelUsed, &t.DurationSec, &t.StartedAt, &t.EndedAt); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func (r *Repository) CreateRoute(ctx context.Context, ro Route) (Route, error) {
	const q = `
INSERT INTO routes (tenant_id, name, waypoints, distance_km)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text, created_at`
	err := r.db.QueryRow(ctx, q, ro.TenantID, ro.Name, ro.Waypoints, ro.DistanceKM).Scan(&ro.ID, &ro.CreatedAt)
	return ro, err
}

func (r *Repository) ListRoutes(ctx context.Context, tenantID string) ([]Route, error) {
	const q = `SELECT id::text, tenant_id::text, name, waypoints, distance_km, created_at FROM routes WHERE tenant_id = $1::uuid ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rs []Route
	for rows.Next() {
		var ro Route
		if err := rows.Scan(&ro.ID, &ro.TenantID, &ro.Name, &ro.Waypoints, &ro.DistanceKM, &ro.CreatedAt); err != nil {
			return nil, err
		}
		rs = append(rs, ro)
	}
	return rs, nil
}

func nullable(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
