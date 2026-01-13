package reviewer

import (
	"errors"
	"fmt"
)

// ReviewerError represents an error during reviewer pass execution.
type ReviewerError struct {
	Phase   string // "config", "prompt", or "execute"
	Message string
	Err     error
}

func (e *ReviewerError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("reviewer %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("reviewer %s: %s", e.Phase, e.Message)
}

func (e *ReviewerError) Unwrap() error {
	return e.Err
}

// IsReviewerError checks if an error is a ReviewerError.
func IsReviewerError(err error) bool {
	var re *ReviewerError
	return errors.As(err, &re)
}

// Predefined errors.
var (
	// ErrNoReviewPrompt indicates no review prompt was provided.
	ErrNoReviewPrompt = &ReviewerError{Phase: "prompt", Message: "no review prompt provided"}
)
