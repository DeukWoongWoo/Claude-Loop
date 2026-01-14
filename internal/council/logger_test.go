package council

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecisionLogger_Log(t *testing.T) {
	t.Run("writes decision to file when enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, ".claude", "decisions.log")
		logger := NewDecisionLogger(logFile, true)

		decision := &Decision{
			Timestamp:      time.Date(2026, 1, 14, 10, 30, 0, 0, time.UTC),
			Iteration:      5,
			Decision:       "Use startup preset",
			Rationale:      "Speed is prioritized",
			Preset:         config.PresetStartup,
			CouncilInvoked: true,
		}

		err := logger.Log(decision)
		require.NoError(t, err)

		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, "---")
		assert.Contains(t, contentStr, `timestamp: "2026-01-14T10:30:00Z"`)
		assert.Contains(t, contentStr, "iteration: 5")
		assert.Contains(t, contentStr, `decision: "Use startup preset"`)
		assert.Contains(t, contentStr, `rationale: "Speed is prioritized"`)
		assert.Contains(t, contentStr, `preset: "startup"`)
		assert.Contains(t, contentStr, "council_invoked: true")
	})

	t.Run("returns nil when disabled", func(t *testing.T) {
		logger := NewDecisionLogger("/nonexistent/path", false)

		decision := &Decision{
			Decision:  "Test",
			Rationale: "Test",
		}

		err := logger.Log(decision)
		assert.NoError(t, err)
	})

	t.Run("returns nil for nil decision", func(t *testing.T) {
		logger := NewDecisionLogger("/nonexistent/path", true)

		err := logger.Log(nil)
		assert.NoError(t, err)
	})

	t.Run("returns nil for empty decision", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "decisions.log")
		logger := NewDecisionLogger(logFile, true)

		decision := &Decision{
			Iteration: 1,
			// Empty Decision and Rationale
		}

		err := logger.Log(decision)
		assert.NoError(t, err)

		// File should not be created
		_, err = os.Stat(logFile)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("creates directory if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "deep", "nested", "dir", "decisions.log")
		logger := NewDecisionLogger(logFile, true)

		decision := &Decision{
			Decision: "Test decision",
		}

		err := logger.Log(decision)
		require.NoError(t, err)

		_, err = os.Stat(logFile)
		assert.NoError(t, err)
	})

	t.Run("appends to existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "decisions.log")
		logger := NewDecisionLogger(logFile, true)

		// Log first decision
		err := logger.Log(&Decision{
			Iteration: 1,
			Decision:  "First decision",
		})
		require.NoError(t, err)

		// Log second decision
		err = logger.Log(&Decision{
			Iteration: 2,
			Decision:  "Second decision",
		})
		require.NoError(t, err)

		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Equal(t, 2, strings.Count(contentStr, "---"))
		assert.Contains(t, contentStr, "First decision")
		assert.Contains(t, contentStr, "Second decision")
	})

	t.Run("escapes special characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "decisions.log")
		logger := NewDecisionLogger(logFile, true)

		decision := &Decision{
			Decision:  `Decision with "quotes" and\nnewlines`,
			Rationale: "Tab\there",
		}

		err := logger.Log(decision)
		require.NoError(t, err)

		content, err := os.ReadFile(logFile)
		require.NoError(t, err)

		contentStr := string(content)
		assert.Contains(t, contentStr, `\"quotes\"`)
		assert.Contains(t, contentStr, `\\n`)
	})
}

func TestDecisionLogger_IsEnabled(t *testing.T) {
	t.Run("returns true when enabled", func(t *testing.T) {
		logger := NewDecisionLogger("test.log", true)
		assert.True(t, logger.IsEnabled())
	})

	t.Run("returns false when disabled", func(t *testing.T) {
		logger := NewDecisionLogger("test.log", false)
		assert.False(t, logger.IsEnabled())
	})
}

func TestNewDecisionLogger(t *testing.T) {
	logger := NewDecisionLogger("/path/to/log", true)

	assert.NotNil(t, logger)
	assert.Equal(t, "/path/to/log", logger.logFile)
	assert.True(t, logger.enabled)
}

func TestEscapeYAMLString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "simple text",
			expected: "simple text",
		},
		{
			name:     "double quotes",
			input:    `text with "quotes"`,
			expected: `text with \"quotes\"`,
		},
		{
			name:     "newline",
			input:    "line1\nline2",
			expected: `line1\nline2`,
		},
		{
			name:     "tab",
			input:    "col1\tcol2",
			expected: `col1\tcol2`,
		},
		{
			name:     "carriage return",
			input:    "text\rmore",
			expected: `text\rmore`,
		},
		{
			name:     "backslash",
			input:    `path\to\file`,
			expected: `path\\to\\file`,
		},
		{
			name:     "mixed special chars",
			input:    "\"hello\"\n\tworld\\",
			expected: `\"hello\"\n\tworld\\`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeYAMLString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
