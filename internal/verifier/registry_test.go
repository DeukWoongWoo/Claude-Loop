package verifier

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCheckerRegistry(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	require.NotNil(t, registry)
	assert.Equal(t, 4, registry.Count()) // FileExists, Build, Test, ContentMatch
}

func TestNewEmptyRegistry(t *testing.T) {
	registry := NewEmptyRegistry()

	require.NotNil(t, registry)
	assert.Equal(t, 0, registry.Count())
}

func TestCheckerRegistry_Register(t *testing.T) {
	registry := NewEmptyRegistry()

	// Register a checker
	registry.Register(NewFileExistsChecker())
	assert.Equal(t, 1, registry.Count())

	// Register another checker
	registry.Register(NewBuildChecker(nil))
	assert.Equal(t, 2, registry.Count())
}

func TestCheckerRegistry_FindChecker(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	tests := []struct {
		criterion    string
		expectedType string
	}{
		{"file `test.go` exists", "file_exists"},
		{"build passes", "build"},
		{"test passes", "test"},
		{"file `test.go` contains `pattern`", "content_match"},
		{"unknown criterion", ""}, // no checker found
	}

	for _, tt := range tests {
		t.Run(tt.criterion, func(t *testing.T) {
			checker := registry.FindChecker(tt.criterion)
			if tt.expectedType == "" {
				assert.Nil(t, checker)
			} else {
				require.NotNil(t, checker)
				assert.Equal(t, tt.expectedType, checker.Type())
			}
		})
	}
}

func TestCheckerRegistry_FindChecker_Priority(t *testing.T) {
	// Test that the first matching checker is returned
	registry := NewEmptyRegistry()

	// Register file exists first
	registry.Register(NewFileExistsChecker())
	registry.Register(NewBuildChecker(nil))

	// "build passes" should match build checker, not file exists
	checker := registry.FindChecker("build passes")
	require.NotNil(t, checker)
	assert.Equal(t, "build", checker.Type())
}

func TestCheckerRegistry_AllCheckers(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	checkers := registry.AllCheckers()

	assert.Len(t, checkers, 4)

	// Verify types
	types := make([]string, len(checkers))
	for i, c := range checkers {
		types[i] = c.Type()
	}
	assert.Contains(t, types, "file_exists")
	assert.Contains(t, types, "build")
	assert.Contains(t, types, "test")
	assert.Contains(t, types, "content_match")
}

func TestCheckerRegistry_AllCheckers_ReturnsCopy(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	checkers1 := registry.AllCheckers()
	checkers2 := registry.AllCheckers()

	// Modifying one should not affect the other
	assert.Equal(t, len(checkers1), len(checkers2))

	// They should be different slices
	checkers1[0] = nil
	assert.NotNil(t, checkers2[0])
}

func TestCheckerRegistry_ConcurrentAccess(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = registry.FindChecker("file `test.go` exists")
			_ = registry.AllCheckers()
			_ = registry.Count()
		}()
	}

	// Test concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			registry.Register(NewFileExistsChecker())
		}()
	}

	wg.Wait()

	// After all registrations, we should have original 4 + numGoroutines
	assert.Equal(t, 4+numGoroutines, registry.Count())
}

// MockChecker for testing custom checkers
type MockChecker struct {
	typeName   string
	canHandle  bool
	checkFunc  func(ctx context.Context, criterion string, workDir string) *CheckResult
}

func (m *MockChecker) Type() string {
	return m.typeName
}

func (m *MockChecker) CanHandle(criterion string) bool {
	return m.canHandle
}

func (m *MockChecker) Check(ctx context.Context, criterion string, workDir string) *CheckResult {
	if m.checkFunc != nil {
		return m.checkFunc(ctx, criterion, workDir)
	}
	return &CheckResult{Criterion: criterion, Passed: true}
}

func TestCheckerRegistry_CustomChecker(t *testing.T) {
	registry := NewEmptyRegistry()

	customChecker := &MockChecker{
		typeName:  "custom",
		canHandle: true,
		checkFunc: func(ctx context.Context, criterion string, workDir string) *CheckResult {
			return &CheckResult{
				Criterion:   criterion,
				CheckerType: "custom",
				Passed:      true,
			}
		},
	}

	registry.Register(customChecker)

	// Find the custom checker
	found := registry.FindChecker("any criterion")
	require.NotNil(t, found)
	assert.Equal(t, "custom", found.Type())

	// Use the custom checker
	result := found.Check(context.Background(), "test", "")
	assert.True(t, result.Passed)
	assert.Equal(t, "custom", result.CheckerType)
}

func TestCheckerRegistry_EmptyRegistry_FindReturnsNil(t *testing.T) {
	registry := NewEmptyRegistry()

	checker := registry.FindChecker("file `test.go` exists")
	assert.Nil(t, checker)
}

func TestCheckerRegistry_NilExecutorUsesDefault(t *testing.T) {
	registry := NewCheckerRegistry(nil)

	// Should create default executor internally
	assert.Equal(t, 4, registry.Count())

	// Build checker should work
	buildChecker := registry.FindChecker("build passes")
	require.NotNil(t, buildChecker)
	assert.Equal(t, "build", buildChecker.Type())
}
