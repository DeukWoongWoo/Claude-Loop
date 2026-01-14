package council

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCouncilError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CouncilError
		expected string
	}{
		{
			name: "with wrapped error",
			err: &CouncilError{
				Phase:   "resolve",
				Message: "something failed",
				Err:     errors.New("underlying error"),
			},
			expected: "council resolve: something failed: underlying error",
		},
		{
			name: "without wrapped error",
			err: &CouncilError{
				Phase:   "detect",
				Message: "no conflict found",
			},
			expected: "council detect: no conflict found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestCouncilError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &CouncilError{
		Phase:   "log",
		Message: "failed",
		Err:     underlying,
	}

	assert.Equal(t, underlying, err.Unwrap())
}

func TestIsCouncilError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "council error",
			err:      &CouncilError{Phase: "test", Message: "test"},
			expected: true,
		},
		{
			name:     "wrapped council error",
			err:      errors.Join(errors.New("wrapper"), &CouncilError{Phase: "test", Message: "test"}),
			expected: true,
		},
		{
			name:     "regular error",
			err:      errors.New("regular error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsCouncilError(tt.err))
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	assert.NotNil(t, ErrNoPrinciples)
	assert.Equal(t, "resolve", ErrNoPrinciples.Phase)
	assert.Contains(t, ErrNoPrinciples.Error(), "no principles configured")
}
