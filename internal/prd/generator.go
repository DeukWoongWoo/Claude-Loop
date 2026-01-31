package prd

import (
	"context"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// DefaultGenerator implements the Generator interface.
type DefaultGenerator struct {
	config        *Config
	client        ClaudeClient
	promptBuilder *planner.PromptBuilder
	parser        *Parser
	validator     *Validator
}

// NewGenerator creates a new DefaultGenerator.
// If config is nil, DefaultConfig() is used.
// Returns nil if client is nil.
func NewGenerator(config *Config, client ClaudeClient) *DefaultGenerator {
	if client == nil {
		return nil
	}
	if config == nil {
		config = DefaultConfig()
	}
	return &DefaultGenerator{
		config:        config,
		client:        client,
		promptBuilder: planner.NewPromptBuilder(),
		parser:        NewParser(),
		validator:     NewValidator(),
	}
}

// Generate creates a PRD from the given user prompt.
func (g *DefaultGenerator) Generate(ctx context.Context, userPrompt string) (*PRD, error) {
	if userPrompt == "" {
		return nil, ErrEmptyPrompt
	}

	startTime := time.Now()
	prompt := g.promptBuilder.BuildPRDPrompt(userPrompt)

	result, err := g.client.Execute(ctx, prompt)
	if err != nil {
		return nil, &PRDError{
			Phase:   "generate",
			Message: "claude execution failed",
			Err:     err,
		}
	}

	prd, err := g.parser.Parse(result.Output)
	if err != nil {
		return nil, err
	}

	// Add metadata
	prd.CreatedAt = startTime
	prd.Cost = result.Cost
	prd.Duration = time.Since(startTime)

	if g.config.ValidateOutput {
		if err := g.validator.Validate(prd); err != nil {
			return nil, &PRDError{
				Phase:   "validate",
				Message: "generated PRD validation failed",
				Err:     err,
			}
		}
	}

	return prd, nil
}

// Config returns the generator's configuration.
func (g *DefaultGenerator) Config() *Config {
	return g.config
}

// Compile-time interface compliance check.
var _ Generator = (*DefaultGenerator)(nil)
