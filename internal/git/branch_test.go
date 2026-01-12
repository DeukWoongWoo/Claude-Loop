package git

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBranchManager_GenerateBranchName(t *testing.T) {
	bm := NewBranchManager(nil)

	t.Run("generates name with default prefix", func(t *testing.T) {
		name, err := bm.GenerateBranchName("")
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(name, "claude-loop/"))
		// Check date format: claude-loop/YYYY-MM-DD-XXXXXX
		dateStr := time.Now().Format("2006-01-02")
		assert.Contains(t, name, dateStr)
		// Check hash part (6 hex chars)
		parts := strings.Split(name, "-")
		assert.Len(t, parts[len(parts)-1], 6)
	})

	t.Run("generates name with custom prefix", func(t *testing.T) {
		name, err := bm.GenerateBranchName("feature/")
		require.NoError(t, err)

		assert.True(t, strings.HasPrefix(name, "feature/"))
	})

	t.Run("generates unique names", func(t *testing.T) {
		names := make(map[string]bool)
		for i := 0; i < 100; i++ {
			name, err := bm.GenerateBranchName("")
			require.NoError(t, err)
			assert.False(t, names[name], "duplicate name generated: %s", name)
			names[name] = true
		}
	})
}

func TestBranchManager_CreateBranch(t *testing.T) {
	t.Run("creates branch successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1},    // BranchExists (show-ref) - not found (exit 1)
				{Stdout: "main"}, // GetCurrentBranch
				{Stdout: ""},     // git branch
			},
		}
		bm := NewBranchManager(mock)

		err := bm.CreateBranch(context.Background(), "new-branch", nil)
		require.NoError(t, err)
	})

	t.Run("error when branch exists", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 0}, // BranchExists - branch found (exit 0)
			},
		}
		bm := NewBranchManager(mock)

		err := bm.CreateBranch(context.Background(), "existing-branch", nil)
		assert.Error(t, err)
		assert.True(t, IsBranchError(err))
		assert.Contains(t, err.Error(), "already exists")
	})

	t.Run("uses custom base branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1}, // BranchExists - not found
				{Stdout: ""},  // git branch (base branch specified, no GetCurrentBranch call)
			},
		}
		bm := NewBranchManager(mock)

		opts := &BranchOptions{
			BaseBranch: "develop",
		}
		err := bm.CreateBranch(context.Background(), "new-branch", opts)
		require.NoError(t, err)
	})
}

func TestBranchManager_CreateIterationBranch(t *testing.T) {
	t.Run("creates branch with auto-generated name", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1},    // BranchExists - not found
				{Stdout: "main"}, // GetCurrentBranch
				{Stdout: ""},     // git branch
			},
		}
		bm := NewBranchManager(mock)

		name, err := bm.CreateIterationBranch(context.Background(), nil)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(name, "claude-loop/"))
	})

	t.Run("uses custom prefix", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1},    // BranchExists - not found
				{Stdout: "main"}, // GetCurrentBranch
				{Stdout: ""},     // git branch
			},
		}
		bm := NewBranchManager(mock)

		opts := &BranchOptions{Prefix: "auto/"}
		name, err := bm.CreateIterationBranch(context.Background(), opts)
		require.NoError(t, err)
		assert.True(t, strings.HasPrefix(name, "auto/"))
	})
}

func TestBranchManager_BranchExists(t *testing.T) {
	t.Run("branch exists", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 0}, // show-ref exits 0 when found
			},
		}
		bm := NewBranchManager(mock)

		exists, err := bm.BranchExists(context.Background(), "main")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("branch does not exist", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1}, // show-ref exits 1 when not found
			},
		}
		bm := NewBranchManager(mock)

		exists, err := bm.BranchExists(context.Background(), "nonexistent")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestBranchManager_DeleteBranch(t *testing.T) {
	t.Run("deletes branch successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 0}, // BranchExists
				{Stdout: ""},  // git branch -d
			},
		}
		bm := NewBranchManager(mock)

		err := bm.DeleteBranch(context.Background(), "old-branch", false)
		require.NoError(t, err)
	})

	t.Run("force deletes branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 0}, // BranchExists
				{Stdout: ""},  // git branch -D
			},
		}
		bm := NewBranchManager(mock)

		err := bm.DeleteBranch(context.Background(), "old-branch", true)
		require.NoError(t, err)
	})

	t.Run("error when branch not found", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1}, // BranchExists - not found
			},
		}
		bm := NewBranchManager(mock)

		err := bm.DeleteBranch(context.Background(), "nonexistent", false)
		assert.Error(t, err)
		assert.True(t, IsBranchError(err))
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestBranchManager_Checkout(t *testing.T) {
	t.Run("checkout succeeds", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		bm := NewBranchManager(mock)

		err := bm.Checkout(context.Background(), "feature/test")
		require.NoError(t, err)
	})

	t.Run("checkout fails", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "error: pathspec 'nonexistent' did not match any file"},
			},
		}
		bm := NewBranchManager(mock)

		err := bm.Checkout(context.Background(), "nonexistent")
		assert.Error(t, err)
		assert.True(t, IsBranchError(err))
	})
}

func TestBranchManager_ListBranches(t *testing.T) {
	t.Run("lists all branches", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "main\nfeature/a\nfeature/b\ndevelop"},
			},
		}
		bm := NewBranchManager(mock)

		branches, err := bm.ListBranches(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"main", "feature/a", "feature/b", "develop"}, branches)
	})

	t.Run("handles single branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "main"},
			},
		}
		bm := NewBranchManager(mock)

		branches, err := bm.ListBranches(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []string{"main"}, branches)
	})

	t.Run("handles empty output", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		bm := NewBranchManager(mock)

		branches, err := bm.ListBranches(context.Background())
		require.NoError(t, err)
		assert.Empty(t, branches)
	})
}

func TestNewBranchManager(t *testing.T) {
	t.Run("uses default executor when nil", func(t *testing.T) {
		bm := NewBranchManager(nil)
		assert.NotNil(t, bm)
		assert.NotNil(t, bm.executor)
		assert.NotNil(t, bm.repo)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		bm := NewBranchManager(mock)
		assert.Equal(t, mock, bm.executor)
	})
}

func TestDefaultBranchOptions(t *testing.T) {
	opts := DefaultBranchOptions()
	assert.Equal(t, "claude-loop/", opts.Prefix)
	assert.Empty(t, opts.BaseBranch)
}
