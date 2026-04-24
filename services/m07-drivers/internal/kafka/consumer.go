package kafka

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
)

type DriverService interface {
	UpdateDriverBehaviorScore(ctx context.Context, driverID string, points float64) error
}

type Consumer struct {
	reader  *kafka.Reader
	service DriverService
}

func NewConsumer(brokers string, service DriverService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m07-drivers-group",
			Topic:   shared.TopicAlertTriggered, // Assume bad behavior creates an alert
		}),
		service: service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("kafka consumer error: %v", err)
			continue
		}

		if msg.Topic == shared.TopicAlertTriggered {
			var evt sharedtypes.AlertTriggeredEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				if evt.DriverID != "" && evt.Type == "BEHAVIOR" {
					// example: deduct 2 points per behavior alert
					_ = c.service.UpdateDriverBehaviorScore(ctx, evt.DriverID, 2.0)
				}
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
