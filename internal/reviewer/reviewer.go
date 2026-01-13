package reviewer

import (
	"context"
)

// DefaultReviewer implements the Reviewer interface.
type DefaultReviewer struct {
	config        *Config
	client        ClaudeClient
	promptBuilder *PromptBuilder
}

// NewReviewer creates a new DefaultReviewer with the given configuration.
// If config is nil, DefaultConfig() is used.
func NewReviewer(config *Config, client ClaudeClient) *DefaultReviewer {
	if config == nil {
		config = DefaultConfig()
	}
	return &DefaultReviewer{
		config:        config,
		client:        client,
		promptBuilder: NewPromptBuilder(),
	}
}

// Run executes a reviewer pass by building a prompt and executing Claude.
func (r *DefaultReviewer) Run(ctx context.Context) (*Result, error) {
	buildResult, err := r.promptBuilder.Build(BuildContext{
		UserReviewPrompt: r.config.ReviewPrompt,
	})
	if err != nil {
		// ErrNoReviewPrompt is already a ReviewerError, wrap others
		if IsReviewerError(err) {
			return nil, err
		}
		return nil, &ReviewerError{Phase: "prompt", Message: "failed to build prompt", Err: err}
	}

	iterResult, err := r.client.Execute(ctx, buildResult.Prompt)
	if err != nil {
		return nil, &ReviewerError{Phase: "execute", Message: "claude execution failed", Err: err}
	}

	return &Result{
		Output:                iterResult.Output,
		Cost:                  iterResult.Cost,
		Duration:              iterResult.Duration,
		CompletionSignalFound: iterResult.CompletionSignalFound,
	}, nil
}

// Config returns the reviewer's configuration.
func (r *DefaultReviewer) Config() *Config {
	return r.config
}
