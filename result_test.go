package lcl

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var errSample = errors.New("something went wrong")

func TestResultOf(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name    string
			value   int
			err     error
			wantErr bool
		}{
			{"with value", 42, nil, false},
			{"with error", 0, errSample, true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				r := ResultOf(tc.value, tc.err)
				assert.Equal(t, tc.value, r.Val())
				if tc.wantErr {
					assert.ErrorIs(t, r.Err(), tc.err)
				} else {
					assert.NoError(t, r.Err())
				}
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name    string
			value   string
			err     error
			wantErr bool
		}{
			{"with value", "hello", nil, false},
			{"with error", "", errSample, true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				r := ResultOf(tc.value, tc.err)
				assert.Equal(t, tc.value, r.Val())
				if tc.wantErr {
					assert.ErrorIs(t, r.Err(), tc.err)
				} else {
					assert.NoError(t, r.Err())
				}
			})
		}
	})
}

func TestResultVal(t *testing.T) {
	r := ResultVal("hello")
	assert.Equal(t, "hello", r.Val())
}


func TestUnwrap(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *ResultE[int]
			wantValue int
			wantErr   error
		}{
			{"returns value", ResultOf(42, nil), 42, nil},
			{"returns error", ResultOf(0, errSample), 0, errSample},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.result.Unwrap()
				assert.Equal(t, tc.wantValue, got)
				assert.ErrorIs(t, err, tc.wantErr)
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *ResultE[string]
			wantValue string
			wantErr   error
		}{
			{"returns value", ResultOf("hello", nil), "hello", nil},
			{"returns error", ResultOf("", errSample), "", errSample},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				got, err := tc.result.Unwrap()
				assert.Equal(t, tc.wantValue, got)
				assert.ErrorIs(t, err, tc.wantErr)
			})
		}
	})
}

func TestBind(t *testing.T) {
	double := func(v int) *ResultE[int] { return ResultOf(v*2, nil) }

	tests := []struct {
		name      string
		result    *ResultE[int]
		wantValue int
		wantErr   bool
	}{
		{"applies on success, value changes", ResultOf(5, nil), 10, false},
		{"skips on error, value unchanged", ResultOf(5, errSample), 5, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := tc.result.Bind(double)
			assert.Equal(t, tc.wantValue, r.Val())
			if tc.wantErr {
				assert.ErrorIs(t, r.Err(), errSample)
			} else {
				assert.NoError(t, r.Err())
			}
		})
	}
}

func TestResultMustPass(t *testing.T) {
	tests := []struct {
		name      string
		result    *Result[error]
		wantPanic bool
		wantMsg   string
	}{
		{"nil error does not panic", ResultVal[error](nil), false, ""},
		{"non-nil error panics", ResultVal(errSample), true, fmt.Sprintf("context: %s", errSample)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assert.PanicsWithValue(t, tc.wantMsg, func() { tc.result.MustPass("context") })
			} else {
				assert.NotPanics(t, func() { tc.result.MustPass("context") })
			}
		})
	}
}

func TestMustPass(t *testing.T) {
	tests := []struct {
		name      string
		result    *ResultE[any]
		wantPanic bool
		wantMsg   string
	}{
		{"nil error does not panic", ResultOf[any](nil, nil), false, ""},
		{"error panics with message", ResultOf[any](nil, errSample), true, fmt.Sprintf("context: %s", errSample)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assert.PanicsWithValue(t, tc.wantMsg, func() { tc.result.MustPass("context") })
			} else {
				assert.NotPanics(t, func() { tc.result.MustPass("context") })
			}
		})
	}
}

func TestMustGet(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *ResultE[int]
			want      int
			wantPanic bool
		}{
			{"returns value on success", ResultOf(7, nil), 7, false},
			{"panics on error", ResultOf(0, errSample), 0, true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.wantPanic {
					assert.Panics(t, func() { tc.result.MustGet("context") })
				} else {
					assert.Equal(t, tc.want, tc.result.MustGet("context"))
				}
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *ResultE[string]
			want      string
			wantPanic bool
		}{
			{"returns value on success", ResultOf("hello", nil), "hello", false},
			{"panics on error", ResultOf("", errSample), "", true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.wantPanic {
					assert.Panics(t, func() { tc.result.MustGet("context") })
				} else {
					assert.Equal(t, tc.want, tc.result.MustGet("context"))
				}
			})
		}
	})
}

func TestResultEMustPresent(t *testing.T) {
	tests := []struct {
		name      string
		result    *ResultE[string]
		wantPanic bool
	}{
		{"value present, no error", ResultOf("hello", nil), false},
		{"zero value, no error", ResultOf("", nil), true},
		{"value present, has error", ResultOf("hello", errSample), true},
		{"zero value, has error", ResultOf("", errSample), true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.wantPanic {
				assert.Panics(t, func() { tc.result.MustPresent("msg") })
			} else {
				assert.NotPanics(t, func() { tc.result.MustPresent("msg") })
			}
		})
	}
}

func TestMustPresent(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *Result[int]
			wantPanic bool
		}{
			{"non-zero does not panic", ResultVal(1), false},
			{"zero value panics", ResultVal(0), true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.wantPanic {
					assert.Panics(t, func() { tc.result.MustPresent("msg") })
				} else {
					assert.NotPanics(t, func() { tc.result.MustPresent("msg") })
				}
			})
		}
	})

	t.Run("string", func(t *testing.T) {
		tests := []struct {
			name      string
			result    *Result[string]
			wantPanic bool
		}{
			{"non-empty string does not panic", ResultVal("hello"), false},
			{"empty string panics", ResultVal(""), true},
		}
		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				if tc.wantPanic {
					assert.Panics(t, func() { tc.result.MustPresent("msg") })
				} else {
					assert.NotPanics(t, func() { tc.result.MustPresent("msg") })
				}
			})
		}
	})
}
