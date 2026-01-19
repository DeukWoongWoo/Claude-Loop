package claude

import (
	"context"
	"os"
	"os/exec"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockExecutor simulates command execution for testing.
type MockExecutor struct {
	// Script is the shell script content to execute
	Script string
	// ExitCode is the exit code to return
	ExitCode int
}

// CommandContext creates a command that executes the mock script.
func (m *MockExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", m.Script}
	cmd := exec.CommandContext(ctx, os.Args[0], cs...)
	cmd.Env = append(os.Environ(),
		"GO_WANT_HELPER_PROCESS=1",
		"MOCK_EXIT_CODE="+strconv.Itoa(m.ExitCode),
	)
	return cmd
}

// TestHelperProcess is a helper for mocking exec.Command.
// It's not a real test - it's called by MockExecutor.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) > 0 {
		// Output the script content as stdout
		os.Stdout.WriteString(args[0])
	}

	exitCode := 0
	if code := os.Getenv("MOCK_EXIT_CODE"); code != "" {
		if parsed, err := strconv.Atoi(code); err == nil {
			exitCode = parsed
		}
	}
	os.Exit(exitCode)
}

func TestClient_NewClient(t *testing.T) {
	t.Run("with nil options uses defaults", func(t *testing.T) {
		client := NewClient(nil)
		assert.NotNil(t, client)
		assert.Equal(t, "claude", client.opts.ClaudePath)
		assert.NotEmpty(t, client.opts.AdditionalFlags)
	})

	t.Run("with custom options", func(t *testing.T) {
		opts := &ClientOptions{
			ClaudePath:      "/custom/claude",
			AdditionalFlags: []string{"--flag"},
		}
		client := NewClient(opts)
		assert.Equal(t, "/custom/claude", client.opts.ClaudePath)
		assert.Equal(t, []string{"--flag"}, client.opts.AdditionalFlags)
	})

	t.Run("fills in defaults for empty fields", func(t *testing.T) {
		opts := &ClientOptions{}
		client := NewClient(opts)
		assert.Equal(t, "claude", client.opts.ClaudePath)
		assert.NotNil(t, client.opts.Executor)
		assert.NotEmpty(t, client.opts.AdditionalFlags)
	})
}

func TestClient_DefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	assert.Equal(t, "claude", opts.ClaudePath)
	assert.Contains(t, opts.AdditionalFlags, "--dangerously-skip-permissions")
	assert.Contains(t, opts.AdditionalFlags, "--output-format")
	assert.Contains(t, opts.AdditionalFlags, "stream-json")
	assert.Contains(t, opts.AdditionalFlags, "--verbose")
	assert.NotNil(t, opts.Executor)
}

func TestClient_Execute_Success(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]}}
{"type":"result","result":"Done","total_cost_usd":0.05,"is_error":false}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		ClaudePath: "echo", // Will be overridden by MockExecutor
		Executor:   mockExec,
	})

	result, err := client.Execute(context.Background(), "test prompt")

	require.NoError(t, err)
	assert.Equal(t, "Hello!", result.Output)
	assert.InDelta(t, 0.05, result.Cost, 0.001)
	assert.False(t, result.CompletionSignalFound)
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestClient_Execute_Error(t *testing.T) {
	output := `{"type":"result","result":"API error occurred","total_cost_usd":0.01,"is_error":true}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 1}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.Execute(context.Background(), "test prompt")

	assert.Nil(t, result)
	require.Error(t, err)

	var claudeErr *ClaudeError
	assert.ErrorAs(t, err, &claudeErr)
	assert.Equal(t, "API error occurred", claudeErr.ResultText)
}

func TestClient_Execute_IsErrorOnExitZero(t *testing.T) {
	// Sometimes claude returns exit 0 but is_error=true
	output := `{"type":"result","result":"Soft error","total_cost_usd":0.01,"is_error":true}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.Execute(context.Background(), "test prompt")

	assert.Nil(t, result)
	require.Error(t, err)

	var claudeErr *ClaudeError
	assert.ErrorAs(t, err, &claudeErr)
	assert.Equal(t, "Soft error", claudeErr.ResultText)
}

