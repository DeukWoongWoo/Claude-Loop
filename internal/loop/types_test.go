package loop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	before := time.Now()
	state := NewState()
	after := time.Now()

	assert.NotNil(t, state)
	assert.Equal(t, 0, state.SuccessfulIterations)
	assert.Equal(t, 0, state.TotalIterations)
	assert.Equal(t, 0, state.ErrorCount)
	assert.Equal(t, 0, state.CompletionSignalCount)
	assert.Equal(t, 0.0, state.TotalCost)
	assert.True(t, state.StartTime.After(before) || state.StartTime.Equal(before))
	assert.True(t, state.StartTime.Before(after) || state.StartTime.Equal(after))
}

func TestState_Elapsed(t *testing.T) {
	state := NewState()
	time.Sleep(10 * time.Millisecond)
	elapsed := state.Elapsed()

	assert.True(t, elapsed >= 10*time.Millisecond)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "CONTINUOUS_CLAUDE_PROJECT_COMPLETE", config.CompletionSignal)
	assert.Equal(t, 3, config.CompletionThreshold)
	assert.Equal(t, 3, config.MaxConsecutiveErrors)
	assert.Empty(t, config.Prompt)
	assert.Equal(t, 0, config.MaxRuns)
	assert.Equal(t, 0.0, config.MaxCost)
	assert.Equal(t, time.Duration(0), config.MaxDuration)
	assert.False(t, config.DryRun)
	assert.Nil(t, config.OnProgress)
}

func TestStopReason_Values(t *testing.T) {
	// Verify all stop reasons are distinct and have expected values
	reasons := []StopReason{
		StopReasonNone,
		StopReasonMaxRuns,
		StopReasonMaxCost,
		StopReasonMaxDuration,
		StopReasonCompletionSignal,
		StopReasonConsecutiveErrors,
		StopReasonContextCancelled,
	}

	// Check uniqueness
	seen := make(map[StopReason]bool)
	for _, r := range reasons {
		assert.False(t, seen[r], "duplicate stop reason: %s", r)
		seen[r] = true
	}

	// Check specific values
	assert.Equal(t, StopReason(""), StopReasonNone)
	assert.Equal(t, StopReason("max_runs_reached"), StopReasonMaxRuns)
	assert.Equal(t, StopReason("max_cost_reached"), StopReasonMaxCost)
	assert.Equal(t, StopReason("max_duration_reached"), StopReasonMaxDuration)
	assert.Equal(t, StopReason("completion_signal"), StopReasonCompletionSignal)
	assert.Equal(t, StopReason("consecutive_errors"), StopReasonConsecutiveErrors)
	assert.Equal(t, StopReason("context_cancelled"), StopReasonContextCancelled)
}

func TestLoopResult(t *testing.T) {
	state := NewState()
	state.SuccessfulIterations = 5

	result := &LoopResult{
		State:      state,
		StopReason: StopReasonMaxRuns,
		LastError:  nil,
	}

	assert.Equal(t, 5, result.State.SuccessfulIterations)
	assert.Equal(t, StopReasonMaxRuns, result.StopReason)
	assert.Nil(t, result.LastError)
}
