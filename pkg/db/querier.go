// Package db provides shared database interfaces and context helpers.
package db

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// contextKey is an unexported type for Gin context keys in this package.
type contextKey string

const txContextKey contextKey = "pg_tx"

// Querier is implemented by both *pgxpool.Pool and pgx.Tx, allowing handlers
// to be agnostic about whether they are running inside a transaction or not.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// SetTx stores a pgx.Tx on the Gin context so downstream handlers can use it.
// Called by the RLS middleware after BEGIN + SET LOCAL.
func SetTx(c *gin.Context, tx pgx.Tx) {
	c.Set(string(txContextKey), tx)
}

// TxFromContext retrieves the pgx.Tx stored by the RLS middleware.
// Returns nil if no transaction is present (e.g., health check routes).
func TxFromContext(c *gin.Context) pgx.Tx {
	if v, ok := c.Get(string(txContextKey)); ok {
		if tx, ok := v.(pgx.Tx); ok {
			return tx
		}
	}
	return nil
}

// KeyRevoked returns the Redis key used to denylist a revoked refresh token.
// TTL on the key matches the token's remaining lifetime.
func KeyRevoked(jti string) string {
	return "auth:revoked:" + jti
}
