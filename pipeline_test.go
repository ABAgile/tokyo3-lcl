package lcl

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeContext_SendsValues(t *testing.T) {
	in := PipeContext(context.Background(), 1, 2, 3)

	var got []int
	for n := range in {
		got = append(got, n)
	}

	assert.Equal(t, []int{1, 2, 3}, got)
}

func TestPipeContext_StopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	in := PipeContext(ctx, 1, 2, 3)

	select {
	case _, ok := <-in:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("pipeline input did not stop after context cancellation")
	}
}

func TestPipe_ComposesSteps(t *testing.T) {
	ctx := context.Background()
	in := PipeContext(ctx, 1, 2, 3)

	doubled := Pipe(ctx, in, func(_ context.Context, n int) int {
		return n * 2
	})
	labels := Pipe(ctx, doubled, func(_ context.Context, n int) string {
		return fmt.Sprintf("n=%d", n)
	})

	var got []string
	for label := range labels {
		got = append(got, label)
	}

	assert.Equal(t, []string{"n=2", "n=4", "n=6"}, got)
}

func TestPipe_PassesContextToTransform(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "sentinel")
	in := make(chan int, 1)
	in <- 1
	close(in)

	out := Pipe(ctx, in, func(ctx context.Context, _ int) string {
		value, _ := ctx.Value(ctxKey{}).(string)
		return value
	})

	got, ok := <-out
	require.True(t, ok)
	assert.Equal(t, "sentinel", got)

	_, ok = <-out
	assert.False(t, ok)
}

func TestPipe_ClosesOutputWhenInputCloses(t *testing.T) {
	in := make(chan int)
	close(in)

	out := Pipe(context.Background(), in, func(_ context.Context, n int) int {
		return n
	})

	_, ok := <-out
	assert.False(t, ok)
}

func TestPipe_StopsWhenContextIsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	out := Pipe(ctx, make(chan int), func(_ context.Context, n int) int {
		return n
	})

	select {
	case _, ok := <-out:
		assert.False(t, ok)
	case <-time.After(time.Second):
		t.Fatal("pipeline did not stop after context cancellation")
	}
}

func TestPipeError_StopsAtFirstError(t *testing.T) {
	wantErr := errors.New("boom")
	in := make(chan int, 3)
	in <- 1
	in <- 2
	in <- 3
	close(in)

	out, errs := PipeError(context.Background(), in, func(_ context.Context, n int) (int, error) {
		if n == 2 {
			return 0, wantErr
		}
		return n * 10, nil
	})

	got, ok := <-out
	require.True(t, ok)
	assert.Equal(t, 10, got)

	_, ok = <-out
	assert.False(t, ok)

	err, ok := <-errs
	require.True(t, ok)
	assert.ErrorIs(t, err, wantErr)

	_, ok = <-errs
	assert.False(t, ok)
}

func TestPipeError_ClosesErrorChannelWithoutError(t *testing.T) {
	in := make(chan int, 2)
	in <- 1
	in <- 2
	close(in)

	out, errs := PipeError(context.Background(), in, func(_ context.Context, n int) (int, error) {
		return n + 1, nil
	})

	var got []int
	for n := range out {
		got = append(got, n)
	}

	assert.Equal(t, []int{2, 3}, got)

	err, ok := <-errs
	assert.False(t, ok)
	assert.NoError(t, err)
}

func TestPipeReceiveAndPipeSend(t *testing.T) {
	ctx := context.Background()
	ch := make(chan int, 1)

	assert.True(t, pipeSend(ctx, ch, 42))

	got, ok := pipeReceive(ctx, ch)
	require.True(t, ok)
	assert.Equal(t, 42, got)
}

func TestPipeReceiveAndPipeSend_CanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, ok := pipeReceive(ctx, make(chan int))
	assert.False(t, ok)
	assert.False(t, pipeSend(ctx, make(chan int), 1))
}
