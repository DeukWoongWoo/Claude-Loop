package architecture

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArchitectureError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *ArchitectureError
		expected string
	}{
		{
			name:     "without wrapped error",
			err:      &ArchitectureError{Phase: "parse", Message: "no components found"},
			expected: "architecture parse: no components found",
		},
		{
			name:     "with wrapped error",
			err:      &ArchitectureError{Phase: "generate", Message: "failed", Err: errors.New("network error")},
			expected: "architecture generate: failed: network error",
		},
		{
			name:     "config phase",
			err:      &ArchitectureError{Phase: "config", Message: "invalid config"},
			expected: "architecture config: invalid config",
		},
		{
			name:     "validate phase",
			err:      &ArchitectureError{Phase: "validate", Message: "missing fields"},
			expected: "architecture validate: missing fields",
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

func TestArchitectureError_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("with wrapped error", func(t *testing.T) {
		t.Parallel()
		inner := errors.New("inner error")
		err := &ArchitectureError{Phase: "generate", Message: "failed", Err: inner}
		assert.Equal(t, inner, err.Unwrap())
		assert.True(t, errors.Is(err, inner))
	})

	t.Run("without wrapped error", func(t *testing.T) {
		t.Parallel()
		err := &ArchitectureError{Phase: "parse", Message: "no components"}
		assert.Nil(t, err.Unwrap())
	})
}

func TestIsArchitectureError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is ArchitectureError",
			err:      &ArchitectureError{Phase: "parse", Message: "test"},
			expected: true,
		},
		{
			name:     "wrapped ArchitectureError",
			err:      errors.New("wrapped: " + (&ArchitectureError{Phase: "parse", Message: "test"}).Error()),
			expected: false,
		},
		{
			name:     "not ArchitectureError",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, IsArchitectureError(tt.err))
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name:     "components field",
			err:      &ValidationError{Field: "components", Message: "at least one required"},
			expected: "validation error on components: at least one required",
		},
		{
			name:     "file_structure field",
			err:      &ValidationError{Field: "file_structure", Message: "missing"},
			expected: "validation error on file_structure: missing",
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

func TestIsValidationError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "is ValidationError",
			err:      &ValidationError{Field: "components", Message: "test"},
			expected: true,
		},
		{
			name:     "not ValidationError",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "ArchitectureError is not ValidationError",
			err:      &ArchitectureError{Phase: "parse", Message: "test"},
			expected: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, IsValidationError(tt.err))
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrNilClient", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrNilClient)
		assert.Equal(t, "config", ErrNilClient.Phase)
		assert.Contains(t, ErrNilClient.Error(), "claude client is nil")
	})

	t.Run("ErrNilPRD", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrNilPRD)
		assert.Equal(t, "generate", ErrNilPRD.Phase)
		assert.Contains(t, ErrNilPRD.Error(), "PRD is nil")
	})

	t.Run("ErrParseNoComponents", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrParseNoComponents)
		assert.Equal(t, "parse", ErrParseNoComponents.Phase)
		assert.Contains(t, ErrParseNoComponents.Error(), "no components")
	})

	t.Run("ErrParseNoFileStructure", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrParseNoFileStructure)
		assert.Equal(t, "parse", ErrParseNoFileStructure.Phase)
		assert.Contains(t, ErrParseNoFileStructure.Error(), "no file structure")
	})
}
