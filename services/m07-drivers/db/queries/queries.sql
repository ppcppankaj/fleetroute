-- name: CreateDriver :one
INSERT INTO drivers (
    tenant_id, name, phone, email, license_no, license_expiry, group_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetDriver :one
SELECT * FROM drivers WHERE id = $1 AND tenant_id = $2;

-- name: ListDrivers :many
SELECT * FROM drivers WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: UpdateDriverBehaviorScore :exec
UPDATE drivers SET behavior_score = behavior_score - $1, updated_at = NOW() WHERE id = $2;

-- name: DeleteDriver :exec
DELETE FROM drivers WHERE id = $1 AND tenant_id = $2;

-- name: CreateBehaviorEvent :one
INSERT INTO behavior_events (
    driver_id, vehicle_id, type, severity, points
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;
