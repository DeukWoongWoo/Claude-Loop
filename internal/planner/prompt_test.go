package planner

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptBuilder(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()
	require.NotNil(t, builder)
}

func TestPromptBuilder_BuildPRDPrompt(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()

	t.Run("basic prompt", func(t *testing.T) {
		t.Parallel()
		prompt := builder.BuildPRDPrompt("Add payment feature")

		assert.Contains(t, prompt, "Add payment feature")
		assert.Contains(t, prompt, "PRODUCT REQUIREMENTS DOCUMENT")
		assert.Contains(t, prompt, "Goals")
		assert.Contains(t, prompt, "Requirements")
		assert.Contains(t, prompt, "Constraints")
		assert.Contains(t, prompt, "Success Criteria")
		assert.Contains(t, prompt, "PLANNING CONSTRAINTS")
	})

	t.Run("empty prompt", func(t *testing.T) {
		t.Parallel()
		prompt := builder.BuildPRDPrompt("")

		assert.Contains(t, prompt, "PRODUCT REQUIREMENTS DOCUMENT")
		// Should still produce valid template
	})

	t.Run("multiline prompt", func(t *testing.T) {
		t.Parallel()
		userPrompt := `Add authentication feature:
1. Login with email
2. Password reset
3. Session management`

		prompt := builder.BuildPRDPrompt(userPrompt)

		assert.Contains(t, prompt, "Login with email")
		assert.Contains(t, prompt, "Password reset")
		assert.Contains(t, prompt, "Session management")
	})
}

func TestPromptBuilder_BuildArchitecturePrompt(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()

	t.Run("with valid PRD", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{
			Goals:           []string{"Goal 1", "Goal 2"},
			Requirements:    []string{"Req 1", "Req 2"},
			Constraints:     []string{"Constraint 1"},
			SuccessCriteria: []string{"Criteria 1"},
			RawOutput:       "raw",
		}

		prompt, err := builder.BuildArchitecturePrompt(prd)

		require.NoError(t, err)
		assert.Contains(t, prompt, "ARCHITECTURE DESIGN PHASE")
		assert.Contains(t, prompt, "Goal 1")
		assert.Contains(t, prompt, "Goal 2")
		assert.Contains(t, prompt, "Req 1")
		assert.Contains(t, prompt, "Constraint 1")
		assert.Contains(t, prompt, "Criteria 1")
		assert.Contains(t, prompt, "Components")
		assert.Contains(t, prompt, "Dependencies")
		assert.Contains(t, prompt, "PLANNING CONSTRAINTS")
	})

	t.Run("with nil PRD", func(t *testing.T) {
		t.Parallel()
		prompt, err := builder.BuildArchitecturePrompt(nil)

		require.Error(t, err)
		assert.Empty(t, prompt)
		assert.True(t, IsPlannerError(err))

		var pe *PlannerError
		require.ErrorAs(t, err, &pe)
		assert.Equal(t, "prompt", pe.Phase)
	})

	t.Run("with empty PRD", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{}

		prompt, err := builder.BuildArchitecturePrompt(prd)

		require.NoError(t, err)
		assert.Contains(t, prompt, "ARCHITECTURE DESIGN PHASE")
		// Empty lists should still work
	})
}

