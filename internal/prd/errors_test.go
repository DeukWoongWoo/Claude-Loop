package prd

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPRDError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *PRDError
		expected string
	}{
		{
			name:     "with wrapped error",
			err:      &PRDError{Phase: "generate", Message: "test message", Err: errors.New("wrapped")},
			expected: "prd generate: test message: wrapped",
		},
		{
			name:     "without wrapped error",
			err:      &PRDError{Phase: "parse", Message: "parse failed"},
			expected: "prd parse: parse failed",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestPRDError_Unwrap(t *testing.T) {
	t.Parallel()

	wrapped := errors.New("inner error")
	err := &PRDError{Phase: "test", Message: "msg", Err: wrapped}

	assert.Equal(t, wrapped, err.Unwrap())
	assert.True(t, errors.Is(err, wrapped))
}

func TestIsPRDError(t *testing.T) {
	t.Parallel()

	t.Run("is PRDError", func(t *testing.T) {
		t.Parallel()
		err := &PRDError{Phase: "test", Message: "msg"}
		assert.True(t, IsPRDError(err))
	})

	t.Run("not PRDError", func(t *testing.T) {
		t.Parallel()
		err := errors.New("regular error")
		assert.False(t, IsPRDError(err))
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsPRDError(nil))
	})
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	err := &ValidationError{Field: "goals", Message: "empty"}
	assert.Equal(t, "validation error on goals: empty", err.Error())
}

func TestIsValidationError(t *testing.T) {
	t.Parallel()

	t.Run("is ValidationError", func(t *testing.T) {
		t.Parallel()
		err := &ValidationError{Field: "goals", Message: "empty"}
		assert.True(t, IsValidationError(err))
	})

	t.Run("not ValidationError", func(t *testing.T) {
		t.Parallel()
		err := errors.New("other")
		assert.False(t, IsValidationError(err))
	})

	t.Run("nil error", func(t *testing.T) {
		t.Parallel()
		assert.False(t, IsValidationError(nil))
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		err   *PRDError
		phase string
	}{
		{"ErrNilClient", ErrNilClient, "config"},
		{"ErrEmptyPrompt", ErrEmptyPrompt, "generate"},
		{"ErrParseNoGoals", ErrParseNoGoals, "parse"},
		{"ErrParseNoRequirements", ErrParseNoRequirements, "parse"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.phase, tt.err.Phase)
			assert.True(t, IsPRDError(tt.err))
		})
	}
}
