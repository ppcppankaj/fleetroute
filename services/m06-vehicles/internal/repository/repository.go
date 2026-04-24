package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Vehicle struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	PlateNumber string    `json:"plate_number"`
	Make        string    `json:"make"`
	Model       string    `json:"model"`
	Year        int       `json:"year"`
	Color       *string   `json:"color"`
	VIN         string    `json:"vin"`
	FuelType    string    `json:"fuel_type"`
	GroupID     *string   `json:"group_id"`
	Status      string    `json:"status"`
	Odometer    float64   `json:"odometer"`
	Photos      []string  `json:"photos"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateVehicle(ctx context.Context, v Vehicle) (Vehicle, error) {
	const q = `
INSERT INTO vehicles (tenant_id, plate_number, make, model, year, color, vin, fuel_type, group_id, status)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7, $8, $9::uuid, $10)
RETURNING id::text, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, v.TenantID, v.PlateNumber, v.Make, v.Model, v.Year, v.Color, v.VIN, v.FuelType, v.GroupID, v.Status).
		Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}

func (r *Repository) GetVehicle(ctx context.Context, id, tenantID string) (Vehicle, error) {
	const q = `SELECT id::text, tenant_id::text, plate_number, make, model, year, color, vin, fuel_type, group_id::text, status, odometer, photos, created_at, updated_at FROM vehicles WHERE id = $1::uuid AND tenant_id = $2::uuid`
	var v Vehicle
	err := r.db.QueryRow(ctx, q, id, tenantID).Scan(&v.ID, &v.TenantID, &v.PlateNumber, &v.Make, &v.Model, &v.Year, &v.Color, &v.VIN, &v.FuelType, &v.GroupID, &v.Status, &v.Odometer, &v.Photos, &v.CreatedAt, &v.UpdatedAt)
	return v, err
}

func (r *Repository) ListVehicles(ctx context.Context, tenantID string) ([]Vehicle, error) {
	const q = `SELECT id::text, tenant_id::text, plate_number, make, model, year, color, vin, fuel_type, group_id::text, status, odometer, photos, created_at, updated_at FROM vehicles WHERE tenant_id = $1::uuid ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var vs []Vehicle
	for rows.Next() {
		var v Vehicle
		if err := rows.Scan(&v.ID, &v.TenantID, &v.PlateNumber, &v.Make, &v.Model, &v.Year, &v.Color, &v.VIN, &v.FuelType, &v.GroupID, &v.Status, &v.Odometer, &v.Photos, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		vs = append(vs, v)
	}
	return vs, nil
}

func (r *Repository) UpdateVehicleOdometer(ctx context.Context, vehicleID string, distance float64) error {
	const q = `UPDATE vehicles SET odometer = odometer + $1, updated_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, distance, vehicleID)
	return err
}

func (r *Repository) DeleteVehicle(ctx context.Context, id, tenantID string) error {
	const q = `DELETE FROM vehicles WHERE id = $1::uuid AND tenant_id = $2::uuid`
	_, err := r.db.Exec(ctx, q, id, tenantID)
	return err
}
