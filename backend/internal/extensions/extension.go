package extensions

import (
	"context"

	"fifriz/internal/events"
)

// ExtensionHandler bridges the extension registry to the event bus.
type ExtensionHandler struct {
	registry *Registry
}

func NewExtensionHandler(registry *Registry) *ExtensionHandler {
	return &ExtensionHandler{registry: registry}
}

func (h *ExtensionHandler) Handle(ctx context.Context, event events.Event) {
	if h == nil || h.registry == nil || event == nil {
		return
	}
	h.registry.Handle(ctx, event)
}

func (h *ExtensionHandler) Register(bus events.EventBus, eventNames ...string) {
	if h == nil || bus == nil {
		return
	}
	for _, name := range eventNames {
		bus.Subscribe(name, h.Handle)
	}
}
