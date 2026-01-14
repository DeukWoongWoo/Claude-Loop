package integration

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/cli"
	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/DeukWoongWoo/claude-loop/test/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_LoadPrinciples_Integration(t *testing.T) {
	fixturesDir := testutil.GetFixturePath("principles")

	t.Run("loads valid principles file", func(t *testing.T) {
		path := filepath.Join(fixturesDir, "valid.yaml")
		p, err := config.LoadFromFile(path)

		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, "2.3", p.Version)
		assert.Equal(t, config.PresetStartup, p.Preset)
		assert.Equal(t, "2026-01-13", p.CreatedAt)

		// Check Layer0 values
		assert.Equal(t, 5, p.Layer0.TrustArchitecture)
		assert.Equal(t, 6, p.Layer0.CurationModel)
		assert.Equal(t, 8, p.Layer0.Auditability)

		// Check Layer1 values
		assert.Equal(t, 6, p.Layer1.SpeedCorrectness)
		assert.Equal(t, 8, p.Layer1.ClarityOfIntent)
	})

	t.Run("returns error for invalid YAML syntax", func(t *testing.T) {
		path := filepath.Join(fixturesDir, "invalid_syntax.yaml")
		_, err := config.LoadFromFile(path)

		assert.Error(t, err)
		assert.True(t, config.IsLoadError(err))

		loadErr := err.(*config.LoadError)
		assert.Contains(t, loadErr.Message, "invalid YAML")
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := config.LoadFromFile("/nonexistent/path.yaml")

		assert.Error(t, err)
		assert.True(t, config.IsLoadError(err))

		loadErr := err.(*config.LoadError)
		assert.Contains(t, loadErr.Message, "file not found")
	})

	t.Run("loads defaults when file not found", func(t *testing.T) {
		p, err := config.LoadOrDefault("/nonexistent/path.yaml", config.PresetStartup)

		require.NoError(t, err)
		assert.NotNil(t, p)
		assert.Equal(t, config.PresetStartup, p.Preset)
	})
}

func TestConfig_DefaultPrinciples_Integration(t *testing.T) {
	presets := []config.Preset{
		config.PresetStartup,
		config.PresetEnterprise,
		config.PresetOpenSource,
	}

	for _, preset := range presets {
		t.Run(string(preset), func(t *testing.T) {
			p := config.DefaultPrinciples(preset)

			assert.NotNil(t, p)
			assert.Equal(t, preset, p.Preset)
			assert.NotEmpty(t, p.Version)

			// Verify all Layer0 values are in valid range (1-10)
			assert.GreaterOrEqual(t, p.Layer0.TrustArchitecture, 1)
			assert.LessOrEqual(t, p.Layer0.TrustArchitecture, 10)
			assert.GreaterOrEqual(t, p.Layer0.CurationModel, 1)
			assert.LessOrEqual(t, p.Layer0.CurationModel, 10)

			// Verify all Layer1 values are in valid range (1-10)
			assert.GreaterOrEqual(t, p.Layer1.SpeedCorrectness, 1)
			assert.LessOrEqual(t, p.Layer1.SpeedCorrectness, 10)
		})
	}
}

func TestConfig_FlagsToLoopConfig_Integration(t *testing.T) {
	t.Run("converts all fields correctly", func(t *testing.T) {
		flags := &cli.Flags{
			Prompt:              "test prompt",
			MaxRuns:             10,
			MaxCost:             5.0,
			MaxDuration:         time.Hour,
			CompletionSignal:    "DONE",
			CompletionThreshold: 5,
			ReviewPrompt:        "run tests",
			NotesFile:           "NOTES.md",
			DryRun:              true,
		}

		loopConfig := loop.ConfigFromFlags(flags)

		assert.Equal(t, "test prompt", loopConfig.Prompt)
		assert.Equal(t, 10, loopConfig.MaxRuns)
		assert.Equal(t, 5.0, loopConfig.MaxCost)
		assert.Equal(t, time.Hour, loopConfig.MaxDuration)
		assert.Equal(t, "DONE", loopConfig.CompletionSignal)
		assert.Equal(t, 5, loopConfig.CompletionThreshold)
		assert.Equal(t, "run tests", loopConfig.ReviewPrompt)
		assert.Equal(t, "NOTES.md", loopConfig.NotesFile)
		assert.True(t, loopConfig.DryRun)
	})

	t.Run("sets default values", func(t *testing.T) {
		flags := &cli.Flags{
			Prompt:  "test",
			MaxRuns: 5,
		}

		loopConfig := loop.ConfigFromFlags(flags)

		assert.Equal(t, 3, loopConfig.MaxConsecutiveErrors) // Default
	})
}

func TestConfig_Validation_Integration(t *testing.T) {
	t.Run("validates preset values", func(t *testing.T) {
		assert.True(t, config.IsValidPreset(config.PresetStartup))
		assert.True(t, config.IsValidPreset(config.PresetEnterprise))
		assert.True(t, config.IsValidPreset(config.PresetOpenSource))
		assert.True(t, config.IsValidPreset(config.PresetCustom))
		assert.False(t, config.IsValidPreset(config.Preset("invalid")))
	})
}

func TestConfig_WriteAndLoad_Integration(t *testing.T) {
	t.Run("can write and reload principles", func(t *testing.T) {
		tmpDir := t.TempDir()
		original := config.DefaultPrinciples(config.PresetStartup)

		yamlContent := `version: "2.3"
preset: startup
created_at: "2026-01-13"

layer0:
  trust_architecture: 5
  curation_model: 6
  scope_philosophy: 4
  monetization_model: 5
  privacy_posture: 7
  ux_philosophy: 6
  authority_stance: 5
  auditability: 8
  interoperability: 7

layer1:
  speed_correctness: 6
  innovation_stability: 5
  blast_radius: 4
  clarity_of_intent: 8
  reversibility_priority: 7
  security_posture: 8
  urgency_tiers: 5
  cost_efficiency: 6
  migration_burden: 4
`
		path := testutil.WriteFile(t, tmpDir, "principles.yaml", yamlContent)

		// Reload and verify
		loaded, err := config.LoadFromFile(path)
		require.NoError(t, err)

		assert.Equal(t, original.Preset, loaded.Preset)
		assert.Equal(t, 5, loaded.Layer0.TrustArchitecture)
		assert.Equal(t, 6, loaded.Layer1.SpeedCorrectness)
	})
}

func TestConfig_LoopDefaultConfig_Integration(t *testing.T) {
	t.Run("default config has expected values", func(t *testing.T) {
		cfg := loop.DefaultConfig()

		assert.Equal(t, "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", cfg.CompletionSignal)
		assert.Equal(t, 3, cfg.CompletionThreshold)
		assert.Equal(t, 3, cfg.MaxConsecutiveErrors)
		assert.False(t, cfg.DryRun)
	})
}
