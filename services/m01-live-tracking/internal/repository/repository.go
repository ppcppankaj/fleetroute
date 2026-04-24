package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	sharedtypes "gpsgo/shared/types"
)

type Breadcrumb struct {
	ID         string    `json:"id"`
	VehicleID  string    `json:"vehicle_id"`
	TenantID   string    `json:"tenant_id"`
	Lat        float64   `json:"lat"`
	Lng        float64   `json:"lng"`
	Speed      float64   `json:"speed"`
	Heading    float64   `json:"heading"`
	Altitude   float64   `json:"altitude"`
	Ignition   bool      `json:"ignition"`
	Timestamp  time.Time `json:"timestamp"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) InsertBreadcrumb(ctx context.Context, evt sharedtypes.LocationUpdatedEvent) error {
	const q = `
INSERT INTO breadcrumbs (vehicle_id, tenant_id, lat, lng, speed, heading, altitude, ignition, timestamp)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8, $9)`
	_, err := r.db.Exec(ctx, q, evt.VehicleID, evt.TenantID, evt.Lat, evt.Lng, evt.Speed, evt.Heading, evt.Altitude, evt.Ignition, evt.Timestamp)
	return err
}

func (r *Repository) GetRecentBreadcrumbs(ctx context.Context, vehicleID, tenantID string, limit int) ([]Breadcrumb, error) {
	const q = `
SELECT id::text, vehicle_id::text, tenant_id::text, lat, lng, speed, heading, altitude, ignition, timestamp, created_at
FROM breadcrumbs
WHERE vehicle_id = $1::uuid AND tenant_id = $2::uuid
ORDER BY timestamp DESC
LIMIT $3`
	
	rows, err := r.db.Query(ctx, q, vehicleID, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bs []Breadcrumb
	for rows.Next() {
		var b Breadcrumb
		err := rows.Scan(&b.ID, &b.VehicleID, &b.TenantID, &b.Lat, &b.Lng, &b.Speed, &b.Heading, &b.Altitude, &b.Ignition, &b.Timestamp, &b.CreatedAt)
		if err != nil {
			return nil, err
		}
		bs = append(bs, b)
	}
	return bs, nil
}
