package prompt

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlaceholderConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		placeholder string
		expected    string
	}{
		{
			name:        "completion signal placeholder",
			placeholder: PlaceholderCompletionSignal,
			expected:    "COMPLETION_SIGNAL_PLACEHOLDER",
		},
		{
			name:        "principles YAML placeholder",
			placeholder: PlaceholderPrinciplesYAML,
			expected:    "PRINCIPLES_YAML_PLACEHOLDER",
		},
		{
			name:        "notes file placeholder",
			placeholder: PlaceholderNotesFile,
			expected:    "NOTES_FILE_PLACEHOLDER",
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.placeholder)
		})
	}
}

func TestTemplatesNotEmpty(t *testing.T) {
	t.Parallel()

	templates := map[string]string{
		"TemplateWorkflowContext":      TemplateWorkflowContext,
		"TemplatePrincipleCollection":  TemplatePrincipleCollection,
		"TemplateDecisionPrinciples":   TemplateDecisionPrinciples,
		"TemplateNotesUpdateExisting":  TemplateNotesUpdateExisting,
		"TemplateNotesCreateNew":       TemplateNotesCreateNew,
		"TemplateNotesGuidelines":      TemplateNotesGuidelines,
		"TemplateNotesContext":         TemplateNotesContext,
		"TemplateIterationNotes":       TemplateIterationNotes,
		"TemplateReviewerContext":      TemplateReviewerContext,
		"TemplateCIFixContext":         TemplateCIFixContext,
	}

	for name, template := range templates {
		name, template := name, template // capture range variables
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.NotEmpty(t, template, "%s should not be empty", name)
		})
	}
}

func TestTemplatesContainExpectedPlaceholders(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		template    string
		placeholder string
	}{
		{
			name:        "TemplateWorkflowContext contains completion signal",
			template:    TemplateWorkflowContext,
			placeholder: PlaceholderCompletionSignal,
		},
		{
			name:        "TemplateDecisionPrinciples contains principles YAML",
			template:    TemplateDecisionPrinciples,
			placeholder: PlaceholderPrinciplesYAML,
		},
		{
			name:        "TemplateNotesUpdateExisting contains notes file",
			template:    TemplateNotesUpdateExisting,
			placeholder: PlaceholderNotesFile,
		},
		{
			name:        "TemplateNotesCreateNew contains notes file",
			template:    TemplateNotesCreateNew,
			placeholder: PlaceholderNotesFile,
		},
		{
			name:        "TemplateNotesContext contains notes file",
			template:    TemplateNotesContext,
			placeholder: PlaceholderNotesFile,
		},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Contains(t, tt.template, tt.placeholder,
				"%s should contain placeholder %s", tt.name, tt.placeholder)
		})
	}
}

func TestTemplateWorkflowContextContent(t *testing.T) {
	t.Parallel()

	// Verify key content from bash script
	assert.Contains(t, TemplateWorkflowContext, "CONTINUOUS WORKFLOW CONTEXT")
	assert.Contains(t, TemplateWorkflowContext, "continuous development loop")
	assert.Contains(t, TemplateWorkflowContext, "Project Completion Signal")
	assert.Contains(t, TemplateWorkflowContext, "PRIMARY GOAL")
}

func TestTemplatePrincipleCollectionContent(t *testing.T) {
	t.Parallel()

	// Verify key content from bash script
	assert.Contains(t, TemplatePrincipleCollection, "PRINCIPLE COLLECTION REQUIRED")
	assert.Contains(t, TemplatePrincipleCollection, "AskUserQuestion")
	assert.Contains(t, TemplatePrincipleCollection, "Project Type")
	assert.Contains(t, TemplatePrincipleCollection, "Startup/MVP")
	assert.Contains(t, TemplatePrincipleCollection, "Enterprise/Production")
	assert.Contains(t, TemplatePrincipleCollection, "Open Source/Library")
	assert.Contains(t, TemplatePrincipleCollection, "PRINCIPLES_COLLECTED_SUCCESSFULLY")
}

func TestTemplateDecisionPrinciplesContent(t *testing.T) {
	t.Parallel()

	// Verify key content from bash script
	assert.Contains(t, TemplateDecisionPrinciples, "DECISION PRINCIPLES")
	assert.Contains(t, TemplateDecisionPrinciples, "Decision Protocol")
	assert.Contains(t, TemplateDecisionPrinciples, "Compatibility Check")
	assert.Contains(t, TemplateDecisionPrinciples, "Type Classification")
	assert.Contains(t, TemplateDecisionPrinciples, "Priority Resolution")
	assert.Contains(t, TemplateDecisionPrinciples, "```yaml")
}

func TestTemplateNotesGuidelinesContent(t *testing.T) {
	t.Parallel()

	// Verify key content from bash script
	assert.Contains(t, TemplateNotesGuidelines, "coordinate work across iterations")
	assert.Contains(t, TemplateNotesGuidelines, "concise and actionable")
	assert.Contains(t, TemplateNotesGuidelines, "should NOT include")
}

func TestTemplateCIFixContextContent(t *testing.T) {
	t.Parallel()

	// Verify key content from bash script
	assert.Contains(t, TemplateCIFixContext, "CI FAILURE FIX CONTEXT")
	assert.Contains(t, TemplateCIFixContext, "gh run list")
	assert.Contains(t, TemplateCIFixContext, "gh run view")
	assert.Contains(t, TemplateCIFixContext, "--log-failed")
}

func TestTemplatesDoNotContainUnwantedCharacters(t *testing.T) {
	t.Parallel()

	templates := []string{
		TemplateWorkflowContext,
		TemplatePrincipleCollection,
		TemplateDecisionPrinciples,
		TemplateNotesUpdateExisting,
		TemplateNotesCreateNew,
		TemplateNotesGuidelines,
		TemplateNotesContext,
		TemplateIterationNotes,
	}

	for i, template := range templates {
		i, template := i, template // capture range variables
		t.Run("template_"+string(rune('0'+i)), func(t *testing.T) {
			t.Parallel()
			// Check for common escape sequence issues
			assert.NotContains(t, template, "\\n", "should not contain literal \\n")
			assert.NotContains(t, template, "\\t", "should not contain literal \\t")
		})
	}
}

func TestPlaceholderReplacementCompatibility(t *testing.T) {
	t.Parallel()

	// Ensure placeholders can be replaced with strings.ReplaceAll
	signal := "MY_CUSTOM_SIGNAL"
	result := strings.ReplaceAll(TemplateWorkflowContext, PlaceholderCompletionSignal, signal)

	assert.Contains(t, result, signal)
	assert.NotContains(t, result, PlaceholderCompletionSignal)
}
