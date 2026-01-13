package reviewer

import (
	"github.com/DeukWoongWoo/claude-loop/internal/prompt"
)

// PromptBuilder builds prompts for reviewer passes.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildContext contains inputs for building a reviewer prompt.
type BuildContext struct {
	UserReviewPrompt string // User's review instructions from -r flag
}

// BuildResult contains the built prompt.
type BuildResult struct {
	Prompt string
}

// Build constructs a reviewer prompt by combining the reviewer context template
// with the user's review instructions.
func (b *PromptBuilder) Build(ctx BuildContext) (*BuildResult, error) {
	if ctx.UserReviewPrompt == "" {
		return nil, ErrNoReviewPrompt
	}

	fullPrompt := prompt.TemplateReviewerContext +
		"\n\n## USER REVIEW INSTRUCTIONS\n\n" +
		ctx.UserReviewPrompt

	return &BuildResult{Prompt: fullPrompt}, nil
}
