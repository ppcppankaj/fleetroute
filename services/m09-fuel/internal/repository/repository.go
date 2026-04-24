package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FuelLog struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	VehicleID    string    `json:"vehicle_id"`
	DriverID     *string   `json:"driver_id"`
	TripID       *string   `json:"trip_id"`
	Liters       float64   `json:"liters"`
	TotalCost    float64   `json:"total_cost"`
	PricePerLiter *float64 `json:"price_per_liter"`
	Odometer     *float64  `json:"odometer"`
	Station      *string   `json:"station"`
	FuelType     *string   `json:"fuel_type"`
	LoggedAt     time.Time `json:"logged_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type TankReading struct {
	ID         string    `json:"id"`
	VehicleID  string    `json:"vehicle_id"`
	TenantID   string    `json:"tenant_id"`
	LevelPct   float64   `json:"level_pct"`
	Liters     *float64  `json:"liters"`
	RecordedAt time.Time `json:"recorded_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateFuelLog(ctx context.Context, f FuelLog) (FuelLog, error) {
	const q = `
INSERT INTO fuel_logs (tenant_id, vehicle_id, driver_id, trip_id, liters, total_cost, price_per_liter, odometer, station, fuel_type)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $8, $9, $10)
RETURNING id::text, logged_at, created_at`
	err := r.db.QueryRow(ctx, q, f.TenantID, f.VehicleID, nullable(f.DriverID), nullable(f.TripID),
		f.Liters, f.TotalCost, f.PricePerLiter, f.Odometer, f.Station, f.FuelType).
		Scan(&f.ID, &f.LoggedAt, &f.CreatedAt)
	return f, err
}

func (r *Repository) ListFuelLogs(ctx context.Context, tenantID string, limit int) ([]FuelLog, error) {
	const q = `
SELECT id::text, tenant_id::text, vehicle_id::text, driver_id::text, trip_id::text, liters, total_cost,
       price_per_liter, odometer, station, fuel_type, logged_at, created_at
FROM fuel_logs WHERE tenant_id = $1::uuid ORDER BY logged_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fs []FuelLog
	for rows.Next() {
		var f FuelLog
		if err := rows.Scan(&f.ID, &f.TenantID, &f.VehicleID, &f.DriverID, &f.TripID,
			&f.Liters, &f.TotalCost, &f.PricePerLiter, &f.Odometer, &f.Station, &f.FuelType,
			&f.LoggedAt, &f.CreatedAt); err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func (r *Repository) InsertTankReading(ctx context.Context, tr TankReading) error {
	const q = `INSERT INTO fuel_tank_readings (vehicle_id, tenant_id, level_pct, liters) VALUES ($1::uuid, $2::uuid, $3, $4)`
	_, err := r.db.Exec(ctx, q, tr.VehicleID, tr.TenantID, tr.LevelPct, tr.Liters)
	return err
}

func (r *Repository) GetLatestTankReading(ctx context.Context, vehicleID string) (TankReading, error) {
	const q = `SELECT id::text, vehicle_id::text, tenant_id::text, level_pct, liters, recorded_at FROM fuel_tank_readings WHERE vehicle_id = $1::uuid ORDER BY recorded_at DESC LIMIT 1`
	var tr TankReading
	err := r.db.QueryRow(ctx, q, vehicleID).Scan(&tr.ID, &tr.VehicleID, &tr.TenantID, &tr.LevelPct, &tr.Liters, &tr.RecordedAt)
	return tr, err
}

func nullable(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