func TestPromptBuilder_BuildTasksPrompt(t *testing.T) {
	t.Parallel()

	builder := NewPromptBuilder()

	t.Run("with valid Architecture", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Components: []Component{
				{Name: "PaymentService", Description: "Handles payments", Files: []string{"payment.go"}},
				{Name: "AuthService", Description: "Handles auth", Files: []string{"auth.go"}},
			},
			Dependencies:  []string{"stripe-go", "jwt-go"},
			FileStructure: []string{"internal/services/payment/", "internal/services/auth/"},
			TechDecisions: []string{"Use Stripe for payments"},
			RawOutput:     "raw",
		}

		prompt, err := builder.BuildTasksPrompt(arch)

		require.NoError(t, err)
		assert.Contains(t, prompt, "TASK DECOMPOSITION PHASE")
		assert.Contains(t, prompt, "PaymentService")
		assert.Contains(t, prompt, "Handles payments")
		assert.Contains(t, prompt, "payment.go")
		assert.Contains(t, prompt, "stripe-go")
		assert.Contains(t, prompt, "internal/services/payment/")
		assert.Contains(t, prompt, "Use Stripe for payments")
		assert.Contains(t, prompt, "T001")
		assert.Contains(t, prompt, "PLANNING CONSTRAINTS")
	})

	t.Run("with nil Architecture", func(t *testing.T) {
		t.Parallel()
		prompt, err := builder.BuildTasksPrompt(nil)

		require.Error(t, err)
		assert.Empty(t, prompt)
		assert.True(t, IsPlannerError(err))

		var pe *PlannerError
		require.ErrorAs(t, err, &pe)
		assert.Equal(t, "prompt", pe.Phase)
	})

	t.Run("with empty Architecture", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{}

		prompt, err := builder.BuildTasksPrompt(arch)

		require.NoError(t, err)
		assert.Contains(t, prompt, "TASK DECOMPOSITION PHASE")
	})

	t.Run("component without files", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Components: []Component{
				{Name: "Service", Description: "A service", Files: nil},
			},
		}

		prompt, err := builder.BuildTasksPrompt(arch)

		require.NoError(t, err)
		assert.Contains(t, prompt, "Service")
		assert.Contains(t, prompt, "A service")
		// Should not have "Files:" line for empty files
	})
}

func TestFormatPRDSummary(t *testing.T) {
	t.Parallel()

	prd := &PRD{
		Goals:           []string{"Increase revenue", "Improve UX"},
		Requirements:    []string{"Payment processing", "User dashboard"},
		Constraints:     []string{"No breaking changes"},
		SuccessCriteria: []string{"100% test coverage"},
	}

	summary := formatPRDSummary(prd)

	assert.Contains(t, summary, "### Goals")
	assert.Contains(t, summary, "- Increase revenue")
	assert.Contains(t, summary, "- Improve UX")
	assert.Contains(t, summary, "### Requirements")
	assert.Contains(t, summary, "- Payment processing")
	assert.Contains(t, summary, "### Constraints")
	assert.Contains(t, summary, "- No breaking changes")
	assert.Contains(t, summary, "### Success Criteria")
	assert.Contains(t, summary, "- 100% test coverage")
}

func TestFormatArchitectureSummary(t *testing.T) {
	t.Parallel()

	arch := &Architecture{
		Components: []Component{
			{Name: "API", Description: "REST API layer", Files: []string{"api.go", "routes.go"}},
			{Name: "DB", Description: "Database layer", Files: []string{}},
		},
		Dependencies:  []string{"gin", "gorm"},
		FileStructure: []string{"internal/api/", "internal/db/"},
		TechDecisions: []string{"Use Gin for routing"},
	}

	summary := formatArchitectureSummary(arch)

	assert.Contains(t, summary, "### Components")
	assert.Contains(t, summary, "**API**: REST API layer")
	assert.Contains(t, summary, "Files: api.go, routes.go")
	assert.Contains(t, summary, "**DB**: Database layer")
	assert.Contains(t, summary, "### Dependencies")
	assert.Contains(t, summary, "- gin")
	assert.Contains(t, summary, "### File Structure")
	assert.Contains(t, summary, "- internal/api/")
	assert.Contains(t, summary, "### Technical Decisions")
	assert.Contains(t, summary, "- Use Gin for routing")
}

func TestTemplatePlanningConstraints(t *testing.T) {
	t.Parallel()

	assert.Contains(t, TemplatePlanningConstraints, "PLANNING CONSTRAINTS")
	assert.Contains(t, TemplatePlanningConstraints, "specific and actionable")
	assert.Contains(t, TemplatePlanningConstraints, "NEVER guess or assume")
	assert.Contains(t, TemplatePlanningConstraints, "OUTPUT FORMAT")
}

func TestTemplates_ContainConstraints(t *testing.T) {
	t.Parallel()

	templates := []struct {
		name     string
		template string
	}{
		{"TemplatePRDPhase", TemplatePRDPhase},
		{"TemplateArchitecturePhase", TemplateArchitecturePhase},
		{"TemplateTasksPhase", TemplateTasksPhase},
	}

	for _, tt := range templates {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, tt.template, "PLANNING CONSTRAINTS")
			// Check template has placeholder
			assert.True(t, strings.Contains(tt.template, "%s"))
		})
	}
}
