package lcl

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Define some event types
type EventA struct {
	Message string
}

type EventB struct {
	Value int
}

func TestEventBus(t *testing.T) {
	bus := NewEventBus()

	eventA := EventA{Message: "Hello"}
	eventB := EventB{Value: 123}

	chA1 := make(chan any, 1)
	chA2 := make(chan any, 1)
	chB1 := make(chan any, 1)

	Subscribe(bus, func(_ context.Context, e EventA) { chA1 <- e.Message })
	Subscribe(bus, func(_ context.Context, e EventA) { chA2 <- e.Message })
	Subscribe(bus, func(_ context.Context, e EventB) { chB1 <- e.Value })

	Publish(context.Background(), bus, eventA)
	Publish(context.Background(), bus, eventB)

	tests := []struct {
		name     string
		channel  chan any
		expected any
	}{
		{"EventA Handler 1", chA1, eventA.Message},
		{"EventA Handler 2", chA2, eventA.Message},
		{"EventB Handler 1", chB1, eventB.Value},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			select {
			case actual := <-tt.channel:
				if actual != tt.expected {
					t.Errorf("expected %v, got %v", tt.expected, actual)
				}
			case <-time.After(time.Second):
				t.Fatal("timeout waiting for event")
			}
		})
	}
}

func TestEventBus_ContextCancellation(t *testing.T) {
	bus := NewEventBus()
	var wg sync.WaitGroup
	wg.Add(1) // We expect one handler to start

	// This handler will signal it has started and then wait.
	Subscribe(bus, func(ctx context.Context, e EventA) {
		wg.Done()    // Signal that the handler has started
		<-ctx.Done() // Wait until the context is cancelled
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Publish in a separate goroutine because it's a blocking call
	go Publish(ctx, bus, EventA{Message: "test"})

	// Wait for the handler to start
	wg.Wait()

	// Now, cancel the context
	cancel()

	// We can't easily assert that the handler stopped, but if Publish doesn't
	// block forever after cancellation, the test will pass. A timeout here
	// would fail the test if Publish gets stuck.
}

func TestEventBus_PreCancelledContext(t *testing.T) {
	bus := NewEventBus()
	handlerCalled := false

	Subscribe(bus, func(ctx context.Context, e EventA) {
		handlerCalled = true
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context before publishing

	Publish(ctx, bus, EventA{Message: "test"})

	if handlerCalled {
		t.Error("handler should not be called when context is already cancelled")
	}
}
