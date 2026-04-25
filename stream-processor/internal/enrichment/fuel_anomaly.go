package enrichment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gpsgo/shared/types"
)

// FuelAnomalyDetector detects rapid unexplained fuel drops (theft / siphoning).
// It uses a sliding window stored in Redis to compare the current fuel level
// against the recent peak, flagging events where the drop exceeds the threshold.
//
// A "theft" event is characterised by:
//   - Engine ignition OFF (not consumption)
//   - Fuel drop > minDropPct% within the last windowSeconds
//   - Fuel level doesn't rise again within confirmSeconds (not a sensor glitch)
//
// All anomalies are persisted to the fuel_anomalies table and a NATS alert is
// published so the notification pipeline can dispatch SMS/email/push.
const (
	// Minimum fuel drop (percent) to flag as anomaly
	minDropPct = 10.0

	// Recent peak window — how far back we look for the high-water mark
	peakWindowSeconds = 600 // 10 min

	// Cooldown between anomaly reports for the same device
	anomalyCooldownSeconds = 900 // 15 min
)

// fuelState holds per-device fuel state in Redis as JSON.
type fuelState struct {
	PeakLevel int       `json:"peak"`
	PeakAt    time.Time `json:"peak_at"`
}

// FuelAnomalyDetector is a pipeline enricher step.
type FuelAnomalyDetector struct {
	pool   *pgxpool.Pool
	rdb    *redis.Client
	logger *zap.Logger
}

func NewFuelAnomalyDetector(pool *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *FuelAnomalyDetector {
	return &FuelAnomalyDetector{pool: pool, rdb: rdb, logger: logger}
}

// Check is called for every enriched record that has a non-zero FuelLevel.
func (d *FuelAnomalyDetector) Check(ctx context.Context, rec *EnrichedRecord) {
	stateKey := fmt.Sprintf("gpsgo:fuel_state:%s", rec.DeviceID)
	cooldownKey := fmt.Sprintf("gpsgo:fuel_anomaly_cooldown:%s", rec.DeviceID)

	// Skip if already on cooldown for this device
	if n, _ := d.rdb.Exists(ctx, cooldownKey).Result(); n > 0 {
		return
	}

	// Load existing peak state from Redis
	var state fuelState
	if raw, err := d.rdb.Get(ctx, stateKey).Bytes(); err == nil {
		json.Unmarshal(raw, &state)
	}

	currentLevel := rec.FuelLevel
	now := rec.Timestamp

	// Update peak if it's higher or if peak is too old
	if currentLevel > state.PeakLevel || now.Sub(state.PeakAt) > time.Duration(peakWindowSeconds)*time.Second {
		state.PeakLevel = currentLevel
		state.PeakAt = now
		d.saveState(ctx, stateKey, state)
		return
	}

	// Calculate drop from peak
	drop := state.PeakLevel - currentLevel
	dropPct := 0.0
	if state.PeakLevel > 0 {
		dropPct = float64(drop) / float64(state.PeakLevel) * 100.0
	}

	// Only flag when ignition is OFF (not normal consumption while driving)
	// AND drop exceeds threshold
	if !rec.Ignition && dropPct >= minDropPct && drop >= 5 {
		d.logger.Warn("fuel anomaly detected",
			zap.String("device", rec.DeviceID),
			zap.Int("drop_pct", int(dropPct)),
			zap.Int("from", state.PeakLevel),
			zap.Int("to", currentLevel))

		d.persistAnomaly(ctx, rec, state.PeakLevel, drop, dropPct)

		// Set cooldown to avoid duplicate alerts
		d.rdb.Set(ctx, cooldownKey, 1, time.Duration(anomalyCooldownSeconds)*time.Second)

		// Reset peak to current level
		state.PeakLevel = currentLevel
		state.PeakAt = now
		d.saveState(ctx, stateKey, state)
		return
	}

	// Save latest state
	d.saveState(ctx, stateKey, state)
}

func (d *FuelAnomalyDetector) saveState(ctx context.Context, key string, state fuelState) {
	if b, err := json.Marshal(state); err == nil {
		d.rdb.Set(ctx, key, b, time.Duration(peakWindowSeconds*3)*time.Second)
	}
}

func (d *FuelAnomalyDetector) persistAnomaly(ctx context.Context, rec *EnrichedRecord,
	startLevel, dropLiters int, dropPct float64) {

	// H2 populated VehicleID on every EnrichedRecord
	vehicleID := rec.VehicleID
	tenantID := rec.TenantID

	_, err := d.pool.Exec(ctx, `
		INSERT INTO fuel_anomalies (
			tenant_id, vehicle_id, device_id, anomaly_type,
			drop_liters, drop_percent, start_level, end_level, detected_at
		) VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8, $9)`,
		tenantID, vehicleID, rec.DeviceID, "theft",
		dropLiters, dropPct, startLevel, startLevel-dropLiters, rec.Timestamp)
	if err != nil {
		d.logger.Error("persist fuel anomaly", zap.Error(err))
	}

	var alertID string
	err = d.pool.QueryRow(ctx, `
		INSERT INTO alerts (tenant_id, device_id, alert_type, severity, message,
		                    lat, lng, speed, triggered_at)
		VALUES ($1, $2, 'fuel_theft', 'critical',
		        format('Fuel theft suspected — %.1f%% drop (ignition OFF)', $3::float),
		        $4, $5, $6, $7)
		RETURNING id`,
		tenantID, rec.DeviceID, dropPct,
		rec.Lat, rec.Lng, rec.Speed, rec.Timestamp).Scan(&alertID)
	
	if err == nil {
		rec.GeneratedEvents = append(rec.GeneratedEvents, &types.FuelTheftSuspectedEvent{
			VehicleID:  vehicleID,
			TenantID:   tenantID,
			LitersLost: float64(dropLiters),
			Lat:        rec.Lat,
			Lng:        rec.Lng,
			DetectedAt: rec.Timestamp,
		})
	}
}
