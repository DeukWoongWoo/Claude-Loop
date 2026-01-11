package loop

import "time"

// LimitChecker checks if any execution limits have been reached.
type LimitChecker struct {
	config *Config
}

// NewLimitChecker creates a new LimitChecker.
func NewLimitChecker(config *Config) *LimitChecker {
	return &LimitChecker{config: config}
}

// CheckResult represents the result of a limit check.
type CheckResult struct {
	LimitReached bool
	Reason       StopReason
}

// Check evaluates all limits against the current state.
// Returns the first limit reached, or a result with LimitReached=false if none exceeded.
func (lc *LimitChecker) Check(state *State) *CheckResult {
	if result := lc.checkRunsLimit(state); result.LimitReached {
		return result
	}
	if result := lc.checkCostLimit(state); result.LimitReached {
		return result
	}
	if result := lc.checkDurationLimit(state); result.LimitReached {
		return result
	}
	return &CheckResult{LimitReached: false}
}

func (lc *LimitChecker) checkRunsLimit(state *State) *CheckResult {
	if lc.config.MaxRuns > 0 && state.SuccessfulIterations >= lc.config.MaxRuns {
		return &CheckResult{
			LimitReached: true,
			Reason:       StopReasonMaxRuns,
		}
	}
	return &CheckResult{LimitReached: false}
}

func (lc *LimitChecker) checkCostLimit(state *State) *CheckResult {
	if lc.config.MaxCost > 0 && state.TotalCost >= lc.config.MaxCost {
		return &CheckResult{
			LimitReached: true,
			Reason:       StopReasonMaxCost,
		}
	}
	return &CheckResult{LimitReached: false}
}

func (lc *LimitChecker) checkDurationLimit(state *State) *CheckResult {
	if lc.config.MaxDuration > 0 && time.Since(state.StartTime) >= lc.config.MaxDuration {
		return &CheckResult{
			LimitReached: true,
			Reason:       StopReasonMaxDuration,
		}
	}
	return &CheckResult{LimitReached: false}
}

// RemainingBudget returns how much cost budget remains.
// Returns -1 if no cost limit is set.
func (lc *LimitChecker) RemainingBudget(state *State) float64 {
	if lc.config.MaxCost <= 0 {
		return -1
	}
	remaining := lc.config.MaxCost - state.TotalCost
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RemainingTime returns how much time remains.
// Returns -1 if no duration limit is set.
func (lc *LimitChecker) RemainingTime(state *State) time.Duration {
	if lc.config.MaxDuration <= 0 {
		return -1
	}
	elapsed := time.Since(state.StartTime)
	remaining := lc.config.MaxDuration - elapsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RemainingRuns returns how many runs remain.
// Returns -1 if no run limit is set.
func (lc *LimitChecker) RemainingRuns(state *State) int {
	if lc.config.MaxRuns <= 0 {
		return -1
	}
	remaining := lc.config.MaxRuns - state.SuccessfulIterations
	if remaining < 0 {
		return 0
	}
	return remaining
}
