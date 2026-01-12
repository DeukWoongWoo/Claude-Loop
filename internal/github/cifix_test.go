package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testWaitOptions returns WaitOptions suitable for testing (no delays).
func testWaitOptions() *WaitOptions {
	return &WaitOptions{
		InitialWait:   0,
		PollInterval:  time.Millisecond,
		MaxIterations: 1,
	}
}

// MockClaudeClient implements loop.ClaudeClient for testing.
type MockClaudeClient struct {
	Results    []*loop.IterationResult
	Errors     []error
	CallCount  int
	LastPrompt string
}

func (m *MockClaudeClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	m.LastPrompt = prompt
	idx := m.CallCount
	m.CallCount++

	if idx < len(m.Errors) && m.Errors[idx] != nil {
		return nil, m.Errors[idx]
	}
	if idx < len(m.Results) {
		return m.Results[idx], nil
	}
	return &loop.IterationResult{Output: "Fixed the issue"}, nil
}

// MockGitExecutor wraps MockExecutor for git operations testing.
type MockGitExecutor struct {
	MockExecutor
}

func TestCIFixManager_AttemptFix(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("successful fix on first attempt", func(t *testing.T) {
		ghMock := &MockExecutor{
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
				{Stdout: "Error: Build failed"},
				// WaitForChecks - GetCheckStatus
				{Stdout: `[{"name":"CI","state":"SUCCESS","bucket":"pass"}]`},
			},
		}

		gitMock := &MockGitExecutor{
			MockExecutor: MockExecutor{
				Commands: []MockCommand{
					{Stdout: ""},  // git add -A
					{ExitCode: 1}, // git diff --cached --quiet (has changes)
					{Stdout: ""},  // git commit
					{Stdout: ""},  // git push
				},
			},
		}

		claudeMock := &MockClaudeClient{
			Results: []*loop.IterationResult{{Output: "Fixed"}},
		}

		config := &CIFixConfig{
			MaxRetries:  1,
			PRNumber:    42,
			BranchName:  "feature/test",
			WaitOptions: testWaitOptions(),
		}

		manager := NewCIFixManager(ghMock, &gitMock.MockExecutor, repo, claudeMock, config)
		result, err := manager.AttemptFix(context.Background())

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 1, result.Attempts)
		assert.Equal(t, 1, result.FixedOnAttempt)
		assert.Nil(t, result.LastError)
		assert.Equal(t, 1, claudeMock.CallCount)
	})

	t.Run("fix succeeds on second attempt", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				// First attempt - analyze
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[{"name":"build","conclusion":"failure","steps":[]}]}`},
				{Stdout: "Error logs"},
				// First attempt - WaitForChecks fails
				{Stdout: `[{"name":"CI","state":"FAILURE","bucket":"fail"}]`},
				// Second attempt - analyze
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12346}]`},
				{Stdout: `{"databaseId":12346,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:01:00Z","jobs":[{"name":"build","conclusion":"failure","steps":[]}]}`},
				{Stdout: "New error logs"},
				// Second attempt - WaitForChecks passes
				{Stdout: `[{"name":"CI","state":"SUCCESS","bucket":"pass"}]`},
			},
		}

		gitMock := &MockGitExecutor{
			MockExecutor: MockExecutor{
				Commands: []MockCommand{
					// First attempt
					{Stdout: ""},  // git add -A
					{ExitCode: 1}, // git diff --cached (has changes)
					{Stdout: ""},  // git commit
					{Stdout: ""},  // git push
					// Second attempt
					{Stdout: ""},  // git add -A
					{ExitCode: 1}, // git diff --cached (has changes)
					{Stdout: ""},  // git commit
					{Stdout: ""},  // git push
				},
			},
		}

		claudeMock := &MockClaudeClient{}

		config := &CIFixConfig{
			MaxRetries:  3,
			PRNumber:    42,
			WaitOptions: testWaitOptions(),
		}

		manager := NewCIFixManager(ghMock, &gitMock.MockExecutor, repo, claudeMock, config)
		result, err := manager.AttemptFix(context.Background())

		require.NoError(t, err)
		assert.True(t, result.Success)
		assert.Equal(t, 2, result.Attempts)
		assert.Equal(t, 2, result.FixedOnAttempt)
		assert.Equal(t, 2, claudeMock.CallCount)
	})

	t.Run("all attempts fail", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				// First attempt
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[]}`},
				{Stdout: "Error"},
				{Stdout: `[{"name":"CI","state":"FAILURE","bucket":"fail"}]`},
				// Second attempt
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12346}]`},
				{Stdout: `{"databaseId":12346,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:01:00Z","jobs":[]}`},
				{Stdout: "Error"},
				{Stdout: `[{"name":"CI","state":"FAILURE","bucket":"fail"}]`},
			},
		}

		gitMock := &MockGitExecutor{
			MockExecutor: MockExecutor{
				Commands: []MockCommand{
					{Stdout: ""}, {ExitCode: 1}, {Stdout: ""}, {Stdout: ""}, // First
					{Stdout: ""}, {ExitCode: 1}, {Stdout: ""}, {Stdout: ""}, // Second
				},
			},
		}

		claudeMock := &MockClaudeClient{}

		config := &CIFixConfig{
			MaxRetries:  2,
			PRNumber:    42,
			WaitOptions: testWaitOptions(),
		}

		manager := NewCIFixManager(ghMock, &gitMock.MockExecutor, repo, claudeMock, config)
		result, err := manager.AttemptFix(context.Background())

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, 2, result.Attempts)
		assert.Equal(t, 0, result.FixedOnAttempt)
		assert.NotNil(t, result.LastError)
	})

	t.Run("disabled retry returns immediately", func(t *testing.T) {
		config := &CIFixConfig{
			DisableRetry: true,
		}

		manager := NewCIFixManager(nil, nil, repo, nil, config)
		result, err := manager.AttemptFix(context.Background())

		require.NoError(t, err)
		assert.False(t, result.Success)
		assert.Equal(t, 0, result.Attempts)
		assert.ErrorIs(t, result.LastError, ErrCIFixDisabled)
	})

	t.Run("no CI failures found", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[]`}, // No failed runs
			},
		}

		config := &CIFixConfig{
			MaxRetries: 1,
			PRNumber:   42,
		}

		manager := NewCIFixManager(ghMock, nil, repo, nil, config)
		result, err := manager.AttemptFix(context.Background())

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.True(t, IsCIFixError(result.LastError))
		assert.Contains(t, result.LastError.Error(), "no CI failures found")
	})

	t.Run("Claude execution fails", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[]}`},
				{Stdout: "Error"},
			},
		}

		claudeMock := &MockClaudeClient{
			Errors: []error{errors.New("Claude API error")},
		}

		config := &CIFixConfig{
			MaxRetries: 1,
			PRNumber:   42,
		}

		manager := NewCIFixManager(ghMock, nil, repo, claudeMock, config)
		result, err := manager.AttemptFix(context.Background())

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.True(t, IsCIFixError(result.LastError))
		assert.Contains(t, result.LastError.Error(), "Claude execution failed")
	})

	t.Run("Claude makes no changes", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[]}`},
				{Stdout: "Error"},
			},
		}

		gitMock := &MockGitExecutor{
			MockExecutor: MockExecutor{
				Commands: []MockCommand{
					{Stdout: ""},  // git add -A
					{ExitCode: 0}, // git diff --cached (no changes)
				},
			},
		}

		claudeMock := &MockClaudeClient{}

		config := &CIFixConfig{
			MaxRetries: 1,
			PRNumber:   42,
		}

		manager := NewCIFixManager(ghMock, &gitMock.MockExecutor, repo, claudeMock, config)
		result, err := manager.AttemptFix(context.Background())

		assert.Error(t, err)
		assert.False(t, result.Success)
		assert.True(t, IsCIFixError(result.LastError))
		assert.Contains(t, result.LastError.Error(), "Claude made no changes")
	})

	t.Run("progress callbacks are called", func(t *testing.T) {
		ghMock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[]}`},
				{Stdout: "Error"},
				{Stdout: `[{"name":"CI","state":"SUCCESS","bucket":"pass"}]`},
			},
		}

		gitMock := &MockGitExecutor{
			MockExecutor: MockExecutor{
				Commands: []MockCommand{
					{Stdout: ""}, {ExitCode: 1}, {Stdout: ""}, {Stdout: ""},
				},
			},
		}

		var progressMessages []string
		var attemptCalls []int

		config := &CIFixConfig{
			MaxRetries:  1,
			PRNumber:    42,
			WaitOptions: testWaitOptions(),
			OnProgress: func(status string) {
				progressMessages = append(progressMessages, status)
			},
			OnAttempt: func(attempt, max int) {
				attemptCalls = append(attemptCalls, attempt)
			},
		}

		manager := NewCIFixManager(ghMock, &gitMock.MockExecutor, repo, &MockClaudeClient{}, config)
		_, err := manager.AttemptFix(context.Background())

		require.NoError(t, err)
		assert.Contains(t, progressMessages, "Analyzing CI failure...")
		assert.Contains(t, progressMessages, "Building fix prompt...")
		assert.Contains(t, progressMessages, "Running Claude to fix CI failure...")
		assert.Contains(t, progressMessages, "Committing and pushing fix...")
		assert.Contains(t, progressMessages, "Waiting for CI checks...")
		assert.Equal(t, []int{1}, attemptCalls)
	})

	t.Run("context cancellation stops fix attempts", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		ghMock := &MockExecutor{
			Commands: []MockCommand{
				{Stdout: `{"headRefOid":"abc123"}`},
				{Stdout: `[{"databaseId":12345}]`},
				{Stdout: `{"databaseId":12345,"name":"CI","conclusion":"failure","url":"","createdAt":"2026-01-12T10:00:00Z","jobs":[]}`},
				{Stdout: "Error"},
			},
		}

		config := &CIFixConfig{
			MaxRetries: 3,
			PRNumber:   42,
		}

		claudeMock := &MockClaudeClient{
			Errors: []error{context.Canceled},
		}

		manager := NewCIFixManager(ghMock, nil, repo, claudeMock, config)
		result, err := manager.AttemptFix(ctx)

		assert.Error(t, err)
		assert.False(t, result.Success)
		// Should not have retried after cancellation
		assert.LessOrEqual(t, result.Attempts, 1)
	})
}

