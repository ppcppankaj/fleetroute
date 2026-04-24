package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	TenantID     string
	Email        string
	Name         string
	RoleName     string
	PasswordHash string
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindUserByEmail(ctx context.Context, email string) (User, error) {
	const q = `
SELECT u.id::text, u.tenant_id::text, u.email, u.name, COALESCE(ro.name, 'USER') as role_name, u.password_hash
FROM users u
LEFT JOIN roles ro ON ro.id = u.role_id
WHERE u.email = $1 AND u.status = 'ACTIVE'
LIMIT 1`
	var u User
	err := r.db.QueryRow(ctx, q, email).Scan(&u.ID, &u.TenantID, &u.Email, &u.Name, &u.RoleName, &u.PasswordHash)
	return u, err
}

func (r *Repository) UpsertSession(ctx context.Context, userID, refreshToken, ipAddress, userAgent string, expiresAt time.Time) error {
	const q = `
INSERT INTO sessions (user_id, refresh_token, ip_address, user_agent, expires_at)
VALUES ($1::uuid, $2, $3::inet, $4, $5)
ON CONFLICT (refresh_token)
DO UPDATE SET user_id = EXCLUDED.user_id, ip_address = EXCLUDED.ip_address, user_agent = EXCLUDED.user_agent, expires_at = EXCLUDED.expires_at`
	_, err := r.db.Exec(ctx, q, userID, refreshToken, nullableIP(ipAddress), userAgent, expiresAt)
	return err
}

func (r *Repository) FindSessionByRefreshToken(ctx context.Context, refreshToken string) (string, string, string, error) {
	const q = `
SELECT u.id::text, u.tenant_id::text, COALESCE(ro.name, 'USER') as role_name
FROM sessions s
JOIN users u ON u.id = s.user_id
LEFT JOIN roles ro ON ro.id = u.role_id
WHERE s.refresh_token = $1 AND s.expires_at > NOW() AND u.status = 'ACTIVE'
LIMIT 1`
	var userID, tenantID, role string
	err := r.db.QueryRow(ctx, q, refreshToken).Scan(&userID, &tenantID, &role)
	return userID, tenantID, role, err
}

func (r *Repository) MarkLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx, `UPDATE users SET last_login = NOW() WHERE id = $1::uuid`, userID)
	return err
}

func nullableIP(ip string) any {
	if ip == "" {
		return nil
	}
	return ip
}
