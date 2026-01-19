package principles

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollector(t *testing.T) {
	t.Parallel()

	path := "/some/path/principles.yaml"

	collector := NewCollector(path)

	assert.NotNil(t, collector)
	assert.Equal(t, path, collector.principlesPath)
	assert.Equal(t, os.Stdin, collector.reader)
}

func TestNewCollectorWithReader(t *testing.T) {
	t.Parallel()

	path := "/some/path/principles.yaml"
	reader := strings.NewReader("test")

	collector := NewCollectorWithReader(path, reader)

	assert.NotNil(t, collector)
	assert.Equal(t, path, collector.principlesPath)
	assert.Equal(t, reader, collector.reader)
}

func TestCollector_NeedsCollection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		fileExists bool
		forceReset bool
		expected   bool
	}{
		{
			name:       "file exists, no reset",
			fileExists: true,
			forceReset: false,
			expected:   false,
		},
		{
			name:       "file exists, force reset",
			fileExists: true,
			forceReset: true,
			expected:   true,
		},
		{
			name:       "file does not exist, no reset",
			fileExists: false,
			forceReset: false,
			expected:   true,
		},
		{
			name:       "file does not exist, force reset",
			fileExists: false,
			forceReset: true,
			expected:   true,
		},
	}

	for _, tt := range tests {
		tc := tt // capture loop variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			principlesPath := filepath.Join(tmpDir, "principles.yaml")

			if tc.fileExists {
				err := os.WriteFile(principlesPath, []byte("version: \"2.3\""), 0644)
				require.NoError(t, err)
			}

			collector := NewCollector(principlesPath)

			result := collector.NeedsCollection(tc.forceReset)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCollector_Collect_Startup(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, ".claude", "principles.yaml")

	// Simulate user input: "1" for startup, "5" for scope, "6" for speed
	input := "1\n5\n6\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	// Verify file was created
	_, statErr := os.Stat(principlesPath)
	assert.NoError(t, statErr)

	// Load and verify content
	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)
	assert.Equal(t, config.PresetStartup, loaded.Preset)
	assert.Equal(t, 5, loaded.Layer0.ScopePhilosophy)
	assert.Equal(t, 6, loaded.Layer1.SpeedCorrectness)
}

func TestCollector_Collect_Enterprise(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Simulate user input: "2" for enterprise, "8" for blast radius, "7" for innovation
	input := "2\n8\n7\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)
	assert.Equal(t, config.PresetEnterprise, loaded.Preset)
	assert.Equal(t, 8, loaded.Layer1.BlastRadius)
	assert.Equal(t, 7, loaded.Layer1.InnovationStability)
}

func TestCollector_Collect_OpenSource(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Simulate user input: "3" for opensource, "4" for curation, "6" for UX
	input := "3\n4\n6\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)
	assert.Equal(t, config.PresetOpenSource, loaded.Preset)
	assert.Equal(t, 4, loaded.Layer0.CurationModel)
	assert.Equal(t, 6, loaded.Layer0.UXPhilosophy)
}

func TestCollector_Collect_DefaultsOnEmpty(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Simulate user input: empty lines (use defaults)
	input := "\n\n\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)
	// Default should be startup
	assert.Equal(t, config.PresetStartup, loaded.Preset)
	// Should use default values
	assert.Equal(t, 3, loaded.Layer0.ScopePhilosophy) // default for startup
	assert.Equal(t, 4, loaded.Layer1.SpeedCorrectness) // default for startup
}

func TestCollector_Collect_InvalidInput(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Simulate user input: invalid preset, out of range values
	input := "invalid\nabc\n15\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)
	// Should fall back to defaults
	assert.Equal(t, config.PresetStartup, loaded.Preset)
	assert.Equal(t, 3, loaded.Layer0.ScopePhilosophy) // default
	assert.Equal(t, 4, loaded.Layer1.SpeedCorrectness) // default
}

func TestCollector_Collect_CreatesParentDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "nested", "dir", "principles.yaml")

	input := "1\n\n\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	// Verify directory was created
	dirInfo, statErr := os.Stat(filepath.Dir(principlesPath))
	require.NoError(t, statErr)
	assert.True(t, dirInfo.IsDir())
}

func TestCollector_Collect_ValidNumberRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"minimum value", "1\n1\n1\n", 1},
		{"maximum value", "1\n10\n10\n", 10},
		{"mid value", "1\n5\n5\n", 5},
		{"below minimum uses default", "1\n0\n0\n", 3}, // default
		{"above maximum uses default", "1\n11\n11\n", 3}, // default
	}

	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			principlesPath := filepath.Join(tmpDir, "principles.yaml")

			collector := NewCollectorWithReader(principlesPath, strings.NewReader(tc.input))
			err := collector.Collect(context.Background())

			require.NoError(t, err)

			loaded, loadErr := config.LoadFromFile(principlesPath)
			require.NoError(t, loadErr)
			assert.Equal(t, tc.expected, loaded.Layer0.ScopePhilosophy)
		})
	}
}

func TestCollector_Collect_PreservesOtherDefaults(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Only specify the asked questions, others should use defaults
	input := "1\n7\n8\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(context.Background())

	require.NoError(t, err)

	loaded, loadErr := config.LoadFromFile(principlesPath)
	require.NoError(t, loadErr)

	// Check that other defaults are preserved (startup defaults)
	assert.Equal(t, "2.3", loaded.Version)
	assert.Equal(t, 7, loaded.Layer0.TrustArchitecture) // startup default
	assert.Equal(t, 6, loaded.Layer0.CurationModel)     // startup default
	assert.Equal(t, 7, loaded.Layer1.BlastRadius)       // startup default
}

func TestCollectorError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *CollectorError
		expected string
	}{
		{
			name: "with underlying error",
			err: &CollectorError{
				Message: "collection failed",
				Err:     errors.New("network error"),
			},
			expected: "collection failed: network error",
		},
		{
			name: "without underlying error",
			err: &CollectorError{
				Message: "file not found",
				Err:     nil,
			},
			expected: "file not found",
		},
	}

	for _, tt := range tests {
		tc := tt // capture loop variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, tc.err.Error())
		})
	}
}

func TestCollectorError_Unwrap(t *testing.T) {
	t.Parallel()

	underlying := errors.New("underlying error")
	err := &CollectorError{
		Message: "wrapper",
		Err:     underlying,
	}

	assert.Equal(t, underlying, err.Unwrap())
	assert.ErrorIs(t, err, underlying)
}

func TestCollectorError_Unwrap_Nil(t *testing.T) {
	t.Parallel()

	err := &CollectorError{
		Message: "no underlying",
		Err:     nil,
	}

	assert.Nil(t, err.Unwrap())
}

func TestCollector_Collect_ContextCancellation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	input := "1\n5\n6\n"
	collector := NewCollectorWithReader(principlesPath, strings.NewReader(input))

	err := collector.Collect(ctx)

	require.Error(t, err)
	var collectorErr *CollectorError
	require.ErrorAs(t, err, &collectorErr)
	assert.Equal(t, "cancelled", collectorErr.Message)
	assert.ErrorIs(t, collectorErr.Err, context.Canceled)
}
