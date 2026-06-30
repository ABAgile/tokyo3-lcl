package lcl

import (
	"errors"
	"fmt"
)

func MustGet[T any](value T, err error) T {
	if err != nil {
		MustPass(err)
	}
	return value
}

func MustPass(err error, msg ...any) {
	if err != nil {
		if message := formatMsg("", msg...); message != "" {
			panic(fmt.Errorf("%s: %w", message, err))
		}
		panic(err)
	}
}

func MustPresent[T comparable](value T, msg ...any) T {
	if IsEmpty(value) {
		panic(errors.New(formatMsg("value is empty", msg...)))
	}
	return value
}

func formatMsg(defaultMsg string, msg ...any) string {
	if len(msg) == 0 {
		return defaultMsg
	}

	format, ok := msg[0].(string)
	if !ok {
		// Use a full-slice expression to skip go vet wrapper inference.
		return fmt.Sprint(msg[:]...)
	}
	if len(msg) == 1 {
		return format
	}
	return fmt.Sprintf(format, msg[1:]...)
}
