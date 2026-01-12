package prompt

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCIFixBuilder_Build(t *testing.T) {
	builder := NewCIFixBuilder()

	t.Run("builds prompt with all fields", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: &CIFailureInfo{
				RunID:        "12345",
				WorkflowName: "CI",
				JobName:      "build",
				FailedSteps:  []string{"Build", "Test"},
				ErrorLogs:    "Error: compilation failed",
				URL:          "https://github.com/owner/repo/actions/runs/12345",
			},
			PRNumber:    42,
			BranchName:  "feature/test",
			Attempt:     1,
			MaxAttempts: 3,
		}

		result, err := builder.Build(ctx)
		require.NoError(t, err)

		// Check that CI fix template is included
		assert.Contains(t, result.Prompt, "CI FAILURE FIX CONTEXT")
		assert.Contains(t, result.Prompt, "gh run list --status failure")

		// Check failure details
		assert.Contains(t, result.Prompt, "## Failure Details")
		assert.Contains(t, result.Prompt, "12345")
		assert.Contains(t, result.Prompt, "CI")
		assert.Contains(t, result.Prompt, "build")
		assert.Contains(t, result.Prompt, "Build, Test")
		assert.Contains(t, result.Prompt, "https://github.com/owner/repo/actions/runs/12345")

		// Check error logs
		assert.Contains(t, result.Prompt, "## Error Logs")
		assert.Contains(t, result.Prompt, "compilation failed")

		// Check instructions
		assert.Contains(t, result.Prompt, "## Instructions")
		assert.Contains(t, result.Prompt, "gh run view 12345 --log")

		// First attempt should not have retry note
		assert.NotContains(t, result.Prompt, "This is attempt")
	})

	t.Run("includes retry note for subsequent attempts", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: &CIFailureInfo{
				RunID:        "12345",
				WorkflowName: "CI",
				ErrorLogs:    "Error",
			},
			Attempt:     2,
			MaxAttempts: 3,
		}

		result, err := builder.Build(ctx)
		require.NoError(t, err)

		assert.Contains(t, result.Prompt, "This is attempt 2 of 3")
		assert.Contains(t, result.Prompt, "Previous fix attempts did not resolve the issue")
	})

	t.Run("handles missing optional fields", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: &CIFailureInfo{
				RunID:        "12345",
				WorkflowName: "CI",
				ErrorLogs:    "Error",
				// JobName, FailedSteps, URL are empty
			},
			Attempt:     1,
			MaxAttempts: 1,
		}

		result, err := builder.Build(ctx)
		require.NoError(t, err)

		// Should not contain empty fields
		assert.NotContains(t, result.Prompt, "Failed Job:")
		assert.NotContains(t, result.Prompt, "Failed Steps:")
		assert.NotContains(t, result.Prompt, "Run URL:")
	})

	t.Run("error when FailureInfo is nil", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: nil,
			Attempt:     1,
			MaxAttempts: 1,
		}

		_, err := builder.Build(ctx)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "FailureInfo is required")
	})

	t.Run("handles empty error logs", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: &CIFailureInfo{
				RunID:        "12345",
				WorkflowName: "CI",
				ErrorLogs:    "",
			},
			Attempt:     1,
			MaxAttempts: 1,
		}

		result, err := builder.Build(ctx)
		require.NoError(t, err)

		assert.Contains(t, result.Prompt, "## Error Logs")
		assert.Contains(t, result.Prompt, "```\n\n```")
	})

	t.Run("handles single failed step", func(t *testing.T) {
		ctx := CIFixContext{
			FailureInfo: &CIFailureInfo{
				RunID:        "12345",
				WorkflowName: "CI",
				FailedSteps:  []string{"Build"},
				ErrorLogs:    "Error",
			},
			Attempt:     1,
			MaxAttempts: 1,
		}

		result, err := builder.Build(ctx)
		require.NoError(t, err)

		// Single step should appear without comma separator in the step list
		assert.Contains(t, result.Prompt, "**Failed Steps**: Build\n")
	})
}

func TestNewCIFixBuilder(t *testing.T) {
	builder := NewCIFixBuilder()
	assert.NotNil(t, builder)
}
