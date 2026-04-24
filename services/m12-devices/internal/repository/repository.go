package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Device struct {
	ID          string    `json:"id"`
	TenantID    *string   `json:"tenant_id"`
	IMEI        string    `json:"imei"`
	SimNumber   *string   `json:"sim_number"`
	SimICCID    *string   `json:"sim_iccid"`
	Model       string    `json:"model"`
	FirmwareVer *string   `json:"firmware_ver"`
	VehicleID   *string   `json:"vehicle_id"`
	Status      string    `json:"status"`
	Config      any       `json:"config"`
	LastSeen    *time.Time `json:"last_seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDevice(ctx context.Context, d Device) (Device, error) {
	const q = `
INSERT INTO devices (tenant_id, imei, sim_number, sim_iccid, model, firmware_ver, vehicle_id, status, config)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid, $8, $9)
RETURNING id::text, status, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, nullable(d.TenantID), d.IMEI, d.SimNumber, d.SimICCID, d.Model, d.FirmwareVer, nullable(d.VehicleID), d.Status, d.Config).
		Scan(&d.ID, &d.Status, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *Repository) GetDevice(ctx context.Context, id, tenantID string) (Device, error) {
	const q = `SELECT id::text, tenant_id::text, imei, sim_number, sim_iccid, model, firmware_ver, vehicle_id::text, status, config, last_seen, created_at, updated_at FROM devices WHERE id = $1::uuid AND tenant_id = $2::uuid`
	var d Device
	err := r.db.QueryRow(ctx, q, id, tenantID).Scan(&d.ID, &d.TenantID, &d.IMEI, &d.SimNumber, &d.SimICCID, &d.Model, &d.FirmwareVer, &d.VehicleID, &d.Status, &d.Config, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *Repository) ListDevices(ctx context.Context, tenantID string) ([]Device, error) {
	const q = `SELECT id::text, tenant_id::text, imei, sim_number, sim_iccid, model, firmware_ver, vehicle_id::text, status, config, last_seen, created_at, updated_at FROM devices WHERE tenant_id = $1::uuid ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ds []Device
	for rows.Next() {
		var d Device
		if err := rows.Scan(&d.ID, &d.TenantID, &d.IMEI, &d.SimNumber, &d.SimICCID, &d.Model, &d.FirmwareVer, &d.VehicleID, &d.Status, &d.Config, &d.LastSeen, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return ds, nil
}

func (r *Repository) UpdateDeviceStatus(ctx context.Context, deviceID, status string) error {
	const q = `UPDATE devices SET status = $1, last_seen = NOW(), updated_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, status, deviceID)
	return err
}

func (r *Repository) DeleteDevice(ctx context.Context, id, tenantID string) error {
	const q = `DELETE FROM devices WHERE id = $1::uuid AND tenant_id = $2::uuid`
	_, err := r.db.Exec(ctx, q, id, tenantID)
	return err
}

func nullable(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
