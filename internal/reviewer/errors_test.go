package reviewer

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReviewerError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *ReviewerError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &ReviewerError{
				Phase:   "execute",
				Message: "claude execution failed",
				Err:     errors.New("connection timeout"),
			},
			expected: "reviewer execute: claude execution failed: connection timeout",
		},
		{
			name: "without wrapped error",
			err: &ReviewerError{
				Phase:   "config",
				Message: "no review prompt provided",
				Err:     nil,
			},
			expected: "reviewer config: no review prompt provided",
		},
		{
			name: "prompt phase error",
			err: &ReviewerError{
				Phase:   "prompt",
				Message: "failed to build prompt",
				Err:     errors.New("template error"),
			},
			expected: "reviewer prompt: failed to build prompt: template error",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestReviewerError_Unwrap(t *testing.T) {
	t.Parallel()

	originalErr := errors.New("original error")
	reviewerErr := &ReviewerError{
		Phase:   "execute",
		Message: "failed",
		Err:     originalErr,
	}

	assert.Equal(t, originalErr, reviewerErr.Unwrap())
	assert.True(t, errors.Is(reviewerErr, originalErr))
}

func TestReviewerError_Unwrap_NilError(t *testing.T) {
	t.Parallel()

	reviewerErr := &ReviewerError{
		Phase:   "config",
		Message: "no review prompt",
		Err:     nil,
	}

	assert.Nil(t, reviewerErr.Unwrap())
}

func TestIsReviewerError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "reviewer error",
			err:      &ReviewerError{Phase: "test", Message: "test"},
			expected: true,
		},
		{
			name:     "wrapped reviewer error",
			err:      errors.New("wrapped: " + ErrNoReviewPrompt.Error()),
			expected: false,
		},
		{
			name:     "standard error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "predefined error",
			err:      ErrNoReviewPrompt,
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, IsReviewerError(tt.err))
		})
	}
}

func TestErrNoReviewPrompt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "prompt", ErrNoReviewPrompt.Phase)
	assert.Equal(t, "no review prompt provided", ErrNoReviewPrompt.Message)
	assert.Nil(t, ErrNoReviewPrompt.Err)
	assert.Equal(t, "reviewer prompt: no review prompt provided", ErrNoReviewPrompt.Error())
}
