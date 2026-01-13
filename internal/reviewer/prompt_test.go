package reviewer

import (
	"strings"
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptBuilder(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()
	assert.NotNil(t, builder)
}

func TestPromptBuilder_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		ctx           BuildContext
		expectError   bool
		errorContains string
		checkPrompt   func(t *testing.T, result *BuildResult)
	}{
		{
			name: "valid review prompt",
			ctx: BuildContext{
				UserReviewPrompt: "Run npm test and fix any failures",
			},
			expectError: false,
			checkPrompt: func(t *testing.T, result *BuildResult) {
				assert.Contains(t, result.Prompt, prompt.TemplateReviewerContext)
				assert.Contains(t, result.Prompt, "## USER REVIEW INSTRUCTIONS")
				assert.Contains(t, result.Prompt, "Run npm test and fix any failures")
			},
		},
		{
			name: "empty review prompt",
			ctx: BuildContext{
				UserReviewPrompt: "",
			},
			expectError:   true,
			errorContains: "no review prompt provided",
		},
		{
			name: "complex review prompt",
			ctx: BuildContext{
				UserReviewPrompt: `1. Run npm test
2. Run npm run lint
3. Check for any TypeScript errors
4. Fix any issues found`,
			},
			expectError: false,
			checkPrompt: func(t *testing.T, result *BuildResult) {
				assert.Contains(t, result.Prompt, "1. Run npm test")
				assert.Contains(t, result.Prompt, "4. Fix any issues found")
			},
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			builder := NewPromptBuilder()
			result, err := builder.Build(tt.ctx)

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.NotEmpty(t, result.Prompt)
				if tt.checkPrompt != nil {
					tt.checkPrompt(t, result)
				}
			}
		})
	}
}

func TestPromptBuilder_Build_PromptFormat(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()
	result, err := builder.Build(BuildContext{
		UserReviewPrompt: "test instructions",
	})

	require.NoError(t, err)

	// Verify the prompt structure
	parts := strings.Split(result.Prompt, "\n\n## USER REVIEW INSTRUCTIONS\n\n")
	require.Len(t, parts, 2)

	// First part should be the reviewer context
	assert.Equal(t, prompt.TemplateReviewerContext, parts[0])

	// Second part should be the user's review instructions
	assert.Equal(t, "test instructions", parts[1])
}

func TestPromptBuilder_Build_ContainsReviewerContext(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()
	result, err := builder.Build(BuildContext{
		UserReviewPrompt: "any prompt",
	})

	require.NoError(t, err)

	// Check for key phrases from TemplateReviewerContext
	assert.Contains(t, result.Prompt, "CODE REVIEW CONTEXT")
	assert.Contains(t, result.Prompt, "review pass")
	assert.Contains(t, result.Prompt, "git commands")
}
