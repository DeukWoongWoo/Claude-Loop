package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		flags   *Flags
		wantErr string // empty string means no error expected
	}{
		{
			name:    "missing prompt",
			flags:   &Flags{MaxRuns: 5},
			wantErr: "prompt is required",
		},
		{
			name:    "missing all limits",
			flags:   &Flags{Prompt: "test"},
			wantErr: "at least one limit required",
		},
		{
			name: "only max-runs provided",
			flags: &Flags{
				Prompt:  "test",
				MaxRuns: 5,
			},
			wantErr: "",
		},
		{
			name: "only max-cost provided",
			flags: &Flags{
				Prompt:  "test",
				MaxCost: 10.0,
			},
			wantErr: "",
		},
		{
			name: "only max-duration provided",
			flags: &Flags{
				Prompt:      "test",
				MaxDuration: 2 * time.Hour,
			},
			wantErr: "",
		},
		{
			name: "multiple limits provided",
			flags: &Flags{
				Prompt:      "test",
				MaxRuns:     5,
				MaxCost:     10.0,
				MaxDuration: 1 * time.Hour,
			},
			wantErr: "",
		},
		{
			name: "invalid merge strategy",
			flags: &Flags{
				Prompt:        "test",
				MaxRuns:       5,
				MergeStrategy: "invalid",
			},
			wantErr: "merge-strategy must be squash, merge, or rebase",
		},
		{
			name: "valid merge strategy squash",
			flags: &Flags{
				Prompt:        "test",
				MaxRuns:       5,
				MergeStrategy: "squash",
			},
			wantErr: "",
		},
		{
			name: "valid merge strategy merge",
			flags: &Flags{
				Prompt:        "test",
				MaxRuns:       5,
				MergeStrategy: "merge",
			},
			wantErr: "",
		},
		{
			name: "valid merge strategy rebase",
			flags: &Flags{
				Prompt:        "test",
				MaxRuns:       5,
				MergeStrategy: "rebase",
			},
			wantErr: "",
		},
		{
			name: "list-worktrees bypasses validation",
			flags: &Flags{
				ListWorktrees: true,
				// No prompt or limits - should still pass
			},
			wantErr: "",
		},
		{
			name: "empty flags",
			flags: &Flags{},
			wantErr: "prompt is required",
		},
		{
			name: "zero max-runs without other limit",
			flags: &Flags{
				Prompt:  "test",
				MaxRuns: 0, // explicitly zero, no other limits
			},
			wantErr: "at least one limit required",
		},
		{
			name: "zero max-runs with max-cost",
			flags: &Flags{
				Prompt:  "test",
				MaxRuns: 0,
				MaxCost: 5.0,
			},
			wantErr: "",
		},
		{
			name: "zero max-runs with max-duration",
			flags: &Flags{
				Prompt:      "test",
				MaxRuns:     0,
				MaxDuration: 30 * time.Minute,
			},
			wantErr: "",
		},
		{
			name: "negative max-runs",
			flags: &Flags{
				Prompt:  "test",
				MaxRuns: -5,
			},
			wantErr: "max-runs cannot be negative",
		},
		{
			name: "negative max-cost",
			flags: &Flags{
				Prompt:  "test",
				MaxCost: -1.0,
			},
			wantErr: "max-cost cannot be negative",
		},
		{
			name: "negative max-duration",
			flags: &Flags{
				Prompt:      "test",
				MaxRuns:     5,
				MaxDuration: -1 * time.Hour,
			},
			wantErr: "max-duration cannot be negative",
		},
		{
			name: "negative ci-retry-max",
			flags: &Flags{
				Prompt:     "test",
				MaxRuns:    5,
				CIRetryMax: -1,
			},
			wantErr: "ci-retry-max cannot be negative",
		},
		{
			name: "negative completion-threshold",
			flags: &Flags{
				Prompt:              "test",
				MaxRuns:             5,
				CompletionThreshold: -1,
			},
			wantErr: "completion-threshold cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.flags.Validate()

			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	tests := []struct {
		name       string
		flags      *Flags
		wantErrors int
	}{
		{
			name:       "no errors",
			flags:      &Flags{Prompt: "test", MaxRuns: 5, MergeStrategy: "squash"},
			wantErrors: 0,
		},
		{
			name:       "single error - missing prompt",
			flags:      &Flags{MaxRuns: 5, MergeStrategy: "squash"},
			wantErrors: 1,
		},
		{
			name:       "single error - missing limit",
			flags:      &Flags{Prompt: "test", MergeStrategy: "squash"},
			wantErrors: 1,
		},
		{
			name:       "single error - invalid merge strategy",
			flags:      &Flags{Prompt: "test", MaxRuns: 5, MergeStrategy: "invalid"},
			wantErrors: 1,
		},
		{
			name:       "multiple errors - missing prompt and limit",
			flags:      &Flags{MergeStrategy: "squash"},
			wantErrors: 2,
		},
		{
			name:       "all errors - missing prompt, limit, invalid merge",
			flags:      &Flags{MergeStrategy: "invalid"},
			wantErrors: 3,
		},
		{
			name:       "list-worktrees bypasses validation",
			flags:      &Flags{ListWorktrees: true},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := tt.flags.ValidateAll()
			assert.Len(t, errs, tt.wantErrors)
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test-field",
		Message: "test error message",
	}

	assert.Equal(t, "test error message", err.Error())
	assert.True(t, IsValidationError(err))
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is validation error",
			err:  &ValidationError{Field: "test", Message: "test"},
			want: true,
		},
		{
			name: "is not validation error",
			err:  assert.AnError,
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
			result := IsValidationError(tt.err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestValidationErrorFields(t *testing.T) {
	// Test that ValidationError has correct field values
	f := &Flags{MaxRuns: 5} // Missing prompt
	err := f.Validate()

	require.Error(t, err)

	var ve *ValidationError
	require.ErrorAs(t, err, &ve)

	assert.Equal(t, "prompt", ve.Field)
	assert.Contains(t, ve.Message, "prompt is required")
}
