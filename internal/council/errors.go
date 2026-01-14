package council

import (
	"errors"
	"fmt"
)

// CouncilError represents an error during council operation.
type CouncilError struct {
	Phase   string // "detect", "resolve", "log", "prompt"
	Message string
	Err     error
}

func (e *CouncilError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("council %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("council %s: %s", e.Phase, e.Message)
}

func (e *CouncilError) Unwrap() error {
	return e.Err
}

// IsCouncilError checks if an error is a CouncilError.
func IsCouncilError(err error) bool {
	var ce *CouncilError
	return errors.As(err, &ce)
}

// Predefined errors.
var (
	ErrNoPrinciples = &CouncilError{Phase: "resolve", Message: "no principles configured"}
)
