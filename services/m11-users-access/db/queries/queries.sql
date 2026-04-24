-- name: GetUserByEmail :one
SELECT u.id, u.tenant_id, u.email, u.password_hash, u.name, COALESCE(r.name, 'USER') as role_name
FROM users u
LEFT JOIN roles r ON r.id = u.role_id
WHERE u.email = $1 AND u.status = 'ACTIVE'
LIMIT 1;

-- name: UpsertSession :exec
INSERT INTO sessions (user_id, refresh_token, ip_address, user_agent, expires_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (refresh_token) DO UPDATE
SET user_id = EXCLUDED.user_id, ip_address = EXCLUDED.ip_address, user_agent = EXCLUDED.user_agent, expires_at = EXCLUDED.expires_at;

-- name: GetSessionByRefreshToken :one
SELECT s.user_id, u.tenant_id, COALESCE(r.name, 'USER') as role_name
FROM sessions s
JOIN users u ON u.id = s.user_id
LEFT JOIN roles r ON r.id = u.role_id
WHERE s.refresh_token = $1 AND s.expires_at > NOW();
