package planner

import (
	"context"

	"github.com/DeukWoongWoo/claude-loop/internal/loop"
)

// ClaudeClientAdapter adapts loop.ClaudeClient to planner.ClaudeClient interface.
// This enables using claude.Client (which implements loop.ClaudeClient) with planner packages.
type ClaudeClientAdapter struct {
	client loop.ClaudeClient
}

// NewClaudeClientAdapter creates a new adapter that wraps a loop.ClaudeClient.
// The adapter converts between loop.IterationResult and planner.IterationResult.
func NewClaudeClientAdapter(client loop.ClaudeClient) *ClaudeClientAdapter {
	if client == nil {
		return nil
	}
	return &ClaudeClientAdapter{client: client}
}

// Execute implements planner.ClaudeClient interface.
// It delegates to the wrapped loop.ClaudeClient and converts the result type.
func (a *ClaudeClientAdapter) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	if a.client == nil {
		return nil, &PlannerError{
			Phase:   "adapter",
			Message: "claude client is nil",
		}
	}

	result, err := a.client.Execute(ctx, prompt)
	if err != nil {
		return nil, err
	}

	return &IterationResult{
		Output:                result.Output,
		Cost:                  result.Cost,
		Duration:              result.Duration,
		CompletionSignalFound: result.CompletionSignalFound,
	}, nil
}

// Compile-time interface compliance check.
var _ ClaudeClient = (*ClaudeClientAdapter)(nil)
