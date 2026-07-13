package events

import (
	"context"

	"fifriz/internal/tx"
)

type Publisher interface {
	Publish(ctx context.Context, event Event)
	PublishAfterCommit(ctx context.Context, event Event)
}

type BusPublisher struct {
	bus EventBus
}

func NewPublisher(bus EventBus) Publisher {
	return &BusPublisher{bus: bus}
}

func (p *BusPublisher) Publish(ctx context.Context, event Event) {
	if p == nil || p.bus == nil || event == nil {
		return
	}
	p.bus.Publish(ctx, event)
}

func (p *BusPublisher) PublishAfterCommit(ctx context.Context, event Event) {
	if p == nil || p.bus == nil || event == nil {
		return
	}
	tx.AfterCommit(ctx, func() {
		p.bus.Publish(ctx, event)
	})
}
