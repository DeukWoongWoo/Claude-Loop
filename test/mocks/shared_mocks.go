// Package mocks provides mock implementations for integration testing.
package mocks

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
)

// ClientResponse represents a response from the mock Claude client.
type ClientResponse struct {
	Output    string
	Cost      float64
	Duration  time.Duration
	Err       error
	HasSignal bool
}

// RecordedCall records a single call to the mock client.
type RecordedCall struct {
	Prompt    string
	Timestamp time.Time
}

// ConfigurableClaudeClient is a mock that returns configured responses in sequence.
type ConfigurableClaudeClient struct {
	mu            sync.Mutex
	Responses     []ClientResponse
	CurrentIndex  int
	RecordedCalls []RecordedCall
	DefaultCost   float64
	DefaultDelay  time.Duration
}

// NewConfigurableClaudeClient creates a client with default cost and delay.
func NewConfigurableClaudeClient() *ConfigurableClaudeClient {
	return &ConfigurableClaudeClient{
		DefaultCost:  0.01,
		DefaultDelay: 100 * time.Millisecond,
	}
}

// Execute implements loop.ClaudeClient.
func (c *ConfigurableClaudeClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RecordedCalls = append(c.RecordedCalls, RecordedCall{
		Prompt:    prompt,
		Timestamp: time.Now(),
	})

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	if c.CurrentIndex >= len(c.Responses) {
		return &loop.IterationResult{
			Output:   "default output",
			Cost:     c.DefaultCost,
			Duration: c.DefaultDelay,
		}, nil
	}

	resp := c.Responses[c.CurrentIndex]
	c.CurrentIndex++

	if resp.Err != nil {
		return nil, resp.Err
	}

	return &loop.IterationResult{
		Output:                resp.Output,
		Cost:                  resp.Cost,
		Duration:              resp.Duration,
		CompletionSignalFound: resp.HasSignal,
	}, nil
}

// CallCount returns the number of Execute calls.
func (c *ConfigurableClaudeClient) CallCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.RecordedCalls)
}

// Reset clears recorded calls and resets the response index.
func (c *ConfigurableClaudeClient) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.CurrentIndex = 0
	c.RecordedCalls = nil
}

// LastPrompt returns the most recent prompt, or empty string if none.
func (c *ConfigurableClaudeClient) LastPrompt() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.RecordedCalls) == 0 {
		return ""
	}
	return c.RecordedCalls[len(c.RecordedCalls)-1].Prompt
}

// AllPromptsContain checks if all recorded prompts contain the given substring.
func (c *ConfigurableClaudeClient) AllPromptsContain(substr string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, call := range c.RecordedCalls {
		if !strings.Contains(call.Prompt, substr) {
			return false
		}
	}
	return true
}

var _ loop.ClaudeClient = (*ConfigurableClaudeClient)(nil)

// SequentialClaudeClient returns predetermined responses in sequence.
type SequentialClaudeClient struct {
	mu        sync.Mutex
	sequences []ClientResponse
	index     int
	calls     []RecordedCall
}

// NewSequentialClaudeClient creates a client with the given responses.
func NewSequentialClaudeClient(responses ...ClientResponse) *SequentialClaudeClient {
	return &SequentialClaudeClient{sequences: responses}
}

// Execute implements loop.ClaudeClient.
func (s *SequentialClaudeClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.calls = append(s.calls, RecordedCall{Prompt: prompt, Timestamp: time.Now()})

	if s.index >= len(s.sequences) {
		return &loop.IterationResult{Output: "exhausted", Cost: 0.01}, nil
	}

	resp := s.sequences[s.index]
	s.index++

	if resp.Err != nil {
		return nil, resp.Err
	}

	return &loop.IterationResult{
		Output:                resp.Output,
		Cost:                  resp.Cost,
		Duration:              resp.Duration,
		CompletionSignalFound: resp.HasSignal,
	}, nil
}

// CallCount returns the number of Execute calls.
func (s *SequentialClaudeClient) CallCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.calls)
}

var _ loop.ClaudeClient = (*SequentialClaudeClient)(nil)

// ErroringClaudeClient always returns the configured error.
type ErroringClaudeClient struct {
	Err       error
	CallCount int
}

// Execute implements loop.ClaudeClient.
func (e *ErroringClaudeClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	e.CallCount++
	return nil, e.Err
}

var _ loop.ClaudeClient = (*ErroringClaudeClient)(nil)
