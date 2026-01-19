package principles

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockInteractiveClient is a mock implementation of InteractiveClient for testing.
type MockInteractiveClient struct {
	ExecuteInteractiveFunc func(ctx context.Context, prompt string) error
	Calls                  []string
}

func (m *MockInteractiveClient) ExecuteInteractive(ctx context.Context, prompt string) error {
	m.Calls = append(m.Calls, prompt)
	if m.ExecuteInteractiveFunc != nil {
		return m.ExecuteInteractiveFunc(ctx, prompt)
	}
	return nil
}

func TestNewCollector(t *testing.T) {
	t.Parallel()

	client := &MockInteractiveClient{}
	path := "/some/path/principles.yaml"

	collector := NewCollector(client, path)

	assert.NotNil(t, collector)
	assert.Equal(t, client, collector.client)
	assert.Equal(t, path, collector.principlesPath)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			principlesPath := filepath.Join(tmpDir, "principles.yaml")

			if tt.fileExists {
				err := os.WriteFile(principlesPath, []byte("version: \"2.3\""), 0644)
				require.NoError(t, err)
			}

			client := &MockInteractiveClient{}
			collector := NewCollector(client, principlesPath)

			result := collector.NeedsCollection(tt.forceReset)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollector_Collect_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, ".claude", "principles.yaml")

	client := &MockInteractiveClient{
		ExecuteInteractiveFunc: func(ctx context.Context, prompt string) error {
			// Simulate Claude creating the file
			dir := filepath.Dir(principlesPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			return os.WriteFile(principlesPath, []byte("version: \"2.3\"\npreset: startup"), 0644)
		},
	}

	collector := NewCollector(client, principlesPath)
	err := collector.Collect(context.Background())

	require.NoError(t, err)
	assert.Len(t, client.Calls, 1)
	assert.Contains(t, client.Calls[0], "PRINCIPLE COLLECTION REQUIRED")

	// Verify file exists
	_, statErr := os.Stat(principlesPath)
	assert.NoError(t, statErr)
}

func TestCollector_Collect_InteractiveError(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	expectedErr := errors.New("claude execution failed")
	client := &MockInteractiveClient{
		ExecuteInteractiveFunc: func(ctx context.Context, prompt string) error {
			return expectedErr
		},
	}

	collector := NewCollector(client, principlesPath)
	err := collector.Collect(context.Background())

	require.Error(t, err)
	var collectorErr *CollectorError
	require.ErrorAs(t, err, &collectorErr)
	assert.Equal(t, "interactive principles collection failed", collectorErr.Message)
	assert.ErrorIs(t, err, expectedErr)
}

func TestCollector_Collect_FileNotCreated(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	// Mock that doesn't create the file
	client := &MockInteractiveClient{
		ExecuteInteractiveFunc: func(ctx context.Context, prompt string) error {
			// Simulate Claude returning without creating the file
			return nil
		},
	}

	collector := NewCollector(client, principlesPath)
	err := collector.Collect(context.Background())

	require.Error(t, err)
	var collectorErr *CollectorError
	require.ErrorAs(t, err, &collectorErr)
	assert.Equal(t, "principles file was not created after collection", collectorErr.Message)
	assert.Nil(t, collectorErr.Err)
}

func TestCollector_Collect_CreatesParentDirectory(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "nested", "dir", "principles.yaml")

	client := &MockInteractiveClient{
		ExecuteInteractiveFunc: func(ctx context.Context, prompt string) error {
			// Simulate Claude creating the file (directory should already exist)
			return os.WriteFile(principlesPath, []byte("version: \"2.3\""), 0644)
		},
	}

	collector := NewCollector(client, principlesPath)
	err := collector.Collect(context.Background())

	require.NoError(t, err)

	// Verify directory was created
	dirInfo, statErr := os.Stat(filepath.Dir(principlesPath))
	require.NoError(t, statErr)
	assert.True(t, dirInfo.IsDir())
}

func TestCollector_Collect_ContextCancellation(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	principlesPath := filepath.Join(tmpDir, "principles.yaml")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &MockInteractiveClient{
		ExecuteInteractiveFunc: func(ctx context.Context, prompt string) error {
			return ctx.Err()
		},
	}

	collector := NewCollector(client, principlesPath)
	err := collector.Collect(ctx)

	require.Error(t, err)
	var collectorErr *CollectorError
	require.ErrorAs(t, err, &collectorErr)
	assert.ErrorIs(t, collectorErr.Err, context.Canceled)
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
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.err.Error())
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

func TestBuildCollectionPrompt(t *testing.T) {
	t.Parallel()

	prompt := buildCollectionPrompt(".claude/principles.yaml")

	assert.Contains(t, prompt, "PRINCIPLE COLLECTION REQUIRED")
	assert.Contains(t, prompt, "AskUserQuestion")
	assert.Contains(t, prompt, "Project Type")
	assert.Contains(t, prompt, ".claude/principles.yaml")
}

func TestBuildCollectionPrompt_CustomPath(t *testing.T) {
	t.Parallel()

	customPath := "custom/path/my-principles.yaml"
	prompt := buildCollectionPrompt(customPath)

	assert.Contains(t, prompt, customPath)
	assert.NotContains(t, prompt, "PRINCIPLES_FILE_PLACEHOLDER")
}
