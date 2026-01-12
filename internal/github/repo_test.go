package github

import (
	"context"
	"os/exec"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor simulates command execution for testing.
type MockExecutor struct {
	Commands []MockCommand
	index    int
}

// MockCommand represents a single command mock.
type MockCommand struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (m *MockExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	if m.index >= len(m.Commands) {
		return exec.CommandContext(ctx, "false")
	}

	mock := m.Commands[m.index]
	m.index++

	if mock.ExitCode == 0 {
		return exec.CommandContext(ctx, "echo", "-n", mock.Stdout)
	}

	// For non-zero exit codes, use a shell command
	return exec.CommandContext(ctx, "sh", "-c", "echo -n '"+mock.Stderr+"' >&2; exit "+strconv.Itoa(mock.ExitCode))
}

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantRepo  string
		wantErr   bool
	}{
		{
			name:      "HTTPS URL with .git",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "HTTPS URL without .git",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH URL with .git",
			url:       "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "SSH URL without .git",
			url:       "git@github.com:owner/repo",
			wantOwner: "owner",
			wantRepo:  "repo",
		},
		{
			name:      "organization owner",
			url:       "https://github.com/my-org/my-project.git",
			wantOwner: "my-org",
			wantRepo:  "my-project",
		},
		{
			name:      "repo with dots",
			url:       "https://github.com/owner/repo.name.git",
			wantOwner: "owner",
			wantRepo:  "repo.name",
		},
		{
			name:    "not github URL",
			url:     "https://gitlab.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "invalid URL format",
			url:     "not-a-url",
			wantErr: true,
		},
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseRemoteURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOwner, info.Owner)
			assert.Equal(t, tt.wantRepo, info.Repo)
		})
	}
}

func TestRepoDetector_DetectRepo(t *testing.T) {
	t.Run("detects GitHub repo from HTTPS URL", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/owner/repo.git"},
			},
		}
		detector := NewRepoDetector(mock)

		info, err := detector.DetectRepo(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "owner", info.Owner)
		assert.Equal(t, "repo", info.Repo)
	})

	t.Run("detects GitHub repo from SSH URL", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "git@github.com:owner/repo.git"},
			},
		}
		detector := NewRepoDetector(mock)

		info, err := detector.DetectRepo(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "owner", info.Owner)
		assert.Equal(t, "repo", info.Repo)
	})

	t.Run("returns error when not a git repo", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "fatal: not a git repository"},
			},
		}
		detector := NewRepoDetector(mock)

		_, err := detector.DetectRepo(context.Background())
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
	})

	t.Run("returns error for non-GitHub repo", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://gitlab.com/owner/repo.git"},
			},
		}
		detector := NewRepoDetector(mock)

		_, err := detector.DetectRepo(context.Background())
		assert.Error(t, err)
		assert.Equal(t, ErrNotGitHubRepo, err)
	})
}

func TestRepoDetector_ValidateGHCLI(t *testing.T) {
	t.Run("gh CLI installed and authenticated", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "gh version 2.40.0"},          // gh --version
				{Stdout: "Logged in to github.com as"}, // gh auth status
			},
		}
		detector := NewRepoDetector(mock)

		err := detector.ValidateGHCLI(context.Background())
		assert.NoError(t, err)
	})

	t.Run("gh CLI not installed", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 127, Stderr: "command not found: gh"},
			},
		}
		detector := NewRepoDetector(mock)

		err := detector.ValidateGHCLI(context.Background())
		assert.Equal(t, ErrGHCLINotFound, err)
	})

	t.Run("gh CLI not authenticated", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "gh version 2.40.0"},
				{ExitCode: 1, Stderr: "You are not logged into any GitHub hosts"},
			},
		}
		detector := NewRepoDetector(mock)

		err := detector.ValidateGHCLI(context.Background())
		assert.Equal(t, ErrGHNotAuthenticated, err)
	})
}

func TestRepoInfo_RepoString(t *testing.T) {
	info := &RepoInfo{
		Owner: "owner",
		Repo:  "repo",
	}
	assert.Equal(t, "owner/repo", info.RepoString())
}

func TestNewRepoDetector(t *testing.T) {
	t.Run("uses default executor when nil", func(t *testing.T) {
		detector := NewRepoDetector(nil)
		assert.NotNil(t, detector)
		assert.NotNil(t, detector.executor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		detector := NewRepoDetector(mock)
		assert.Equal(t, mock, detector.executor)
	})
}
