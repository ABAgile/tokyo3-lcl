package lcl

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSample = errors.New("something went wrong")

func TestMust(t *testing.T) {
	got := Must(42, nil)
	assert.Equal(t, 42, got)

	assert.PanicsWithValue(t, errSample, func() {
		Must(0, errSample)
	})
}

func TestMustMsg(t *testing.T) {
	got := MustMsg("hello", nil, "context")
	assert.Equal(t, "hello", got)

	assert.PanicsWithValue(t, fmt.Sprintf("context: %s", errSample), func() {
		MustMsg("", errSample, "context")
	})

	assert.PanicsWithValue(t, fmt.Sprintf("failed for db: %s", errSample), func() {
		MustMsg("", errSample, "failed for %s", "db")
	})
}

func TestMustPass(t *testing.T) {
	assert.NotPanics(t, func() { MustPass(nil, "context") })
	assert.PanicsWithValue(t, fmt.Sprintf("context: %s", errSample), func() {
		MustPass(errSample, "context")
	})
}

func TestMustPresent(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, 1, MustPresent(1, "msg"))
		assert.PanicsWithValue(t, "msg", func() { MustPresent(0, "msg") })
	})

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, "hello", MustPresent("hello", "msg"))
		assert.PanicsWithValue(t, "missing name", func() {
			MustPresent("", "missing %s", "name")
		})
	})
}

func TestFormatWithErrEmptyMessage(t *testing.T) {
	assert.Equal(t, errSample.Error(), formatWithErr("", errSample))
}
