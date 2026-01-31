package prd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	t.Parallel()

	parser := NewParser()
	assert.NotNil(t, parser)
}

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		input           string
		wantGoals       []string
		wantReqs        []string
		wantConstraints []string
		wantCriteria    []string
		wantErr         error
	}{
		{
			name: "valid PRD output with bullet lists",
			input: `### Goals
- Build a user authentication system
- Support OAuth2 providers

### Requirements
- Implement login/logout functionality
- Store sessions securely

### Constraints
- Use existing database schema
- No external dependencies

### Success Criteria
- All tests pass
- Code coverage above 80%`,
			wantGoals:       []string{"Build a user authentication system", "Support OAuth2 providers"},
			wantReqs:        []string{"Implement login/logout functionality", "Store sessions securely"},
			wantConstraints: []string{"Use existing database schema", "No external dependencies"},
			wantCriteria:    []string{"All tests pass", "Code coverage above 80%"},
			wantErr:         nil,
		},
		{
			name: "missing goals",
			input: `### Requirements
- Implement feature

### Success Criteria
- Works`,
			wantErr: ErrParseNoGoals,
		},
		{
			name: "missing requirements",
			input: `### Goals
- Build something

### Success Criteria
- Works`,
			wantErr: ErrParseNoRequirements,
		},
		{
			name: "numbered list format",
			input: `### Goals
1. First goal
2. Second goal

### Requirements
1. First requirement

### Success Criteria
1. Passes`,
			wantGoals:    []string{"First goal", "Second goal"},
			wantReqs:     []string{"First requirement"},
			wantCriteria: []string{"Passes"},
		},
		{
			name: "asterisk list format",
			input: `### Goals
* Asterisk goal one
* Asterisk goal two

### Requirements
* Requirement one

### Success Criteria
* Criterion one`,
			wantGoals:    []string{"Asterisk goal one", "Asterisk goal two"},
			wantReqs:     []string{"Requirement one"},
			wantCriteria: []string{"Criterion one"},
		},
		{
			name: "mixed list formats",
			input: `### Goals
- Dash goal
* Asterisk goal
1. Numbered goal

### Requirements
- Requirement

### Success Criteria
- Criterion`,
			wantGoals:    []string{"Dash goal", "Asterisk goal", "Numbered goal"},
			wantReqs:     []string{"Requirement"},
			wantCriteria: []string{"Criterion"},
		},
		{
			name: "two hash headers",
			input: `## Goals
- Goal one

## Requirements
- Req one

## Success Criteria
- Criterion one`,
			wantGoals:    []string{"Goal one"},
			wantReqs:     []string{"Req one"},
			wantCriteria: []string{"Criterion one"},
		},
		{
			name: "empty constraints allowed",
			input: `### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
			wantGoals:       []string{"Goal"},
			wantReqs:        []string{"Req"},
			wantConstraints: nil,
			wantCriteria:    []string{"Criterion"},
		},
		{
			name: "extra whitespace handling",
			input: `### Goals

  - Goal with spaces

### Requirements
  -   Requirement with spaces

### Success Criteria
-    Criterion   `,
			wantGoals:    []string{"Goal with spaces"},
			wantReqs:     []string{"Requirement with spaces"},
			wantCriteria: []string{"Criterion"},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prd, err := parser.Parse(tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, prd)
			assert.Equal(t, tt.wantGoals, prd.Goals)
			assert.Equal(t, tt.wantReqs, prd.Requirements)
			assert.Equal(t, tt.wantConstraints, prd.Constraints)
			assert.Equal(t, tt.wantCriteria, prd.SuccessCriteria)
			assert.Equal(t, tt.input, prd.RawOutput)
		})
	}
}

func TestParser_ExtractExtendedFields(t *testing.T) {
	t.Parallel()

	input := `### Title: User Authentication System

### Summary
This PRD defines requirements for building a secure user authentication system.
It will support multiple authentication methods.

### Goals
- Implement secure login

### Requirements
- Support password auth

### Out of Scope
- Social login
- SSO integration

### Success Criteria
- Security audit passes`

	parser := NewParser()
	prd, err := parser.Parse(input)

	require.NoError(t, err)
	assert.Equal(t, "User Authentication System", prd.Title)
	assert.Contains(t, prd.Summary, "secure user authentication")
	assert.Equal(t, []string{"Social login", "SSO integration"}, prd.OutOfScope)
}

func TestParser_TitleVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantTitle string
	}{
		{
			name: "title with colon",
			input: `### Title: My Title

### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
			wantTitle: "My Title",
		},
		{
			name: "PRD Title format",
			input: `### PRD Title: Another Title

### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
			wantTitle: "Another Title",
		},
		{
			name: "title with newline",
			input: `### Title
My Newline Title

### Goals
- Goal

### Requirements
- Req

### Success Criteria
- Criterion`,
			wantTitle: "My Newline Title",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			prd, err := parser.Parse(tt.input)
			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, prd.Title)
		})
	}
}

func TestParser_NonGoalsVariation(t *testing.T) {
	t.Parallel()

	input := `### Goals
- Goal

### Requirements
- Req

### Non-Goals
- Not doing this
- Not doing that

### Success Criteria
- Done`

	parser := NewParser()
	prd, err := parser.Parse(input)

	require.NoError(t, err)
	assert.Equal(t, []string{"Not doing this", "Not doing that"}, prd.OutOfScope)
}

func TestParser_ExtractListItems_EdgeCases(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	t.Run("no matching section", func(t *testing.T) {
		t.Parallel()
		items := parser.extractListItems(goalsPattern, "no goals here")
		assert.Nil(t, items)
	})

	t.Run("empty section", func(t *testing.T) {
		t.Parallel()
		items := parser.extractListItems(goalsPattern, "### Goals\n\n### Requirements")
		assert.Nil(t, items)
	})
}

func TestParser_ExtractSingleLine_EdgeCases(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	t.Run("no match", func(t *testing.T) {
		t.Parallel()
		result := parser.extractSingleLine(titlePattern, "no title here")
		assert.Empty(t, result)
	})
}

func TestParser_ExtractSection_EdgeCases(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	t.Run("no match", func(t *testing.T) {
		t.Parallel()
		result := parser.extractSection(summaryPattern, "no summary here")
		assert.Empty(t, result)
	})
}
