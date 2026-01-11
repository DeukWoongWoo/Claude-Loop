package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValidPreset(t *testing.T) {
	tests := []struct {
		name   string
		preset Preset
		want   bool
	}{
		{"startup is valid", PresetStartup, true},
		{"enterprise is valid", PresetEnterprise, true},
		{"opensource is valid", PresetOpenSource, true},
		{"custom is valid", PresetCustom, true},
		{"invalid preset", Preset("invalid"), false},
		{"empty preset", Preset(""), false},
		{"similar but wrong", Preset("Startup"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidPreset(tt.preset)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPresetConstants(t *testing.T) {
	assert.Equal(t, Preset("startup"), PresetStartup)
	assert.Equal(t, Preset("enterprise"), PresetEnterprise)
	assert.Equal(t, Preset("opensource"), PresetOpenSource)
	assert.Equal(t, Preset("custom"), PresetCustom)
}

func TestValidPresetsContainsAllConstants(t *testing.T) {
	assert.Contains(t, ValidPresets, PresetStartup)
	assert.Contains(t, ValidPresets, PresetEnterprise)
	assert.Contains(t, ValidPresets, PresetOpenSource)
	assert.Contains(t, ValidPresets, PresetCustom)
	assert.Len(t, ValidPresets, 4)
}
