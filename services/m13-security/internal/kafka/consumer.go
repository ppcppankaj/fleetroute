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

type SecurityService interface {
	RecordUserLogin(ctx context.Context, evt sharedtypes.UserLoginEvent)
	RecordUserAction(ctx context.Context, evt sharedtypes.UserActionEvent)
}

type Consumer struct {
	readerLogin  *kafka.Reader
	readerAction *kafka.Reader
	service      SecurityService
}

func NewConsumer(brokers string, service SecurityService) *Consumer {
	bl := strings.Split(brokers, ",")
	return &Consumer{
		readerLogin: kafka.NewReader(kafka.ReaderConfig{
			Brokers: bl,
			GroupID: "m13-security-login",
			Topic:   shared.TopicUserLogin,
		}),
		readerAction: kafka.NewReader(kafka.ReaderConfig{
			Brokers: bl,
			GroupID: "m13-security-action",
			Topic:   shared.TopicUserAction,
		}),
		service: service,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.readerLogin.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { return }
				log.Printf("m13 kafka login error: %v", err)
				continue
			}
			var evt sharedtypes.UserLoginEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				c.service.RecordUserLogin(ctx, evt)
			}
		}
	}()

	go func() {
		for {
			msg, err := c.readerAction.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil { return }
				log.Printf("m13 kafka action error: %v", err)
				continue
			}
			var evt sharedtypes.UserActionEvent
			if err := json.Unmarshal(msg.Value, &evt); err == nil {
				c.service.RecordUserAction(ctx, evt)
			}
		}
	}()
}

func (c *Consumer) Close() {
	_ = c.readerLogin.Close()
	_ = c.readerAction.Close()
}
