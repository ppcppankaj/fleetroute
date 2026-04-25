package enrichment

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const (
	// deviceTenantCacheTTL is how long we cache the device → tenant mapping.
	// 5 minutes is aggressive enough for real-time processing but short enough
	// that a re-provisioned device will pick up its new tenant quickly.
	deviceTenantCacheTTL = 5 * time.Minute

	// deviceTenantCachePrefix is the Redis key prefix for tenant lookups.
	deviceTenantCachePrefix = "device:tenant:"
)

// DeviceRegistry resolves a device_id to its tenant_id.
// It uses Redis as a fast cache with PostgreSQL as the authoritative source.
// On cache miss it queries the DB and warms the cache.
// On DB error it returns an empty string so callers can handle the unknown
// tenant gracefully (e.g., drop the record or publish to a dead-letter topic).
type DeviceRegistry struct {
	pool   *pgxpool.Pool
	rdb    *redis.Client
	logger *zap.Logger
}

// NewDeviceRegistry constructs a DeviceRegistry.
func NewDeviceRegistry(pool *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *DeviceRegistry {
	return &DeviceRegistry{pool: pool, rdb: rdb, logger: logger}
}

// Lookup returns the tenant_id for the given device_id.
// Returns an empty string if the device is not found or an error occurs.
func (r *DeviceRegistry) Lookup(ctx context.Context, deviceID string) string {
	key := deviceTenantCachePrefix + deviceID

	// 1. Try Redis cache first (O(1), avoids DB round-trip on hot path).
	if cached, err := r.rdb.Get(ctx, key).Result(); err == nil {
		return cached
	}

	// 2. Cache miss — query the devices table.
	var tenantID string
	err := r.pool.QueryRow(ctx,
		`SELECT tenant_id FROM devices WHERE id = $1 AND deleted_at IS NULL LIMIT 1`,
		deviceID,
	).Scan(&tenantID)
	if err != nil {
		r.logger.Warn("device registry: tenant lookup failed",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
		return ""
	}

	// 3. Warm the cache so subsequent events for this device are fast.
	if err := r.rdb.Set(ctx, key, tenantID, deviceTenantCacheTTL).Err(); err != nil {
		// Non-fatal: log and continue. Next request will re-query.
		r.logger.Warn("device registry: failed to cache tenant_id",
			zap.String("device_id", deviceID),
			zap.Error(err),
		)
	}

	return tenantID
}

// Invalidate removes a device's cached tenant mapping.
// Call this when a device is re-assigned to a different tenant.
func (r *DeviceRegistry) Invalidate(ctx context.Context, deviceID string) {
	r.rdb.Del(ctx, deviceTenantCachePrefix+deviceID) //nolint:errcheck
}

// LookupVehicleID returns the vehicle_id for the given device_id.
func (r *DeviceRegistry) LookupVehicleID(ctx context.Context, deviceID string) string {
	var vehicleID string
	r.pool.QueryRow(ctx,
		`SELECT COALESCE(vehicle_id::text, '') FROM devices WHERE id=$1 OR imei=$1 LIMIT 1`,
		deviceID,
	).Scan(&vehicleID)
	return vehicleID
}
