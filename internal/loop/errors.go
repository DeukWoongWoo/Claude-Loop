package loop

import (
	"errors"
	"fmt"
)

// LoopError represents a loop-level error.
type LoopError struct {
	Field   string
	Message string
	Err     error
}

func (e *LoopError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Field, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func (e *LoopError) Unwrap() error {
	return e.Err
}

// IterationError represents an error in a single iteration.
type IterationError struct {
	Iteration int
	Message   string
	Err       error
}

func (e *IterationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("iteration %d: %s: %v", e.Iteration, e.Message, e.Err)
	}
	return fmt.Sprintf("iteration %d: %s", e.Iteration, e.Message)
}

func (e *IterationError) Unwrap() error {
	return e.Err
}

// IsLoopError checks if an error is a LoopError.
func IsLoopError(err error) bool {
	var le *LoopError
	return errors.As(err, &le)
}

// IsIterationError checks if an error is an IterationError.
func IsIterationError(err error) bool {
	var ie *IterationError
	return errors.As(err, &ie)
}

// Predefined errors for common stop conditions.
var (
	ErrMaxRunsReached     = &LoopError{Field: "max_runs", Message: "maximum runs reached"}
	ErrMaxCostReached     = &LoopError{Field: "max_cost", Message: "maximum cost reached"}
	ErrMaxDurationReached = &LoopError{Field: "max_duration", Message: "maximum duration reached"}
	ErrConsecutiveErrors  = &LoopError{Field: "errors", Message: "too many consecutive errors"}
	ErrCompletionSignal   = &LoopError{Field: "completion", Message: "completion signal threshold reached"}
)
