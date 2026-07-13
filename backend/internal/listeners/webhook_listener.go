package listeners

import (
	"context"

	"fifriz/internal/events"
)

// WebhookListener is reserved for future outbound webhook integrations.
type WebhookListener struct{}

func NewWebhookListener() *WebhookListener { return &WebhookListener{} }

func (l *WebhookListener) Register(_ events.EventBus) {}

func (l *WebhookListener) Handle(_ context.Context, _ events.Event) {}
