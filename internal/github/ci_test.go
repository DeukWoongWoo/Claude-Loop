package github

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCIAnalyzer_GetFailureLogs(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("parses failure info correctly", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// gh run view --json
				{Stdout: `{
					"databaseId": 12345,
					"name": "CI",
					"conclusion": "failure",
					"url": "https://github.com/owner/repo/actions/runs/12345",
					"createdAt": "2026-01-12T10:00:00Z",
					"jobs": [
						{
							"name": "build",
							"conclusion": "failure",
							"steps": [
								{"name": "Checkout", "conclusion": "success"},
								{"name": "Build", "conclusion": "failure"},
								{"name": "Test", "conclusion": "skipped"}
							]
						}
					]
				}`},
				// gh run view --log-failed
				{Stdout: "Error: Build failed\nsome error message"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetFailureLogs(context.Background(), "12345")
		require.NoError(t, err)
		assert.Equal(t, "12345", info.RunID)
		assert.Equal(t, "CI", info.WorkflowName)
		assert.Equal(t, "build", info.JobName)
		assert.Equal(t, []string{"Build"}, info.FailedSteps)
		assert.Contains(t, info.ErrorLogs, "Build failed")
		assert.Equal(t, "https://github.com/owner/repo/actions/runs/12345", info.URL)
	})

	t.Run("handles multiple failed steps", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{
					"databaseId": 12345,
					"name": "CI",
					"conclusion": "failure",
					"url": "https://github.com/owner/repo/actions/runs/12345",
					"createdAt": "2026-01-12T10:00:00Z",
					"jobs": [
						{
							"name": "test",
							"conclusion": "failure",
							"steps": [
								{"name": "Unit Tests", "conclusion": "failure"},
								{"name": "Integration Tests", "conclusion": "failure"}
							]
						}
					]
				}`},
				{Stdout: "Test failures"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetFailureLogs(context.Background(), "12345")
		require.NoError(t, err)
		assert.Equal(t, []string{"Unit Tests", "Integration Tests"}, info.FailedSteps)
	})

	t.Run("handles log retrieval failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{
					"databaseId": 12345,
					"name": "CI",
					"conclusion": "failure",
					"url": "https://github.com/owner/repo/actions/runs/12345",
					"createdAt": "2026-01-12T10:00:00Z",
					"jobs": []
				}`},
				// gh run view --log-failed fails
				{ExitCode: 1, Stderr: "failed to get logs"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetFailureLogs(context.Background(), "12345")
		require.NoError(t, err)
		assert.Equal(t, "(failed to retrieve logs)", info.ErrorLogs)
	})

	t.Run("truncates long logs", func(t *testing.T) {
		longLog := make([]byte, 10000)
		for i := range longLog {
			longLog[i] = 'x'
		}
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{
					"databaseId": 12345,
					"name": "CI",
					"conclusion": "failure",
					"url": "https://github.com/owner/repo/actions/runs/12345",
					"createdAt": "2026-01-12T10:00:00Z",
					"jobs": []
				}`},
				{Stdout: string(longLog)},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetFailureLogs(context.Background(), "12345")
		require.NoError(t, err)
		assert.True(t, len(info.ErrorLogs) <= 5020) // 5000 + truncation prefix
		assert.Contains(t, info.ErrorLogs, "...(truncated)...")
	})

	t.Run("error on run info failure", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{ExitCode: 1, Stderr: "run not found"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		_, err := analyzer.GetFailureLogs(context.Background(), "99999")
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
		assert.Contains(t, err.Error(), "failed to get run info")
	})

	t.Run("error on invalid JSON", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: "invalid json"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		_, err := analyzer.GetFailureLogs(context.Background(), "12345")
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
		assert.Contains(t, err.Error(), "failed to parse run info")
	})

	t.Run("error on empty runID", func(t *testing.T) {
		analyzer := NewCIAnalyzer(nil, repo)

		_, err := analyzer.GetFailureLogs(context.Background(), "")
		assert.Error(t, err)
		assert.True(t, IsGitHubError(err))
		assert.Contains(t, err.Error(), "runID cannot be empty")
	})
}

func TestCIAnalyzer_GetLatestFailure(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("returns failure info for PR", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// GetFailedRunID - get PR head SHA
				{Stdout: `{"headRefOid":"abc123"}`},
				// GetFailedRunID - list failed runs
				{Stdout: `[{"databaseId":12345}]`},
				// GetFailureLogs - run view --json
				{Stdout: `{
					"databaseId": 12345,
					"name": "CI",
					"conclusion": "failure",
					"url": "https://github.com/owner/repo/actions/runs/12345",
					"createdAt": "2026-01-12T10:00:00Z",
					"jobs": [{"name": "build", "conclusion": "failure", "steps": []}]
				}`},
				// GetFailureLogs - log-failed
				{Stdout: "Error logs"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetLatestFailure(context.Background(), 42)
		require.NoError(t, err)
		require.NotNil(t, info)
		assert.Equal(t, "12345", info.RunID)
		assert.Equal(t, "CI", info.WorkflowName)
	})

	t.Run("returns nil when no failures", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// GetFailedRunID - get PR head SHA
				{Stdout: `{"headRefOid":"abc123"}`},
				// GetFailedRunID - no failed runs
				{Stdout: `[]`},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		info, err := analyzer.GetLatestFailure(context.Background(), 42)
		require.NoError(t, err)
		assert.Nil(t, info)
	})

	t.Run("error when GetFailedRunID fails", func(t *testing.T) {
		mock := &MockExecutor{
			Commands: []MockCommand{
				// GetFailedRunID fails
				{ExitCode: 1, Stderr: "not found"},
			},
		}
		analyzer := NewCIAnalyzer(mock, repo)

		_, err := analyzer.GetLatestFailure(context.Background(), 42)
		assert.Error(t, err)
	})
}

func TestNewCIAnalyzer(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("uses default executor when nil", func(t *testing.T) {
		analyzer := NewCIAnalyzer(nil, repo)
		assert.NotNil(t, analyzer)
		assert.NotNil(t, analyzer.executor)
		assert.NotNil(t, analyzer.checkMonitor)
	})

	t.Run("uses provided executor", func(t *testing.T) {
		mock := &MockExecutor{}
		analyzer := NewCIAnalyzer(mock, repo)
		assert.Equal(t, mock, analyzer.executor)
	})
}
