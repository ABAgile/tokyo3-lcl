package lcl

import "context"

// PipeContext sends values into a new pipeline input channel.
//
// The output channel is closed after all values are sent or ctx is done.
func PipeContext[T any](ctx context.Context, values ...T) <-chan T {
	out := make(chan T)

	go func() {
		defer close(out)
		for _, val := range values {
			if !pipeSend(ctx, out, val) {
				return
			}
		}
	}()

	return out
}

// Pipe transforms values from in and returns a channel of transformed values.
//
// The output channel is closed when the input channel is closed or ctx is done.
// Callers should either drain the output channel or cancel ctx to let the
// pipeline goroutine exit.
func Pipe[I, O any](ctx context.Context, in <-chan I, transform func(context.Context, I) O) <-chan O {
	out := make(chan O)

	go func() {
		defer close(out)
		for {
			val, ok := pipeReceive(ctx, in)
			if !ok {
				return
			}

			if !pipeSend(ctx, out, transform(ctx, val)) {
				return
			}
		}
	}()

	return out
}

// PipeError is like Pipe, but stops at the first transform error.
//
// The error channel is buffered for one error and is closed when the pipeline
// exits. It receives transform errors only; cancellation is exposed through ctx.
func PipeError[I, O any](ctx context.Context, in <-chan I, transform func(context.Context, I) (O, error)) (<-chan O, <-chan error) {
	out := make(chan O)
	errs := make(chan error, 1)

	go func() {
		defer close(out)
		defer close(errs)
		for {
			val, ok := pipeReceive(ctx, in)
			if !ok {
				return
			}

			newVal, err := transform(ctx, val)
			if err != nil {
				errs <- err
				return
			}

			if !pipeSend(ctx, out, newVal) {
				return
			}
		}
	}()

	return out, errs
}

func pipeReceive[T any](ctx context.Context, in <-chan T) (T, bool) {
	var zero T
	if ctx.Err() != nil {
		return zero, false
	}

	select {
	case <-ctx.Done():
		return zero, false
	case val, ok := <-in:
		return val, ok
	}
}

func pipeSend[T any](ctx context.Context, out chan<- T, val T) bool {
	if ctx.Err() != nil {
		return false
	}

	select {
	case <-ctx.Done():
		return false
	case out <- val:
		return true
	}
}
