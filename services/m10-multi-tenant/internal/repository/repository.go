package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Tenant struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	PlanID       *string   `json:"plan_id"`
	Status       string    `json:"status"`
	Branding     any       `json:"branding"`
	FeatureFlags any       `json:"feature_flags"`
	MaxVehicles  int       `json:"max_vehicles"`
	MaxUsers     int       `json:"max_users"`
	Timezone     string    `json:"timezone"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTenant(ctx context.Context, t Tenant) (Tenant, error) {
	const q = `
INSERT INTO tenants (name, slug, plan_id, branding, feature_flags, max_vehicles, max_users, timezone)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id::text, status, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, t.Name, t.Slug, t.PlanID, t.Branding, t.FeatureFlags, t.MaxVehicles, t.MaxUsers, t.Timezone).
		Scan(&t.ID, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *Repository) GetTenant(ctx context.Context, id string) (Tenant, error) {
	const q = `SELECT id::text, name, slug, plan_id, status, branding, feature_flags, max_vehicles, max_users, timezone, created_at, updated_at FROM tenants WHERE id = $1::uuid`
	var t Tenant
	err := r.db.QueryRow(ctx, q, id).Scan(&t.ID, &t.Name, &t.Slug, &t.PlanID, &t.Status, &t.Branding, &t.FeatureFlags, &t.MaxVehicles, &t.MaxUsers, &t.Timezone, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *Repository) ListTenants(ctx context.Context) ([]Tenant, error) {
	const q = `SELECT id::text, name, slug, plan_id, status, branding, feature_flags, max_vehicles, max_users, timezone, created_at, updated_at FROM tenants ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ts []Tenant
	for rows.Next() {
		var t Tenant
		if err := rows.Scan(&t.ID, &t.Name, &t.Slug, &t.PlanID, &t.Status, &t.Branding, &t.FeatureFlags, &t.MaxVehicles, &t.MaxUsers, &t.Timezone, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func (r *Repository) UpdateTenantStatus(ctx context.Context, id, status string) error {
	const q = `UPDATE tenants SET status = $1, updated_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, status, id)
	return err
}
