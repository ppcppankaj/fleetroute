package kafka

import "context"

type Producer struct{}

func (p *Producer) Publish(_ context.Context, _ string, _ []byte) error {
	return nil
}
