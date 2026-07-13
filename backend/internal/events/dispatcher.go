package events

import (
	"context"
	"log"
	"sync"

	"fifriz/internal/tx"
)

type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string][]Handler),
	}
}

func (d *Dispatcher) Subscribe(eventName string, handler Handler) {
	if d == nil || handler == nil || eventName == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventName] = append(d.handlers[eventName], handler)
}

func (d *Dispatcher) Publish(ctx context.Context, event Event) {
	if d == nil || event == nil {
		return
	}
	d.mu.RLock()
	handlers := append([]Handler(nil), d.handlers[event.EventName()]...)
	d.mu.RUnlock()

	for _, handler := range handlers {
		func(h Handler) {
			defer func() {
				if recovered := recover(); recovered != nil {
					log.Printf("event handler panic: event=%s error=%v", event.EventName(), recovered)
				}
			}()
			h(ctx, event)
		}(handler)
	}
}

func (d *Dispatcher) PublishAfterCommit(ctx context.Context, event Event) {
	tx.AfterCommit(ctx, func() {
		d.Publish(ctx, event)
	})
}

func NewBus() EventBus {
	return NewDispatcher()
}
