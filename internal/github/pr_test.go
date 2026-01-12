package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePRNumber(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    int
		wantErr bool
	}{
		{
			name: "standard PR URL",
			url:  "https://github.com/owner/repo/pull/123",
			want: 123,
		},
		{
			name: "PR URL with trailing slash",
			url:  "https://github.com/owner/repo/pull/456/",
			want: 456,
		},
		{
			name: "PR URL with additional path",
			url:  "https://github.com/owner/repo/pull/789/files",
			want: 789,
		},
		{
			name:    "invalid URL",
			url:     "https://github.com/owner/repo",
			wantErr: true,
		},
		{
			name:    "not a PR URL",
			url:     "https://github.com/owner/repo/issues/123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parsePRNumber(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPRManager_Create(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("creates PR successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/owner/repo/pull/42"},
			},
		}
		manager := NewPRManager(mock, repo)

		prNum, url, err := manager.Create(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
			Base:  "main",
		})
		require.NoError(t, err)
		assert.Equal(t, 42, prNum)
		assert.Equal(t, "https://github.com/owner/repo/pull/42", url)
	})

	t.Run("creates draft PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/owner/repo/pull/43"},
			},
		}
		manager := NewPRManager(mock, repo)

		prNum, _, err := manager.Create(context.Background(), &PRCreateOptions{
			Title: "Draft PR",
			Body:  "Test body",
			Draft: true,
		})
		require.NoError(t, err)
		assert.Equal(t, 43, prNum)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "pull request create failed"},
			},
		}
		manager := NewPRManager(mock, repo)

		_, _, err := manager.Create(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
		})
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
	})

	t.Run("returns error when options nil", func(t *testing.T) {
		manager := NewPRManager(&MockExecutor{}, repo)

		_, _, err := manager.Create(context.Background(), nil)
		assert.Error(t, err)
	})
}

func TestPRManager_GetInfo(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("returns PR info", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"number":123,"title":"Test PR","body":"Body","state":"OPEN","headRefName":"feature","baseRefName":"main","headRefOid":"abc123","reviewDecision":"APPROVED","reviewRequests":[],"mergeable":"MERGEABLE","createdAt":"2024-01-01T00:00:00Z","url":"https://github.com/owner/repo/pull/123"}`},
			},
		}
		manager := NewPRManager(mock, repo)

		info, err := manager.GetInfo(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, 123, info.Number)
		assert.Equal(t, "Test PR", info.Title)
		assert.Equal(t, PRStateOpen, info.State)
		assert.Equal(t, "feature", info.HeadBranch)
		assert.Equal(t, "main", info.BaseBranch)
		assert.Equal(t, "APPROVED", info.ReviewDecision)
		assert.True(t, info.IsMergeable)
	})

	t.Run("returns error for not found", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "Could not resolve to a PullRequest: not found"},
			},
		}
		manager := NewPRManager(mock, repo)

		_, err := manager.GetInfo(context.Background(), 999)
		assert.Error(t, err)
		assert.Equal(t, ErrPRNotFound, err)
	})
}

func TestPRManager_UpdateBranch(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("updates branch successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "Updated branch"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.UpdateBranch(context.Background(), 123)
		assert.NoError(t, err)
	})

	t.Run("returns already up-to-date", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "already up to date"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.UpdateBranch(context.Background(), 123)
		assert.Equal(t, ErrPRAlreadyUpToDate, err)
	})

	t.Run("returns merge conflict error", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "merge conflict"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.UpdateBranch(context.Background(), 123)
		assert.Equal(t, ErrPRMergeConflict, err)
	})
}

func TestPRManager_Merge(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	tests := []struct {
		name     string
		strategy MergeStrategy
	}{
		{"squash", MergeStrategySquash},
		{"merge", MergeStrategyMerge},
		{"rebase", MergeStrategyRebase},
	}

	for _, tt := range tests {
		t.Run(tt.name+" strategy", func(t *testing.T) {
			mock := &MockExecutor{
				Commands: []MockCommand{
					{Stdout: "Merged"},
				},
			}
			manager := NewPRManager(mock, repo)

			err := manager.Merge(context.Background(), 123, tt.strategy, true)
			assert.NoError(t, err)
		})
	}

	t.Run("returns merge conflict error", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "merge conflict detected"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.Merge(context.Background(), 123, MergeStrategySquash, false)
		assert.Equal(t, ErrPRMergeConflict, err)
	})

	t.Run("returns not mergeable error", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "Pull request is not mergeable"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.Merge(context.Background(), 123, MergeStrategySquash, false)
		assert.Equal(t, ErrPRMergeFailed, err)
	})
}

func TestPRManager_Close(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("closes PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "Closed"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.Close(context.Background(), 123, false)
		assert.NoError(t, err)
	})

	t.Run("closes PR and deletes branch", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "Closed and deleted branch"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.Close(context.Background(), 123, true)
		assert.NoError(t, err)
	})

	t.Run("returns error on failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "PR close failed"},
			},
		}
		manager := NewPRManager(mock, repo)

		err := manager.Close(context.Background(), 123, false)
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
	})
}

func TestNewPRManager(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("uses default executor when nil", func(t *testing.T) {
		manager := NewPRManager(nil, repo)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.executor)
		assert.NotNil(t, manager.monitor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		manager := NewPRManager(mock, repo)
		assert.Equal(t, mock, manager.executor)
	})
}

func TestPRManager_GetCheckMonitor(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}
	manager := NewPRManager(nil, repo)
	monitor := manager.GetCheckMonitor()
	assert.NotNil(t, monitor)
}
