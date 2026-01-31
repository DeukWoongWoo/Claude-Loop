package verifier

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockClaudeClient for testing AI verification.
type MockClaudeClient struct {
	ExecuteFunc func(ctx context.Context, prompt string) (*IterationResult, error)
	calls       []string
}

func (m *MockClaudeClient) Execute(ctx context.Context, prompt string) (*IterationResult, error) {
	m.calls = append(m.calls, prompt)
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, prompt)
	}
	return &IterationResult{Output: "VERIFICATION_PASS"}, nil
}

func TestNewVerifier(t *testing.T) {
	verifier := NewVerifier(nil, nil)

	require.NotNil(t, verifier)
	assert.NotNil(t, verifier.config)
	assert.NotNil(t, verifier.registry)
	assert.NotNil(t, verifier.promptBuilder)
	assert.Nil(t, verifier.client)
}

func TestNewVerifier_WithConfig(t *testing.T) {
	config := &Config{
		Level:    VerificationLevelStrict,
		EnableAI: true,
		Timeout:  10 * time.Minute,
	}

	verifier := NewVerifier(config, nil)

	assert.Equal(t, VerificationLevelStrict, verifier.config.Level)
	assert.True(t, verifier.config.EnableAI)
	assert.Equal(t, 10*time.Minute, verifier.config.Timeout)
}

func TestNewVerifierWithRegistry(t *testing.T) {
	registry := NewEmptyRegistry()
	verifier := NewVerifierWithRegistry(nil, nil, registry)

	require.NotNil(t, verifier)
	assert.Equal(t, 0, verifier.registry.Count())
}

func TestNewVerifierWithRegistry_NilRegistry(t *testing.T) {
	verifier := NewVerifierWithRegistry(nil, nil, nil)

	require.NotNil(t, verifier)
	assert.Equal(t, 4, verifier.registry.Count()) // Default checkers
}

func TestDefaultVerifier_Verify_NilTask(t *testing.T) {
	verifier := NewVerifier(nil, nil)

	result, err := verifier.Verify(context.Background(), nil)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, IsVerifierError(err))
	assert.Equal(t, ErrNilTask, err)
}

func TestDefaultVerifier_Verify_NoCriteria(t *testing.T) {
	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{},
	}

	result, err := verifier.Verify(context.Background(), task)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.Equal(t, ErrNoCriteria, err)
}

func TestDefaultVerifier_Verify_FileExists(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	err := os.WriteFile(tmpFile, []byte("package main"), 0644)
	require.NoError(t, err)

	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		Title:           "Test Task",
		SuccessCriteria: []string{"file `test.go` exists"},
		WorkDir:         tmpDir,
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "T001", result.TaskID)
	assert.True(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.True(t, result.Checks[0].Passed)
	assert.Equal(t, "file_exists", result.Checks[0].CheckerType)
}

func TestDefaultVerifier_Verify_FileNotExists(t *testing.T) {
	tmpDir := t.TempDir()

	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"file `nonexistent.go` exists"},
		WorkDir:         tmpDir,
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.False(t, result.Checks[0].Passed)
}

func TestDefaultVerifier_Verify_MultipleCriteria(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("pkg"), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("pkg"), 0644)
	require.NoError(t, err)

	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID: "T001",
		SuccessCriteria: []string{
			"file `file1.go` exists",
			"file `file2.go` exists",
		},
		WorkDir: tmpDir,
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.Len(t, result.Checks, 2)
	assert.True(t, result.Checks[0].Passed)
	assert.True(t, result.Checks[1].Passed)
}

func TestDefaultVerifier_Verify_PartialFailure(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "exists.go"), []byte("pkg"), 0644)
	require.NoError(t, err)

	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID: "T001",
		SuccessCriteria: []string{
			"file `exists.go` exists",
			"file `missing.go` exists",
		},
		WorkDir: tmpDir,
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 2)
	assert.True(t, result.Checks[0].Passed)
	assert.False(t, result.Checks[1].Passed)

	// Check FailedChecks
	failed := result.FailedChecks()
	assert.Len(t, failed, 1)
	assert.Contains(t, failed[0].Criterion, "missing.go")
}

func TestDefaultVerifier_Verify_UnknownCriterion(t *testing.T) {
	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"some unknown criterion"},
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.Equal(t, "unknown", result.Checks[0].CheckerType)
	assert.Contains(t, result.Checks[0].Error, "no suitable checker found")
}

