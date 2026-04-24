-- name: CreateVehicle :one
INSERT INTO vehicles (
    tenant_id, plate_number, make, model, year, color, vin, fuel_type, group_id, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetVehicle :one
SELECT * FROM vehicles WHERE id = $1 AND tenant_id = $2;

-- name: ListVehicles :many
SELECT * FROM vehicles WHERE tenant_id = $1 ORDER BY created_at DESC;

-- name: UpdateVehicleOdometer :exec
UPDATE vehicles SET odometer = $1, updated_at = NOW() WHERE id = $2;

-- name: DeleteVehicle :exec
DELETE FROM vehicles WHERE id = $1 AND tenant_id = $2;

-- name: CreateVehicleGroup :one
INSERT INTO vehicle_groups (
    tenant_id, name
) VALUES (
    $1, $2
) RETURNING *;

-- name: ListVehicleGroups :many
SELECT * FROM vehicle_groups WHERE tenant_id = $1;
