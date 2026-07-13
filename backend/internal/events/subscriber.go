package events

type Subscriber struct {
	EventName string
	Handler   Handler
}

func (s Subscriber) Register(bus EventBus) {
	if bus == nil || s.Handler == nil || s.EventName == "" {
		return
	}
	bus.Subscribe(s.EventName, s.Handler)
}
