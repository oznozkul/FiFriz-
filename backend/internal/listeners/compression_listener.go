package listeners

import (
	"context"

	"fifriz/internal/events"
	domain "fifriz/internal/events/events"
	"fifriz/internal/jobs"
)

type CompressionListener struct {
	queue jobs.Enqueuer
}

func NewCompressionListener(queue jobs.Enqueuer) *CompressionListener {
	return &CompressionListener{queue: queue}
}

func (l *CompressionListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	bus.Subscribe(domain.DocumentCreatedName, l.Handle)
}

func (l *CompressionListener) Handle(ctx context.Context, event events.Event) {
	if l == nil || l.queue == nil {
		return
	}
	created, ok := event.(domain.DocumentCreated)
	if !ok {
		return
	}
	_ = jobs.EnqueueCompressDocument(ctx, l.queue, created.TenantID(), created.DocumentID)
}
