package verifier

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifierError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *VerifierError
		expected string
	}{
		{
			name: "without wrapped error",
			err: &VerifierError{
				Phase:   "config",
				Message: "task is nil",
			},
			expected: "verifier config: task is nil",
		},
		{
			name: "with wrapped error",
			err: &VerifierError{
				Phase:   "execute",
				Message: "verification failed",
				Err:     errors.New("timeout"),
			},
			expected: "verifier execute: verification failed: timeout",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestVerifierError_Unwrap(t *testing.T) {
	wrapped := errors.New("underlying error")
	err := &VerifierError{
		Phase:   "check",
		Message: "check failed",
		Err:     wrapped,
	}

	assert.Equal(t, wrapped, err.Unwrap())

	// Test with nil wrapped error
	errNoWrap := &VerifierError{
		Phase:   "check",
		Message: "check failed",
	}
	assert.Nil(t, errNoWrap.Unwrap())
}

func TestIsVerifierError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "verifier error",
			err:  &VerifierError{Phase: "test", Message: "msg"},
			want: true,
		},
		{
			name: "wrapped verifier error",
			err:  errors.Join(&VerifierError{Phase: "test", Message: "msg"}, errors.New("other")),
			want: true,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "predefined error",
			err:  ErrNilTask,
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsVerifierError(tt.err))
		})
	}
}

func TestCheckError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CheckError
		expected string
	}{
		{
			name: "without wrapped error",
			err: &CheckError{
				CheckerType: "file_exists",
				Criterion:   "file test.go exists",
				Message:     "file not found",
			},
			expected: "check file_exists failed for 'file test.go exists': file not found",
		},
		{
			name: "with wrapped error",
			err: &CheckError{
				CheckerType: "build",
				Criterion:   "build passes",
				Message:     "compilation error",
				Err:         errors.New("syntax error"),
			},
			expected: "check build failed for 'build passes': compilation error: syntax error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestCheckError_Unwrap(t *testing.T) {
	wrapped := errors.New("underlying error")
	err := &CheckError{
		CheckerType: "test",
		Criterion:   "test criterion",
		Message:     "failed",
		Err:         wrapped,
	}

	assert.Equal(t, wrapped, err.Unwrap())
}

func TestIsCheckError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "check error",
			err:  &CheckError{CheckerType: "test", Criterion: "c", Message: "m"},
			want: true,
		},
		{
			name: "verifier error",
			err:  &VerifierError{Phase: "test", Message: "msg"},
			want: false,
		},
		{
			name: "standard error",
			err:  errors.New("standard error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsCheckError(tt.err))
		})
	}
}

func TestPredefinedErrors(t *testing.T) {
	// Test ErrNilTask
	assert.True(t, IsVerifierError(ErrNilTask))
	assert.Equal(t, "config", ErrNilTask.Phase)
	assert.Contains(t, ErrNilTask.Message, "nil")

	// Test ErrNoCriteria
	assert.True(t, IsVerifierError(ErrNoCriteria))
	assert.Equal(t, "config", ErrNoCriteria.Phase)
	assert.Contains(t, ErrNoCriteria.Message, "criteria")

	// Test ErrNilClient
	assert.True(t, IsVerifierError(ErrNilClient))
	assert.Equal(t, "config", ErrNilClient.Phase)
	assert.Contains(t, ErrNilClient.Message, "client")

	// Test ErrTimeout
	assert.True(t, IsVerifierError(ErrTimeout))
	assert.Equal(t, "execute", ErrTimeout.Phase)
	assert.Contains(t, ErrTimeout.Message, "timed out")

	// Test ErrNoCheckerFound
	assert.True(t, IsVerifierError(ErrNoCheckerFound))
	assert.Equal(t, "parse", ErrNoCheckerFound.Phase)
	assert.Contains(t, ErrNoCheckerFound.Message, "checker")
}

func TestErrorsAs(t *testing.T) {
	// Test errors.As with VerifierError
	err := &VerifierError{Phase: "test", Message: "msg"}
	var ve *VerifierError
	assert.True(t, errors.As(err, &ve))
	assert.Equal(t, "test", ve.Phase)

	// Test errors.As with CheckError
	checkErr := &CheckError{CheckerType: "test", Criterion: "c", Message: "m"}
	var ce *CheckError
	assert.True(t, errors.As(checkErr, &ce))
	assert.Equal(t, "test", ce.CheckerType)
}
