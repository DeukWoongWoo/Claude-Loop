package architecture

import (
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.True(t, config.ValidateOutput)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		expected bool
	}{
		{
			name:     "nil config",
			config:   nil,
			expected: false,
		},
		{
			name:     "non-nil config",
			config:   &Config{},
			expected: true,
		},
		{
			name:     "default config",
			config:   DefaultConfig(),
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.config.IsEnabled())
		})
	}
}

func TestArchitecture_EmbeddedFields(t *testing.T) {
	t.Parallel()

	arch := &Architecture{
		Architecture: planner.Architecture{
			Components: []planner.Component{
				{Name: "Parser", Description: "Parses output", Files: []string{"parser.go"}},
			},
			Dependencies:  []string{"regexp"},
			FileStructure: []string{"internal/architecture/parser.go"},
			TechDecisions: []string{"Use regex for parsing"},
			RawOutput:     "raw output here",
		},
		ID:        "arch-001",
		Title:     "Test Architecture",
		Summary:   "Summary of architecture",
		Rationale: "Why we chose this approach",
	}

	// Verify embedded fields are accessible
	assert.Len(t, arch.Components, 1)
	assert.Equal(t, "Parser", arch.Components[0].Name)
	assert.Equal(t, "Parses output", arch.Components[0].Description)
	assert.Equal(t, []string{"parser.go"}, arch.Components[0].Files)
	assert.Equal(t, []string{"regexp"}, arch.Dependencies)
	assert.Equal(t, []string{"internal/architecture/parser.go"}, arch.FileStructure)
	assert.Equal(t, []string{"Use regex for parsing"}, arch.TechDecisions)
	assert.Equal(t, "raw output here", arch.RawOutput)

	// Verify extended fields
	assert.Equal(t, "arch-001", arch.ID)
	assert.Equal(t, "Test Architecture", arch.Title)
	assert.Equal(t, "Summary of architecture", arch.Summary)
	assert.Equal(t, "Why we chose this approach", arch.Rationale)
}

func TestResult(t *testing.T) {
	t.Parallel()

	arch := &Architecture{
		ID: "arch-001",
	}

	result := &Result{
		Architecture: arch,
		Cost:         0.05,
		Duration:     1000000000, // 1 second in nanoseconds
	}

	assert.Equal(t, arch, result.Architecture)
	assert.Equal(t, 0.05, result.Cost)
	assert.Equal(t, int64(1000000000), int64(result.Duration))
}
