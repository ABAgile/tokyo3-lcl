package lcl

import (
	"fmt"
	"reflect"
)

type Result[T any] struct {
	value T
	err   error
}

func NewResult[T any](v T, err error) *Result[T] {
	return &Result[T]{value: v, err: err}
}

func NewValueResult[T any](v T) *Result[T] {
	return NewResult(v, nil)
}

func NewErrorResult[T any](err error) *Result[T] {
	return &Result[T]{err: err}
}

func (r *Result[T]) Val() T {
	return r.value
}

func (r *Result[T]) Err() error {
	return r.err
}

func (r *Result[T]) Unwrap() (T, error) {
	return r.value, r.err
}

func (r *Result[T]) Bind(f func(v T) *Result[T]) *Result[T] {
	if r.err != nil {
		return r
	}
	return f(r.value)
}

func (r *Result[T]) MustPass(msg string, v ...any) {
	if r.err != nil {
		panic(fmt.Sprintf(msg+": "+r.err.Error(), v...))
	}
}

func (r *Result[T]) MustGet(msg string, v ...any) T {
	r.MustPass(msg, v...)
	return r.value
}

func (r *Result[T]) MustPresent(msg string, v ...any) {
	var zero T
	if reflect.DeepEqual(r.value, zero) {
		panic(fmt.Sprintf(msg, v...))
	}
}
