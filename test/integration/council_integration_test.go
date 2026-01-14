package integration

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/DeukWoongWoo/claude-loop/internal/loop"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// councilMockClient is a mock client for council integration tests.
type councilMockClient struct {
	responses []*loop.IterationResult
	callCount int
}

func (m *councilMockClient) Execute(ctx context.Context, prompt string) (*loop.IterationResult, error) {
	if m.callCount >= len(m.responses) {
		return &loop.IterationResult{
			Output:   "Default response",
			Cost:     0.01,
			Duration: 100 * time.Millisecond,
		}, nil
	}
	result := m.responses[m.callCount]
	m.callCount++
	return result, nil
}

func TestCouncilIntegration_NoConflict(t *testing.T) {
	// Test that council doesn't invoke when no conflict is detected
	client := &councilMockClient{
		responses: []*loop.IterationResult{
			{
				Output:                "Task completed successfully.\n**Decision**: Use startup preset\n**Rationale**: Speed is priority",
				Cost:                  0.05,
				Duration:              1 * time.Second,
				CompletionSignalFound: true,
			},
		},
	}

	cfg := &loop.Config{
		Prompt:               "Test task",
		MaxRuns:              1,
		CompletionSignal:     "COMPLETE",
		CompletionThreshold:  1,
		MaxConsecutiveErrors: 3,
		Principles: &config.Principles{
			Version: "2.3",
			Preset:  config.PresetStartup,
		},
		LogDecisions: false,
	}

	executor := loop.NewExecutor(cfg, client)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 1, result.State.SuccessfulIterations)
	// No council invocations since no conflict
	assert.Equal(t, 0, result.State.CouncilInvocations)
	// Only one call (the main iteration)
	assert.Equal(t, 1, client.callCount)
}

func TestCouncilIntegration_ConflictDetected(t *testing.T) {
	// Test that council is invoked when conflict is detected
	client := &councilMockClient{
		responses: []*loop.IterationResult{
			{
				// First iteration with conflict
				Output:   "I found a PRINCIPLE_CONFLICT_UNRESOLVED between speed and correctness",
				Cost:     0.05,
				Duration: 1 * time.Second,
			},
			{
				// Council resolution response
				Output:   "**Decision**: Prioritize speed\n**Rationale**: Startup preset favors iteration speed",
				Cost:     0.03,
				Duration: 500 * time.Millisecond,
			},
		},
	}

	cfg := &loop.Config{
		Prompt:               "Test task",
		MaxRuns:              1,
		CompletionSignal:     "COMPLETE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
		Principles: &config.Principles{
			Version: "2.3",
			Preset:  config.PresetStartup,
		},
		LogDecisions: false,
	}

	executor := loop.NewExecutor(cfg, client)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 1, result.State.SuccessfulIterations)
	// Council was invoked once
	assert.Equal(t, 1, result.State.CouncilInvocations)
	// Cost includes both iteration and council
	assert.Equal(t, 0.08, result.State.TotalCost)
	assert.Equal(t, 0.03, result.State.CouncilCost)
	// Two calls: main iteration + council
	assert.Equal(t, 2, client.callCount)
}

func TestCouncilIntegration_DecisionLogging(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, ".claude", "principles-decisions.log")

	client := &councilMockClient{
		responses: []*loop.IterationResult{
			{
				Output:   "**Decision**: Use enterprise config\n**Rationale**: Security requirements",
				Cost:     0.05,
				Duration: 1 * time.Second,
			},
		},
	}

	// We need to change to tmpDir for the log file path to work correctly
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tmpDir)

	cfg := &loop.Config{
		Prompt:               "Test task",
		MaxRuns:              1,
		CompletionSignal:     "COMPLETE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
		Principles: &config.Principles{
			Version: "2.3",
			Preset:  config.PresetEnterprise,
		},
		LogDecisions: true,
	}

	executor := loop.NewExecutor(cfg, client)
	_, err := executor.Run(context.Background())

	require.NoError(t, err)

	// Check that log file was created
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	contentStr := string(content)
	assert.Contains(t, contentStr, "Use enterprise config")
	assert.Contains(t, contentStr, "Security requirements")
	assert.Contains(t, contentStr, "preset: \"enterprise\"")
	assert.Contains(t, contentStr, "council_invoked: false")
}

