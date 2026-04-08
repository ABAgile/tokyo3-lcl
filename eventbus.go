package lcl

import (
	"context"
	"reflect"
	"sync"
)

type Handler[T any] func(context.Context, T)

type EventBus struct {
	lock     sync.RWMutex
	handlers map[reflect.Type][]any
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[reflect.Type][]any),
	}
}

func Subscribe[T any](bus *EventBus, handler Handler[T]) {
	bus.lock.Lock()
	defer bus.lock.Unlock()

	// Get the static type from the generic parameter T, not the concrete event value.
	eventType := reflect.TypeFor[T]()
	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
}

// Publish sends an event to all registered handlers for that event type.
// It passes a context that can be used for cancellation.
func Publish[T any](ctx context.Context, bus *EventBus, event T) {
	if ctx.Err() != nil {
		return
	}

	// Get the static type from the generic parameter T, not the concrete event value.
	eventType := reflect.TypeFor[T]()

	bus.lock.RLock()
	handlers := append([]any(nil), bus.handlers[eventType]...)
	bus.lock.RUnlock()

	var wg sync.WaitGroup
	for _, handler := range handlers {
		wg.Add(1)
		go func(h any) {
			defer wg.Done()
			h.(Handler[T])(ctx, event)
		}(handler)
	}
	wg.Wait()
}
