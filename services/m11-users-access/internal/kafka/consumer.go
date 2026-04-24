package kafka

import "context"

func StartConsumer(ctx context.Context, handler func(context.Context, []byte) error) {
	go func() {
		<-ctx.Done()
	}()
	_ = handler
}