func TestClient_Execute_ContextCancellation(t *testing.T) {
	// This test uses a real command that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	client := NewClient(&ClientOptions{
		ClaudePath: "sleep",
		AdditionalFlags: []string{}, // Clear flags for sleep command
	})
	client.opts.Executor = &DefaultExecutor{}

	_, err := client.Execute(ctx, "10") // sleep 10 seconds

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestClient_Execute_WithStreamHandler(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Part1"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Part2"}]}}
{"type":"result","result":"Done","total_cost_usd":0.05,"is_error":false}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	handler := &mockStreamHandler{}
	client := NewClient(&ClientOptions{
		Executor:      mockExec,
		StreamHandler: handler,
	})

	result, err := client.Execute(context.Background(), "test")

	require.NoError(t, err)
	assert.Equal(t, "Part1Part2", result.Output)
	assert.Equal(t, []string{"Part1", "Part2"}, handler.texts)
}

func TestClient_Execute_NoCost(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","result":"Done","is_error":false}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.Execute(context.Background(), "test")

	require.NoError(t, err)
	assert.Zero(t, result.Cost)
}

func TestDefaultExecutor_CommandContext(t *testing.T) {
	executor := &DefaultExecutor{}
	ctx := context.Background()

	cmd := executor.CommandContext(ctx, "echo", "hello")

	assert.NotNil(t, cmd)
	// cmd.Path may be the full path (e.g., "/bin/echo") depending on the system
	assert.Contains(t, cmd.Path, "echo")
}

// Compile-time interface check
func TestClient_ImplementsClaudeClient(t *testing.T) {
	// This test just ensures the file compiles with the interface check
	// The actual check is: var _ loop.ClaudeClient = (*Client)(nil)
	_ = NewClient(nil)
}

func TestClient_ExecuteWithSession_Success(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello!"}]}}
{"type":"result","result":"Done","total_cost_usd":0.05,"is_error":false,"session_id":"test-session-123"}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		ClaudePath: "echo",
		Executor:   mockExec,
	})

	result, err := client.ExecuteWithSession(context.Background(), "test prompt", "")

	require.NoError(t, err)
	assert.Equal(t, "Hello!", result.Output)
	assert.Equal(t, "test-session-123", result.SessionID)
	assert.InDelta(t, 0.05, result.Cost, 0.001)
}

func TestClient_ExecuteWithSession_WithResume(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Resumed session!"}]}}
{"type":"result","result":"Done","total_cost_usd":0.02,"is_error":false,"session_id":"test-session-123"}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.ExecuteWithSession(context.Background(), "user answer", "test-session-123")

	require.NoError(t, err)
	assert.Equal(t, "Resumed session!", result.Output)
	assert.Equal(t, "test-session-123", result.SessionID)
}

func TestClient_ExecuteWithSession_Error(t *testing.T) {
	output := `{"type":"result","result":"API error occurred","total_cost_usd":0.01,"is_error":true}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 1}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.ExecuteWithSession(context.Background(), "test prompt", "")

	assert.Nil(t, result)
	require.Error(t, err)

	var claudeErr *ClaudeError
	assert.ErrorAs(t, err, &claudeErr)
	assert.Equal(t, "API error occurred", claudeErr.ResultText)
}

func TestClient_ExecuteWithSession_IsErrorOnExitZero(t *testing.T) {
	output := `{"type":"result","result":"Soft error","total_cost_usd":0.01,"is_error":true}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.ExecuteWithSession(context.Background(), "test prompt", "")

	assert.Nil(t, result)
	require.Error(t, err)

	var claudeErr *ClaudeError
	assert.ErrorAs(t, err, &claudeErr)
	assert.Equal(t, "Soft error", claudeErr.ResultText)
}

func TestClient_ExecuteWithSession_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := NewClient(&ClientOptions{
		ClaudePath:      "sleep",
		AdditionalFlags: []string{},
	})
	client.opts.Executor = &DefaultExecutor{}

	_, err := client.ExecuteWithSession(ctx, "10", "") // sleep 10 seconds

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestClient_ExecuteWithSession_NoSessionID(t *testing.T) {
	output := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","result":"Done","is_error":false}
`
	mockExec := &MockExecutor{Script: output, ExitCode: 0}

	client := NewClient(&ClientOptions{
		Executor: mockExec,
	})

	result, err := client.ExecuteWithSession(context.Background(), "test", "")

	require.NoError(t, err)
	assert.Empty(t, result.SessionID)
}
