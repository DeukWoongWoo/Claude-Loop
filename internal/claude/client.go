package claude

import (
	"bytes"
	"context"
	"os/exec"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
)

// CommandExecutor abstracts exec.Command for testing.
type CommandExecutor interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// DefaultExecutor uses the real exec.CommandContext.
type DefaultExecutor struct{}

// CommandContext creates a new exec.Cmd with the given context.
func (e *DefaultExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// ClientOptions configures the Claude client.
type ClientOptions struct {
	// ClaudePath is the path to the claude CLI binary (default: "claude")
	ClaudePath string

	// AdditionalFlags are extra flags to pass to claude CLI.
	// Default: ["--dangerously-skip-permissions", "--output-format", "stream-json", "--verbose"]
	AdditionalFlags []string

	// StreamHandler receives real-time text output (optional).
	StreamHandler StreamHandler

	// Executor for command creation (for testing).
	Executor CommandExecutor
}

// DefaultOptions returns ClientOptions with default values.
func DefaultOptions() *ClientOptions {
	return &ClientOptions{
		ClaudePath: "claude",
		AdditionalFlags: []string{
			"--dangerously-skip-permissions",
			"--output-format", "stream-json",
			"--verbose",
			"--include-partial-messages",
		},
		Executor: &DefaultExecutor{},
	}
}

// Client implements loop.ClaudeClient by executing the claude CLI.
type Client struct {
	opts   *ClientOptions
	parser *Parser
}

// NewClient creates a new Client with the given options.
// If opts is nil, default options are used.
func NewClient(opts *ClientOptions) *Client {
	if opts == nil {
		opts = DefaultOptions()
	}

	defaults := DefaultOptions()
	if opts.ClaudePath == "" {
		opts.ClaudePath = defaults.ClaudePath
	}
	if opts.Executor == nil {
		opts.Executor = defaults.Executor
	}
	if opts.AdditionalFlags == nil {
		opts.AdditionalFlags = defaults.AdditionalFlags
	}

	return &Client{
		opts:   opts,
		parser: NewParser(opts.StreamHandler),
	}
}

// execResult holds the result of running a command.
type execResult struct {
	parsed *ParsedResult
	stderr string
}

// runCommand executes the claude CLI with the given arguments and returns the parsed result.
// This is the common execution logic shared by Execute and ExecuteWithSession.
func (c *Client) runCommand(ctx context.Context, args []string) (*execResult, error) {
	cmd := c.opts.Executor.CommandContext(ctx, c.opts.ClaudePath, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, &ClaudeError{Message: "failed to create stdout pipe", Err: err}
	}

	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	if err := cmd.Start(); err != nil {
		return nil, &ClaudeError{Message: "failed to start claude", Err: err}
	}

	parsed, parseErr := c.parser.Parse(stdout)
	cmdErr := cmd.Wait()
	stderr := stderrBuf.String()

	if parseErr != nil {
		return nil, &ClaudeError{
			Message: "failed to parse output",
			Err:     parseErr,
			Stderr:  stderr,
		}
	}

	if err := c.checkExecutionError(ctx, cmdErr, parsed, stderr); err != nil {
		return nil, err
	}

	return &execResult{parsed: parsed, stderr: stderr}, nil
}

// checkExecutionError checks for command and result errors.
// Returns nil if execution was successful, or an appropriate error otherwise.
func (c *Client) checkExecutionError(ctx context.Context, cmdErr error, parsed *ParsedResult, stderr string) error {
	if cmdErr != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if parsed != nil && parsed.IsError {
			return &ClaudeError{
				Message:    "claude returned error",
				ResultText: parsed.ResultText,
				Stderr:     stderr,
			}
		}
		return &ClaudeError{
			Message: "claude exited with error",
			Err:     cmdErr,
			Stderr:  stderr,
		}
	}

	if parsed.IsError {
		return &ClaudeError{
			Message:    "claude returned error",
			ResultText: parsed.ResultText,
			Stderr:     stderr,
		}
	}

	return nil
}

// Execute implements loop.ClaudeClient interface.
// It runs the claude CLI with the given prompt and returns the result.
func (c *Client) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	startTime := time.Now()

	args := []string{"-p", prompt}
	args = append(args, c.opts.AdditionalFlags...)

	result, err := c.runCommand(ctx, args)
	if err != nil {
		return nil, err
	}

	return &loop.IterationResult{
		Output:                result.parsed.Output,
		Cost:                  result.parsed.TotalCostUSD,
		Duration:              time.Since(startTime),
		CompletionSignalFound: false, // Detected by loop package
	}, nil
}

// ExecuteWithSession executes a prompt with optional session resume.
// Used for multi-turn interactions like principles collection.
// If sessionID is empty, starts a new session. Otherwise resumes the specified session.
func (c *Client) ExecuteWithSession(ctx context.Context, prompt string, sessionID string) (*SessionResult, error) {
	args := []string{"-p", prompt}
	if sessionID != "" {
		args = append(args, "--resume", sessionID)
	}
	args = append(args, c.opts.AdditionalFlags...)

	result, err := c.runCommand(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SessionResult{
		Output:    result.parsed.Output,
		SessionID: result.parsed.SessionID,
		Cost:      result.parsed.TotalCostUSD,
	}, nil
}

// Verify interface compliance at compile time.
var _ loop.ClaudeClient = (*Client)(nil)