func TestDefaultVerifier_Verify_AIVerification(t *testing.T) {
	mockClient := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output:   "VERIFICATION_PASS: Criterion met",
				Cost:     0.01,
				Duration: 100 * time.Millisecond,
			}, nil
		},
	}

	config := &Config{
		EnableAI: true,
	}

	verifier := NewVerifier(config, mockClient)
	task := &VerificationTask{
		TaskID:          "T001",
		Title:           "Test",
		SuccessCriteria: []string{"some semantic criterion"},
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.Equal(t, "ai", result.Checks[0].CheckerType)
	assert.Len(t, mockClient.calls, 1)
}

func TestDefaultVerifier_Verify_AIVerificationFail(t *testing.T) {
	mockClient := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return &IterationResult{
				Output: "VERIFICATION_FAIL: Criterion not met",
			}, nil
		},
	}

	config := &Config{
		EnableAI: true,
	}

	verifier := NewVerifier(config, mockClient)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"some criterion"},
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.False(t, result.Checks[0].Passed)
	assert.Contains(t, result.Checks[0].Error, "AI verification determined criterion not met")
}

func TestDefaultVerifier_Verify_AIVerificationError(t *testing.T) {
	mockClient := &MockClaudeClient{
		ExecuteFunc: func(ctx context.Context, prompt string) (*IterationResult, error) {
			return nil, errors.New("API error")
		},
	}

	config := &Config{
		EnableAI: true,
	}

	verifier := NewVerifier(config, mockClient)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"some criterion"},
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.False(t, result.Passed)
	assert.Len(t, result.Checks, 1)
	assert.False(t, result.Checks[0].Passed)
	assert.Contains(t, result.Checks[0].Error, "AI verification failed")
}

func TestDefaultVerifier_Verify_ContextCancelled(t *testing.T) {
	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID: "T001",
		SuccessCriteria: []string{
			"file `test1.go` exists",
			"file `test2.go` exists",
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := verifier.Verify(ctx, task)

	assert.Nil(t, result)
	require.Error(t, err)
	assert.True(t, IsVerifierError(err))
	assert.Contains(t, err.Error(), "context cancelled")
}

func TestDefaultVerifier_Verify_WorkDirFromConfig(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("pkg"), 0644)
	require.NoError(t, err)

	config := &Config{
		WorkDir: tmpDir,
	}

	verifier := NewVerifier(config, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"file `test.go` exists"},
		// Note: WorkDir not set on task, should use config.WorkDir
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
}

func TestDefaultVerifier_Verify_TaskWorkDirOverridesConfig(t *testing.T) {
	configDir := t.TempDir()
	taskDir := t.TempDir()

	// File only exists in taskDir
	err := os.WriteFile(filepath.Join(taskDir, "test.go"), []byte("pkg"), 0644)
	require.NoError(t, err)

	config := &Config{
		WorkDir: configDir,
	}

	verifier := NewVerifier(config, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"file `test.go` exists"},
		WorkDir:         taskDir, // Should override config.WorkDir
	}

	result, err := verifier.Verify(context.Background(), task)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Passed)
}

func TestDefaultVerifier_Config(t *testing.T) {
	config := &Config{Level: VerificationLevelStrict}
	verifier := NewVerifier(config, nil)

	assert.Equal(t, config, verifier.Config())
}

func TestDefaultVerifier_Registry(t *testing.T) {
	verifier := NewVerifier(nil, nil)

	registry := verifier.Registry()

	require.NotNil(t, registry)
	assert.Equal(t, 4, registry.Count())
}

func TestDefaultVerifier_Verify_Timestamp(t *testing.T) {
	tmpDir := t.TempDir()
	err := os.WriteFile(filepath.Join(tmpDir, "test.go"), []byte("pkg"), 0644)
	require.NoError(t, err)

	before := time.Now()

	verifier := NewVerifier(nil, nil)
	task := &VerificationTask{
		TaskID:          "T001",
		SuccessCriteria: []string{"file `test.go` exists"},
		WorkDir:         tmpDir,
	}

	result, err := verifier.Verify(context.Background(), task)

	after := time.Now()

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.Timestamp.After(before) || result.Timestamp.Equal(before))
	assert.True(t, result.Timestamp.Before(after) || result.Timestamp.Equal(after))
	assert.True(t, result.Duration > 0)
}

func TestDefaultVerifier_InterfaceCompliance(t *testing.T) {
	var _ Verifier = (*DefaultVerifier)(nil)
}
