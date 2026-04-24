package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Driver struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	Email         *string   `json:"email"`
	LicenseNo     string    `json:"license_no"`
	LicenseExpiry time.Time `json:"license_expiry"`
	GroupID       *string   `json:"group_id"`
	BehaviorScore float64   `json:"behavior_score"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDriver(ctx context.Context, d Driver) (Driver, error) {
	const q = `
INSERT INTO drivers (tenant_id, name, phone, email, license_no, license_expiry, group_id)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7::uuid)
RETURNING id::text, behavior_score, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, d.TenantID, d.Name, d.Phone, d.Email, d.LicenseNo, d.LicenseExpiry, d.GroupID).
		Scan(&d.ID, &d.BehaviorScore, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *Repository) GetDriver(ctx context.Context, id, tenantID string) (Driver, error) {
	const q = `SELECT id::text, tenant_id::text, name, phone, email, license_no, license_expiry, group_id::text, behavior_score, created_at, updated_at FROM drivers WHERE id = $1::uuid AND tenant_id = $2::uuid`
	var d Driver
	err := r.db.QueryRow(ctx, q, id, tenantID).Scan(&d.ID, &d.TenantID, &d.Name, &d.Phone, &d.Email, &d.LicenseNo, &d.LicenseExpiry, &d.GroupID, &d.BehaviorScore, &d.CreatedAt, &d.UpdatedAt)
	return d, err
}

func (r *Repository) ListDrivers(ctx context.Context, tenantID string) ([]Driver, error) {
	const q = `SELECT id::text, tenant_id::text, name, phone, email, license_no, license_expiry, group_id::text, behavior_score, created_at, updated_at FROM drivers WHERE tenant_id = $1::uuid ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ds []Driver
	for rows.Next() {
		var d Driver
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Name, &d.Phone, &d.Email, &d.LicenseNo, &d.LicenseExpiry, &d.GroupID, &d.BehaviorScore, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return ds, nil
}

func (r *Repository) UpdateDriverBehaviorScore(ctx context.Context, driverID string, points float64) error {
	const q = `UPDATE drivers SET behavior_score = behavior_score - $1, updated_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, points, driverID)
	return err
}

func (r *Repository) DeleteDriver(ctx context.Context, id, tenantID string) error {
	const q = `DELETE FROM drivers WHERE id = $1::uuid AND tenant_id = $2::uuid`
	_, err := r.db.Exec(ctx, q, id, tenantID)
	return err
}
