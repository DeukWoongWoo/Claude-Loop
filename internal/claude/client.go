package claude

import (
	"bytes"
	"context"
	"os"
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

// Execute implements loop.ClaudeClient interface.
// It runs the claude CLI with the given prompt and returns the result.
func (c *Client) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	startTime := time.Now()

	// Build command arguments
	args := []string{"-p", prompt}
	args = append(args, c.opts.AdditionalFlags...)

	// Create command
	cmd := c.opts.Executor.CommandContext(ctx, c.opts.ClaudePath, args...)

	// Get stdout pipe for streaming
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, &ClaudeError{Message: "failed to create stdout pipe", Err: err}
	}

	// Capture stderr for error context
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, &ClaudeError{Message: "failed to start claude", Err: err}
	}

	// Parse output stream
	parsed, parseErr := c.parser.Parse(stdout)

	// Wait for command to complete
	cmdErr := cmd.Wait()

	duration := time.Since(startTime)

	// Handle parse error
	if parseErr != nil {
		return nil, &ClaudeError{
			Message: "failed to parse output",
			Err:     parseErr,
			Stderr:  stderrBuf.String(),
		}
	}

	// Handle command error
	if cmdErr != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		// Check for is_error in JSON response
		if parsed != nil && parsed.IsError {
			return nil, &ClaudeError{
				Message:    "claude returned error",
				ResultText: parsed.ResultText,
				Stderr:     stderrBuf.String(),
			}
		}

		return nil, &ClaudeError{
			Message: "claude exited with error",
			Err:     cmdErr,
			Stderr:  stderrBuf.String(),
		}
	}

	// Check is_error flag even on exit code 0
	if parsed.IsError {
		return nil, &ClaudeError{
			Message:    "claude returned error",
			ResultText: parsed.ResultText,
			Stderr:     stderrBuf.String(),
		}
	}

	return &loop.IterationResult{
		Output:                parsed.Output,
		Cost:                  parsed.TotalCostUSD,
		Duration:              duration,
		CompletionSignalFound: false, // Detected by loop package
	}, nil
}

// ExecuteInteractive runs claude in interactive mode.
// This connects stdin/stdout/stderr directly to the terminal, allowing
// Claude to use AskUserQuestion for interactive prompts.
// Uses --print flag to send an initial prompt while keeping the session interactive.
func (c *Client) ExecuteInteractive(ctx context.Context, prompt string) error {
	args := []string{"--print", prompt}

	// Add compatible flags from AdditionalFlags (skip --output-format which breaks interactive)
	args = append(args, filterInteractiveFlags(c.opts.AdditionalFlags)...)

	// Create command
	cmd := c.opts.Executor.CommandContext(ctx, c.opts.ClaudePath, args...)

	// Connect to terminal directly for interactive mode
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run and wait for completion
	if err := cmd.Run(); err != nil {
		// Check if it's a context cancellation
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return &ClaudeError{Message: "interactive claude execution failed", Err: err}
	}

	return nil
}

// filterInteractiveFlags removes flags incompatible with interactive mode.
// In particular, --output-format breaks terminal interaction.
func filterInteractiveFlags(flags []string) []string {
	var result []string
	skipNext := false
	for _, flag := range flags {
		if skipNext {
			skipNext = false
			continue
		}
		// Skip --output-format and its value
		if flag == "--output-format" {
			skipNext = true
			continue
		}
		// Skip --output-format=value style
		if len(flag) > 15 && flag[:15] == "--output-format" {
			continue
		}
		result = append(result, flag)
	}
	return result
}

// Verify interface compliance at compile time.
var _ loop.ClaudeClient = (*Client)(nil)
