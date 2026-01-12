package github

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckMonitor_GetCheckStatus(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("parses check statuses correctly", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"},{"name":"test","state":"SUCCESS","bucket":"pass"},{"name":"lint","state":"PENDING","bucket":"pending"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		summary, err := monitor.GetCheckStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, 3, summary.Total)
		assert.Equal(t, 2, summary.Success)
		assert.Equal(t, 1, summary.Pending)
		assert.Equal(t, 0, summary.Failed)
		assert.False(t, summary.AllCompleted)
		assert.False(t, summary.AllPassed)
	})

	t.Run("all checks passed", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"},{"name":"test","state":"SUCCESS","bucket":"pass"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		summary, err := monitor.GetCheckStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.True(t, summary.AllCompleted)
		assert.True(t, summary.AllPassed)
	})

	t.Run("some checks failed", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"FAILURE","bucket":"fail"},{"name":"test","state":"SUCCESS","bucket":"pass"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		summary, err := monitor.GetCheckStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, 1, summary.Failed)
		assert.Equal(t, 1, summary.Success)
		assert.True(t, summary.AllCompleted)
		assert.False(t, summary.AllPassed)
	})

	t.Run("no checks configured", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		summary, err := monitor.GetCheckStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.True(t, summary.NoChecks)
		assert.True(t, summary.AllCompleted)
		assert.True(t, summary.AllPassed)
	})

	t.Run("handles invalid JSON", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "invalid json"},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		_, err := monitor.GetCheckStatus(context.Background(), 123)
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
	})
}

func TestCheckMonitor_GetReviewStatus(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("approved review", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"reviewDecision":"APPROVED","reviewRequests":[]}`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		decision, pending, err := monitor.GetReviewStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, "APPROVED", decision)
		assert.Equal(t, 0, pending)
	})

	t.Run("changes requested", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"reviewDecision":"CHANGES_REQUESTED","reviewRequests":[]}`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		decision, _, err := monitor.GetReviewStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, "CHANGES_REQUESTED", decision)
	})

	t.Run("pending review requests", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"reviewDecision":"","reviewRequests":[{"requestedReviewer":{"login":"reviewer1"}},{"requestedReviewer":{"login":"reviewer2"}}]}`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		decision, pending, err := monitor.GetReviewStatus(context.Background(), 123)
		require.NoError(t, err)
		assert.Empty(t, decision)
		assert.Equal(t, 2, pending)
	})
}

func TestCheckMonitor_WaitForChecks(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("checks pass immediately", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		opts := &WaitOptions{
			MaxIterations: 1,
			PollInterval:  1 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 123, opts)
		assert.NoError(t, err)
	})

	t.Run("checks fail", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"FAILURE","bucket":"fail"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		opts := &WaitOptions{
			MaxIterations: 1,
			PollInterval:  1 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 123, opts)
		assert.Error(t, err)
		assert.True(t, IsCheckError(err))
	})

	t.Run("times out waiting for checks", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"PENDING","bucket":"pending"}]`},
				{Stdout: `[{"name":"build","state":"PENDING","bucket":"pending"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		opts := &WaitOptions{
			MaxIterations: 2,
			PollInterval:  1 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 123, opts)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout")
	})

	t.Run("respects context cancellation during initial wait", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{},
		}
		monitor := NewCheckMonitor(mock, repo)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		opts := &WaitOptions{
			MaxIterations: 100,
			PollInterval:  1 * time.Second,
			InitialWait:   1 * time.Second, // Non-zero so cancellation happens during wait
		}

		err := monitor.WaitForChecks(ctx, 123, opts)
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("calls status change callback", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[{"name":"build","state":"SUCCESS","bucket":"pass"}]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		callbackCalled := false
		opts := &WaitOptions{
			MaxIterations: 1,
			PollInterval:  1 * time.Millisecond,
			InitialWait:   0,
			OnStatusChange: func(summary *CheckSummary, reviewStatus string) {
				callbackCalled = true
			},
		}

		err := monitor.WaitForChecks(context.Background(), 123, opts)
		assert.NoError(t, err)
		assert.True(t, callbackCalled)
	})

	t.Run("no checks configured passes", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `[]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		opts := &WaitOptions{
			MaxIterations: 1,
			PollInterval:  1 * time.Millisecond,
			InitialWait:   0,
		}

		err := monitor.WaitForChecks(context.Background(), 123, opts)
		assert.NoError(t, err)
	})
}

func TestCheckMonitor_GetFailedRunID(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("returns failed run ID", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},                                             // gh pr view
				{Stdout: `[{"databaseId":456789,"status":"completed","conclusion":"failure"}]`}, // gh run list
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		runID, err := monitor.GetFailedRunID(context.Background(), 123)
		require.NoError(t, err)
		assert.Equal(t, "456789", runID)
	})

	t.Run("returns empty when no failed runs", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[]`},
			},
		}
		monitor := NewCheckMonitor(mock, repo)

		runID, err := monitor.GetFailedRunID(context.Background(), 123)
		require.NoError(t, err)
		assert.Empty(t, runID)
	})
}

func TestNewCheckMonitor(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("uses default executor when nil", func(t *testing.T) {
		monitor := NewCheckMonitor(nil, repo)
		assert.NotNil(t, monitor)
		assert.NotNil(t, monitor.executor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		monitor := NewCheckMonitor(mock, repo)
		assert.Equal(t, mock, monitor.executor)
	})
}
