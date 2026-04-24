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

type DeviceService interface {
	UpdateDeviceStatusFromOfflineEvent(ctx context.Context, deviceID string) error
}

type Consumer struct {
	reader  *kafka.Reader
	service DeviceService
}

func NewConsumer(brokers string, service DeviceService) *Consumer {
	return &Consumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: strings.Split(brokers, ","),
			GroupID: "m12-devices-group",
			Topic:   shared.TopicDeviceOffline,
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

		if msg.Topic == shared.TopicDeviceOffline {
			var evt sharedtypes.DeviceOfflineEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				_ = c.service.UpdateDeviceStatusFromOfflineEvent(ctx, evt.DeviceID)
			}
		}
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
