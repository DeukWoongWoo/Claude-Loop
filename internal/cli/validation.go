package cli

import (
	"errors"
	"fmt"
)

// ValidationError represents a CLI validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// validatePrompt checks if prompt is provided when required.
func (f *Flags) validatePrompt() *ValidationError {
	if f.Prompt == "" {
		return &ValidationError{
			Field:   "prompt",
			Message: "prompt is required: use -p or --prompt",
		}
	}
	return nil
}

// validateLimit checks if at least one execution limit is provided.
func (f *Flags) validateLimit() *ValidationError {
	hasLimit := f.MaxRuns > 0 || f.MaxCost > 0 || f.MaxDuration > 0
	if !hasLimit {
		return &ValidationError{
			Field:   "limit",
			Message: "at least one limit required: use -m/--max-runs, --max-cost, or --max-duration",
		}
	}
	return nil
}

// validateMergeStrategy checks if merge strategy is valid.
func (f *Flags) validateMergeStrategy() *ValidationError {
	if f.MergeStrategy == "" {
		return nil
	}

	switch f.MergeStrategy {
	case "squash", "merge", "rebase":
		return nil
	default:
		return &ValidationError{
			Field:   "merge-strategy",
			Message: fmt.Sprintf("merge-strategy must be squash, merge, or rebase (got %q)", f.MergeStrategy),
		}
	}
}

// validatePlanningFlags checks planning mode flag combinations.
func (f *Flags) validatePlanningFlags() *ValidationError {
	// --plan-only and --resume are mutually exclusive
	if f.PlanOnly && f.Resume != "" {
		return &ValidationError{
			Field:   "planning",
			Message: "--plan-only and --resume cannot be used together",
		}
	}

	// --plan-only requires --prompt
	if f.PlanOnly && f.Prompt == "" {
		return &ValidationError{
			Field:   "planning",
			Message: "--plan-only requires --prompt",
		}
	}

	// --plan requires --prompt (unless resuming)
	if f.Plan && f.Prompt == "" && f.Resume == "" {
		return &ValidationError{
			Field:   "planning",
			Message: "--plan requires --prompt (or use --resume to continue a saved plan)",
		}
	}

	return nil
}

// validateNonNegative checks that numeric values are not negative.
func (f *Flags) validateNonNegative() *ValidationError {
	if f.MaxRuns < 0 {
		return &ValidationError{
			Field:   "max-runs",
			Message: "max-runs cannot be negative",
		}
	}
	if f.MaxCost < 0 {
		return &ValidationError{
			Field:   "max-cost",
			Message: "max-cost cannot be negative",
		}
	}
	if f.MaxDuration < 0 {
		return &ValidationError{
			Field:   "max-duration",
			Message: "max-duration cannot be negative",
		}
	}
	if f.CIRetryMax < 0 {
		return &ValidationError{
			Field:   "ci-retry-max",
			Message: "ci-retry-max cannot be negative",
		}
	}
	if f.CompletionThreshold < 0 {
		return &ValidationError{
			Field:   "completion-threshold",
			Message: "completion-threshold cannot be negative",
		}
	}
	return nil
}

// Validate checks the Flags for validity according to CLI_CONTRACT.md rules.
// Returns nil if valid, or the first error encountered.
func (f *Flags) Validate() error {
	if f.ListWorktrees {
		return nil
	}

	// Planning mode has different validation rules
	if f.isPlanningMode() {
		return f.validateForPlanningMode()
	}

	if err := f.validatePrompt(); err != nil {
		return err
	}
	if err := f.validateNonNegative(); err != nil {
		return err
	}
	if err := f.validateLimit(); err != nil {
		return err
	}
	if err := f.validateMergeStrategy(); err != nil {
		return err
	}

	return nil
}

// isPlanningMode checks if any planning mode flag is set.
func (f *Flags) isPlanningMode() bool {
	return f.Plan || f.PlanOnly || f.Resume != ""
}

// validateForPlanningMode validates flags specific to planning mode.
func (f *Flags) validateForPlanningMode() error {
	// Check planning-specific flags first
	if err := f.validatePlanningFlags(); err != nil {
		return err
	}

	// --resume doesn't require --prompt
	if f.Resume != "" {
		// Only validate non-negative values
		if err := f.validateNonNegative(); err != nil {
			return err
		}
		return nil
	}

	// --plan and --plan-only require --prompt
	if err := f.validatePrompt(); err != nil {
		return err
	}
	if err := f.validateNonNegative(); err != nil {
		return err
	}

	// Planning mode doesn't require limits (plan phases handle their own execution)
	// But if limits are provided, they should be valid
	if err := f.validateMergeStrategy(); err != nil {
		return err
	}

	return nil
}

// ValidateAll runs all validation checks and returns all errors found.
func (f *Flags) ValidateAll() []error {
	if f.ListWorktrees {
		return nil
	}

	var errs []error

	// Planning mode validation
	if f.isPlanningMode() {
		if err := f.validatePlanningFlags(); err != nil {
			errs = append(errs, err)
		}
		// For --resume, skip prompt validation
		if f.Resume == "" {
			if err := f.validatePrompt(); err != nil {
				errs = append(errs, err)
			}
		}
		if err := f.validateNonNegative(); err != nil {
			errs = append(errs, err)
		}
		if err := f.validateMergeStrategy(); err != nil {
			errs = append(errs, err)
		}
		return errs
	}

	if err := f.validatePrompt(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateNonNegative(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateLimit(); err != nil {
		errs = append(errs, err)
	}
	if err := f.validateMergeStrategy(); err != nil {
		errs = append(errs, err)
	}

	return errs
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
