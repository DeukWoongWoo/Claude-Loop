package council

import (
	"context"
	"time"
)

// DefaultCouncil implements the Council interface.
type DefaultCouncil struct {
	config        *Config
	client        ClaudeClient
	detector      *ConflictDetector
	promptBuilder *PromptBuilder
	logger        *DecisionLogger
}

// NewCouncil creates a new DefaultCouncil.
func NewCouncil(cfg *Config, client ClaudeClient) *DefaultCouncil {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &DefaultCouncil{
		config:        cfg,
		client:        client,
		detector:      NewConflictDetector(),
		promptBuilder: NewPromptBuilder(),
		logger:        NewDecisionLogger(cfg.LogFile, cfg.LogDecisions),
	}
}

// DetectConflict checks if output contains unresolved principle conflicts.
func (c *DefaultCouncil) DetectConflict(output string) bool {
	return c.detector.Detect(output)
}

// Resolve invokes the LLM council for conflict resolution.
func (c *DefaultCouncil) Resolve(ctx context.Context, conflictContext string) (*Result, error) {
	if c.config.Principles == nil {
		return nil, ErrNoPrinciples
	}

	buildResult, err := c.promptBuilder.Build(BuildContext{
		ConflictContext: conflictContext,
		Principles:      c.config.Principles,
	})
	if err != nil {
		return nil, err
	}

	startTime := time.Now()
	iterResult, err := c.client.Execute(ctx, buildResult.Prompt)
	if err != nil {
		return nil, &CouncilError{
			Phase:   "resolve",
			Message: "council invocation failed",
			Err:     err,
		}
	}

	// Extract decision and rationale from council output
	decision, rationale := c.detector.ExtractDecision(iterResult.Output)

	return &Result{
		Output:     iterResult.Output,
		Cost:       iterResult.Cost,
		Duration:   time.Since(startTime),
		Resolution: decision,
		Rationale:  rationale,
	}, nil
}

// LogDecision logs a decision to the decision log file.
func (c *DefaultCouncil) LogDecision(decision *Decision) error {
	return c.logger.Log(decision)
}

// Config returns the council's configuration.
func (c *DefaultCouncil) Config() *Config {
	return c.config
}

// ExtractDecisionFromOutput extracts decision info from output for logging.
func (c *DefaultCouncil) ExtractDecisionFromOutput(output string) (decision, rationale string) {
	return c.detector.ExtractDecision(output)
}
