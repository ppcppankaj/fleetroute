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

type FuelService interface {
	ProcessTripCompleted(ctx context.Context, evt sharedtypes.TripCompletedEvent)
}

type Consumer struct {
	reader  *kafka.Reader
	service FuelService
}

func NewConsumer(brokers string, service FuelService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m09-fuel-group",
			Topic:   shared.TopicTripCompleted,
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
		var evt sharedtypes.TripCompletedEvent
		if err := json.Unmarshal(msg.Value, &evt); err == nil {
			c.service.ProcessTripCompleted(ctx, evt)
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
