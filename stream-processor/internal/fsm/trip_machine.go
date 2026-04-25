package fsm

import (
	"context"
	"fmt"
	"math"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"gpsgo/stream-processor/internal/enrichment"
)

const (
	StateIdle   = "IDLE"
	StateActive = "ACTIVE"
	StateEnded  = "ENDED"
)

type TripMachine struct {
	pool   *pgxpool.Pool
	rdb    *redis.Client
	logger *zap.Logger
}

func NewTripMachine(pool *pgxpool.Pool, rdb *redis.Client, logger *zap.Logger) *TripMachine {
	return &TripMachine{pool: pool, rdb: rdb, logger: logger}
}

func (m *TripMachine) Process(ctx context.Context, rec *enrichment.EnrichedRecord) {
	stateKey := fmt.Sprintf("gpsgo:tripstate:%s", rec.DeviceID)
	tripIdKey := fmt.Sprintf("gpsgo:tripstate:%s:tripid", rec.DeviceID)
	metricsKey := fmt.Sprintf("gpsgo:tripstate:%s:metrics", rec.DeviceID)

	state, err := m.rdb.Get(ctx, stateKey).Result()
	if err == redis.Nil {
		state = StateIdle
	}

	if state == StateIdle && rec.Speed > 5 {
		// Transition to ACTIVE
		var newTripID string
		err := m.pool.QueryRow(ctx, `
			INSERT INTO trips (tenant_id, device_id, started_at, start_lat, start_lng)
			VALUES ($1, $2, $3, $4, $5) RETURNING id
		`, rec.TenantID, rec.DeviceID, rec.Timestamp, rec.Lat, rec.Lng).Scan(&newTripID)

		if err != nil {
			m.logger.Error("failed to create trip", zap.Error(err))
			return
		}

		m.rdb.Set(ctx, stateKey, StateActive, 0)
		m.rdb.Set(ctx, tripIdKey, newTripID, 0)

		m.rdb.HSet(ctx, metricsKey,
			"distance", 0.0,
			"max_speed", rec.Speed,
			"harsh_accel", 0,
			"harsh_brake", 0,
			"overspeed", 0,
			"idle_s", 0,
			"last_lat", rec.Lat,
			"last_lng", rec.Lng,
			"last_speed", rec.Speed,
			"last_ts", rec.Timestamp.Unix(),
		)
		m.logger.Info("started trip", zap.String("trip_id", newTripID))

	} else if state == StateActive {
		if rec.Speed < 2 {
			// Transition to ENDED logic
			tripID, _ := m.rdb.Get(ctx, tripIdKey).Result()
			if tripID != "" {
				metrics, _ := m.rdb.HGetAll(ctx, metricsKey).Result()

				distance := parseFloat(metrics["distance"])
				maxSpeed := parseInt(metrics["max_speed"])
				hAccel := parseInt(metrics["harsh_accel"])
				hBrake := parseInt(metrics["harsh_brake"])
				overspeed := parseInt(metrics["overspeed"])
				idleS := parseInt(metrics["idle_s"])

				_, err := m.pool.Exec(ctx, `
					UPDATE trips 
					SET ended_at = $1, end_lat = $2, end_lng = $3,
					    distance_m = $5, duration_s = EXTRACT(EPOCH FROM ($1 - started_at)),
					    max_speed = $6, harsh_accel = $7, harsh_brake = $8, overspeed_count = $9, idle_time_s = $10
					WHERE id = $4
				`, rec.Timestamp, rec.Lat, rec.Lng, tripID, distance, maxSpeed, hAccel, hBrake, overspeed, idleS)

				if err != nil {
					m.logger.Error("failed to end trip", zap.Error(err))
				} else {
					m.logger.Info("ended trip", zap.String("trip_id", tripID))
				}
			}

			m.rdb.Set(ctx, stateKey, StateIdle, 0)
			m.rdb.Del(ctx, tripIdKey, metricsKey)
		} else {
			// Update ongoing metrics
			metrics, err := m.rdb.HGetAll(ctx, metricsKey).Result()
			if err == nil && len(metrics) > 0 {
				lastLat := parseFloat(metrics["last_lat"])
				lastLng := parseFloat(metrics["last_lng"])
				lastSpeed := parseInt(metrics["last_speed"])
				lastTs := parseInt(metrics["last_ts"])

				distDelta := haversine(lastLat, lastLng, rec.Lat, rec.Lng)
				timeDelta := rec.Timestamp.Unix() - int64(lastTs)

				pipe := m.rdb.Pipeline()
				pipe.HIncrByFloat(ctx, metricsKey, "distance", distDelta)

				if int(rec.Speed) > parseInt(metrics["max_speed"]) {
					pipe.HSet(ctx, metricsKey, "max_speed", rec.Speed)
				}

				if timeDelta > 0 && timeDelta < 300 {
					speedDiff := int(rec.Speed) - lastSpeed
					if speedDiff > 15 {
						pipe.HIncrBy(ctx, metricsKey, "harsh_accel", 1)
					} else if speedDiff < -15 {
						pipe.HIncrBy(ctx, metricsKey, "harsh_brake", 1)
					}
				}

				if rec.Speed > 80 {
					pipe.HIncrBy(ctx, metricsKey, "overspeed", 1)
				}

				if rec.Speed == 0 && rec.Ignition && timeDelta > 0 && timeDelta < 300 {
					pipe.HIncrBy(ctx, metricsKey, "idle_s", timeDelta)
				}

				pipe.HSet(ctx, metricsKey,
					"last_lat", rec.Lat,
					"last_lng", rec.Lng,
					"last_speed", rec.Speed,
					"last_ts", rec.Timestamp.Unix(),
				)
				pipe.Exec(ctx) //nolint:errcheck
			}
		}
	}
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // Earth radius in meters
	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
