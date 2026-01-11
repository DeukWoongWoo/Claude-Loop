package config

// DefaultVersion is the current schema version for principles files.
const DefaultVersion = "2.3"

// DefaultPrinciples returns a Principles struct with default values for the given preset.
// Note: CreatedAt is not set by default as it should be set to the actual creation time.
// Callers must set CreatedAt before calling Validate().
func DefaultPrinciples(preset Preset) *Principles {
	switch preset {
	case PresetStartup:
		return startupDefaults()
	case PresetEnterprise:
		return enterpriseDefaults()
	case PresetOpenSource:
		return opensourceDefaults()
	default:
		return startupDefaults()
	}
}

func startupDefaults() *Principles {
	return &Principles{
		Version: DefaultVersion,
		Preset:  PresetStartup,
		Layer0: Layer0{
			TrustArchitecture: 7,
			CurationModel:     6,
			ScopePhilosophy:   3,
			MonetizationModel: 5,
			PrivacyPosture:    7,
			UXPhilosophy:      4,
			AuthorityStance:   6,
			Auditability:      5,
			Interoperability:  7,
		},
		Layer1: Layer1{
			SpeedCorrectness:      4,
			InnovationStability:   6,
			BlastRadius:           7,
			ClarityOfIntent:       6,
			ReversibilityPriority: 7,
			SecurityPosture:       7,
			UrgencyTiers:          3,
			CostEfficiency:        6,
			MigrationBurden:       5,
		},
	}
}

func enterpriseDefaults() *Principles {
	return &Principles{
		Version: DefaultVersion,
		Preset:  PresetEnterprise,
		Layer0: Layer0{
			TrustArchitecture: 9,
			CurationModel:     8,
			ScopePhilosophy:   7,
			MonetizationModel: 8,
			PrivacyPosture:    9,
			UXPhilosophy:      6,
			AuthorityStance:   7,
			Auditability:      9,
			Interoperability:  6,
		},
		Layer1: Layer1{
			SpeedCorrectness:      8,
			InnovationStability:   8,
			BlastRadius:           9,
			ClarityOfIntent:       8,
			ReversibilityPriority: 9,
			SecurityPosture:       9,
			UrgencyTiers:          5,
			CostEfficiency:        5,
			MigrationBurden:       7,
		},
	}
}

func opensourceDefaults() *Principles {
	return &Principles{
		Version: DefaultVersion,
		Preset:  PresetOpenSource,
		Layer0: Layer0{
			TrustArchitecture: 6,
			CurationModel:     5,
			ScopePhilosophy:   6,
			MonetizationModel: 2,
			PrivacyPosture:    8,
			UXPhilosophy:      5,
			AuthorityStance:   4,
			Auditability:      7,
			Interoperability:  9,
		},
		Layer1: Layer1{
			SpeedCorrectness:      7,
			InnovationStability:   7,
			BlastRadius:           8,
			ClarityOfIntent:       9,
			ReversibilityPriority: 8,
			SecurityPosture:       8,
			UrgencyTiers:          4,
			CostEfficiency:        7,
			MigrationBurden:       8,
		},
	}
}

// NewPrinciples creates an empty Principles struct for custom configuration.
func NewPrinciples() *Principles {
	return &Principles{
		Version: DefaultVersion,
		Preset:  PresetCustom,
	}
}
