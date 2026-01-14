package council

import (
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_Build(t *testing.T) {
	builder := NewPromptBuilder()

	t.Run("success with valid principles", func(t *testing.T) {
		ctx := BuildContext{
			ConflictContext: "Cannot decide between speed and correctness",
			Principles: &config.Principles{
				Version: "2.3",
				Preset:  config.PresetStartup,
				Layer0: config.Layer0{
					TrustArchitecture: 7,
					CurationModel:     5,
				},
			},
		}

		result, err := builder.Build(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, result.Prompt, "Conflict Context")
		assert.Contains(t, result.Prompt, "Cannot decide between speed and correctness")
		assert.Contains(t, result.Prompt, "Current Principles")
		assert.Contains(t, result.Prompt, "R10 3-step resolution protocol")
		assert.Contains(t, result.Prompt, "Compatibility Check")
		assert.Contains(t, result.Prompt, "Type Classification")
		assert.Contains(t, result.Prompt, "Priority Resolution")
		assert.Contains(t, result.Prompt, "**Decision**")
		assert.Contains(t, result.Prompt, "**Rationale**")
	})

	t.Run("error with nil principles", func(t *testing.T) {
		ctx := BuildContext{
			ConflictContext: "Some conflict",
			Principles:      nil,
		}

		result, err := builder.Build(ctx)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, ErrNoPrinciples, err)
	})

	t.Run("includes YAML-formatted principles", func(t *testing.T) {
		ctx := BuildContext{
			ConflictContext: "Test conflict",
			Principles: &config.Principles{
				Version: "2.3",
				Preset:  config.PresetEnterprise,
				Layer1: config.Layer1{
					SpeedCorrectness:    3,
					InnovationStability: 4,
				},
			},
		}

		result, err := builder.Build(ctx)

		require.NoError(t, err)
		assert.Contains(t, result.Prompt, "version: \"2.3\"")
		assert.Contains(t, result.Prompt, "preset: enterprise")
	})
}

func TestNewPromptBuilder(t *testing.T) {
	builder := NewPromptBuilder()
	assert.NotNil(t, builder)
}

func TestTemplateCouncilResolution(t *testing.T) {
	// Verify template contains key sections
	assert.Contains(t, TemplateCouncilResolution, "Conflict Context")
	assert.Contains(t, TemplateCouncilResolution, "Current Principles")
	assert.Contains(t, TemplateCouncilResolution, "R10 3-step resolution protocol")
	assert.Contains(t, TemplateCouncilResolution, "Compatibility Check")
	assert.Contains(t, TemplateCouncilResolution, "Type Classification")
	assert.Contains(t, TemplateCouncilResolution, "Priority Resolution")
	assert.Contains(t, TemplateCouncilResolution, "Response Format")
}
