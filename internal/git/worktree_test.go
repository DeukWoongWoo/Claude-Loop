package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorktreeManager_List(t *testing.T) {
	t.Run("lists worktrees successfully", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project
HEAD abc1234567890def
branch refs/heads/main

worktree /home/user/worktrees/feature-1
HEAD def4567890123abc
branch refs/heads/feature-1

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: porcelainOutput},
			},
		}
		wm := NewWorktreeManager(mock)

		worktrees, err := wm.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, worktrees, 2)

		// First worktree (main)
		assert.Equal(t, "/home/user/project", worktrees[0].Path)
		assert.Equal(t, "main", worktrees[0].Branch)
		assert.Equal(t, "abc1234", worktrees[0].CommitHash)
		assert.True(t, worktrees[0].IsMain)

		// Second worktree
		assert.Equal(t, "/home/user/worktrees/feature-1", worktrees[1].Path)
		assert.Equal(t, "feature-1", worktrees[1].Branch)
		assert.Equal(t, "def4567", worktrees[1].CommitHash)
		assert.False(t, worktrees[1].IsMain)
	})

	t.Run("handles bare worktree", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project.git
bare

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: porcelainOutput},
			},
		}
		wm := NewWorktreeManager(mock)

		worktrees, err := wm.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, worktrees, 1)
		assert.True(t, worktrees[0].IsBare)
	})

	t.Run("handles detached HEAD", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project
HEAD abc1234567890def
detached

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: porcelainOutput},
			},
		}
		wm := NewWorktreeManager(mock)

		worktrees, err := wm.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, worktrees, 1)
		assert.Equal(t, "(detached)", worktrees[0].Branch)
	})

	t.Run("handles empty list", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		wm := NewWorktreeManager(mock)

		worktrees, err := wm.List(context.Background())
		require.NoError(t, err)
		assert.Empty(t, worktrees)
	})
}

func TestWorktreeManager_parseWorktreeList(t *testing.T) {
	wm := NewWorktreeManager(nil)

	t.Run("parses multiple worktrees", func(t *testing.T) {
		output := `worktree /path/to/main
HEAD 1234567890abcdef
branch refs/heads/main

worktree /path/to/feature
HEAD abcdef1234567890
branch refs/heads/feature/new

`
		worktrees, err := wm.parseWorktreeList(output)
		require.NoError(t, err)
		assert.Len(t, worktrees, 2)

		assert.Equal(t, "/path/to/main", worktrees[0].Path)
		assert.Equal(t, "main", worktrees[0].Branch)
		assert.True(t, worktrees[0].IsMain)

		assert.Equal(t, "/path/to/feature", worktrees[1].Path)
		assert.Equal(t, "feature/new", worktrees[1].Branch)
		assert.False(t, worktrees[1].IsMain)
	})

	t.Run("handles short commit hash", func(t *testing.T) {
		output := `worktree /path
HEAD abc123
branch refs/heads/main

`
		worktrees, err := wm.parseWorktreeList(output)
		require.NoError(t, err)
		assert.Equal(t, "abc123", worktrees[0].CommitHash)
	})
}

