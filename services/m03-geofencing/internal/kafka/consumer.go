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

type GeofenceService interface {
	EvaluateLocation(ctx context.Context, evt sharedtypes.LocationUpdatedEvent)
}

type Consumer struct {
	reader  *kafka.Reader
	service GeofenceService
}

func NewConsumer(brokers string, service GeofenceService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m03-geofencing-group",
			Topic:   shared.TopicLocationUpdated,
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

		if msg.Topic == shared.TopicLocationUpdated {
			var evt sharedtypes.LocationUpdatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				c.service.EvaluateLocation(ctx, evt)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
