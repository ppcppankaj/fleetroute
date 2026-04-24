package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditLog struct {
	ID         string    `json:"id"`
	TenantID   *string   `json:"tenant_id"`
	UserID     *string   `json:"user_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID *string   `json:"resource_id"`
	IPAddress  *string   `json:"ip_address"`
	UserAgent  *string   `json:"user_agent"`
	Metadata   any       `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}

type SecurityIncident struct {
	ID          string    `json:"id"`
	TenantID    *string   `json:"tenant_id"`
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description *string   `json:"description"`
	Resolved    bool      `json:"resolved"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAuditLog(ctx context.Context, a AuditLog) error {
	const q = `
INSERT INTO audit_logs (tenant_id, user_id, action, resource, resource_id, ip_address, user_agent, metadata)
VALUES ($1::uuid, $2::uuid, $3, $4, $5, $6, $7, $8)`
	_, err := r.db.Exec(ctx, q, nullUUID(a.TenantID), nullUUID(a.UserID), a.Action, a.Resource, a.ResourceID, a.IPAddress, a.UserAgent, a.Metadata)
	return err
}

func (r *Repository) ListAuditLogs(ctx context.Context, tenantID string, limit int) ([]AuditLog, error) {
	const q = `
SELECT id::text, tenant_id::text, user_id::text, action, resource, resource_id, ip_address, user_agent, metadata, created_at
FROM audit_logs WHERE tenant_id = $1::uuid ORDER BY created_at DESC LIMIT $2`
	rows, err := r.db.Query(ctx, q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []AuditLog
	for rows.Next() {
		var a AuditLog
		if err := rows.Scan(&a.ID, &a.TenantID, &a.UserID, &a.Action, &a.Resource, &a.ResourceID, &a.IPAddress, &a.UserAgent, &a.Metadata, &a.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, a)
	}
	return logs, nil
}

func (r *Repository) CreateIncident(ctx context.Context, inc SecurityIncident) (SecurityIncident, error) {
	const q = `
INSERT INTO security_incidents (tenant_id, type, severity, description)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text, resolved, created_at`
	err := r.db.QueryRow(ctx, q, nullUUID(inc.TenantID), inc.Type, inc.Severity, inc.Description).
		Scan(&inc.ID, &inc.Resolved, &inc.CreatedAt)
	return inc, err
}

func (r *Repository) ListIncidents(ctx context.Context, tenantID string) ([]SecurityIncident, error) {
	const q = `
SELECT id::text, tenant_id::text, type, severity, description, resolved, created_at
FROM security_incidents WHERE tenant_id = $1::uuid OR $1::uuid IS NULL ORDER BY created_at DESC LIMIT 50`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var incs []SecurityIncident
	for rows.Next() {
		var inc SecurityIncident
		if err := rows.Scan(&inc.ID, &inc.TenantID, &inc.Type, &inc.Severity, &inc.Description, &inc.Resolved, &inc.CreatedAt); err != nil {
			return nil, err
		}
		incs = append(incs, inc)
	}
	return incs, nil
}

func nullUUID(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
