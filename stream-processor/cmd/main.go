// Package main is the entry point for the stream processor service.
// It consumes raw AVL records from NATS JetStream, runs enrichment, writes to
// TimescaleDB, updates Redis live state, and re-publishes enriched events.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"


	pkgdb "gpsgo/pkg/db"
	natsclient "gpsgo/pkg/nats"
	"gpsgo/pkg/protocol"
	"gpsgo/shared/kafka"
	"gpsgo/shared/types"
	"gpsgo/stream-processor/internal/enrichment"
	"gpsgo/stream-processor/internal/writer"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() //nolint:errcheck

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ── NATS ──────────────────────────────────────────────────────────────────
	nc, err := natsclient.New(envStr("NATS_URL", "nats://localhost:4222"))
	if err != nil {
		logger.Fatal("NATS connect", zap.Error(err))
	}
	defer nc.Close()

	// ── TimescaleDB ───────────────────────────────────────────────────────────
	pool, err := pkgdb.NewPool(ctx, envStr("TIMESCALE_DSN", ""))
	if err != nil {
		logger.Fatal("TimescaleDB connect", zap.Error(err))
	}
	defer pool.Close()

	// ── Redis ─────────────────────────────────────────────────────────────────
	rdb, err := pkgdb.NewRedis(ctx, pkgdb.RedisConfig{
		Addr:     envStr("REDIS_ADDR", "localhost:6379"),
		Password: envStr("REDIS_PASSWORD", ""),
	})
	if err != nil {
		logger.Fatal("Redis connect", zap.Error(err))
	}
	defer rdb.Close()

	// ── Services ──────────────────────────────────────────────────────────────
	tsWriter := writer.NewTimescaleWriter(pool, logger)
	redisWriter := writer.NewRedisWriter(rdb, logger)
	pipeline := enrichment.NewPipeline(pool, rdb, logger)

	// ── Kafka Producer (M2 Bridge) ────────────────────────────────────────────
	kafkaBrokers := []string{envStr("KAFKA_BROKERS", "localhost:29092")}
	kafkaProducer := kafka.NewProducer(kafkaBrokers)
	defer kafkaProducer.Close()

	// ── NATS JetStream Consumer ───────────────────────────────────────────────
	js := nc.JetStream()
	consumer, err := js.CreateOrUpdateConsumer(ctx, natsclient.StreamAVL, jetstream.ConsumerConfig{
		Name:          "stream-processor",
		Durable:       "stream-processor",
		FilterSubject: natsclient.SubjectRawAVL,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    5,
		AckWait:       30 * time.Second,
	})
	if err != nil {
		logger.Fatal("create consumer", zap.Error(err))
	}

	logger.Info("stream-processor started, consuming from NATS")

	iter, err := consumer.Messages()
	if err != nil {
		logger.Fatal("messages iterator", zap.Error(err))
	}

	go func() {
		<-ctx.Done()
		iter.Stop()
	}()

	for {
		msg, err := iter.Next()
		if err != nil {
			break
		}
		processMessage(ctx, msg, pipeline, tsWriter, redisWriter, nc, kafkaProducer, logger)
	}

	logger.Info("stream processor stopped")
}

func processMessage(
	ctx context.Context,
	msg jetstream.Msg,
	pipeline *enrichment.Pipeline,
	tsWriter *writer.TimescaleWriter,
	redisWriter *writer.RedisWriter,
	nc *natsclient.Client,
	kafkaProducer *kafka.Producer,
	logger *zap.Logger,
) {
	var raw protocol.ParsedRecord
	if err := json.Unmarshal(msg.Data(), &raw); err != nil {
		logger.Error("unmarshal raw record", zap.Error(err))
		msg.Ack() //nolint:errcheck
		return
	}

	// ── Enrichment Pipeline ────────────────────────────────────────────────────
	enriched := pipeline.Process(ctx, raw)
	if enriched == nil {
		msg.Ack() //nolint:errcheck
		return
	}

	// ── Write to TimescaleDB ──────────────────────────────────────────────────
	if err := tsWriter.Write(ctx, enriched); err != nil {
		logger.Error("timescale write", zap.String("device_id", raw.DeviceID), zap.Error(err))
		// Don't NAK — will be retried up to MaxDeliver times
		msg.Nak() //nolint:errcheck
		return
	}

	// ── Update Redis live state ───────────────────────────────────────────────
	if err := redisWriter.UpdateLive(ctx, enriched); err != nil {
		logger.Warn("redis update", zap.String("device_id", raw.DeviceID), zap.Error(err))
		// Non-fatal: continue with ACK
	}

	// ── Publish enriched event ────────────────────────────────────────────────
	enrichedData, err := json.Marshal(enriched)
	if err == nil {
		msgID := fmt.Sprintf("%s-%d-enriched", enriched.DeviceID, enriched.Timestamp.UnixMilli())
		nc.Publish(ctx, natsclient.SubjectEnrichedAVL, msgID, enrichedData) //nolint:errcheck
	}

	// ── Publish to Redis Pub/Sub for WebSocket fan-out ────────────────────────
	if err := redisWriter.PublishLive(ctx, enriched); err != nil {
		logger.Warn("redis pubsub publish", zap.Error(err))
	}

	// ── Bridge to Kafka (M2) ──────────────────────────────────────────────────
	locEvent := &types.LocationUpdatedEvent{
		VehicleID:      enriched.DeviceID, // using device ID as fallback if vehicle ID not available
		TenantID:       enriched.TenantID,
		Lat:            enriched.Lat,
		Lng:            enriched.Lng,
		Speed:          float64(enriched.Speed),
		Heading:        float64(enriched.Heading),
		Altitude:       float64(enriched.Altitude),
		Ignition:       enriched.Ignition,
		SignalStrength: enriched.GSMSignal,
		Timestamp:      enriched.Timestamp,
	}
	locEventData, _ := json.Marshal(locEvent)
	if err := kafkaProducer.Publish(ctx, types.TopicLocationUpdated, locEvent.VehicleID, locEventData); err != nil {
		logger.Warn("kafka bridge publish", zap.Error(err))
	}

	msg.Ack() //nolint:errcheck
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}


