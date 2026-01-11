package loop

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionDetector_Detect(t *testing.T) {
	tests := []struct {
		name             string
		completionSignal string
		output           string
		want             bool
	}{
		{
			name:             "signal found",
			completionSignal: "COMPLETE",
			output:           "Task is COMPLETE now",
			want:             true,
		},
		{
			name:             "signal not found",
			completionSignal: "COMPLETE",
			output:           "Still working on it",
			want:             false,
		},
		{
			name:             "empty signal",
			completionSignal: "",
			output:           "COMPLETE",
			want:             false,
		},
		{
			name:             "signal at start",
			completionSignal: "DONE",
			output:           "DONE with the task",
			want:             true,
		},
		{
			name:             "signal at end",
			completionSignal: "FINISHED",
			output:           "All work FINISHED",
			want:             true,
		},
		{
			name:             "case sensitive - exact match",
			completionSignal: "Complete",
			output:           "Complete",
			want:             true,
		},
		{
			name:             "case sensitive - no match",
			completionSignal: "COMPLETE",
			output:           "complete",
			want:             false,
		},
		{
			name:             "default signal",
			completionSignal: "CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
			output:           "CONTINUOUS_CLAUDE_PROJECT_COMPLETE",
			want:             true,
		},
		{
			name:             "empty output",
			completionSignal: "DONE",
			output:           "",
			want:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{CompletionSignal: tt.completionSignal}
			detector := NewCompletionDetector(config)

			got := detector.Detect(tt.output)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCompletionDetector_CheckThreshold(t *testing.T) {
	tests := []struct {
		name        string
		threshold   int
		signalCount int
		wantReached bool
		wantReason  StopReason
	}{
		{
			name:        "under threshold",
			threshold:   3,
			signalCount: 1,
			wantReached: false,
		},
		{
			name:        "at threshold",
			threshold:   3,
			signalCount: 3,
			wantReached: true,
			wantReason:  StopReasonCompletionSignal,
		},
		{
			name:        "over threshold",
			threshold:   3,
			signalCount: 5,
			wantReached: true,
			wantReason:  StopReasonCompletionSignal,
		},
		{
			name:        "threshold disabled (0)",
			threshold:   0,
			signalCount: 100,
			wantReached: false,
		},
		{
			name:        "zero signal count",
			threshold:   3,
			signalCount: 0,
			wantReached: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{CompletionThreshold: tt.threshold}
			detector := NewCompletionDetector(config)
			state := &State{CompletionSignalCount: tt.signalCount}

			result := detector.CheckThreshold(state)

			assert.Equal(t, tt.wantReached, result.LimitReached)
			if tt.wantReached {
				assert.Equal(t, tt.wantReason, result.Reason)
			}
		})
	}
}

func TestCompletionDetector_UpdateState(t *testing.T) {
	tests := []struct {
		name             string
		initialCount     int
		signalFound      bool
		expectedCount    int
	}{
		{
			name:          "signal found increments",
			initialCount:  2,
			signalFound:   true,
			expectedCount: 3,
		},
		{
			name:          "signal not found resets",
			initialCount:  2,
			signalFound:   false,
			expectedCount: 0,
		},
		{
			name:          "first signal",
			initialCount:  0,
			signalFound:   true,
			expectedCount: 1,
		},
		{
			name:          "reset from zero",
			initialCount:  0,
			signalFound:   false,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{}
			detector := NewCompletionDetector(config)
			state := &State{CompletionSignalCount: tt.initialCount}

			detector.UpdateState(state, tt.signalFound)

			assert.Equal(t, tt.expectedCount, state.CompletionSignalCount)
		})
	}
}

func TestCompletionDetector_Integration(t *testing.T) {
	// Test a realistic sequence of outputs
	config := &Config{
		CompletionSignal:    "DONE",
		CompletionThreshold: 3,
	}
	detector := NewCompletionDetector(config)
	state := &State{}

	// Sequence: no signal, signal, no signal (reset), signal, signal, signal
	outputs := []struct {
		output      string
		expectCount int
	}{
		{"Working...", 0},
		{"DONE with part 1", 1},
		{"Working on part 2...", 0}, // Reset
		{"DONE with part 2", 1},
		{"DONE with part 3", 2},
		{"DONE with everything", 3},
	}

	for i, out := range outputs {
		found := detector.Detect(out.output)
		detector.UpdateState(state, found)
		assert.Equal(t, out.expectCount, state.CompletionSignalCount, "step %d", i)
	}

	// Now threshold should be reached
	result := detector.CheckThreshold(state)
	assert.True(t, result.LimitReached)
	assert.Equal(t, StopReasonCompletionSignal, result.Reason)
}

func TestNewCompletionDetector(t *testing.T) {
	config := &Config{CompletionSignal: "TEST"}
	detector := NewCompletionDetector(config)

	assert.NotNil(t, detector)
	assert.Equal(t, config, detector.config)
}
