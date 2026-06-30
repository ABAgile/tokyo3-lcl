package lcl

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errSample = errors.New("something went wrong")

func TestMustGet(t *testing.T) {
	got := MustGet(42, nil)
	assert.Equal(t, 42, got)

	err := panicError(t, func() {
		MustGet(0, errSample)
	})
	assert.Same(t, errSample, err)
}

func TestMustPass(t *testing.T) {
	assert.NotPanics(t, func() { MustPass(nil) })
	assert.NotPanics(t, func() { MustPass(nil, "context") })

	err := panicError(t, func() {
		MustPass(errSample)
	})
	assert.Same(t, errSample, err)

	err = panicError(t, func() {
		MustPass(errSample, "context")
	})
	assert.EqualError(t, err, "context: something went wrong")
	assert.ErrorIs(t, err, errSample)

	format := "failed for %s"
	err = panicError(t, func() {
		MustPass(errSample, format, "db")
	})
	assert.EqualError(t, err, "failed for db: something went wrong")
	assert.ErrorIs(t, err, errSample)
}

func TestMustPresent(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		assert.Equal(t, 1, MustPresent(1))
		err := panicError(t, func() { MustPresent(0, "msg") })
		assert.EqualError(t, err, "msg")
	})

	t.Run("string", func(t *testing.T) {
		assert.Equal(t, "hello", MustPresent("hello"))
		format := "missing %s"
		err := panicError(t, func() {
			MustPresent("", format, "name")
		})
		assert.EqualError(t, err, "missing name")
	})

	t.Run("default message", func(t *testing.T) {
		err := panicError(t, func() { MustPresent("") })
		assert.EqualError(t, err, "value is empty")
	})
}

func TestFormatMsgNonString(t *testing.T) {
	assert.Equal(t, "1 2x", formatMsg("default", 1, 2, "x"))
}

func panicError(t *testing.T, fn func()) (err error) {
	t.Helper()
	defer func() {
		r := recover()
		require.NotNil(t, r)
		var ok bool
		err, ok = r.(error)
		require.Truef(t, ok, "panic value should be an error, got %T", r)
	}()
	fn()
	return nil
}
