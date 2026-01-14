// Package mocks provides mock implementations for integration testing.
package mocks

import (
	"context"
	"os/exec"
	"strconv"
	"sync"
)

// CommandMock represents a mock command response.
type CommandMock struct {
	ExpectedName string
	ExpectedArgs []string
	Stdout       string
	Stderr       string
	ExitCode     int
}

// CommandCall records a command invocation.
type CommandCall struct {
	Name string
	Args []string
}

// CommandSequence returns mock commands in order.
type CommandSequence struct {
	mu       sync.Mutex
	Commands []CommandMock
	index    int
	calls    []CommandCall
}

// NewCommandSequence creates a sequence with the given mocks.
func NewCommandSequence(commands ...CommandMock) *CommandSequence {
	return &CommandSequence{Commands: commands}
}

// CommandContext implements github.CommandExecutor.
func (s *CommandSequence) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, CommandCall{Name: name, Args: args})

	if s.index >= len(s.Commands) {
		return exec.CommandContext(ctx, "false")
	}

	mock := s.Commands[s.index]
	s.index++

	if mock.ExitCode == 0 {
		return exec.CommandContext(ctx, "echo", "-n", mock.Stdout)
	}

	script := "echo -n '" + mock.Stderr + "' >&2; exit " + strconv.Itoa(mock.ExitCode)
	return exec.CommandContext(ctx, "sh", "-c", script)
}

// Reset clears recorded calls and resets the command index.
func (s *CommandSequence) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.index = 0
	s.calls = nil
}

// AllCommandsUsed returns true if all commands have been consumed.
func (s *CommandSequence) AllCommandsUsed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.index == len(s.Commands)
}

// CallCount returns the number of commands invoked.
func (s *CommandSequence) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

// GetCall returns the nth call (0-indexed), or nil if out of range.
func (s *CommandSequence) GetCall(n int) *CommandCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	if n < 0 || n >= len(s.calls) {
		return nil
	}
	call := s.calls[n]
	return &call
}

// SuccessCommand creates a mock that succeeds with the given output.
func SuccessCommand(stdout string) CommandMock {
	return CommandMock{Stdout: stdout}
}

// FailCommand creates a mock that fails with the given stderr and exit code.
func FailCommand(stderr string, exitCode int) CommandMock {
	return CommandMock{Stderr: stderr, ExitCode: exitCode}
}

// EmptySuccessCommand creates a mock that succeeds with empty output.
func EmptySuccessCommand() CommandMock {
	return CommandMock{}
}
