package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AlertRule struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	EventType  string    `json:"event_type"`
	Severity   string    `json:"severity"`
	Conditions any       `json:"conditions"` // jsonb
	IsActive   bool      `json:"is_active"`
}

type ActiveAlert struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	RuleID     *string   `json:"rule_id"`
	VehicleID  string    `json:"vehicle_id"`
	DriverID   *string   `json:"driver_id"`
	Type       string    `json:"type"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Metadata   any       `json:"metadata"`
	Status     string    `json:"status"`
	ResolvedBy *string   `json:"resolved_by"`
	ResolvedAt *time.Time `json:"resolved_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListAlertRulesByEvent(ctx context.Context, eventType string) ([]AlertRule, error) {
	const q = `SELECT id::text, tenant_id::text, name, event_type, severity, conditions, is_active FROM alert_rules WHERE event_type = $1 AND is_active = true`
	rows, err := r.db.Query(ctx, q, eventType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rs []AlertRule
	for rows.Next() {
		var a AlertRule
		if err := rows.Scan(&a.ID, &a.TenantID, &a.Name, &a.EventType, &a.Severity, &a.Conditions, &a.IsActive); err != nil {
			return nil, err
		}
		rs = append(rs, a)
	}
	return rs, nil
}

func (r *Repository) CreateAlert(ctx context.Context, a ActiveAlert) (ActiveAlert, error) {
	const q = `
INSERT INTO active_alerts (tenant_id, rule_id, vehicle_id, driver_id, type, severity, message, metadata)
VALUES ($1::uuid, $2::uuid, $3::uuid, $4::uuid, $5, $6, $7, $8)
RETURNING id::text, status, created_at`
	err := r.db.QueryRow(ctx, q, a.TenantID, nullable(a.RuleID), a.VehicleID, nullable(a.DriverID), a.Type, a.Severity, a.Message, a.Metadata).
		Scan(&a.ID, &a.Status, &a.CreatedAt)
	return a, err
}

func (r *Repository) ResolveAlert(ctx context.Context, id, tenantID, resolvedBy string) error {
	const q = `UPDATE active_alerts SET status = 'RESOLVED', resolved_by = $1::uuid, resolved_at = NOW() WHERE id = $2::uuid AND tenant_id = $3::uuid`
	_, err := r.db.Exec(ctx, q, resolvedBy, id, tenantID)
	return err
}

func (r *Repository) ListActiveAlerts(ctx context.Context, tenantID string) ([]ActiveAlert, error) {
	const q = `
SELECT id::text, tenant_id::text, rule_id::text, vehicle_id::text, driver_id::text, type, severity, message, metadata, status, resolved_by::text, resolved_at, created_at
FROM active_alerts WHERE tenant_id = $1::uuid AND status = 'TRIGGERED' ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var as []ActiveAlert
	for rows.Next() {
		var a ActiveAlert
		if err := rows.Scan(&a.ID, &a.TenantID, &a.RuleID, &a.VehicleID, &a.DriverID, &a.Type, &a.Severity, &a.Message, &a.Metadata, &a.Status, &a.ResolvedBy, &a.ResolvedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		as = append(as, a)
	}
	return as, nil
}

func nullable(s *string) any {
	if s == nil || *s == "" {
		return nil
	}
	return *s
}
