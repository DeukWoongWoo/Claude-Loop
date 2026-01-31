package verifier

import (
	"errors"
	"fmt"
)

// VerifierError represents an error during verification operations.
type VerifierError struct {
	Phase   string // "config", "parse", "check", "execute"
	Message string
	Err     error
}

func (e *VerifierError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("verifier %s: %s: %v", e.Phase, e.Message, e.Err)
	}
	return fmt.Sprintf("verifier %s: %s", e.Phase, e.Message)
}

func (e *VerifierError) Unwrap() error {
	return e.Err
}

// IsVerifierError checks if an error is a VerifierError.
func IsVerifierError(err error) bool {
	var ve *VerifierError
	return errors.As(err, &ve)
}

// CheckError represents a specific check failure.
type CheckError struct {
	CheckerType string
	Criterion   string
	Message     string
	Err         error
}

func (e *CheckError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("check %s failed for '%s': %s: %v", e.CheckerType, e.Criterion, e.Message, e.Err)
	}
	return fmt.Sprintf("check %s failed for '%s': %s", e.CheckerType, e.Criterion, e.Message)
}

func (e *CheckError) Unwrap() error {
	return e.Err
}

// IsCheckError checks if an error is a CheckError.
func IsCheckError(err error) bool {
	var ce *CheckError
	return errors.As(err, &ce)
}

// Predefined errors.
var (
	ErrNilTask        = &VerifierError{Phase: "config", Message: "task is nil"}
	ErrNoCriteria     = &VerifierError{Phase: "config", Message: "no success criteria to verify"}
	ErrNilClient      = &VerifierError{Phase: "config", Message: "claude client is nil (required for AI verification)"}
	ErrTimeout        = &VerifierError{Phase: "execute", Message: "verification timed out"}
	ErrNoCheckerFound = &VerifierError{Phase: "parse", Message: "no suitable checker found for criterion"}
)
