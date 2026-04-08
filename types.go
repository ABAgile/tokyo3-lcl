package lcl

import (
	"fmt"
	"strconv"
	"strings"
)

func Empty[T any]() T {
	var zero T
	return zero
}

func IsEmpty[T comparable](v T) bool {
	var zero T
	return zero == v
}

func IsNotEmpty[T comparable](v T) bool {
	var zero T
	return zero != v
}

func Coalesce[T comparable](values ...T) (T, bool) {
	var zero T
	for i := range values {
		if values[i] != zero {
			return values[i], true
		}
	}
	return zero, false
}

func Coalesced[T comparable](values ...T) T {
	result, _ := Coalesce(values...)
	return result
}

func ToPtr[T comparable](v T) *T {
	if IsEmpty(v) {
		return nil
	}
	return &v
}

func FromPtr[T any](ptr *T, fallback ...T) T {
	if ptr != nil {
		return *ptr
	}
	if len(fallback) > 0 {
		return fallback[0]
	}
	return *new(T)
}

func ToAnySlice[T any](in []T) []any {
	result := make([]any, len(in))
	for i := range in {
		result[i] = in[i]
	}
	return result
}

func FromAnySlice[T any](in []any) ([]T, bool) {
	out := make([]T, len(in))
	for i := range in {
		t, ok := in[i].(T)
		if !ok {
			return []T{}, false
		}
		out[i] = t
	}
	return out, true
}

func GetIn(data any, path string) (any, error) {
	current := data
	paths := strings.SplitSeq(path, ".")
	for key := range paths {
		switch container := current.(type) {
		case map[string]any:
			if val, ok := container[key]; ok {
				current = val
			} else {
				return nil, fmt.Errorf("key %q not found in map", key)
			}
		case map[any]any:
			found := false
			for k, v := range container {
				if fmt.Sprintf("%v", k) == key {
					current = v
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("key %q not found in map", key)
			}
		case []any:
			index, err := strconv.Atoi(key)
			if err != nil {
				return nil, fmt.Errorf("invalid index %q for slice", key)
			}
			if index < 0 || index >= len(container) {
				return nil, fmt.Errorf("index %d out of bounds for slice", index)
			}
			current = container[index]
		default:
			return nil, fmt.Errorf("unexpected type %T at %q", current, key)
		}
	}
	return current, nil
}

func SetIn(data any, path string, value any) error {
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}
	paths := strings.Split(path, ".")
	current := data
	for _, key := range paths[:len(paths)-1] {
		switch container := current.(type) {
		case map[string]any:
			if next, ok := container[key]; ok {
				current = next
			} else {
				newMap := make(map[string]any)
				container[key] = newMap
				current = newMap
			}
		case []any:
			index, err := strconv.Atoi(key)
			if err != nil {
				return fmt.Errorf("invalid index %q for slice", key)
			}
			if index < 0 || index >= len(container) {
				return fmt.Errorf("index %d out of bounds for slice", index)
			}
			current = container[index]
		default:
			return fmt.Errorf("unexpected type %T at %q", current, key)
		}
	}
	lastKey := paths[len(paths)-1]
	switch container := current.(type) {
	case map[string]any:
		container[lastKey] = value
	case []any:
		index, err := strconv.Atoi(lastKey)
		if err != nil {
			return fmt.Errorf("invalid index %q for slice", lastKey)
		}
		if index < 0 || index >= len(container) {
			return fmt.Errorf("index %d out of bounds for slice", index)
		}
		container[index] = value
	default:
		return fmt.Errorf("unexpected type %T at %q", current, lastKey)
	}
	return nil
}

func Id[T any](v T) T {
	return v
}
