package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadFromFile(t *testing.T) {
	t.Run("valid startup example file", func(t *testing.T) {
		p, err := LoadFromFile("../../examples/principles-startup.yaml")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, PresetStartup, p.Preset)
		assert.Equal(t, "2.3", p.Version)
	})

	t.Run("valid enterprise example file", func(t *testing.T) {
		p, err := LoadFromFile("../../examples/principles-enterprise.yaml")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, PresetEnterprise, p.Preset)
	})

	t.Run("valid opensource example file", func(t *testing.T) {
		p, err := LoadFromFile("../../examples/principles-opensource.yaml")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, PresetOpenSource, p.Preset)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadFromFile("/nonexistent/path.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "file not found")
		assert.True(t, IsLoadError(err))
	})

	t.Run("invalid yaml syntax", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidFile, []byte("invalid: yaml: content: ["), 0644)
		require.NoError(t, err)

		_, err = LoadFromFile(invalidFile)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid YAML syntax")
		assert.True(t, IsLoadError(err))
	})
}

func TestLoadFromBytes(t *testing.T) {
	t.Run("valid yaml", func(t *testing.T) {
		input := `
version: "2.3"
preset: startup
created_at: "2026-01-11"
layer0:
  trust_architecture: 7
  curation_model: 6
  scope_philosophy: 3
  monetization_model: 5
  privacy_posture: 7
  ux_philosophy: 4
  authority_stance: 6
  auditability: 5
  interoperability: 7
layer1:
  speed_correctness: 4
  innovation_stability: 6
  blast_radius: 7
  clarity_of_intent: 6
  reversibility_priority: 7
  security_posture: 7
  urgency_tiers: 3
  cost_efficiency: 6
  migration_burden: 5
`
		p, err := LoadFromBytes([]byte(input), "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Equal(t, "2.3", p.Version)
		assert.Equal(t, PresetStartup, p.Preset)
		assert.Equal(t, "2026-01-11", p.CreatedAt)
		assert.Equal(t, 7, p.Layer0.TrustArchitecture)
		assert.Equal(t, 4, p.Layer1.SpeedCorrectness)
	})

	t.Run("empty input", func(t *testing.T) {
		p, err := LoadFromBytes([]byte(""), "empty.yaml")
		require.NoError(t, err)
		require.NotNil(t, p)
		assert.Empty(t, p.Version)
	})

	t.Run("partial yaml", func(t *testing.T) {
		input := `
version: "2.3"
preset: enterprise
`
		p, err := LoadFromBytes([]byte(input), "partial.yaml")
		require.NoError(t, err)
		assert.Equal(t, "2.3", p.Version)
		assert.Equal(t, PresetEnterprise, p.Preset)
		assert.Empty(t, p.CreatedAt)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		_, err := LoadFromBytes([]byte("invalid: [yaml"), "invalid.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid YAML syntax")
	})
}

func TestLoadOrDefault(t *testing.T) {
	t.Run("file exists returns file content", func(t *testing.T) {
		p, err := LoadOrDefault("../../examples/principles-startup.yaml", PresetEnterprise)
		require.NoError(t, err)
		assert.Equal(t, PresetStartup, p.Preset) // file content, not default
	})

	t.Run("file not exists returns default", func(t *testing.T) {
		p, err := LoadOrDefault("/nonexistent/path.yaml", PresetEnterprise)
		require.NoError(t, err)
		assert.Equal(t, PresetEnterprise, p.Preset) // default preset
	})

	t.Run("invalid yaml returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		invalidFile := filepath.Join(tmpDir, "invalid.yaml")
		err := os.WriteFile(invalidFile, []byte("invalid: yaml: ["), 0644)
		require.NoError(t, err)

		_, err = LoadOrDefault(invalidFile, PresetStartup)
		require.Error(t, err)
	})
}

func TestLoadError(t *testing.T) {
	t.Run("error with wrapped error", func(t *testing.T) {
		innerErr := os.ErrNotExist
		err := &LoadError{
			Path:    "/test.yaml",
			Message: "file not found",
			Err:     innerErr,
		}
		assert.Contains(t, err.Error(), "/test.yaml")
		assert.Contains(t, err.Error(), "file not found")
		assert.Equal(t, innerErr, err.Unwrap())
	})

	t.Run("error without wrapped error", func(t *testing.T) {
		err := &LoadError{
			Path:    "/test.yaml",
			Message: "test error",
		}
		assert.Contains(t, err.Error(), "/test.yaml")
		assert.Contains(t, err.Error(), "test error")
		assert.Nil(t, err.Unwrap())
	})
}

func TestIsLoadError(t *testing.T) {
	t.Run("LoadError returns true", func(t *testing.T) {
		err := &LoadError{Path: "/test.yaml", Message: "test"}
		assert.True(t, IsLoadError(err))
	})

	t.Run("other error returns false", func(t *testing.T) {
		err := os.ErrNotExist
		assert.False(t, IsLoadError(err))
	})

	t.Run("nil returns false", func(t *testing.T) {
		assert.False(t, IsLoadError(nil))
	})
}

func TestLoadExampleFilesValidation(t *testing.T) {
	examples := []struct {
		path   string
		preset Preset
	}{
		{"../../examples/principles-startup.yaml", PresetStartup},
		{"../../examples/principles-enterprise.yaml", PresetEnterprise},
		{"../../examples/principles-opensource.yaml", PresetOpenSource},
	}

	for _, ex := range examples {
		t.Run(ex.path, func(t *testing.T) {
			p, err := LoadFromFile(ex.path)
			require.NoError(t, err)

			assert.Equal(t, ex.preset, p.Preset)

			err = p.Validate()
			assert.NoError(t, err, "example file should pass validation")
		})
	}
}
