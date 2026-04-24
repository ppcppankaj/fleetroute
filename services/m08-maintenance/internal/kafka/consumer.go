package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
)

type MaintenanceService interface {
	CheckOverdueAndEmit(ctx context.Context) error
}

type Consumer struct {
	reader  *kafka.Reader
	service MaintenanceService
}

func NewConsumer(brokers string, service MaintenanceService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m08-maintenance-group",
			Topic:   shared.TopicTripCompleted,
		}),
		service: service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	// Also run a ticker to check for overdue tasks every hour
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = c.service.CheckOverdueAndEmit(ctx)
			}
		}
	}()

	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka consumer error: %v", err)
			continue
		}
		var evt sharedtypes.TripCompletedEvent
		if err := json.Unmarshal(msg.Value, &evt); err == nil {
			// Could check odometer-based maintenance due after trip completion
			_ = c.service.CheckOverdueAndEmit(ctx)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
