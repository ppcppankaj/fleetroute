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

type AlertService interface {
	CheckSpeeding(ctx context.Context, evt sharedtypes.LocationUpdatedEvent)
	CheckGeofence(ctx context.Context, evt sharedtypes.GeofenceBreachEvent)
}

type Consumer struct {
	readerLocation *kafka.Reader
	readerGeofence *kafka.Reader
	service        AlertService
}

func NewConsumer(brokers string, service AlertService) *Consumer {
	brokersList := strings.Split(brokers, ",")
	return &Consumer{
		readerLocation: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokersList,
			GroupID: "m04-alerts-group",
			Topic:   shared.TopicLocationUpdated,
		}),
		readerGeofence: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokersList,
			GroupID: "m04-alerts-group-geo", // different topic needs different or same group, we use a separate reader
			Topic:   shared.TopicGeofenceBreach,
		}),
		service: service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.readerLocation.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { return }
				log.Printf("kafka reader error: %v", err)
				continue
			}
			var evt sharedtypes.LocationUpdatedEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				c.service.CheckSpeeding(ctx, evt)
			}
		}
	}()

	go func() {
		for {
			msg, err := c.readerGeofence.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { return }
				log.Printf("kafka reader error: %v", err)
				continue
			}
			var evt sharedtypes.GeofenceBreachEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				c.service.CheckGeofence(ctx, evt)
			}
		}
	}()
}

func (c *Consumer) Close() {
	_ = c.readerLocation.Close()
	_ = c.readerGeofence.Close()
}
