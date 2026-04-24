package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Subscription struct {
	ID                  string     `json:"id"`
	TenantID            string     `json:"tenant_id"`
	PlanID              string     `json:"plan_id"`
	StripeSubID         *string    `json:"stripe_sub_id"`
	StripeCusID         *string    `json:"stripe_cus_id"`
	Status              string     `json:"status"`
	CurrentPeriodStart  *time.Time `json:"current_period_start"`
	CurrentPeriodEnd    *time.Time `json:"current_period_end"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type Invoice struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	StripeInvID *string    `json:"stripe_inv_id"`
	Amount      float64    `json:"amount"`
	Currency    string     `json:"currency"`
	Status      string     `json:"status"`
	DueDate     *time.Time `json:"due_date"`
	PaidAt      *time.Time `json:"paid_at"`
	PDFURL      *string    `json:"pdf_url"`
	CreatedAt   time.Time  `json:"created_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertSubscription(ctx context.Context, s Subscription) error {
	const q = `
INSERT INTO subscriptions (tenant_id, plan_id, stripe_sub_id, stripe_cus_id, status, current_period_start, current_period_end)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
ON CONFLICT (tenant_id) DO UPDATE SET
  plan_id = EXCLUDED.plan_id,
  stripe_sub_id = EXCLUDED.stripe_sub_id,
  status = EXCLUDED.status,
  current_period_start = EXCLUDED.current_period_start,
  current_period_end = EXCLUDED.current_period_end,
  updated_at = NOW()`
	_, err := r.db.Exec(ctx, q, s.TenantID, s.PlanID, s.StripeSubID, s.StripeCusID, s.Status, s.CurrentPeriodStart, s.CurrentPeriodEnd)
	return err
}

func (r *Repository) GetSubscription(ctx context.Context, tenantID string) (Subscription, error) {
	const q = `SELECT id::text, tenant_id::text, plan_id, stripe_sub_id, stripe_cus_id, status, current_period_start, current_period_end, created_at, updated_at FROM subscriptions WHERE tenant_id = $1::uuid`
	var s Subscription
	err := r.db.QueryRow(ctx, q, tenantID).Scan(&s.ID, &s.TenantID, &s.PlanID, &s.StripeSubID, &s.StripeCusID, &s.Status, &s.CurrentPeriodStart, &s.CurrentPeriodEnd, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) CreateInvoice(ctx context.Context, inv Invoice) (Invoice, error) {
	const q = `
INSERT INTO invoices (tenant_id, stripe_inv_id, amount, currency, status, due_date, pdf_url)
VALUES ($1::uuid, $2, $3, $4, $5, $6, $7)
RETURNING id::text, created_at`
	err := r.db.QueryRow(ctx, q, inv.TenantID, inv.StripeInvID, inv.Amount, inv.Currency, inv.Status, inv.DueDate, inv.PDFURL).
		Scan(&inv.ID, &inv.CreatedAt)
	return inv, err
}

func (r *Repository) ListInvoices(ctx context.Context, tenantID string) ([]Invoice, error) {
	const q = `SELECT id::text, tenant_id::text, stripe_inv_id, amount, currency, status, due_date, paid_at, pdf_url, created_at FROM invoices WHERE tenant_id = $1::uuid ORDER BY created_at DESC LIMIT 24`
	rows, err := r.db.Query(ctx, q, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invs []Invoice
	for rows.Next() {
		var inv Invoice
		if err := rows.Scan(&inv.ID, &inv.TenantID, &inv.StripeInvID, &inv.Amount, &inv.Currency, &inv.Status, &inv.DueDate, &inv.PaidAt, &inv.PDFURL, &inv.CreatedAt); err != nil {
			return nil, err
		}
		invs = append(invs, inv)
	}
	return invs, nil
}
