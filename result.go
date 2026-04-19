package lcl

import (
	"fmt"
	"reflect"
)

// Result holds a value with no error path. Use MustPresent to assert presence.
type Result[T any] struct {
	value T
}

func ResultVal[T any](v T) *Result[T] {
	return &Result[T]{value: v}
}

func (r *Result[T]) Val() T {
	return r.value
}

func (r *Result[T]) MustPass(msg string, v ...any) {
	if err, ok := any(r.value).(error); ok && err != nil {
		panic(fmt.Sprintf(msg+": "+err.Error(), v...))
	}
}

func (r *Result[T]) MustPresent(msg string, v ...any) {
	var zero T
	if reflect.DeepEqual(r.value, zero) {
		panic(fmt.Sprintf(msg, v...))
	}
}

// ResultE holds a value and an error. Use MustPass/MustGet to assert success.
type ResultE[T any] struct {
	value T
	err   error
}

func ResultOf[T any](v T, err error) *ResultE[T] {
	return &ResultE[T]{value: v, err: err}
}

func (r *ResultE[T]) Val() T {
	return r.value
}

func (r *ResultE[T]) Err() error {
	return r.err
}

func (r *ResultE[T]) Unwrap() (T, error) {
	return r.value, r.err
}

func (r *ResultE[T]) Bind(f func(v T) *ResultE[T]) *ResultE[T] {
	if r.err != nil {
		return r
	}
	return f(r.value)
}

func (r *ResultE[T]) MustGet(msg string, v ...any) T {
	r.MustPass(msg, v...)
	return r.value
}

func (r *ResultE[T]) MustPass(msg string, v ...any) {
	if r.err != nil {
		panic(fmt.Sprintf(msg+": "+r.err.Error(), v...))
	}
}

func (r *ResultE[T]) MustPresent(msg string, v ...any) {
	r.MustPass(msg, v...)
	var zero T
	if reflect.DeepEqual(r.value, zero) {
		panic(fmt.Sprintf(msg, v...))
	}
}
