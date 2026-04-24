package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Feature struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description"`
	Status      string    `json:"status"`
	Category    *string   `json:"category"`
	Votes       int       `json:"votes"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Repository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListFeatures(ctx context.Context) ([]Feature, error) {
	const q = `SELECT id::text, title, description, status, category, votes, created_at, updated_at FROM roadmap_features ORDER BY votes DESC, created_at DESC`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fs []Feature
	for rows.Next() {
		var f Feature
		if err := rows.Scan(&f.ID, &f.Title, &f.Description, &f.Status, &f.Category, &f.Votes, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		fs = append(fs, f)
	}
	return fs, nil
}

func (r *Repository) CreateFeature(ctx context.Context, f Feature) (Feature, error) {
	const q = `
INSERT INTO roadmap_features (title, description, status, category)
VALUES ($1, $2, $3, $4)
RETURNING id::text, votes, created_at, updated_at`
	err := r.db.QueryRow(ctx, q, f.Title, f.Description, f.Status, f.Category).
		Scan(&f.ID, &f.Votes, &f.CreatedAt, &f.UpdatedAt)
	return f, err
}

func (r *Repository) CastVote(ctx context.Context, featureID, tenantID, userID string) error {
	// insert vote row; feature votes counter updated via trigger or manual
	const qVote = `INSERT INTO feature_votes (feature_id, tenant_id, user_id) VALUES ($1::uuid, $2::uuid, $3::uuid) ON CONFLICT DO NOTHING`
	if _, err := r.db.Exec(ctx, qVote, featureID, tenantID, userID); err != nil {
		return err
	}
	const qCount = `UPDATE roadmap_features SET votes = (SELECT COUNT(*) FROM feature_votes WHERE feature_id = $1::uuid) WHERE id = $1::uuid`
	_, err := r.db.Exec(ctx, qCount, featureID)
	return err
}
