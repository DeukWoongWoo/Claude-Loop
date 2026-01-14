package council

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
)

// TemplateCouncilResolution is the prompt template for council resolution.
// Matches bash implementation (claude_loop.sh lines 1216-1232).
const TemplateCouncilResolution = `A principle conflict was detected that requires Council resolution.

## Conflict Context
%s

## Current Principles
` + "```yaml" + `
%s
` + "```" + `

## Instructions
Please analyze this conflict using the R10 3-step resolution protocol:
1. Compatibility Check: Are conflicting principles about different dimensions?
2. Type Classification: Is this Constraint vs Objective?
3. Priority Resolution: Apply Layer 3 hierarchy if needed.

Provide a clear recommendation with rationale.

## Response Format
**Decision**: <your recommendation>
**Rationale**: <which principle(s) applied and why>`

// PromptBuilder builds prompts for council resolution.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildContext contains inputs for building a council prompt.
type BuildContext struct {
	ConflictContext string             // The conflict context from Claude's output
	Principles      *config.Principles // Current principles
}

// BuildResult contains the built prompt.
type BuildResult struct {
	Prompt string
}

// Build constructs a council resolution prompt.
func (b *PromptBuilder) Build(ctx BuildContext) (*BuildResult, error) {
	if ctx.Principles == nil {
		return nil, ErrNoPrinciples
	}

	principlesYAML, err := yaml.Marshal(ctx.Principles)
	if err != nil {
		return nil, &CouncilError{
			Phase:   "prompt",
			Message: "failed to marshal principles",
			Err:     err,
		}
	}

	prompt := fmt.Sprintf(TemplateCouncilResolution, ctx.ConflictContext, string(principlesYAML))

	return &BuildResult{Prompt: prompt}, nil
}
