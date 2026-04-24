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

type VehicleService interface {
	UpdateVehicleOdometer(ctx context.Context, vehicleID string, distanceKm float64) error
}

type Consumer struct {
	reader  *kafka.Reader
	service VehicleService
}

func NewConsumer(brokers string, service VehicleService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m06-vehicles-group",
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

		if msg.Topic == shared.TopicTripCompleted {
			var evt sharedtypes.TripCompletedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				_ = c.service.UpdateVehicleOdometer(ctx, evt.VehicleID, evt.DistanceKM)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
