package claude

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaudeError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *ClaudeError
		expected string
	}{
		{
			name:     "with result text",
			err:      &ClaudeError{Message: "execution failed", ResultText: "API error"},
			expected: "claude: execution failed: API error",
		},
		{
			name:     "with underlying error",
			err:      &ClaudeError{Message: "execution failed", Err: errors.New("timeout")},
			expected: "claude: execution failed: timeout",
		},
		{
			name:     "message only",
			err:      &ClaudeError{Message: "execution failed"},
			expected: "claude: execution failed",
		},
		{
			name:     "result text takes precedence over error",
			err:      &ClaudeError{Message: "failed", ResultText: "result", Err: errors.New("err")},
			expected: "claude: failed: result",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestClaudeError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := &ClaudeError{Message: "test", Err: underlying}

	assert.Equal(t, underlying, err.Unwrap())
	assert.True(t, errors.Is(err, underlying))
}

func TestClaudeError_UnwrapNil(t *testing.T) {
	err := &ClaudeError{Message: "test"}
	assert.Nil(t, err.Unwrap())
}

func TestParseError_ErrorMessage(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		expected string
	}{
		{
			name:     "with line",
			err:      &ParseError{Message: "invalid json", Line: `{"broken`},
			expected: `parse error: invalid json (line: "{\"broken")`,
		},
		{
			name:     "with underlying error",
			err:      &ParseError{Message: "scanner failed", Err: errors.New("io error")},
			expected: "parse error: scanner failed: io error",
		},
		{
			name:     "message only",
			err:      &ParseError{Message: "unknown error"},
			expected: "parse error: unknown error",
		},
		{
			name:     "with line and underlying error",
			err:      &ParseError{Message: "decode failed", Line: `{"bad": }`, Err: errors.New("unexpected token")},
			expected: `parse error: decode failed: unexpected token (line: "{\"bad\": }")`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestParseError_Unwrap(t *testing.T) {
	underlying := errors.New("io error")
	err := &ParseError{Message: "test", Err: underlying}

	assert.Equal(t, underlying, err.Unwrap())
	assert.True(t, errors.Is(err, underlying))
}

func TestTruncateLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short line",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long line",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "very long line",
			input:    "this is a very long line that needs truncation",
			maxLen:   20,
			expected: "this is a very lo...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, truncateLine(tt.input, tt.maxLen))
		})
	}
}
