package loop

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoopError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *LoopError
		expected string
	}{
		{
			name: "without wrapped error",
			err: &LoopError{
				Field:   "max_runs",
				Message: "maximum runs reached",
			},
			expected: "max_runs: maximum runs reached",
		},
		{
			name: "with wrapped error",
			err: &LoopError{
				Field:   "config",
				Message: "failed to load",
				Err:     errors.New("file not found"),
			},
			expected: "config: failed to load: file not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestLoopError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	loopErr := &LoopError{
		Field:   "test",
		Message: "test message",
		Err:     innerErr,
	}

	assert.Equal(t, innerErr, loopErr.Unwrap())
	assert.True(t, errors.Is(loopErr, innerErr))
}

func TestIterationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *IterationError
		expected string
	}{
		{
			name: "without wrapped error",
			err: &IterationError{
				Iteration: 5,
				Message:   "execution failed",
			},
			expected: "iteration 5: execution failed",
		},
		{
			name: "with wrapped error",
			err: &IterationError{
				Iteration: 3,
				Message:   "claude error",
				Err:       errors.New("timeout"),
			},
			expected: "iteration 3: claude error: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestIterationError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	iterErr := &IterationError{
		Iteration: 1,
		Message:   "test",
		Err:       innerErr,
	}

	assert.Equal(t, innerErr, iterErr.Unwrap())
	assert.True(t, errors.Is(iterErr, innerErr))
}

func TestIsLoopError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is LoopError",
			err:  &LoopError{Field: "test", Message: "test"},
			want: true,
		},
		{
			name: "is not LoopError",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "wrapped LoopError",
			err:  &LoopError{Field: "outer", Message: "outer", Err: &LoopError{Field: "inner", Message: "inner"}},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsLoopError(tt.err))
		})
	}
}

func TestIsIterationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is IterationError",
			err:  &IterationError{Iteration: 1, Message: "test"},
			want: true,
		},
		{
			name: "is not IterationError",
			err:  errors.New("regular error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "LoopError is not IterationError",
			err:  &LoopError{Field: "test", Message: "test"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsIterationError(tt.err))
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      *LoopError
		field    string
		contains string
	}{
		{
			name:     "ErrMaxRunsReached",
			err:      ErrMaxRunsReached,
			field:    "max_runs",
			contains: "maximum runs",
		},
		{
			name:     "ErrMaxCostReached",
			err:      ErrMaxCostReached,
			field:    "max_cost",
			contains: "maximum cost",
		},
		{
			name:     "ErrMaxDurationReached",
			err:      ErrMaxDurationReached,
			field:    "max_duration",
			contains: "maximum duration",
		},
		{
			name:     "ErrConsecutiveErrors",
			err:      ErrConsecutiveErrors,
			field:    "errors",
			contains: "consecutive errors",
		},
		{
			name:     "ErrCompletionSignal",
			err:      ErrCompletionSignal,
			field:    "completion",
			contains: "completion signal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			assert.Equal(t, tt.field, tt.err.Field)
			assert.Contains(t, tt.err.Message, tt.contains)
			assert.True(t, IsLoopError(tt.err))
		})
	}
}

func TestErrorAs(t *testing.T) {
	loopErr := &LoopError{Field: "test", Message: "test message"}

	var target *LoopError
	assert.True(t, errors.As(loopErr, &target))
	assert.Equal(t, "test", target.Field)

	iterErr := &IterationError{Iteration: 5, Message: "test"}
	var iterTarget *IterationError
	assert.True(t, errors.As(iterErr, &iterTarget))
	assert.Equal(t, 5, iterTarget.Iteration)
}
