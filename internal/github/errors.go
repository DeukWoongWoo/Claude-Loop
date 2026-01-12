package github

import (
	"errors"
	"fmt"
)

// GitHubError represents a GitHub operation error.
type GitHubError struct {
	Operation string // Failed operation (e.g., "pr", "checks", "repo")
	Message   string // Error message
	Stderr    string // stderr output from gh command
	Err       error  // Underlying error
}

func (e *GitHubError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("github %s: %s: %s", e.Operation, e.Message, e.Stderr)
	}
	if e.Err != nil {
		return fmt.Sprintf("github %s: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("github %s: %s", e.Operation, e.Message)
}

func (e *GitHubError) Unwrap() error {
	return e.Err
}

// PRError represents a PR-specific error.
type PRError struct {
	PRNumber int
	Message  string
	Err      error
}

func (e *PRError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("PR #%d: %s: %v", e.PRNumber, e.Message, e.Err)
	}
	return fmt.Sprintf("PR #%d: %s", e.PRNumber, e.Message)
}

func (e *PRError) Unwrap() error {
	return e.Err
}

// CheckError represents a CI check error.
type CheckError struct {
	PRNumber int
	Message  string
	Summary  *CheckSummary
	Err      error
}

func (e *CheckError) Error() string {
	if e.Summary != nil {
		return fmt.Sprintf("PR #%d checks: %s (passed: %d, failed: %d, pending: %d)",
			e.PRNumber, e.Message, e.Summary.Success, e.Summary.Failed, e.Summary.Pending)
	}
	if e.Err != nil {
		return fmt.Sprintf("PR #%d checks: %s: %v", e.PRNumber, e.Message, e.Err)
	}
	return fmt.Sprintf("PR #%d checks: %s", e.PRNumber, e.Message)
}

func (e *CheckError) Unwrap() error {
	return e.Err
}

// Predefined errors for common conditions.
var (
	ErrNotGitHubRepo      = &GitHubError{Operation: "repo", Message: "not a GitHub repository"}
	ErrPRNotFound         = &GitHubError{Operation: "pr", Message: "pull request not found"}
	ErrPRChecksFailed     = &GitHubError{Operation: "checks", Message: "PR checks failed"}
	ErrPRChecksTimeout    = &GitHubError{Operation: "checks", Message: "timeout waiting for PR checks"}
	ErrPRMergeFailed      = &GitHubError{Operation: "merge", Message: "failed to merge PR"}
	ErrPRMergeConflict    = &GitHubError{Operation: "merge", Message: "merge conflict"}
	ErrPRUpdateFailed     = &GitHubError{Operation: "update", Message: "failed to update PR branch"}
	ErrPRAlreadyUpToDate  = &GitHubError{Operation: "update", Message: "branch already up to date"}
	ErrChangesRequested   = &GitHubError{Operation: "review", Message: "changes requested in review"}
	ErrReviewPending      = &GitHubError{Operation: "review", Message: "review pending"}
	ErrGHCLINotFound      = &GitHubError{Operation: "cli", Message: "gh CLI not found"}
	ErrGHNotAuthenticated = &GitHubError{Operation: "auth", Message: "gh CLI not authenticated"}
)

// IsGitHubError checks if an error is a GitHubError.
func IsGitHubError(err error) bool {
	var ge *GitHubError
	return errors.As(err, &ge)
}

// IsPRError checks if an error is a PRError.
func IsPRError(err error) bool {
	var pe *PRError
	return errors.As(err, &pe)
}

// IsCheckError checks if an error is a CheckError.
func IsCheckError(err error) bool {
	var ce *CheckError
	return errors.As(err, &ce)
}
