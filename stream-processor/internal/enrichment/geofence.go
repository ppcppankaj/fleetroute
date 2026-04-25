package enrichment

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"gpsgo/shared/types"
)

type GeofenceEngine struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewGeofenceEngine(pool *pgxpool.Pool, logger *zap.Logger) *GeofenceEngine {
	return &GeofenceEngine{pool: pool, logger: logger}
}

func (e *GeofenceEngine) Check(ctx context.Context, rec *EnrichedRecord) {
	rows, err := e.pool.Query(ctx, `
		SELECT id, name FROM geofences 
		WHERE tenant_id = $1 
		AND ST_Contains(geometry, ST_SetSRID(ST_Point($2, $3), 4326))
	`, rec.TenantID, rec.Lng, rec.Lat)
	
	if err != nil {
		e.logger.Error("geofence check error", zap.Error(err))
		return
	}
	defer rows.Close()

	for rows.Next() {
		var gid, name string
		if err := rows.Scan(&gid, &name); err == nil {
			e.logger.Debug("Device inside geofence", 
				zap.String("device_id", rec.DeviceID),
				zap.String("geofence", name))
				
			// TODO: GeofenceEngine currently emits a DWELL event for every frame 
			// where the device is inside the geofence. Real geofencing requires Redis
			// state tracking (per device-per-zone) to emit ENTRY on the first match
			// and EXIT when the device leaves, to avoid flooding Kafka at high GPS frequencies.
			rec.GeneratedEvents = append(rec.GeneratedEvents, &types.GeofenceBreachEvent{
				ZoneID:    gid,
				ZoneName:  name,
				VehicleID: rec.VehicleID,
				TenantID:  rec.TenantID,
				EventType: "DWELL",
				Lat:       rec.Lat,
				Lng:       rec.Lng,
				Speed:     float64(rec.Speed),
				Timestamp: rec.Timestamp,
			})
		}
	}
}
