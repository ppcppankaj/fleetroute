package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Geofence struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Properties any      `json:"properties"`
	Polygon   any       `json:"polygon"` // raw JSON
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListGeofencesByTenant(ctx context.Context, tenantID string) ([]Geofence, error) {
	const q = `SELECT id::text, tenant_id::text, name, type, properties, polygon, created_at, updated_at FROM geofences WHERE tenant_id = $1::uuid`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gs []Geofence
	for rows.Next() {
		var g Geofence
		if err := rows.Scan(&g.ID, &g.TenantID, &g.Name, &g.Type, &g.Properties, &g.Polygon, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		gs = append(gs, g)
	}
	return gs, nil
}

func (r *Repository) GetVehicleState(ctx context.Context, vehicleID, geofenceID string) (string, error) {
	const q = `SELECT state FROM vehicle_geofence_states WHERE vehicle_id = $1::uuid AND geofence_id = $2::uuid`
	var state string
	err := r.db.QueryRow(ctx, q, vehicleID, geofenceID).Scan(&state)
	return state, err
}

func (r *Repository) UpdateVehicleState(ctx context.Context, vehicleID, geofenceID, tenantID, state string) error {
	const q = `
INSERT INTO vehicle_geofence_states (vehicle_id, geofence_id, tenant_id, state, last_event)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4, NOW())
ON CONFLICT (vehicle_id, geofence_id) DO UPDATE 
SET state = EXCLUDED.state, last_event = EXCLUDED.last_event`
	_, err := r.db.Exec(ctx, q, vehicleID, geofenceID, tenantID, state)
	return err
}

func (r *Repository) CreateGeofence(ctx context.Context, g Geofence) (Geofence, error) {
	const q = `
INSERT INTO geofences (tenant_id, name, type, properties, polygon)
VALUES ($1::uuid, $2, $3, $4, $5)
RETURNING id::text, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, g.TenantID, g.Name, g.Type, g.Properties, g.Polygon).
		Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
	return g, err
}