func TestNewCIFixManager(t *testing.T) {
	repo := &RepoInfo{Owner: "owner", Repo: "repo"}

	t.Run("uses default executor when nil", func(t *testing.T) {
		manager := NewCIFixManager(nil, nil, repo, nil, nil)
		assert.NotNil(t, manager)
		assert.NotNil(t, manager.analyzer)
		assert.NotNil(t, manager.commitMgr)
		assert.NotNil(t, manager.checkMonitor)
		assert.NotNil(t, manager.promptBuilder)
		assert.NotNil(t, manager.config)
	})

	t.Run("uses default config when nil", func(t *testing.T) {
		manager := NewCIFixManager(nil, nil, repo, nil, nil)
		assert.Equal(t, 1, manager.config.MaxRetries)
		assert.False(t, manager.config.DisableRetry)
	})

	t.Run("uses provided config", func(t *testing.T) {
		config := &CIFixConfig{
			MaxRetries:   5,
			DisableRetry: true,
		}
		manager := NewCIFixManager(nil, nil, repo, nil, config)
		assert.Equal(t, 5, manager.config.MaxRetries)
		assert.True(t, manager.config.DisableRetry)
	})
}

func TestDefaultCIFixConfig(t *testing.T) {
	config := DefaultCIFixConfig()
	assert.Equal(t, 1, config.MaxRetries)
	assert.False(t, config.DisableRetry)
	assert.NotNil(t, config.WaitOptions)
}

