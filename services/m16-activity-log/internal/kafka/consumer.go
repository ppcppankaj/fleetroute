package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/segmentio/kafka-go"
	sharedtypes "gpsgo/shared/types"
	shared "gpsgo/shared/kafka"
)

type ActivityService interface {
	LogAlert(ctx context.Context, evt sharedtypes.AlertTriggeredEvent)
	LogTrip(ctx context.Context, evt sharedtypes.TripStartedEvent)
	LogTripCompleted(ctx context.Context, evt sharedtypes.TripCompletedEvent)
}

type Consumer struct {
	readers []*kafka.Reader
	service ActivityService
}

func NewConsumer(brokers string, service ActivityService) *Consumer {
	bl := strings.Split(brokers, ",")
	topics := []string{shared.TopicAlertTriggered, shared.TopicTripStarted, shared.TopicTripCompleted}
	var readers []*kafka.Reader
	for i, t := range topics {
		readers = append(readers, kafka.NewReader(kafka.ReaderConfig{
			Brokers: bl,
			GroupID: fmt.Sprintf("m16-activity-group-%d", i),
			Topic:   t,
		}))
	}
	return &Consumer{readers: readers, service: service}
}

func (c *Consumer) Start(ctx context.Context) {
	consume := func(r *kafka.Reader, topic string) {
		for {
			msg, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { return }
				log.Printf("m16 kafka error [%s]: %v", topic, err)
				continue
			}
			switch msg.Topic {
			case shared.TopicAlertTriggered:
				var evt sharedtypes.AlertTriggeredEvent
				if err := json.Unmarshal(msg.Value, &evt); err == nil {
					c.service.LogAlert(ctx, evt)
				}
			case shared.TopicTripStarted:
				var evt sharedtypes.TripStartedEvent
				if err := json.Unmarshal(msg.Value, &evt); err == nil {
					c.service.LogTrip(ctx, evt)
				}
			case shared.TopicTripCompleted:
				var evt sharedtypes.TripCompletedEvent
				if err := json.Unmarshal(msg.Value, &evt); err == nil {
					c.service.LogTripCompleted(ctx, evt)
				}
			}
		}
	}

	for _, r := range c.readers {
		go consume(r, r.Config().Topic)
	}
}

func (c *Consumer) Close() {
	for _, r := range c.readers {
		_ = r.Close()
	}
}
