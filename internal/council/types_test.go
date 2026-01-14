package council

import (
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.NotNil(t, cfg)
	assert.Equal(t, ".claude/principles-decisions.log", cfg.LogFile)
	assert.Nil(t, cfg.Principles)
	assert.False(t, cfg.LogDecisions)
}

func TestConfig_IsEnabled(t *testing.T) {
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
			name:     "nil principles",
			config:   &Config{},
			expected: false,
		},
		{
			name: "with principles",
			config: &Config{
				Principles: &config.Principles{},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.IsEnabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}
