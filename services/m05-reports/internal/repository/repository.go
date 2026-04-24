package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ReportDefinition struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Parameters any       `json:"parameters"`
	Schedule   *string   `json:"schedule"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type ReportRun struct {
	ID           string     `json:"id"`
	DefinitionID string     `json:"definition_id"`
	TenantID     string     `json:"tenant_id"`
	Status       string     `json:"status"`
	FileURL      *string    `json:"file_url"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	Error        *string    `json:"error"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateDefinition(ctx context.Context, d ReportDefinition) (ReportDefinition, error) {
	const q = `
INSERT INTO report_definitions (tenant_id, name, type, parameters, schedule)
VALUES ($1::uuid, $2, $3, $4, $5)
RETURNING id::text, is_active, created_at`
	err := r.db.QueryRow(ctx, q, d.TenantID, d.Name, d.Type, d.Parameters, d.Schedule).
		Scan(&d.ID, &d.IsActive, &d.CreatedAt)
	return d, err
}

func (r *Repository) ListDefinitions(ctx context.Context, tenantID string) ([]ReportDefinition, error) {
	const q = `SELECT id::text, tenant_id::text, name, type, parameters, schedule, is_active, created_at FROM report_definitions WHERE tenant_id = $1::uuid ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ds []ReportDefinition
	for rows.Next() {
		var d ReportDefinition
		if err := rows.Scan(&d.ID, &d.TenantID, &d.Name, &d.Type, &d.Parameters, &d.Schedule, &d.IsActive, &d.CreatedAt); err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return ds, nil
}

func (r *Repository) CreateRun(ctx context.Context, defID, tenantID string) (ReportRun, error) {
	const q = `
INSERT INTO report_runs (definition_id, tenant_id)
VALUES ($1::uuid, $2::uuid)
RETURNING id::text, status, started_at`
	var run ReportRun
	run.DefinitionID = defID
	run.TenantID = tenantID
	err := r.db.QueryRow(ctx, q, defID, tenantID).Scan(&run.ID, &run.Status, &run.StartedAt)
	return run, err
}

func (r *Repository) CompleteRun(ctx context.Context, runID, fileURL string) error {
	const q = `UPDATE report_runs SET status = 'DONE', file_url = $1, completed_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, fileURL, runID)
	return err
}

func (r *Repository) FailRun(ctx context.Context, runID, errMsg string) error {
	const q = `UPDATE report_runs SET status = 'FAILED', error = $1, completed_at = NOW() WHERE id = $2::uuid`
	_, err := r.db.Exec(ctx, q, errMsg, runID)
	return err
}

func (r *Repository) ListRuns(ctx context.Context, tenantID string) ([]ReportRun, error) {
	const q = `SELECT id::text, definition_id::text, tenant_id::text, status, file_url, started_at, completed_at, error FROM report_runs WHERE tenant_id = $1::uuid ORDER BY started_at DESC LIMIT 50`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rs []ReportRun
	for rows.Next() {
		var run ReportRun
		if err := rows.Scan(&run.ID, &run.DefinitionID, &run.TenantID, &run.Status, &run.FileURL, &run.StartedAt, &run.CompletedAt, &run.Error); err != nil {
			return nil, err
		}
		rs = append(rs, run)
	}
	return rs, nil
}
