# Principles YAML Schema

> Schema definition for `.claude/principles.yaml` configuration file.

## Overview

The principles file defines project-specific decision-making guidelines for Claude Loop.
It uses a 1-10 scale for each principle, where values indicate preference direction.

## File Location

- Default: `.claude/principles.yaml`
- Custom: `--principles-file <path>`

---

## Schema Definition

```yaml
# Required: Schema version
version: "2.3"

# Required: Preset type
preset: "startup" | "enterprise" | "opensource" | "custom"

# Required: Creation timestamp
created_at: "YYYY-MM-DD"

# Required: Layer 0 - Product Principles (9 principles)
layer0:
  trust_architecture: 1-10
  curation_model: 1-10
  scope_philosophy: 1-10
  monetization_model: 1-10
  privacy_posture: 1-10
  ux_philosophy: 1-10
  authority_stance: 1-10
  auditability: 1-10
  interoperability: 1-10

# Required: Layer 1 - Development Principles (9 principles)
layer1:
  speed_correctness: 1-10
  innovation_stability: 1-10
  blast_radius: 1-10
  clarity_of_intent: 1-10
  reversibility_priority: 1-10
  security_posture: 1-10
  urgency_tiers: 1-10
  cost_efficiency: 1-10
  migration_burden: 1-10
```

---

## Layer 0: Product Principles

| Principle | Low (1-3) | High (7-10) |
|-----------|-----------|-------------|
| `trust_architecture` | Permissive | Strict verification |
| `curation_model` | Algorithm-driven | Human judgment |
| `scope_philosophy` | MVP focus | Feature-rich |
| `monetization_model` | Free/freemium | Premium/B2B |
| `privacy_posture` | Data collection | Minimal data |
| `ux_philosophy` | Power users | Simple-first |
| `authority_stance` | User freedom | Strong guidance |
| `auditability` | Minimal logging | Full audit trail |
| `interoperability` | Closed system | Open integrations |

---

## Layer 1: Development Principles

| Principle | Low (1-3) | High (7-10) |
|-----------|-----------|-------------|
| `speed_correctness` | Speed first | Correctness first |
| `innovation_stability` | New tech | Proven tech |
| `blast_radius` | Large changes | Small changes |
| `clarity_of_intent` | Implicit | Explicit |
| `reversibility_priority` | Permanent | Easy rollback |
| `security_posture` | Basic | Maximum security |
| `urgency_tiers` | All urgent | Normal pace |
| `cost_efficiency` | Build everything | Use external tools |
| `migration_burden` | Heavy migration OK | Avoid migrations |

---

## Presets

### Startup/MVP
```yaml
preset: startup
layer0:
  trust_architecture: 7
  curation_model: 6
  scope_philosophy: 3      # MVP focus
  monetization_model: 5
  privacy_posture: 7
  ux_philosophy: 4
  authority_stance: 6
  auditability: 5
  interoperability: 7
layer1:
  speed_correctness: 4     # Favor speed
  innovation_stability: 6
  blast_radius: 7
  clarity_of_intent: 6
  reversibility_priority: 7
  security_posture: 7
  urgency_tiers: 3
  cost_efficiency: 6
  migration_burden: 5
```

### Enterprise/Production
```yaml
preset: enterprise
layer0:
  trust_architecture: 9    # High verification
  curation_model: 8
  scope_philosophy: 7      # Feature-rich
  monetization_model: 8
  privacy_posture: 9       # Max privacy
  ux_philosophy: 6
  authority_stance: 7
  auditability: 9          # Full audit
  interoperability: 6
layer1:
  speed_correctness: 8     # Favor correctness
  innovation_stability: 8  # Proven tech
  blast_radius: 9          # Tiny changes
  clarity_of_intent: 8
  reversibility_priority: 9
  security_posture: 9
  urgency_tiers: 5
  cost_efficiency: 5
  migration_burden: 7
```

### Open Source/Library
```yaml
preset: opensource
layer0:
  trust_architecture: 6
  curation_model: 5
  scope_philosophy: 6
  monetization_model: 2    # Free/open
  privacy_posture: 8
  ux_philosophy: 5
  authority_stance: 4      # User freedom
  auditability: 7
  interoperability: 9      # Max openness
layer1:
  speed_correctness: 7
  innovation_stability: 7
  blast_radius: 8
  clarity_of_intent: 9     # Very explicit
  reversibility_priority: 8
  security_posture: 8
  urgency_tiers: 4
  cost_efficiency: 7
  migration_burden: 8      # Avoid breaking changes
```

---

## Validation Rules

1. **version**: Must be string, format "X.Y"
2. **preset**: Must be one of: `startup`, `enterprise`, `opensource`, `custom`
3. **created_at**: Must be string, format "YYYY-MM-DD"
4. **layer0/layer1**: All 9 principles required for each layer
5. **Values**: Must be integers 1-10

---

## Go Type Definition

```go
type Principles struct {
    Version   string `yaml:"version"`
    Preset    string `yaml:"preset"`
    CreatedAt string `yaml:"created_at"`
    Layer0    Layer0 `yaml:"layer0"`
    Layer1    Layer1 `yaml:"layer1"`
}

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

type Layer1 struct {
    SpeedCorrectness     int `yaml:"speed_correctness"`
    InnovationStability  int `yaml:"innovation_stability"`
    BlastRadius          int `yaml:"blast_radius"`
    ClarityOfIntent      int `yaml:"clarity_of_intent"`
    ReversibilityPriority int `yaml:"reversibility_priority"`
    SecurityPosture      int `yaml:"security_posture"`
    UrgencyTiers         int `yaml:"urgency_tiers"`
    CostEfficiency       int `yaml:"cost_efficiency"`
    MigrationBurden      int `yaml:"migration_burden"`
}
```

---

## Decision Log Format

When `--log-decisions` is enabled, decisions are logged to `.claude/principles-decisions.log`:

```
[2026-01-11T14:30:00] DECISION: <description>
  Principles considered: <principle1>=<value>, <principle2>=<value>
  Outcome: <chosen action>
  Rationale: <explanation>
```
