package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MaintenanceTask struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	VehicleID   string     `json:"vehicle_id"`
	Type        string     `json:"type"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	Odometer    *float64   `json:"odometer"`
	DueAt       time.Time  `json:"due_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Cost        *float64   `json:"cost"`
	Vendor      *string    `json:"vendor"`
	Notes       *string    `json:"notes"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTask(ctx context.Context, t MaintenanceTask) (MaintenanceTask, error) {
	const q = `
INSERT INTO maintenance_tasks (tenant_id, vehicle_id, type, title, description, odometer, due_at)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7)
RETURNING id::text, status, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, t.TenantID, t.VehicleID, t.Type, t.Title, t.Description, t.Odometer, t.DueAt).
		Scan(&t.ID, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}

func (r *Repository) ListTasks(ctx context.Context, tenantID string) ([]MaintenanceTask, error) {
	const q = `
SELECT id::text, tenant_id::text, vehicle_id::text, type, title, description, status, odometer, due_at, completed_at, cost, vendor, notes, created_at, updated_at
FROM maintenance_tasks WHERE tenant_id = $1::uuid ORDER BY due_at ASC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ts []MaintenanceTask
	for rows.Next() {
		var t MaintenanceTask
		if err := rows.Scan(&t.ID, &t.TenantID, &t.VehicleID, &t.Type, &t.Title, &t.Description, &t.Status,
			&t.Odometer, &t.DueAt, &t.CompletedAt, &t.Cost, &t.Vendor, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}

func (r *Repository) CompleteTask(ctx context.Context, id, tenantID string, cost float64, vendor, notes string) error {
	const q = `
UPDATE maintenance_tasks
SET status = 'COMPLETED', completed_at = NOW(), cost = $3, vendor = $4, notes = $5, updated_at = NOW()
WHERE id = $1::uuid AND tenant_id = $2::uuid`
	_, err := r.db.Exec(ctx, q, id, tenantID, cost, vendor, notes)
	return err
}

func (r *Repository) ListOverdueTasks(ctx context.Context) ([]MaintenanceTask, error) {
	const q = `
SELECT id::text, tenant_id::text, vehicle_id::text, type, title, description, status, odometer, due_at, completed_at, cost, vendor, notes, created_at, updated_at
FROM maintenance_tasks WHERE due_at < NOW() AND status = 'SCHEDULED'`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ts []MaintenanceTask
	for rows.Next() {
		var t MaintenanceTask
		if err := rows.Scan(&t.ID, &t.TenantID, &t.VehicleID, &t.Type, &t.Title, &t.Description, &t.Status,
			&t.Odometer, &t.DueAt, &t.CompletedAt, &t.Cost, &t.Vendor, &t.Notes, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		ts = append(ts, t)
	}
	return ts, nil
}
