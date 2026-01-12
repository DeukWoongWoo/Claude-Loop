package git

import (
	"errors"
	"fmt"
)

// GitError represents a Git operation error.
type GitError struct {
	Operation string // Failed operation (e.g., "branch", "worktree", "repo")
	Message   string // Error message
	Stderr    string // stderr output from git command
	Err       error  // Underlying error
}

func (e *GitError) Error() string {
	if e.Stderr != "" {
		return fmt.Sprintf("git %s: %s: %s", e.Operation, e.Message, e.Stderr)
	}
	if e.Err != nil {
		return fmt.Sprintf("git %s: %s: %v", e.Operation, e.Message, e.Err)
	}
	return fmt.Sprintf("git %s: %s", e.Operation, e.Message)
}

func (e *GitError) Unwrap() error {
	return e.Err
}

// BranchError represents a branch-specific error.
type BranchError struct {
	Branch  string
	Message string
	Err     error
}

func (e *BranchError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("branch %q: %s: %v", e.Branch, e.Message, e.Err)
	}
	return fmt.Sprintf("branch %q: %s", e.Branch, e.Message)
}

func (e *BranchError) Unwrap() error {
	return e.Err
}

// WorktreeError represents a worktree-specific error.
type WorktreeError struct {
	Path    string
	Message string
	Err     error
}

func (e *WorktreeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("worktree %q: %s: %v", e.Path, e.Message, e.Err)
	}
	return fmt.Sprintf("worktree %q: %s", e.Path, e.Message)
}

func (e *WorktreeError) Unwrap() error {
	return e.Err
}

// Predefined errors for common conditions.
var (
	ErrNotGitRepository = &GitError{Operation: "repo", Message: "not a git repository"}
	ErrBranchExists     = &GitError{Operation: "branch", Message: "branch already exists"}
	ErrBranchNotFound   = &GitError{Operation: "branch", Message: "branch not found"}
	ErrWorktreeExists   = &GitError{Operation: "worktree", Message: "worktree already exists"}
	ErrWorktreeNotFound = &GitError{Operation: "worktree", Message: "worktree not found"}
	ErrDirtyWorkingTree = &GitError{Operation: "repo", Message: "working tree has uncommitted changes"}
)

// IsGitError checks if an error is a GitError.
func IsGitError(err error) bool {
	var ge *GitError
	return errors.As(err, &ge)
}

// IsBranchError checks if an error is a BranchError.
func IsBranchError(err error) bool {
	var be *BranchError
	return errors.As(err, &be)
}

// IsWorktreeError checks if an error is a WorktreeError.
func IsWorktreeError(err error) bool {
	var we *WorktreeError
	return errors.As(err, &we)
}
