-- name: CreateGeofence :one
INSERT INTO geofences (tenant_id, name, type, properties, polygon)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListGeofences :many
SELECT * FROM geofences WHERE tenant_id = $1;

-- name: GetGeofencesByTenant :many
SELECT * FROM geofences WHERE tenant_id = $1;

-- name: GetVehicleState :one
SELECT state FROM vehicle_geofence_states WHERE vehicle_id = $1 AND geofence_id = $2;

-- name: UpdateVehicleState :exec
INSERT INTO vehicle_geofence_states (vehicle_id, geofence_id, tenant_id, state, last_event)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (vehicle_id, geofence_id) DO UPDATE 
SET state = EXCLUDED.state, last_event = EXCLUDED.last_event;
