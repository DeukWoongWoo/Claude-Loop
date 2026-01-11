// Package config provides configuration loading for claude-loop principles.
package config

// Preset represents the principle preset type.
type Preset string

const (
	PresetStartup    Preset = "startup"
	PresetEnterprise Preset = "enterprise"
	PresetOpenSource Preset = "opensource"
	PresetCustom     Preset = "custom"
)

// ValidPresets is the list of valid preset values.
var ValidPresets = []Preset{PresetStartup, PresetEnterprise, PresetOpenSource, PresetCustom}

// Principles represents the complete principles.yaml structure.
type Principles struct {
	Version   string `yaml:"version"`
	Preset    Preset `yaml:"preset"`
	CreatedAt string `yaml:"created_at"`
	Layer0    Layer0 `yaml:"layer0"`
	Layer1    Layer1 `yaml:"layer1"`
}

// Layer0 contains Product Principles (9 principles).
type Layer0 struct {
	TrustArchitecture int `yaml:"trust_architecture"`
	CurationModel     int `yaml:"curation_model"`
	ScopePhilosophy   int `yaml:"scope_philosophy"`
	MonetizationModel int `yaml:"monetization_model"`
	PrivacyPosture    int `yaml:"privacy_posture"`
	UXPhilosophy      int `yaml:"ux_philosophy"`
	AuthorityStance   int `yaml:"authority_stance"`
	Auditability      int `yaml:"auditability"`
	Interoperability  int `yaml:"interoperability"`
}

// Layer1 contains Development Principles (9 principles).
type Layer1 struct {
	SpeedCorrectness      int `yaml:"speed_correctness"`
	InnovationStability   int `yaml:"innovation_stability"`
	BlastRadius           int `yaml:"blast_radius"`
	ClarityOfIntent       int `yaml:"clarity_of_intent"`
	ReversibilityPriority int `yaml:"reversibility_priority"`
	SecurityPosture       int `yaml:"security_posture"`
	UrgencyTiers          int `yaml:"urgency_tiers"`
	CostEfficiency        int `yaml:"cost_efficiency"`
	MigrationBurden       int `yaml:"migration_burden"`
}

// IsValidPreset checks if a preset value is valid.
func IsValidPreset(p Preset) bool {
	for _, valid := range ValidPresets {
		if p == valid {
			return true
		}
	}
	return false
}
