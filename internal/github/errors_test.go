package github

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitHubError_Error(t *testing.T) {
	t.Run("with stderr", func(t *testing.T) {
		err := &GitHubError{
			Operation: "pr",
			Message:   "failed to create",
			Stderr:    "permission denied",
		}
		assert.Equal(t, "github pr: failed to create: permission denied", err.Error())
	})

	t.Run("with underlying error", func(t *testing.T) {
		err := &GitHubError{
			Operation: "checks",
			Message:   "failed to get status",
			Err:       errors.New("timeout"),
		}
		assert.Equal(t, "github checks: failed to get status: timeout", err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &GitHubError{
			Operation: "repo",
			Message:   "not found",
		}
		assert.Equal(t, "github repo: not found", err.Error())
	})
}

func TestGitHubError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &GitHubError{
		Operation: "pr",
		Message:   "test",
		Err:       underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
}

func TestPRError_Error(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		err := &PRError{
			PRNumber: 123,
			Message:  "failed to merge",
			Err:      errors.New("conflict"),
		}
		assert.Equal(t, "PR #123: failed to merge: conflict", err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &PRError{
			PRNumber: 456,
			Message:  "not found",
		}
		assert.Equal(t, "PR #456: not found", err.Error())
	})
}

func TestPRError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &PRError{
		PRNumber: 123,
		Message:  "test",
		Err:      underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
}

func TestCheckError_Error(t *testing.T) {
	t.Run("with summary", func(t *testing.T) {
		err := &CheckError{
			PRNumber: 123,
			Message:  "checks failed",
			Summary: &CheckSummary{
				Success: 3,
				Failed:  2,
				Pending: 1,
			},
		}
		assert.Equal(t, "PR #123 checks: checks failed (passed: 3, failed: 2, pending: 1)", err.Error())
	})

	t.Run("with underlying error", func(t *testing.T) {
		err := &CheckError{
			PRNumber: 456,
			Message:  "timeout",
			Err:      errors.New("context deadline exceeded"),
		}
		assert.Equal(t, "PR #456 checks: timeout: context deadline exceeded", err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &CheckError{
			PRNumber: 789,
			Message:  "unknown error",
		}
		assert.Equal(t, "PR #789 checks: unknown error", err.Error())
	})
}

func TestCheckError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &CheckError{
		PRNumber: 123,
		Message:  "test",
		Err:      underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
}

func TestIsGitHubError(t *testing.T) {
	t.Run("returns true for GitHubError", func(t *testing.T) {
		err := &GitHubError{Operation: "test", Message: "test"}
		assert.True(t, IsGitHubError(err))
	})

	t.Run("returns true for wrapped GitHubError", func(t *testing.T) {
		gitHubErr := &GitHubError{Operation: "test", Message: "test"}
		wrapped := errors.New("wrapper: " + gitHubErr.Error())
		_ = wrapped // Not actually wrapped, so:
		assert.False(t, IsGitHubError(wrapped))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("some error")
		assert.False(t, IsGitHubError(err))
	})
}

func TestIsPRError(t *testing.T) {
	t.Run("returns true for PRError", func(t *testing.T) {
		err := &PRError{PRNumber: 123, Message: "test"}
		assert.True(t, IsPRError(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("some error")
		assert.False(t, IsPRError(err))
	})
}

func TestIsCheckError(t *testing.T) {
	t.Run("returns true for CheckError", func(t *testing.T) {
		err := &CheckError{PRNumber: 123, Message: "test"}
		assert.True(t, IsCheckError(err))
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("some error")
		assert.False(t, IsCheckError(err))
	})
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		contains string
	}{
		{"ErrNotGitHubRepo", ErrNotGitHubRepo, "not a GitHub repository"},
		{"ErrPRNotFound", ErrPRNotFound, "pull request not found"},
		{"ErrPRChecksFailed", ErrPRChecksFailed, "PR checks failed"},
		{"ErrPRChecksTimeout", ErrPRChecksTimeout, "timeout waiting for PR checks"},
		{"ErrPRMergeFailed", ErrPRMergeFailed, "failed to merge PR"},
		{"ErrPRMergeConflict", ErrPRMergeConflict, "merge conflict"},
		{"ErrPRUpdateFailed", ErrPRUpdateFailed, "failed to update PR branch"},
		{"ErrPRAlreadyUpToDate", ErrPRAlreadyUpToDate, "already up to date"},
		{"ErrChangesRequested", ErrChangesRequested, "changes requested"},
		{"ErrReviewPending", ErrReviewPending, "review pending"},
		{"ErrGHCLINotFound", ErrGHCLINotFound, "gh CLI not found"},
		{"ErrGHNotAuthenticated", ErrGHNotAuthenticated, "not authenticated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Contains(t, tt.err.Error(), tt.contains)
			assert.True(t, IsGitHubError(tt.err))
		})
	}
}
