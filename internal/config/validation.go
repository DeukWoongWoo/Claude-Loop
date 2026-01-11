package config

import (
	"errors"
	"fmt"
	"regexp"
)

// ValidationError represents a config validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

const (
	MinPrincipleValue = 1
	MaxPrincipleValue = 10
)

var (
	versionRegex = regexp.MustCompile(`^\d+\.\d+$`)
	dateRegex    = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

// Validate checks the Principles for validity.
// Returns nil if valid, or the first error encountered.
func (p *Principles) Validate() error {
	if err := p.validateVersion(); err != nil {
		return err
	}
	if err := p.validatePreset(); err != nil {
		return err
	}
	if err := p.validateCreatedAt(); err != nil {
		return err
	}
	if err := p.validateLayer0(); err != nil {
		return err
	}
	if err := p.validateLayer1(); err != nil {
		return err
	}
	return nil
}

// ValidateAll runs all validation checks and returns all errors found.
func (p *Principles) ValidateAll() []error {
	var errs []error

	if err := p.validateVersion(); err != nil {
		errs = append(errs, err)
	}
	if err := p.validatePreset(); err != nil {
		errs = append(errs, err)
	}
	if err := p.validateCreatedAt(); err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, p.validateLayer0Fields()...)
	errs = append(errs, p.validateLayer1Fields()...)

	return errs
}

func (p *Principles) validateVersion() *ValidationError {
	if p.Version == "" {
		return &ValidationError{
			Field:   "version",
			Message: "version is required",
		}
	}
	if !versionRegex.MatchString(p.Version) {
		return &ValidationError{
			Field:   "version",
			Message: fmt.Sprintf("version must be in X.Y format (got %q)", p.Version),
		}
	}
	return nil
}

func (p *Principles) validatePreset() *ValidationError {
	if p.Preset == "" {
		return &ValidationError{
			Field:   "preset",
			Message: "preset is required",
		}
	}
	if !IsValidPreset(p.Preset) {
		return &ValidationError{
			Field:   "preset",
			Message: fmt.Sprintf("preset must be one of: startup, enterprise, opensource, custom (got %q)", p.Preset),
		}
	}
	return nil
}

func (p *Principles) validateCreatedAt() *ValidationError {
	if p.CreatedAt == "" {
		return &ValidationError{
			Field:   "created_at",
			Message: "created_at is required",
		}
	}
	if !dateRegex.MatchString(p.CreatedAt) {
		return &ValidationError{
			Field:   "created_at",
			Message: fmt.Sprintf("created_at must be in YYYY-MM-DD format (got %q)", p.CreatedAt),
		}
	}
	return nil
}

func validatePrincipleValue(field string, value int) *ValidationError {
	if value < MinPrincipleValue || value > MaxPrincipleValue {
		return &ValidationError{
			Field:   field,
			Message: fmt.Sprintf("%s must be between %d and %d (got %d)", field, MinPrincipleValue, MaxPrincipleValue, value),
		}
	}
	return nil
}

// principleField represents a named principle value for validation.
type principleField struct {
	name  string
	value int
}

func (p *Principles) layer0Fields() []principleField {
	l := &p.Layer0
	return []principleField{
		{"layer0.trust_architecture", l.TrustArchitecture},
		{"layer0.curation_model", l.CurationModel},
		{"layer0.scope_philosophy", l.ScopePhilosophy},
		{"layer0.monetization_model", l.MonetizationModel},
		{"layer0.privacy_posture", l.PrivacyPosture},
		{"layer0.ux_philosophy", l.UXPhilosophy},
		{"layer0.authority_stance", l.AuthorityStance},
		{"layer0.auditability", l.Auditability},
		{"layer0.interoperability", l.Interoperability},
	}
}

func (p *Principles) layer1Fields() []principleField {
	l := &p.Layer1
	return []principleField{
		{"layer1.speed_correctness", l.SpeedCorrectness},
		{"layer1.innovation_stability", l.InnovationStability},
		{"layer1.blast_radius", l.BlastRadius},
		{"layer1.clarity_of_intent", l.ClarityOfIntent},
		{"layer1.reversibility_priority", l.ReversibilityPriority},
		{"layer1.security_posture", l.SecurityPosture},
		{"layer1.urgency_tiers", l.UrgencyTiers},
		{"layer1.cost_efficiency", l.CostEfficiency},
		{"layer1.migration_burden", l.MigrationBurden},
	}
}

func (p *Principles) validateLayer0() *ValidationError {
	for _, f := range p.layer0Fields() {
		if err := validatePrincipleValue(f.name, f.value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Principles) validateLayer0Fields() []error {
	var errs []error
	for _, f := range p.layer0Fields() {
		if err := validatePrincipleValue(f.name, f.value); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func (p *Principles) validateLayer1() *ValidationError {
	for _, f := range p.layer1Fields() {
		if err := validatePrincipleValue(f.name, f.value); err != nil {
			return err
		}
	}
	return nil
}

func (p *Principles) validateLayer1Fields() []error {
	var errs []error
	for _, f := range p.layer1Fields() {
		if err := validatePrincipleValue(f.name, f.value); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// IsValidationError checks if an error is a ValidationError.
func IsValidationError(err error) bool {
	var ve *ValidationError
	return errors.As(err, &ve)
}
