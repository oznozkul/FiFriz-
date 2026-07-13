package events

import "context"

type Handler func(ctx context.Context, event Event)

type EventBus interface {
	Publish(ctx context.Context, event Event)
	PublishAfterCommit(ctx context.Context, event Event)
	Subscribe(eventName string, handler Handler)
}
