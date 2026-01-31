package prd

import "fmt"

// Validator validates PRD structure and content.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks PRD for completeness and validity.
// Returns nil if valid, or the first error encountered.
func (v *Validator) Validate(prd *PRD) error {
	if prd == nil {
		return &ValidationError{Field: "prd", Message: "PRD is nil"}
	}
	if err := v.validateGoals(prd); err != nil {
		return err
	}
	if err := v.validateRequirements(prd); err != nil {
		return err
	}
	return v.validateSuccessCriteria(prd)
}

// ValidateAll returns all validation errors found.
func (v *Validator) ValidateAll(prd *PRD) []error {
	if prd == nil {
		return []error{&ValidationError{Field: "prd", Message: "PRD is nil"}}
	}

	var errs []error
	if err := v.validateGoals(prd); err != nil {
		errs = append(errs, err)
	}
	if err := v.validateRequirements(prd); err != nil {
		errs = append(errs, err)
	}
	if err := v.validateSuccessCriteria(prd); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, v.validateNoEmptyStrings(prd)...)
	return errs
}

func (v *Validator) validateGoals(prd *PRD) error {
	if len(prd.Goals) == 0 {
		return &ValidationError{Field: "goals", Message: "at least one goal is required"}
	}
	return nil
}

func (v *Validator) validateRequirements(prd *PRD) error {
	if len(prd.Requirements) == 0 {
		return &ValidationError{Field: "requirements", Message: "at least one requirement is required"}
	}
	return nil
}

func (v *Validator) validateSuccessCriteria(prd *PRD) error {
	if len(prd.SuccessCriteria) == 0 {
		return &ValidationError{Field: "success_criteria", Message: "at least one success criterion is required"}
	}
	return nil
}

func (v *Validator) validateNoEmptyStrings(prd *PRD) []error {
	var errs []error
	for i, goal := range prd.Goals {
		if goal == "" {
			errs = append(errs, &ValidationError{
				Field:   "goals",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	for i, req := range prd.Requirements {
		if req == "" {
			errs = append(errs, &ValidationError{
				Field:   "requirements",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	for i, criteria := range prd.SuccessCriteria {
		if criteria == "" {
			errs = append(errs, &ValidationError{
				Field:   "success_criteria",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	return errs
}
