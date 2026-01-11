package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate(t *testing.T) {
	// Helper to create valid principles from default with CreatedAt set
	validPrinciples := func(preset Preset) *Principles {
		p := DefaultPrinciples(preset)
		p.CreatedAt = "2026-01-11"
		return p
	}

	tests := []struct {
		name       string
		principles *Principles
		wantErr    string
	}{
		{
			name:       "valid startup principles",
			principles: validPrinciples(PresetStartup),
			wantErr:    "",
		},
		{
			name:       "valid enterprise principles",
			principles: validPrinciples(PresetEnterprise),
			wantErr:    "",
		},
		{
			name:       "valid opensource principles",
			principles: validPrinciples(PresetOpenSource),
			wantErr:    "",
		},
		{
			name: "missing version",
			principles: &Principles{
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "version is required",
		},
		{
			name: "invalid version format with v prefix",
			principles: &Principles{
				Version:   "v2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "version must be in X.Y format",
		},
		{
			name: "invalid version format single number",
			principles: &Principles{
				Version:   "2",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "version must be in X.Y format",
		},
		{
			name: "missing preset",
			principles: &Principles{
				Version:   "2.3",
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "preset is required",
		},
		{
			name: "invalid preset",
			principles: &Principles{
				Version:   "2.3",
				Preset:    "invalid",
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "preset must be one of",
		},
		{
			name: "missing created_at",
			principles: &Principles{
				Version: "2.3",
				Preset:  PresetStartup,
				Layer0:  DefaultPrinciples(PresetStartup).Layer0,
				Layer1:  DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "created_at is required",
		},
		{
			name: "invalid date format MM-DD-YYYY",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "01-11-2026",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "created_at must be in YYYY-MM-DD format",
		},
		{
			name: "layer0 value too low (0)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    Layer0{TrustArchitecture: 0, CurationModel: 5, ScopePhilosophy: 5, MonetizationModel: 5, PrivacyPosture: 5, UXPhilosophy: 5, AuthorityStance: 5, Auditability: 5, Interoperability: 5},
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "must be between 1 and 10",
		},
		{
			name: "layer0 value too high (11)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    Layer0{TrustArchitecture: 11, CurationModel: 5, ScopePhilosophy: 5, MonetizationModel: 5, PrivacyPosture: 5, UXPhilosophy: 5, AuthorityStance: 5, Auditability: 5, Interoperability: 5},
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			},
			wantErr: "must be between 1 and 10",
		},
		{
			name: "layer1 value too low (0)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    Layer1{SpeedCorrectness: 0, InnovationStability: 5, BlastRadius: 5, ClarityOfIntent: 5, ReversibilityPriority: 5, SecurityPosture: 5, UrgencyTiers: 5, CostEfficiency: 5, MigrationBurden: 5},
			},
			wantErr: "must be between 1 and 10",
		},
		{
			name: "layer1 value too high (11)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    Layer1{SpeedCorrectness: 11, InnovationStability: 5, BlastRadius: 5, ClarityOfIntent: 5, ReversibilityPriority: 5, SecurityPosture: 5, UrgencyTiers: 5, CostEfficiency: 5, MigrationBurden: 5},
			},
			wantErr: "must be between 1 and 10",
		},
		{
			name: "edge case: value at minimum (1)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    Layer0{TrustArchitecture: 1, CurationModel: 1, ScopePhilosophy: 1, MonetizationModel: 1, PrivacyPosture: 1, UXPhilosophy: 1, AuthorityStance: 1, Auditability: 1, Interoperability: 1},
				Layer1:    Layer1{SpeedCorrectness: 1, InnovationStability: 1, BlastRadius: 1, ClarityOfIntent: 1, ReversibilityPriority: 1, SecurityPosture: 1, UrgencyTiers: 1, CostEfficiency: 1, MigrationBurden: 1},
			},
			wantErr: "",
		},
		{
			name: "edge case: value at maximum (10)",
			principles: &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    Layer0{TrustArchitecture: 10, CurationModel: 10, ScopePhilosophy: 10, MonetizationModel: 10, PrivacyPosture: 10, UXPhilosophy: 10, AuthorityStance: 10, Auditability: 10, Interoperability: 10},
				Layer1:    Layer1{SpeedCorrectness: 10, InnovationStability: 10, BlastRadius: 10, ClarityOfIntent: 10, ReversibilityPriority: 10, SecurityPosture: 10, UrgencyTiers: 10, CostEfficiency: 10, MigrationBurden: 10},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.principles.Validate()
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

func TestValidateAll(t *testing.T) {
	t.Run("valid principles returns empty slice", func(t *testing.T) {
		p := DefaultPrinciples(PresetStartup)
		p.CreatedAt = "2026-01-11"
		errs := p.ValidateAll()
		assert.Empty(t, errs)
	})

	t.Run("multiple errors are collected", func(t *testing.T) {
		p := &Principles{
			// missing version, preset, created_at
			Layer0: Layer0{TrustArchitecture: 0}, // invalid
			Layer1: Layer1{SpeedCorrectness: 11}, // invalid
		}

		errs := p.ValidateAll()
		assert.GreaterOrEqual(t, len(errs), 3)
	})

	t.Run("all layer0 errors collected", func(t *testing.T) {
		p := &Principles{
			Version:   "2.3",
			Preset:    PresetStartup,
			CreatedAt: "2026-01-11",
			Layer0:    Layer0{}, // all zeros
			Layer1:    DefaultPrinciples(PresetStartup).Layer1,
		}

		errs := p.ValidateAll()
		assert.Len(t, errs, 9) // 9 layer0 fields with value 0
	})

	t.Run("all layer1 errors collected", func(t *testing.T) {
		p := &Principles{
			Version:   "2.3",
			Preset:    PresetStartup,
			CreatedAt: "2026-01-11",
			Layer0:    DefaultPrinciples(PresetStartup).Layer0,
			Layer1:    Layer1{}, // all zeros
		}

		errs := p.ValidateAll()
		assert.Len(t, errs, 9) // 9 layer1 fields with value 0
	})
}

func TestValidationErrorFields(t *testing.T) {
	tests := []struct {
		name       string
		principles *Principles
		wantField  string
	}{
		{
			name:       "version field",
			principles: &Principles{Preset: PresetStartup, CreatedAt: "2026-01-11"},
			wantField:  "version",
		},
		{
			name:       "preset field",
			principles: &Principles{Version: "2.3", CreatedAt: "2026-01-11"},
			wantField:  "preset",
		},
		{
			name:       "created_at field",
			principles: &Principles{Version: "2.3", Preset: PresetStartup},
			wantField:  "created_at",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.principles.Validate()
			require.Error(t, err)

			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Equal(t, tt.wantField, ve.Field)
		})
	}
}

func TestIsValidationError(t *testing.T) {
	t.Run("ValidationError returns true", func(t *testing.T) {
		ve := &ValidationError{Field: "test", Message: "test error"}
		assert.True(t, IsValidationError(ve))
	})

	t.Run("other error returns false", func(t *testing.T) {
		err := fmt.Errorf("not a validation error")
		assert.False(t, IsValidationError(err))
	})

	t.Run("nil returns false", func(t *testing.T) {
		assert.False(t, IsValidationError(nil))
	})
}

func TestValidationConstants(t *testing.T) {
	assert.Equal(t, 1, MinPrincipleValue)
	assert.Equal(t, 10, MaxPrincipleValue)
}

func TestVersionRegex(t *testing.T) {
	validVersions := []string{"2.3", "1.0", "10.20", "0.1"}
	invalidVersions := []string{"v2.3", "2", "2.3.1", "a.b", "", "2.", ".3"}

	for _, v := range validVersions {
		t.Run("valid: "+v, func(t *testing.T) {
			p := &Principles{
				Version:   v,
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
				Layer0:    DefaultPrinciples(PresetStartup).Layer0,
				Layer1:    DefaultPrinciples(PresetStartup).Layer1,
			}
			err := p.validateVersion()
			assert.Nil(t, err)
		})
	}

	for _, v := range invalidVersions {
		t.Run("invalid: "+v, func(t *testing.T) {
			p := &Principles{
				Version:   v,
				Preset:    PresetStartup,
				CreatedAt: "2026-01-11",
			}
			err := p.validateVersion()
			assert.NotNil(t, err)
		})
	}
}

func TestDateRegex(t *testing.T) {
	validDates := []string{"2026-01-11", "2000-12-31", "1999-01-01"}
	invalidDates := []string{"01-11-2026", "2026/01/11", "2026-1-11", "2026-01-1", ""}

	for _, d := range validDates {
		t.Run("valid: "+d, func(t *testing.T) {
			p := &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: d,
			}
			err := p.validateCreatedAt()
			assert.Nil(t, err)
		})
	}

	for _, d := range invalidDates {
		t.Run("invalid: "+d, func(t *testing.T) {
			p := &Principles{
				Version:   "2.3",
				Preset:    PresetStartup,
				CreatedAt: d,
			}
			err := p.validateCreatedAt()
			assert.NotNil(t, err)
		})
	}
}