func TestCIFixError(t *testing.T) {
	t.Run("error message with underlying error", func(t *testing.T) {
		err := &CIFixError{
			Attempt: 2,
			Phase:   "commit",
			Message: "failed to push",
			Err:     errors.New("network error"),
		}
		assert.Contains(t, err.Error(), "CI fix attempt 2")
		assert.Contains(t, err.Error(), "commit")
		assert.Contains(t, err.Error(), "failed to push")
		assert.Contains(t, err.Error(), "network error")
	})

	t.Run("error message without underlying error", func(t *testing.T) {
		err := &CIFixError{
			Attempt: 1,
			Phase:   "analyze",
			Message: "no failures found",
		}
		assert.Contains(t, err.Error(), "CI fix attempt 1")
		assert.Contains(t, err.Error(), "analyze")
		assert.Contains(t, err.Error(), "no failures found")
		// Format: "CI fix attempt 1 (analyze): no failures found"
		// Note: Format includes colons after phase, so we just verify no ": <error>" suffix
		assert.NotContains(t, err.Error(), ": <nil>")
	})

	t.Run("unwrap returns underlying error", func(t *testing.T) {
		underlying := errors.New("underlying")
		err := &CIFixError{Err: underlying}
		assert.Equal(t, underlying, err.Unwrap())
	})

	t.Run("IsCIFixError returns true for CIFixError", func(t *testing.T) {
		err := &CIFixError{Message: "test"}
		assert.True(t, IsCIFixError(err))
	})

	t.Run("IsCIFixError returns false for other errors", func(t *testing.T) {
		err := errors.New("not a CIFixError")
		assert.False(t, IsCIFixError(err))
	})
}
