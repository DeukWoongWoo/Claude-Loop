package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultPrinciples(t *testing.T) {
	tests := []struct {
		name   string
		preset Preset
		verify func(*testing.T, *Principles)
	}{
		{
			name:   "startup preset",
			preset: PresetStartup,
			verify: func(t *testing.T, p *Principles) {
				assert.Equal(t, PresetStartup, p.Preset)
				assert.Equal(t, "2.3", p.Version)
				assert.Equal(t, 3, p.Layer0.ScopePhilosophy)
				assert.Equal(t, 4, p.Layer1.SpeedCorrectness)
			},
		},
		{
			name:   "enterprise preset",
			preset: PresetEnterprise,
			verify: func(t *testing.T, p *Principles) {
				assert.Equal(t, PresetEnterprise, p.Preset)
				assert.Equal(t, "2.3", p.Version)
				assert.Equal(t, 9, p.Layer0.Auditability)
				assert.Equal(t, 9, p.Layer1.BlastRadius)
			},
		},
		{
			name:   "opensource preset",
			preset: PresetOpenSource,
			verify: func(t *testing.T, p *Principles) {
				assert.Equal(t, PresetOpenSource, p.Preset)
				assert.Equal(t, "2.3", p.Version)
				assert.Equal(t, 9, p.Layer0.Interoperability)
				assert.Equal(t, 9, p.Layer1.ClarityOfIntent)
			},
		},
		{
			name:   "custom falls back to startup",
			preset: PresetCustom,
			verify: func(t *testing.T, p *Principles) {
				startup := startupDefaults()
				assert.Equal(t, startup.Layer0, p.Layer0)
				assert.Equal(t, startup.Layer1, p.Layer1)
			},
		},
		{
			name:   "unknown falls back to startup",
			preset: Preset("unknown"),
			verify: func(t *testing.T, p *Principles) {
				startup := startupDefaults()
				assert.Equal(t, startup.Layer0, p.Layer0)
				assert.Equal(t, startup.Layer1, p.Layer1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DefaultPrinciples(tt.preset)
			require.NotNil(t, p)
			tt.verify(t, p)
		})
	}
}

func TestDefaultPrinciplesAllValuesInRange(t *testing.T) {
	for _, preset := range ValidPresets {
		t.Run(string(preset), func(t *testing.T) {
			p := DefaultPrinciples(preset)
			// Set CreatedAt for validation (defaults don't include it as it's runtime-dependent)
			p.CreatedAt = "2026-01-11"
			err := p.Validate()
			assert.NoError(t, err, "preset %s should have valid defaults", preset)
		})
	}
}

func TestNewPrinciples(t *testing.T) {
	p := NewPrinciples()
	require.NotNil(t, p)
	assert.Equal(t, "2.3", p.Version)
	assert.Equal(t, PresetCustom, p.Preset)
	assert.Empty(t, p.CreatedAt)
	assert.Equal(t, Layer0{}, p.Layer0)
	assert.Equal(t, Layer1{}, p.Layer1)
}

func TestStartupDefaultsMatchSchema(t *testing.T) {
	p := startupDefaults()
	assert.Equal(t, 7, p.Layer0.TrustArchitecture)
	assert.Equal(t, 6, p.Layer0.CurationModel)
	assert.Equal(t, 3, p.Layer0.ScopePhilosophy)
	assert.Equal(t, 5, p.Layer0.MonetizationModel)
	assert.Equal(t, 7, p.Layer0.PrivacyPosture)
	assert.Equal(t, 4, p.Layer0.UXPhilosophy)
	assert.Equal(t, 6, p.Layer0.AuthorityStance)
	assert.Equal(t, 5, p.Layer0.Auditability)
	assert.Equal(t, 7, p.Layer0.Interoperability)

	assert.Equal(t, 4, p.Layer1.SpeedCorrectness)
	assert.Equal(t, 6, p.Layer1.InnovationStability)
	assert.Equal(t, 7, p.Layer1.BlastRadius)
	assert.Equal(t, 6, p.Layer1.ClarityOfIntent)
	assert.Equal(t, 7, p.Layer1.ReversibilityPriority)
	assert.Equal(t, 7, p.Layer1.SecurityPosture)
	assert.Equal(t, 3, p.Layer1.UrgencyTiers)
	assert.Equal(t, 6, p.Layer1.CostEfficiency)
	assert.Equal(t, 5, p.Layer1.MigrationBurden)
}

func TestEnterpriseDefaultsMatchSchema(t *testing.T) {
	p := enterpriseDefaults()
	assert.Equal(t, 9, p.Layer0.TrustArchitecture)
	assert.Equal(t, 8, p.Layer0.CurationModel)
	assert.Equal(t, 7, p.Layer0.ScopePhilosophy)
	assert.Equal(t, 8, p.Layer0.MonetizationModel)
	assert.Equal(t, 9, p.Layer0.PrivacyPosture)
	assert.Equal(t, 6, p.Layer0.UXPhilosophy)
	assert.Equal(t, 7, p.Layer0.AuthorityStance)
	assert.Equal(t, 9, p.Layer0.Auditability)
	assert.Equal(t, 6, p.Layer0.Interoperability)

	assert.Equal(t, 8, p.Layer1.SpeedCorrectness)
	assert.Equal(t, 8, p.Layer1.InnovationStability)
	assert.Equal(t, 9, p.Layer1.BlastRadius)
	assert.Equal(t, 8, p.Layer1.ClarityOfIntent)
	assert.Equal(t, 9, p.Layer1.ReversibilityPriority)
	assert.Equal(t, 9, p.Layer1.SecurityPosture)
	assert.Equal(t, 5, p.Layer1.UrgencyTiers)
	assert.Equal(t, 5, p.Layer1.CostEfficiency)
	assert.Equal(t, 7, p.Layer1.MigrationBurden)
}

func TestOpensourceDefaultsMatchSchema(t *testing.T) {
	p := opensourceDefaults()
	assert.Equal(t, 6, p.Layer0.TrustArchitecture)
	assert.Equal(t, 5, p.Layer0.CurationModel)
	assert.Equal(t, 6, p.Layer0.ScopePhilosophy)
	assert.Equal(t, 2, p.Layer0.MonetizationModel)
	assert.Equal(t, 8, p.Layer0.PrivacyPosture)
	assert.Equal(t, 5, p.Layer0.UXPhilosophy)
	assert.Equal(t, 4, p.Layer0.AuthorityStance)
	assert.Equal(t, 7, p.Layer0.Auditability)
	assert.Equal(t, 9, p.Layer0.Interoperability)

	assert.Equal(t, 7, p.Layer1.SpeedCorrectness)
	assert.Equal(t, 7, p.Layer1.InnovationStability)
	assert.Equal(t, 8, p.Layer1.BlastRadius)
	assert.Equal(t, 9, p.Layer1.ClarityOfIntent)
	assert.Equal(t, 8, p.Layer1.ReversibilityPriority)
	assert.Equal(t, 8, p.Layer1.SecurityPosture)
	assert.Equal(t, 4, p.Layer1.UrgencyTiers)
	assert.Equal(t, 7, p.Layer1.CostEfficiency)
	assert.Equal(t, 8, p.Layer1.MigrationBurden)
}
