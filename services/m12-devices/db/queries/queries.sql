-- name: CreateDevice :one
INSERT INTO devices (
    tenant_id, imei, sim_number, sim_iccid, model, firmware_ver, vehicle_id, status, config
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetDevice :one
SELECT * FROM devices WHERE id = $1 AND tenant_id = $2;

-- name: ListDevices :many
SELECT * FROM devices WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: UpdateDeviceStatus :exec
UPDATE devices SET status = $1, last_seen = NOW(), updated_at = NOW() WHERE id = $2;

-- name: UpdateDeviceConfig :exec
UPDATE devices SET config = $1, updated_at = NOW() WHERE id = $2;

-- name: DeleteDevice :exec
DELETE FROM devices WHERE id = $1 AND tenant_id = $2;

-- name: CreateDeviceCommand :one
INSERT INTO device_commands (
    device_id, command, payload
) VALUES (
    $1, $2, $3
) RETURNING *;
