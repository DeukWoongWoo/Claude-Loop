package loop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLimitChecker_Check_MaxRuns(t *testing.T) {
	tests := []struct {
		name         string
		maxRuns      int
		iterations   int
		wantReached  bool
		wantReason   StopReason
	}{
		{
			name:         "under limit",
			maxRuns:      5,
			iterations:   3,
			wantReached:  false,
			wantReason:   StopReasonNone,
		},
		{
			name:         "at limit",
			maxRuns:      5,
			iterations:   5,
			wantReached:  true,
			wantReason:   StopReasonMaxRuns,
		},
		{
			name:         "over limit",
			maxRuns:      5,
			iterations:   7,
			wantReached:  true,
			wantReason:   StopReasonMaxRuns,
		},
		{
			name:         "unlimited (0)",
			maxRuns:      0,
			iterations:   100,
			wantReached:  false,
			wantReason:   StopReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxRuns: tt.maxRuns}
			checker := NewLimitChecker(config)
			state := &State{SuccessfulIterations: tt.iterations}

			result := checker.Check(state)

			assert.Equal(t, tt.wantReached, result.LimitReached)
			if tt.wantReached {
				assert.Equal(t, tt.wantReason, result.Reason)
			}
		})
	}
}

func TestLimitChecker_Check_MaxCost(t *testing.T) {
	tests := []struct {
		name        string
		maxCost     float64
		totalCost   float64
		wantReached bool
		wantReason  StopReason
	}{
		{
			name:        "under limit",
			maxCost:     10.0,
			totalCost:   5.0,
			wantReached: false,
			wantReason:  StopReasonNone,
		},
		{
			name:        "at limit",
			maxCost:     10.0,
			totalCost:   10.0,
			wantReached: true,
			wantReason:  StopReasonMaxCost,
		},
		{
			name:        "over limit",
			maxCost:     10.0,
			totalCost:   15.0,
			wantReached: true,
			wantReason:  StopReasonMaxCost,
		},
		{
			name:        "unlimited (0)",
			maxCost:     0,
			totalCost:   1000.0,
			wantReached: false,
			wantReason:  StopReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxCost: tt.maxCost}
			checker := NewLimitChecker(config)
			state := &State{TotalCost: tt.totalCost}

			result := checker.Check(state)

			assert.Equal(t, tt.wantReached, result.LimitReached)
			if tt.wantReached {
				assert.Equal(t, tt.wantReason, result.Reason)
			}
		})
	}
}

func TestLimitChecker_Check_MaxDuration(t *testing.T) {
	tests := []struct {
		name        string
		maxDuration time.Duration
		elapsed     time.Duration
		wantReached bool
		wantReason  StopReason
	}{
		{
			name:        "under limit",
			maxDuration: 1 * time.Hour,
			elapsed:     30 * time.Minute,
			wantReached: false,
			wantReason:  StopReasonNone,
		},
		{
			name:        "at limit",
			maxDuration: 1 * time.Hour,
			elapsed:     1 * time.Hour,
			wantReached: true,
			wantReason:  StopReasonMaxDuration,
		},
		{
			name:        "over limit",
			maxDuration: 1 * time.Hour,
			elapsed:     2 * time.Hour,
			wantReached: true,
			wantReason:  StopReasonMaxDuration,
		},
		{
			name:        "unlimited (0)",
			maxDuration: 0,
			elapsed:     100 * time.Hour,
			wantReached: false,
			wantReason:  StopReasonNone,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxDuration: tt.maxDuration}
			checker := NewLimitChecker(config)
			state := &State{StartTime: time.Now().Add(-tt.elapsed)}

			result := checker.Check(state)

			assert.Equal(t, tt.wantReached, result.LimitReached)
			if tt.wantReached {
				assert.Equal(t, tt.wantReason, result.Reason)
			}
		})
	}
}

func TestLimitChecker_Check_Priority(t *testing.T) {
	// When multiple limits are exceeded, check order: runs > cost > duration
	config := &Config{
		MaxRuns:     5,
		MaxCost:     10.0,
		MaxDuration: 1 * time.Hour,
	}
	checker := NewLimitChecker(config)

	// All limits exceeded
	state := &State{
		SuccessfulIterations: 10,
		TotalCost:            20.0,
		StartTime:            time.Now().Add(-2 * time.Hour),
	}

	result := checker.Check(state)

	assert.True(t, result.LimitReached)
	assert.Equal(t, StopReasonMaxRuns, result.Reason) // Runs checked first
}

func TestLimitChecker_RemainingBudget(t *testing.T) {
	tests := []struct {
		name      string
		maxCost   float64
		totalCost float64
		want      float64
	}{
		{
			name:      "has remaining budget",
			maxCost:   10.0,
			totalCost: 3.0,
			want:      7.0,
		},
		{
			name:      "no remaining budget",
			maxCost:   10.0,
			totalCost: 10.0,
			want:      0.0,
		},
		{
			name:      "over budget",
			maxCost:   10.0,
			totalCost: 15.0,
			want:      0.0,
		},
		{
			name:      "unlimited",
			maxCost:   0,
			totalCost: 100.0,
			want:      -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxCost: tt.maxCost}
			checker := NewLimitChecker(config)
			state := &State{TotalCost: tt.totalCost}

			got := checker.RemainingBudget(state)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestLimitChecker_RemainingTime(t *testing.T) {
	tests := []struct {
		name        string
		maxDuration time.Duration
		elapsed     time.Duration
		wantPositive bool
		wantValue   time.Duration // Only checked if wantPositive is false
	}{
		{
			name:         "has remaining time",
			maxDuration:  1 * time.Hour,
			elapsed:      30 * time.Minute,
			wantPositive: true,
		},
		{
			name:         "no remaining time",
			maxDuration:  1 * time.Hour,
			elapsed:      1 * time.Hour,
			wantPositive: false,
			wantValue:    0,
		},
		{
			name:         "over time",
			maxDuration:  1 * time.Hour,
			elapsed:      2 * time.Hour,
			wantPositive: false,
			wantValue:    0,
		},
		{
			name:         "unlimited",
			maxDuration:  0,
			elapsed:      100 * time.Hour,
			wantPositive: false,
			wantValue:    -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxDuration: tt.maxDuration}
			checker := NewLimitChecker(config)
			state := &State{StartTime: time.Now().Add(-tt.elapsed)}

			got := checker.RemainingTime(state)

			if tt.wantPositive {
				assert.True(t, got > 0)
			} else {
				assert.Equal(t, tt.wantValue, got)
			}
		})
	}
}

func TestLimitChecker_RemainingRuns(t *testing.T) {
	tests := []struct {
		name       string
		maxRuns    int
		iterations int
		want       int
	}{
		{
			name:       "has remaining runs",
			maxRuns:    10,
			iterations: 3,
			want:       7,
		},
		{
			name:       "no remaining runs",
			maxRuns:    5,
			iterations: 5,
			want:       0,
		},
		{
			name:       "over runs",
			maxRuns:    5,
			iterations: 7,
			want:       0,
		},
		{
			name:       "unlimited",
			maxRuns:    0,
			iterations: 100,
			want:       -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &Config{MaxRuns: tt.maxRuns}
			checker := NewLimitChecker(config)
			state := &State{SuccessfulIterations: tt.iterations}

			got := checker.RemainingRuns(state)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewLimitChecker(t *testing.T) {
	config := &Config{MaxRuns: 5}
	checker := NewLimitChecker(config)

	assert.NotNil(t, checker)
	assert.Equal(t, config, checker.config)
}
