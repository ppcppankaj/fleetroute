package mqtt

import (
	"context"
	"encoding/json"
	"log"

	paho "github.com/eclipse/paho.mqtt.golang"
	sharedtypes "gpsgo/shared/types"
)

type LocationHandler func(ctx context.Context, evt sharedtypes.LocationUpdatedEvent)

type Client struct {
	client paho.Client
}

func NewClient(brokerURL string, handler LocationHandler) *Client {
	opts := paho.NewClientOptions().AddBroker(brokerURL).SetClientID("m01-live-tracking-ingest")
	
	opts.SetDefaultPublishHandler(func(client paho.Client, msg paho.Message) {
		var evt sharedtypes.LocationUpdatedEvent
		if err := json.Unmarshal(msg.Payload(), &evt); err == nil {
			handler(context.Background(), evt)
		}
	})

	c := paho.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt connect error: %v", token.Error())
	}

	// Devices push to fleet/telemetry/location
	if token := c.Subscribe("fleet/telemetry/location", 1, nil); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt subscribe error: %v", token.Error())
	}

	return &Client{client: c}
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
