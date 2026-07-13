package extensions

import (
	"context"

	"fifriz/internal/events"
)

type Extension interface {
	Name() string
	Enabled() bool
	Handle(ctx context.Context, event events.Event) error
}

type Registry struct {
	extensions []Extension
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) Register(extension Extension) {
	if r == nil || extension == nil || !extension.Enabled() {
		return
	}
	r.extensions = append(r.extensions, extension)
}

func (r *Registry) Handle(ctx context.Context, event events.Event) {
	if r == nil {
		return
	}
	for _, extension := range r.extensions {
		_ = extension.Handle(ctx, event)
	}
}
