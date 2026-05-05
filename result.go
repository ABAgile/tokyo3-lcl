package lcl

import "fmt"

type Result[T comparable] struct {
	value T
}

func ResultVal[T comparable](v T) *Result[T] {
	return &Result[T]{value: v}
}

func (r *Result[T]) Val() T {
	return r.value
}

func (r *Result[T]) MustPass(msg string, v ...any) {
	if err, ok := any(r.value).(error); ok && err != nil {
		panicWithErr(msg, err, v...)
	}
}

func (r *Result[T]) MustPresent(msg string, v ...any) {
	if IsEmpty(r.value) {
		panic(fmt.Sprintf(msg, v...))
	}
}

type ResultE[T comparable] struct {
	value T
	err   error
}

func ResultOf[T comparable](v T, err error) *ResultE[T] {
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
		panicWithErr(msg, r.err, v...)
	}
}

func (r *ResultE[T]) MustPresent(msg string, v ...any) {
	r.MustPass(msg, v...)
	if IsEmpty(r.value) {
		panic(fmt.Sprintf(msg, v...))
	}
}

func panicWithErr(msg string, err error, v ...any) {
	panic(fmt.Sprintf(msg+": "+err.Error(), v...))
}