func TestCouncilIntegration_NoPrinciples(t *testing.T) {
	// Test that council is not initialized when no principles are loaded
	client := &councilMockClient{
		responses: []*loop.IterationResult{
			{
				Output:   "PRINCIPLE_CONFLICT_UNRESOLVED should not trigger council",
				Cost:     0.05,
				Duration: 1 * time.Second,
			},
		},
	}

	cfg := &loop.Config{
		Prompt:               "Test task",
		MaxRuns:              1,
		CompletionSignal:     "COMPLETE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
		Principles:           nil, // No principles loaded
		LogDecisions:         false,
	}

	executor := loop.NewExecutor(cfg, client)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
	// No council invocations since council not initialized
	assert.Equal(t, 0, result.State.CouncilInvocations)
	assert.Equal(t, float64(0), result.State.CouncilCost)
	// Only one call (the main iteration)
	assert.Equal(t, 1, client.callCount)
}

func TestCouncilIntegration_MultipleConflicts(t *testing.T) {
	// Test multiple iterations with conflicts
	client := &councilMockClient{
		responses: []*loop.IterationResult{
			{
				Output:   "First iteration: cannot resolve principle conflict",
				Cost:     0.05,
				Duration: 1 * time.Second,
			},
			{
				Output:   "**Decision**: Choice A\n**Rationale**: Reason A",
				Cost:     0.02,
				Duration: 500 * time.Millisecond,
			},
			{
				Output:   "Second iteration: conflicting principles remain unresolved",
				Cost:     0.05,
				Duration: 1 * time.Second,
			},
			{
				Output:   "**Decision**: Choice B\n**Rationale**: Reason B",
				Cost:     0.02,
				Duration: 500 * time.Millisecond,
			},
		},
	}

	cfg := &loop.Config{
		Prompt:               "Test task",
		MaxRuns:              2,
		CompletionSignal:     "COMPLETE",
		CompletionThreshold:  3,
		MaxConsecutiveErrors: 3,
		Principles: &config.Principles{
			Version: "2.3",
			Preset:  config.PresetStartup,
		},
		LogDecisions: false,
	}

	executor := loop.NewExecutor(cfg, client)
	result, err := executor.Run(context.Background())

	require.NoError(t, err)
	assert.Equal(t, loop.StopReasonMaxRuns, result.StopReason)
	assert.Equal(t, 2, result.State.SuccessfulIterations)
	assert.Equal(t, 2, result.State.CouncilInvocations)
	assert.Equal(t, 0.04, result.State.CouncilCost)
	// 4 calls: 2 iterations + 2 council resolutions
	assert.Equal(t, 4, client.callCount)
}

func TestCouncilIntegration_ConflictPatterns(t *testing.T) {
	patterns := []string{
		"PRINCIPLE_CONFLICT_UNRESOLVED",
		"I cannot resolve this principle issue",
		"There are conflicting principles that remain unresolved",
	}

	for _, pattern := range patterns {
		t.Run(strings.ReplaceAll(pattern, " ", "_"), func(t *testing.T) {
			client := &councilMockClient{
				responses: []*loop.IterationResult{
					{
						Output:   pattern,
						Cost:     0.05,
						Duration: 1 * time.Second,
					},
					{
						Output:   "**Decision**: Resolved\n**Rationale**: Applied priority",
						Cost:     0.02,
						Duration: 500 * time.Millisecond,
					},
				},
			}

			cfg := &loop.Config{
				Prompt:               "Test task",
				MaxRuns:              1,
				CompletionSignal:     "COMPLETE",
				CompletionThreshold:  3,
				MaxConsecutiveErrors: 3,
				Principles: &config.Principles{
					Version: "2.3",
					Preset:  config.PresetStartup,
				},
				LogDecisions: false,
			}

			executor := loop.NewExecutor(cfg, client)
			result, err := executor.Run(context.Background())

			require.NoError(t, err)
			assert.Equal(t, 1, result.State.CouncilInvocations, "Pattern should trigger council: %s", pattern)
		})
	}
}
