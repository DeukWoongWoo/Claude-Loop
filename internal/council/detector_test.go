package council

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConflictDetector_Detect(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name     string
		output   string
		expected bool
	}{
		{
			name:     "PRINCIPLE_CONFLICT_UNRESOLVED uppercase",
			output:   "The task has PRINCIPLE_CONFLICT_UNRESOLVED status",
			expected: true,
		},
		{
			name:     "PRINCIPLE_CONFLICT_UNRESOLVED lowercase",
			output:   "Found principle_conflict_unresolved in response",
			expected: true,
		},
		{
			name:     "cannot resolve principle",
			output:   "I cannot resolve this principle conflict",
			expected: true,
		},
		{
			name:     "Cannot Resolve Principle mixed case",
			output:   "Cannot Resolve the Principle here",
			expected: true,
		},
		{
			name:     "conflicting principles unresolved",
			output:   "There are conflicting principles that remain unresolved",
			expected: true,
		},
		{
			name:     "no conflict pattern",
			output:   "Everything is fine, no conflicts here",
			expected: false,
		},
		{
			name:     "empty string",
			output:   "",
			expected: false,
		},
		{
			name:     "partial match - principle without conflict",
			output:   "This is about principles in general",
			expected: false,
		},
		{
			name:     "partial match - conflict without principle",
			output:   "There is a merge conflict here",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.Detect(tt.output)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConflictDetector_ExtractDecision(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name              string
		output            string
		expectedDecision  string
		expectedRationale string
	}{
		{
			name: "both present",
			output: `Here is my analysis:
**Decision**: Use the startup preset for faster iteration
**Rationale**: Layer 1 speed_correctness favors speed (8/10)`,
			expectedDecision:  "Use the startup preset for faster iteration",
			expectedRationale: "Layer 1 speed_correctness favors speed (8/10)",
		},
		{
			name:              "only decision",
			output:            "**Decision**: Apply enterprise defaults",
			expectedDecision:  "Apply enterprise defaults",
			expectedRationale: "",
		},
		{
			name:              "only rationale",
			output:            "**Rationale**: Based on user requirements",
			expectedDecision:  "",
			expectedRationale: "Based on user requirements",
		},
		{
			name:              "neither present",
			output:            "Just some regular text without patterns",
			expectedDecision:  "",
			expectedRationale: "",
		},
		{
			name:              "empty string",
			output:            "",
			expectedDecision:  "",
			expectedRationale: "",
		},
		{
			name: "with extra whitespace",
			output: `**Decision**:    Trimmed decision
**Rationale**:   Trimmed rationale  `,
			expectedDecision:  "Trimmed decision",
			expectedRationale: "Trimmed rationale",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision, rationale := detector.ExtractDecision(tt.output)
			assert.Equal(t, tt.expectedDecision, decision)
			assert.Equal(t, tt.expectedRationale, rationale)
		})
	}
}

func TestNewConflictDetector(t *testing.T) {
	detector := NewConflictDetector()
	assert.NotNil(t, detector)
}
