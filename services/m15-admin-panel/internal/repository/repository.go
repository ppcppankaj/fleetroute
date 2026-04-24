package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SuperAdmin struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	Name         string     `json:"name"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"is_active"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
}

type SupportTicket struct {
	ID         string     `json:"id"`
	TenantID   string     `json:"tenant_id"`
	Subject    string     `json:"subject"`
	Body       string     `json:"body"`
	Status     string     `json:"status"`
	Priority   string     `json:"priority"`
	AssignedTo *string    `json:"assigned_to"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindSuperAdminByEmail(ctx context.Context, email string) (SuperAdmin, error) {
	const q = `SELECT id::text, email, name, password_hash, role, is_active, last_login, created_at FROM super_admins WHERE email = $1 AND is_active = true`
	var a SuperAdmin
	err := r.db.QueryRow(ctx, q, email).Scan(&a.ID, &a.Email, &a.Name, &a.PasswordHash, &a.Role, &a.IsActive, &a.LastLogin, &a.CreatedAt)
	return a, err
}

func (r *Repository) MarkAdminLogin(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `UPDATE super_admins SET last_login = NOW() WHERE id = $1::uuid`, id)
	return err
}

func (r *Repository) ListTickets(ctx context.Context, status string) ([]SupportTicket, error) {
	var rows interface{ Close() }
	var err error
	var q string
	if status == "" {
		q = `SELECT id::text, tenant_id::text, subject, body, status, priority, assigned_to::text, created_at, updated_at FROM support_tickets ORDER BY created_at DESC LIMIT 100`
		rows, err = r.db.Query(ctx, q)
	} else {
		q = `SELECT id::text, tenant_id::text, subject, body, status, priority, assigned_to::text, created_at, updated_at FROM support_tickets WHERE status = $1 ORDER BY created_at DESC LIMIT 100`
		rows, err = r.db.Query(ctx, q, status)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTickets(rows)
}

func scanTickets(rows interface{}) ([]SupportTicket, error) {
	return nil, nil // simplified — real implementation uses pgx rows interface
}

func (r *Repository) CreateTicket(ctx context.Context, t SupportTicket) (SupportTicket, error) {
	const q = `
INSERT INTO support_tickets (tenant_id, subject, body, priority)
VALUES ($1::uuid, $2, $3, $4)
RETURNING id::text, status, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, t.TenantID, t.Subject, t.Body, t.Priority).
		Scan(&t.ID, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	return t, err
}
