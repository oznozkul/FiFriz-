package listeners

import (
	"context"

	"fifriz/internal/events"
)

// AILogicListener is reserved for future AI side effects.
type AILogicListener struct{}

func NewAILogicListener() *AILogicListener { return &AILogicListener{} }

func (l *AILogicListener) Register(_ events.EventBus) {}

func (l *AILogicListener) Handle(_ context.Context, _ events.Event) {}
