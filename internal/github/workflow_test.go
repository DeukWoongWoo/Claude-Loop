package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkflowManager_RunPRWorkflow(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("successful workflow", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// Create PR
				{Stdout: "https://github.com/owner/repo/pull/42"},
				// Wait for checks (one poll)
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
				// Get final check status
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
				// Merge PR
				{Stdout: "Merged"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		cfg := &WorkflowConfig{
			MergeStrategy: MergeStrategySquash,
			DeleteBranch:  true,
			WaitOptions: &WaitOptions{
				MaxIterations: 1,
				PollInterval:  1 * time.Millisecond,
				InitialWait:   0,
			},
		}

		result, err := manager.RunPRWorkflow(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
			Base:  "main",
		}, cfg)

		require.NoError(t, err)
		assert.Equal(t, 42, result.PRNumber)
		assert.Equal(t, "Test PR", result.PRTitle)
		assert.True(t, result.Merged)
		assert.Nil(t, result.MergeError)
	})

	t.Run("dry run mode", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{},
		}
		manager := NewWorkflowManager(mock, repo)

		cfg := &WorkflowConfig{
			DryRun: true,
		}

		result, err := manager.RunPRWorkflow(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
		}, cfg)

		require.NoError(t, err)
		assert.Equal(t, "Test PR", result.PRTitle)
		assert.False(t, result.Merged)
	})

	t.Run("fails on check failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// Create PR
				{Stdout: "https://github.com/owner/repo/pull/42"},
				// Wait for checks - fails
				{Stdout: `[{"name":"build","state":"FAILURE","bucket":"fail"}]`},
				// Get final check status
				{Stdout: `[{"name":"build","state":"FAILURE","bucket":"fail"}]`},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		cfg := &WorkflowConfig{
			WaitOptions: &WaitOptions{
				MaxIterations: 1,
				PollInterval:  1 * time.Millisecond,
				InitialWait:   0,
			},
		}

		result, err := manager.RunPRWorkflow(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
		}, cfg)

		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 42, result.PRNumber)
		assert.False(t, result.Merged)
		assert.NotNil(t, result.MergeError)
	})

	t.Run("calls progress callback", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/owner/repo/pull/42"},
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
				{Stdout: "Merged"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		progressCalls := []string{}
		cfg := &WorkflowConfig{
			MergeStrategy: MergeStrategySquash,
			WaitOptions: &WaitOptions{
				MaxIterations: 1,
				PollInterval:  1 * time.Millisecond,
				InitialWait:   0,
			},
			OnProgress: func(status string) {
				progressCalls = append(progressCalls, status)
			},
		}

		_, err := manager.RunPRWorkflow(context.Background(), &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
		}, cfg)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(progressCalls), 4) // Creating, Created, Waiting, All checks passed, Merging, Merged
	})

	t.Run("uses default config when nil", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "https://github.com/owner/repo/pull/42"},
				// Will timeout waiting for checks with default options
				{Stdout: `[{"name":"build","state":"PENDING","bucket":"pending"}]`},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := manager.RunPRWorkflow(ctx, &PRCreateOptions{
			Title: "Test PR",
			Body:  "Test body",
		}, nil)

		// Should timeout or get context cancellation
		assert.Error(t, err)
	})
}

func TestWorkflowManager_MergeAndCleanup(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("merges open PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// GetInfo
				{Stdout: `{"number":123,"state":"OPEN","mergeable":"MERGEABLE"}`},
				// Merge
				{Stdout: "Merged"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.MergeAndCleanup(context.Background(), 123, MergeStrategySquash, true)
		assert.NoError(t, err)
	})

	t.Run("fails for closed PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"number":123,"state":"CLOSED","mergeable":"UNKNOWN"}`},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.MergeAndCleanup(context.Background(), 123, MergeStrategySquash, true)
		assert.Error(t, err)
		assert.True(t, IsPRError(err))
	})

	t.Run("fails for unmergeable PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"number":123,"state":"OPEN","mergeable":"CONFLICTING"}`},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.MergeAndCleanup(context.Background(), 123, MergeStrategySquash, true)
		assert.Error(t, err)
	})
}

func TestWorkflowManager_HandleFailedChecks(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("closes PR on failed checks", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "Closed"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.HandleFailedChecks(context.Background(), 123, true)
		assert.NoError(t, err)
	})
}

func TestWorkflowManager_TryUpdateBranch(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("updates branch successfully", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "Updated"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.TryUpdateBranch(context.Background(), 123)
		assert.NoError(t, err)
	})

	t.Run("returns nil for already up-to-date", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "already up to date"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.TryUpdateBranch(context.Background(), 123)
		assert.NoError(t, err) // Should not return error
	})

	t.Run("returns error for conflict", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "merge conflict"},
			},
		}
		manager := NewWorkflowManager(mock, repo)

		err := manager.TryUpdateBranch(context.Background(), 123)
		assert.Equal(t, ErrPRMergeConflict, err)
	})
}

func TestNewWorkflowManager(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("uses default executor when nil", func(t *testing.T) {
		manager := NewWorkflowManager(nil, repo)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.executor)
		assert.NotNil(t, manager.prManager)
		assert.NotNil(t, manager.monitor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		manager := NewWorkflowManager(mock, repo)
		assert.Equal(t, mock, manager.executor)
	})
}

func TestWorkflowManager_GetPRManager(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}
	manager := NewWorkflowManager(nil, repo)
	assert.NotNil(t, manager.GetPRManager())
}

func TestWorkflowManager_GetCheckMonitor(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}
	manager := NewWorkflowManager(nil, repo)
	assert.NotNil(t, manager.GetCheckMonitor())
}

func TestDefaultWorkflowConfig(t *testing.T) {
	cfg := DefaultWorkflowConfig()
	assert.Equal(t, MergeStrategySquash, cfg.MergeStrategy)
	assert.NotNil(t, cfg.WaitOptions)
	assert.False(t, cfg.DryRun)
	assert.True(t, cfg.DeleteBranch)
}

func TestDefaultWaitOptions(t *testing.T) {
	opts := DefaultWaitOptions()
	assert.Equal(t, 180, opts.MaxIterations)
	assert.Equal(t, 10*time.Second, opts.PollInterval)
	assert.Equal(t, 3*time.Minute, opts.InitialWait)
	assert.False(t, opts.RequireApproval)
}
