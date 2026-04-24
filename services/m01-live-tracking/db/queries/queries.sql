-- name: InsertBreadcrumb :exec
INSERT INTO breadcrumbs (
    vehicle_id, tenant_id, lat, lng, speed, heading, altitude, ignition, timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetRecentBreadcrumbs :many
SELECT * FROM breadcrumbs 
WHERE vehicle_id = $1 AND tenant_id = $2
ORDER BY timestamp DESC
LIMIT $3;
