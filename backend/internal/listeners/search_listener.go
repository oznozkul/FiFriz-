package listeners

import (
	"context"

	"fifriz/internal/events"
)

// SearchListener is a placeholder for future search index refresh hooks.
type SearchListener struct{}

func NewSearchListener() *SearchListener { return &SearchListener{} }

func (l *SearchListener) Register(bus events.EventBus) {
	if l == nil || bus == nil {
		return
	}
	// Search is currently PostgreSQL FTS based; no external index refresh required.
}

func (l *SearchListener) Handle(_ context.Context, _ events.Event) {}
