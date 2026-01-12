package git

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGitError_Error(t *testing.T) {
	t.Run("with stderr", func(t *testing.T) {
		err := &GitError{
			Operation: "branch",
			Message:   "failed to create",
			Stderr:    "fatal: already exists",
		}
		assert.Equal(t, "git branch: failed to create: fatal: already exists", err.Error())
	})

	t.Run("with underlying error", func(t *testing.T) {
		err := &GitError{
			Operation: "worktree",
			Message:   "failed to add",
			Err:       errors.New("permission denied"),
		}
		assert.Equal(t, "git worktree: failed to add: permission denied", err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &GitError{
			Operation: "repo",
			Message:   "not a repository",
		}
		assert.Equal(t, "git repo: not a repository", err.Error())
	})
}

func TestGitError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &GitError{
		Operation: "branch",
		Message:   "test",
		Err:       underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
	assert.True(t, errors.Is(err, underlying))
}

func TestBranchError_Error(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		err := &BranchError{
			Branch:  "main",
			Message: "cannot delete",
			Err:     errors.New("protected branch"),
		}
		assert.Equal(t, `branch "main": cannot delete: protected branch`, err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &BranchError{
			Branch:  "feature/test",
			Message: "not found",
		}
		assert.Equal(t, `branch "feature/test": not found`, err.Error())
	})
}

func TestBranchError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying")
	err := &BranchError{
		Branch:  "test",
		Message: "error",
		Err:     underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
}

func TestWorktreeError_Error(t *testing.T) {
	t.Run("with underlying error", func(t *testing.T) {
		err := &WorktreeError{
			Path:    "/tmp/worktree",
			Message: "failed to remove",
			Err:     errors.New("directory not empty"),
		}
		assert.Equal(t, `worktree "/tmp/worktree": failed to remove: directory not empty`, err.Error())
	})

	t.Run("message only", func(t *testing.T) {
		err := &WorktreeError{
			Path:    "/path/to/wt",
			Message: "already exists",
		}
		assert.Equal(t, `worktree "/path/to/wt": already exists`, err.Error())
	})
}

func TestWorktreeError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying")
	err := &WorktreeError{
		Path:    "/test",
		Message: "error",
		Err:     underlying,
	}
	assert.Equal(t, underlying, err.Unwrap())
}

func TestIsGitError(t *testing.T) {
	t.Run("is GitError", func(t *testing.T) {
		err := &GitError{Operation: "test", Message: "test"}
		assert.True(t, IsGitError(err))
	})

	t.Run("wrapped GitError", func(t *testing.T) {
		inner := &GitError{Operation: "test", Message: "test"}
		err := errors.New("wrapped: " + inner.Error())
		// Note: plain string wrapping doesn't preserve type
		assert.False(t, IsGitError(err))
	})

	t.Run("not GitError", func(t *testing.T) {
		err := errors.New("regular error")
		assert.False(t, IsGitError(err))
	})

	t.Run("nil error", func(t *testing.T) {
		assert.False(t, IsGitError(nil))
	})
}

func TestIsBranchError(t *testing.T) {
	t.Run("is BranchError", func(t *testing.T) {
		err := &BranchError{Branch: "test", Message: "test"}
		assert.True(t, IsBranchError(err))
	})

	t.Run("not BranchError", func(t *testing.T) {
		err := &GitError{Operation: "test", Message: "test"}
		assert.False(t, IsBranchError(err))
	})
}

func TestIsWorktreeError(t *testing.T) {
	t.Run("is WorktreeError", func(t *testing.T) {
		err := &WorktreeError{Path: "/test", Message: "test"}
		assert.True(t, IsWorktreeError(err))
	})

	t.Run("not WorktreeError", func(t *testing.T) {
		err := &GitError{Operation: "test", Message: "test"}
		assert.False(t, IsWorktreeError(err))
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Run("ErrNotGitRepository", func(t *testing.T) {
		assert.Equal(t, "git repo: not a git repository", ErrNotGitRepository.Error())
		assert.True(t, IsGitError(ErrNotGitRepository))
	})

	t.Run("ErrBranchExists", func(t *testing.T) {
		assert.Equal(t, "git branch: branch already exists", ErrBranchExists.Error())
		assert.True(t, IsGitError(ErrBranchExists))
	})

	t.Run("ErrBranchNotFound", func(t *testing.T) {
		assert.Equal(t, "git branch: branch not found", ErrBranchNotFound.Error())
		assert.True(t, IsGitError(ErrBranchNotFound))
	})

	t.Run("ErrWorktreeExists", func(t *testing.T) {
		assert.Equal(t, "git worktree: worktree already exists", ErrWorktreeExists.Error())
		assert.True(t, IsGitError(ErrWorktreeExists))
	})

	t.Run("ErrWorktreeNotFound", func(t *testing.T) {
		assert.Equal(t, "git worktree: worktree not found", ErrWorktreeNotFound.Error())
		assert.True(t, IsGitError(ErrWorktreeNotFound))
	})

	t.Run("ErrDirtyWorkingTree", func(t *testing.T) {
		assert.Equal(t, "git repo: working tree has uncommitted changes", ErrDirtyWorkingTree.Error())
		assert.True(t, IsGitError(ErrDirtyWorkingTree))
	})
}
