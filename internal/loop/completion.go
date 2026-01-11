package loop

import "strings"

// CompletionDetector detects completion signals in Claude output.
type CompletionDetector struct {
	config *Config
}

// NewCompletionDetector creates a new CompletionDetector.
func NewCompletionDetector(config *Config) *CompletionDetector {
	return &CompletionDetector{config: config}
}

// Detect checks if the output contains a completion signal.
func (cd *CompletionDetector) Detect(output string) bool {
	if cd.config.CompletionSignal == "" {
		return false
	}
	return strings.Contains(output, cd.config.CompletionSignal)
}

// CheckThreshold evaluates if the completion threshold has been met.
// Returns a CheckResult indicating if the loop should stop.
func (cd *CompletionDetector) CheckThreshold(state *State) *CheckResult {
	if cd.config.CompletionThreshold > 0 &&
		state.CompletionSignalCount >= cd.config.CompletionThreshold {
		return &CheckResult{
			LimitReached: true,
			Reason:       StopReasonCompletionSignal,
		}
	}
	return &CheckResult{LimitReached: false}
}

// UpdateState updates the state based on whether a completion signal was found.
// If found, increments the counter; if not, resets it to zero.
func (cd *CompletionDetector) UpdateState(state *State, signalFound bool) {
	if signalFound {
		state.CompletionSignalCount++
	} else {
		state.CompletionSignalCount = 0
	}
}
