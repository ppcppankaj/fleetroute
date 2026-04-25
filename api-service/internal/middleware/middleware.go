// Package middleware provides Gin middleware for the API service.
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	pkgauth "gpsgo/pkg/auth"
	pkgdb "gpsgo/pkg/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

// RequestLogger logs each HTTP request with method, path, status, and latency.
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("http",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("ip", c.ClientIP()),
		)
	}
}

// RLS begins a PostgreSQL transaction, executes SET LOCAL app.tenant_id, and
// stores the transaction on the Gin context for all downstream handlers to use.
// This activates the per-tenant Row Level Security policies defined in migrations.
//
// The transaction is always rolled back on request completion (via defer).
// Write handlers must call tx.Commit(ctx) themselves via pkgdb.TxFromContext(c).
func RLS(pool *pgxpool.Pool, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := pkgauth.TenantID(c)
		if tenantID == "" {
			c.AbortWithStatus(403)
			return
		}

		ctx := c.Request.Context()

		// Begin a transaction — SET LOCAL scopes to the transaction lifetime.
		tx, err := pool.Begin(ctx)
		if err != nil {
			logger.Error("rls: begin transaction", zap.Error(err))
			c.AbortWithStatus(500)
			return
		}
		// Always roll back on completion. Write handlers commit explicitly.
		defer tx.Rollback(ctx) //nolint:errcheck

		// Activate RLS: PostgreSQL policies read current_setting('app.tenant_id').
		if _, err := tx.Exec(ctx, "SET LOCAL app.tenant_id = $1", tenantID); err != nil {
			logger.Error("rls: set tenant_id", zap.Error(err))
			c.AbortWithStatus(500)
			return
		}

		// Store transaction in Gin context so handlers can retrieve it.
		pkgdb.SetTx(c, tx)
		c.Next()
	}
}

// RateLimit implements a simple per-tenant token bucket using Redis.
// Limits: 1000 req/min per tenant.
// On Redis failure, the middleware fails open (allows the request) but logs
// a warning so the operational team is alerted via Loki/Grafana.
func RateLimit(rdb *redis.Client, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID := pkgauth.TenantID(c)
		if tenantID == "" {
			c.Next()
			return
		}

		key := "ratelimit:api:" + tenantID
		ctx := c.Request.Context()

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			logger.Warn("rate limiter fail-open: Redis unavailable — rate limiting disabled for this request",
				zap.String("tenant_id", tenantID),
				zap.Error(err),
			)
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, time.Minute) //nolint:errcheck
		}

		c.Header("X-RateLimit-Limit", "1000")
		c.Header("X-RateLimit-Remaining", itoa(max(0, 1000-int(count))))

		if count > 1000 {
			c.AbortWithStatusJSON(429, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
