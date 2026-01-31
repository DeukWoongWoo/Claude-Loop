package architecture

import "fmt"

// Validator validates Architecture structure and content.
type Validator struct{}

// NewValidator creates a new Validator.
func NewValidator() *Validator {
	return &Validator{}
}

// Validate checks Architecture for completeness and validity.
// Returns nil if valid, or the first error encountered.
func (v *Validator) Validate(arch *Architecture) error {
	if arch == nil {
		return &ValidationError{Field: "architecture", Message: "Architecture is nil"}
	}
	if err := v.validateComponents(arch); err != nil {
		return err
	}
	if err := v.validateFileStructure(arch); err != nil {
		return err
	}
	return v.validateTechDecisions(arch)
}

// ValidateAll returns all validation errors found.
func (v *Validator) ValidateAll(arch *Architecture) []error {
	if arch == nil {
		return []error{&ValidationError{Field: "architecture", Message: "Architecture is nil"}}
	}

	var errs []error
	if err := v.validateComponents(arch); err != nil {
		errs = append(errs, err)
	}
	if err := v.validateFileStructure(arch); err != nil {
		errs = append(errs, err)
	}
	if err := v.validateTechDecisions(arch); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, v.validateComponentsComplete(arch)...)
	errs = append(errs, v.validateNoEmptyStrings(arch)...)
	return errs
}

func (v *Validator) validateComponents(arch *Architecture) error {
	if len(arch.Components) == 0 {
		return &ValidationError{Field: "components", Message: "at least one component is required"}
	}
	return nil
}

func (v *Validator) validateFileStructure(arch *Architecture) error {
	if len(arch.FileStructure) == 0 {
		return &ValidationError{Field: "file_structure", Message: "at least one file structure entry is required"}
	}
	return nil
}

func (v *Validator) validateTechDecisions(arch *Architecture) error {
	if len(arch.TechDecisions) == 0 {
		return &ValidationError{Field: "tech_decisions", Message: "at least one technical decision is required"}
	}
	return nil
}

// validateComponentsComplete checks each component has required fields.
func (v *Validator) validateComponentsComplete(arch *Architecture) []error {
	var errs []error
	for i, comp := range arch.Components {
		if comp.Name == "" {
			errs = append(errs, &ValidationError{
				Field:   "components",
				Message: fmt.Sprintf("component at index %d has empty name", i),
			})
		}
		if comp.Description == "" {
			errs = append(errs, &ValidationError{
				Field:   "components",
				Message: fmt.Sprintf("component '%s' at index %d has empty description", comp.Name, i),
			})
		}
	}
	return errs
}

// validateNoEmptyStrings checks that lists don't contain empty strings.
func (v *Validator) validateNoEmptyStrings(arch *Architecture) []error {
	var errs []error
	for i, entry := range arch.FileStructure {
		if entry == "" {
			errs = append(errs, &ValidationError{
				Field:   "file_structure",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	for i, decision := range arch.TechDecisions {
		if decision == "" {
			errs = append(errs, &ValidationError{
				Field:   "tech_decisions",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	for i, dep := range arch.Dependencies {
		if dep == "" {
			errs = append(errs, &ValidationError{
				Field:   "dependencies",
				Message: fmt.Sprintf("empty string at index %d", i),
			})
		}
	}
	return errs
}
