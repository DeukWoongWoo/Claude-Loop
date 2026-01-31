package architecture

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

// Generate creates an Architecture from the given PRD.
func (g *DefaultGenerator) Generate(ctx context.Context, prd *planner.PRD) (*Architecture, error) {
	if prd == nil {
		return nil, ErrNilPRD
	}

	startTime := time.Now()

	prompt, err := g.promptBuilder.BuildArchitecturePrompt(prd)
	if err != nil {
		return nil, &ArchitectureError{
			Phase:   "generate",
			Message: "failed to build prompt",
			Err:     err,
		}
	}

	result, err := g.client.Execute(ctx, prompt)
	if err != nil {
		return nil, &ArchitectureError{
			Phase:   "generate",
			Message: "claude execution failed",
			Err:     err,
		}
	}

	arch, err := g.parser.Parse(result.Output)
	if err != nil {
		return nil, err
	}

	// Add metadata
	arch.CreatedAt = startTime
	arch.Cost = result.Cost
	arch.Duration = time.Since(startTime)

	if g.config.ValidateOutput {
		if err := g.validator.Validate(arch); err != nil {
			return nil, &ArchitectureError{
				Phase:   "validate",
				Message: "generated architecture validation failed",
				Err:     err,
			}
		}
	}

	return arch, nil
}

// Config returns the generator's configuration.
func (g *DefaultGenerator) Config() *Config {
	return g.config
}

// Compile-time interface compliance check.
var _ Generator = (*DefaultGenerator)(nil)
