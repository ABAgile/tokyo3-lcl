package lcl

import "fmt"

func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}

func MustMsg[T any](value T, err error, msg string, args ...any) T {
	if err != nil {
		panic(formatWithErr(msg, err, args...))
	}
	return value
}

func MustPass(err error, msg string, args ...any) {
	if err != nil {
		panic(formatWithErr(msg, err, args...))
	}
}

func MustPresent[T comparable](value T, msg string, args ...any) T {
	if IsEmpty(value) {
		panic(fmt.Sprintf(msg, args...))
	}
	return value
}

func formatWithErr(msg string, err error, args ...any) string {
	if msg == "" {
		return err.Error()
	}
	return fmt.Sprintf(msg, args...) + ": " + err.Error()
}
