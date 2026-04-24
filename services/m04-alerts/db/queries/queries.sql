-- name: CreateAlertRule :one
INSERT INTO alert_rules (tenant_id, name, event_type, severity, conditions)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListAlertRulesByEvent :many
SELECT * FROM alert_rules WHERE event_type = $1 AND is_active = true;

-- name: CreateAlert :one
INSERT INTO active_alerts (tenant_id, rule_id, vehicle_id, driver_id, type, severity, message, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ResolveAlert :exec
UPDATE active_alerts SET status = 'RESOLVED', resolved_by = $1, resolved_at = NOW() WHERE id = $2 AND tenant_id = $3;

-- name: ListActiveAlerts :many
SELECT * FROM active_alerts WHERE tenant_id = $1 AND status = 'TRIGGERED' ORDER BY created_at DESC;
