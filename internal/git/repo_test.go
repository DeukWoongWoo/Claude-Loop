package git

import (
	"context"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor simulates git command execution for testing.
type MockExecutor struct {
	Commands []MockCommand
	index    int
}

// MockCommand represents a single command mock.
type MockCommand struct {
	ExpectedName string
	ExpectedArgs []string
	Stdout       string
	Stderr       string
	ExitCode     int
}

func (m *MockExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if m.index >= len(m.Commands) {
		// Return a failing command if we've run out of mocks
		return exec.CommandContext(ctx, "false")
	}

	mock := m.Commands[m.index]
	m.index++

	// Use echo to simulate output
	if mock.ExitCode == 0 {
		return exec.CommandContext(ctx, "echo", "-n", mock.Stdout)
	}

	// For non-zero exit codes, use a shell command
	return exec.CommandContext(ctx, "sh", "-c", "echo -n '"+mock.Stderr+"' >&2; exit "+string(rune('0'+mock.ExitCode)))
}

func TestRepository_IsGitRepository(t *testing.T) {
	t.Run("is a git repository", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "true"},
			},
		}
		repo := NewRepository(mock)

		isRepo, err := repo.IsGitRepository(context.Background())
		require.NoError(t, err)
		assert.True(t, isRepo)
	})

	t.Run("not a git repository", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: not a git repository"},
			},
		}
		repo := NewRepository(mock)

		isRepo, err := repo.IsGitRepository(context.Background())
		require.NoError(t, err) // Should not return error, just false
		assert.False(t, isRepo)
	})
}

func TestRepository_GetRootPath(t *testing.T) {
	t.Run("returns root path", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "/home/user/project"},
			},
		}
		repo := NewRepository(mock)

		path, err := repo.GetRootPath(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "/home/user/project", path)
	})

	t.Run("caches root path", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "/home/user/project"},
			},
		}
		repo := NewRepository(mock)

		// First call
		path1, err := repo.GetRootPath(context.Background())
		require.NoError(t, err)

		// Second call should use cache (mock has no more commands)
		path2, err := repo.GetRootPath(context.Background())
		require.NoError(t, err)

		assert.Equal(t, path1, path2)
	})

	t.Run("error when not in repo", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: not a git repository"},
			},
		}
		repo := NewRepository(mock)

		_, err := repo.GetRootPath(context.Background())
		assert.Error(t, err)
		assert.True(t, IsGitError(err))
	})
}

func TestRepository_GetCurrentBranch(t *testing.T) {
	t.Run("returns current branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "main"},
			},
		}
		repo := NewRepository(mock)

		branch, err := repo.GetCurrentBranch(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "main", branch)
	})

	t.Run("returns feature branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "feature/new-feature"},
			},
		}
		repo := NewRepository(mock)

		branch, err := repo.GetCurrentBranch(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "feature/new-feature", branch)
	})

	t.Run("error on detached HEAD", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: ref HEAD is not a symbolic ref"},
			},
		}
		repo := NewRepository(mock)

		_, err := repo.GetCurrentBranch(context.Background())
		assert.Error(t, err)
	})
}

func TestRepository_GetRemoteURL(t *testing.T) {
	t.Run("returns remote URL", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/user/repo.git"},
			},
		}
		repo := NewRepository(mock)

		url, err := repo.GetRemoteURL(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "https://github.com/user/repo.git", url)
	})

	t.Run("returns SSH URL", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "git@github.com:user/repo.git"},
			},
		}
		repo := NewRepository(mock)

		url, err := repo.GetRemoteURL(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "git@github.com:user/repo.git", url)
	})

	t.Run("returns empty when no origin", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: No such remote 'origin'"},
			},
		}
		repo := NewRepository(mock)

		url, err := repo.GetRemoteURL(context.Background())
		require.NoError(t, err) // Should not return error
		assert.Empty(t, url)
	})
}

func TestRepository_IsClean(t *testing.T) {
	t.Run("clean working directory", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: ""},
			},
		}
		repo := NewRepository(mock)

		isClean, err := repo.IsClean(context.Background())
		require.NoError(t, err)
		assert.True(t, isClean)
	})

	t.Run("dirty working directory", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: " M file.go\n?? newfile.txt"},
			},
		}
		repo := NewRepository(mock)

		isClean, err := repo.IsClean(context.Background())
		require.NoError(t, err)
		assert.False(t, isClean)
	})
}

func TestRepository_GetInfo(t *testing.T) {
	t.Run("returns complete info", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "true"},                             // IsGitRepository
				{Stdout: "/home/user/project"},               // GetRootPath
				{Stdout: "main"},                             // GetCurrentBranch
				{Stdout: "https://github.com/user/repo.git"}, // GetRemoteURL
				{Stdout: ""},                                 // IsClean
			},
		}
		repo := NewRepository(mock)

		info, err := repo.GetInfo(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "/home/user/project", info.RootPath)
		assert.Equal(t, "main", info.CurrentBranch)
		assert.Equal(t, "https://github.com/user/repo.git", info.RemoteURL)
		assert.True(t, info.IsClean)
	})

	t.Run("error when not a repository", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1}, // IsGitRepository returns false
			},
		}
		repo := NewRepository(mock)

		_, err := repo.GetInfo(context.Background())
		assert.Error(t, err)
		assert.Equal(t, ErrNotGitRepository, err)
	})
}

func TestRepository_GetIterationDisplay(t *testing.T) {
	t.Run("returns worktree name when provided", func(t *testing.T) {
		repo := NewRepository(nil)
		display := repo.GetIterationDisplay(context.Background(), "my-worktree")
		assert.Equal(t, "my-worktree", display)
	})

	t.Run("returns branch base name when no worktree", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "feature/my-feature"},
			},
		}
		repo := NewRepository(mock)

		display := repo.GetIterationDisplay(context.Background(), "")
		assert.Equal(t, "my-feature", display)
	})

	t.Run("returns unknown on error", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1},
			},
		}
		repo := NewRepository(mock)

		display := repo.GetIterationDisplay(context.Background(), "")
		assert.Equal(t, "unknown", display)
	})
}

func TestNewRepository(t *testing.T) {
	t.Run("uses default executor when nil", func(t *testing.T) {
		repo := NewRepository(nil)
		assert.NotNil(t, repo)
		assert.NotNil(t, repo.executor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		repo := NewRepository(mock)
		assert.Equal(t, mock, repo.executor)
	})
}
