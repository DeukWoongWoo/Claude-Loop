package reviewer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 3, config.MaxConsecutiveErrors)
	assert.Empty(t, config.ReviewPrompt)
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
			name:     "empty review prompt",
			config:   &Config{ReviewPrompt: ""},
			expected: false,
		},
		{
			name:     "with review prompt",
			config:   &Config{ReviewPrompt: "run tests"},
			expected: true,
		},
		{
			name:     "whitespace-only prompt is enabled",
			config:   &Config{ReviewPrompt: "   "},
			expected: true,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.config.IsEnabled())
		})
	}
}
