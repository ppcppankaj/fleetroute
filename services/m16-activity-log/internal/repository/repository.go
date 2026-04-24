package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityEvent struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	UserID      *string   `json:"user_id"`
	VehicleID   *string   `json:"vehicle_id"`
	DriverID    *string   `json:"driver_id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Metadata    any       `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Insert(ctx context.Context, e ActivityEvent) error {
	const q = `
INSERT INTO activity_events (tenant_id, user_id, vehicle_id, driver_id, type, title, description, metadata)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q,
		e.TenantID, nullStr(e.UserID), nullStr(e.VehicleID), nullStr(e.DriverID),
		e.Type, e.Title, e.Description, e.Metadata)
	return err
}

func (r *Repository) List(ctx context.Context, tenantID string, limit int) ([]ActivityEvent, error) {
	const q = `
SELECT id::text, tenant_id::text, user_id::text, vehicle_id::text, driver_id::text,
       type, title, description, metadata, created_at
FROM activity_events WHERE tenant_id = $1::uuid ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []ActivityEvent
	for rows.Next() {
		var e ActivityEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.UserID, &e.VehicleID, &e.DriverID, &e.Type, &e.Title, &e.Description, &e.Metadata, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func nullStr(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