func TestWorktreeManager_Setup(t *testing.T) {
	t.Run("creates new worktree", func(t *testing.T) {
		// Use temp directory for testing
		tmpDir := t.TempDir()
		projectDir := tmpDir + "/project"

		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: projectDir}, // GetRootPath
				{Stdout: ""},         // List (empty)
				{Stdout: ""},         // git worktree add
			},
		}
		wm := NewWorktreeManager(mock)

		path, err := wm.Setup(context.Background(), "my-worktree", nil)
		require.NoError(t, err)
		assert.Contains(t, path, "my-worktree")
	})

	t.Run("returns existing worktree path", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project
HEAD abc1234567890def
branch refs/heads/main

worktree /home/user/claude-loop-worktrees/my-worktree
HEAD def4567890123abc
branch refs/heads/feature

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "/home/user/project"}, // GetRootPath
				{Stdout: porcelainOutput},      // List
			},
		}
		wm := NewWorktreeManager(mock)

		path, err := wm.Setup(context.Background(), "my-worktree", nil)
		require.NoError(t, err)
		assert.Equal(t, "/home/user/claude-loop-worktrees/my-worktree", path)
	})

	t.Run("creates branch when CreateBranch is true", func(t *testing.T) {
		// Use temp directory for testing
		tmpDir := t.TempDir()

		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: tmpDir + "/project"}, // GetRootPath
				{Stdout: ""},                  // List (empty)
				{Stdout: ""},                  // git worktree add -b
			},
		}
		wm := NewWorktreeManager(mock)

		opts := &WorktreeOptions{
			BaseDir:      tmpDir + "/worktrees",
			CreateBranch: true,
			BaseBranch:   "main",
		}
		path, err := wm.Setup(context.Background(), "new-wt", opts)
		require.NoError(t, err)
		assert.Contains(t, path, "new-wt")
	})

	t.Run("rejects empty name", func(t *testing.T) {
		wm := NewWorktreeManager(nil)

		_, err := wm.Setup(context.Background(), "", nil)
		assert.Error(t, err)
		assert.True(t, IsWorktreeError(err))
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("rejects path traversal in name", func(t *testing.T) {
		wm := NewWorktreeManager(nil)

		_, err := wm.Setup(context.Background(), "../malicious", nil)
		assert.Error(t, err)
		assert.True(t, IsWorktreeError(err))
		assert.Contains(t, err.Error(), "invalid characters")
	})

	t.Run("rejects path separator in name", func(t *testing.T) {
		wm := NewWorktreeManager(nil)

		_, err := wm.Setup(context.Background(), "path/to/wt", nil)
		assert.Error(t, err)
		assert.True(t, IsWorktreeError(err))
		assert.Contains(t, err.Error(), "invalid characters")
	})
}

func TestWorktreeManager_Remove(t *testing.T) {
	t.Run("removes worktree by path", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""}, // git worktree remove
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Remove(context.Background(), "/home/user/worktrees/test", false)
		require.NoError(t, err)
	})

	t.Run("removes worktree by name", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project
HEAD abc1234567890def
branch refs/heads/main

worktree /home/user/worktrees/my-worktree
HEAD def4567890123abc
branch refs/heads/feature

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: porcelainOutput}, // List
				{Stdout: ""},              // git worktree remove
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Remove(context.Background(), "my-worktree", false)
		require.NoError(t, err)
	})

	t.Run("error when worktree not found", func(t *testing.T) {
		porcelainOutput := `worktree /home/user/project
HEAD abc1234567890def
branch refs/heads/main

`
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: porcelainOutput}, // List
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Remove(context.Background(), "nonexistent", false)
		assert.Error(t, err)
		assert.True(t, IsWorktreeError(err))
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("force removes worktree", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""}, // git worktree remove --force
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Remove(context.Background(), "/path/to/wt", true)
		require.NoError(t, err)
	})
}

func TestWorktreeManager_Prune(t *testing.T) {
	t.Run("prunes successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Prune(context.Background())
		require.NoError(t, err)
	})

	t.Run("handles prune error", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "error pruning"},
			},
		}
		wm := NewWorktreeManager(mock)

		err := wm.Prune(context.Background())
		assert.Error(t, err)
		assert.True(t, IsGitError(err))
	})
}

func TestWorktreeManager_FormatList(t *testing.T) {
	wm := NewWorktreeManager(nil)

	t.Run("formats empty list", func(t *testing.T) {
		result := wm.FormatList(nil)
		assert.Equal(t, "No worktrees found.", result)
	})

	t.Run("formats worktrees", func(t *testing.T) {
		worktrees := []WorktreeInfo{
			{
				Path:       "/home/user/project",
				Branch:     "main",
				CommitHash: "abc1234",
				IsMain:     true,
			},
			{
				Path:       "/home/user/worktrees/feature",
				Branch:     "feature/test",
				CommitHash: "def5678",
				IsMain:     false,
			},
		}

		result := wm.FormatList(worktrees)

		assert.Contains(t, result, "Active worktrees:")
		assert.Contains(t, result, "* /home/user/project")
		assert.Contains(t, result, "  /home/user/worktrees/feature")
		assert.Contains(t, result, "Branch: main")
		assert.Contains(t, result, "Branch: feature/test")
		assert.Contains(t, result, "Commit: abc1234")
		assert.Contains(t, result, "Commit: def5678")
	})

	t.Run("handles worktree without commit hash", func(t *testing.T) {
		worktrees := []WorktreeInfo{
			{
				Path:   "/path/to/wt",
				Branch: "main",
				IsMain: true,
			},
		}

		result := wm.FormatList(worktrees)
		assert.Contains(t, result, "Branch: main")
		assert.NotContains(t, result, "Commit:")
	})
}

func TestNewWorktreeManager(t *testing.T) {
	t.Run("uses default executor when nil", func(t *testing.T) {
		wm := NewWorktreeManager(nil)
		assert.NotNil(t, wm)
		assert.NotNil(t, wm.executor)
		assert.NotNil(t, wm.repo)
		assert.NotNil(t, wm.branch)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		wm := NewWorktreeManager(mock)
		assert.Equal(t, mock, wm.executor)
	})
}

func TestDefaultWorktreeOptions(t *testing.T) {
	opts := DefaultWorktreeOptions()
	assert.Equal(t, "../claude-loop-worktrees", opts.BaseDir)
	assert.False(t, opts.CreateBranch)
	assert.Empty(t, opts.BaseBranch)
}
