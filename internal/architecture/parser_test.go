package architecture

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
		name           string
		input          string
		wantComponents int
		wantDeps       int
		wantFiles      int
		wantDecisions  int
		wantTitle      string
		wantErr        error
	}{
		{
			name: "valid architecture output with all sections",
			input: `## Architecture Design

### Title: User Authentication System

### Summary
This architecture implements a secure user authentication system.

### Components
- **AuthService**: Handles authentication logic
  - Files: auth.go, auth_test.go
- **TokenManager**: Manages JWT tokens
  - Files: token.go

### Dependencies
- github.com/golang-jwt/jwt
- golang.org/x/crypto

### File Structure
- internal/auth/auth.go
- internal/auth/token.go
- internal/auth/auth_test.go

### Technical Decisions
- Use JWT for stateless authentication
- Store refresh tokens in database

### Rationale
JWT provides scalability and stateless verification.
`,
			wantComponents: 2,
			wantDeps:       2,
			wantFiles:      3,
			wantDecisions:  2,
			wantTitle:      "User Authentication System",
			wantErr:        nil,
		},
		{
			name: "missing components section",
			input: `## Architecture

### Dependencies
- some-lib

### File Structure
- file.go

### Technical Decisions
- Decision 1
`,
			wantErr: ErrParseNoComponents,
		},
		{
			name: "single component",
			input: `### Components
- **Parser**: Parses the output
  - Files: parser.go

### File Structure
- parser.go

### Technical Decisions
- Use regex
`,
			wantComponents: 1,
			wantDeps:       0,
			wantFiles:      1,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "multiple components with files",
			input: `### Components
- **ComponentA**: First component
  - Files: a.go, a_test.go
- **ComponentB**: Second component
  - Files: b.go
- **ComponentC**: Third component

### File Structure
- a.go
- b.go

### Technical Decisions
- Decision
`,
			wantComponents: 3,
			wantFiles:      2,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "component with inline description",
			input: `### Components
- **Handler**: Processes HTTP requests
- **Router**: Routes requests to handlers
  - Files: router.go

### File Structure
- handler.go

### Technical Decisions
- Use standard library
`,
			wantComponents: 2,
			wantFiles:      1,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "two hash headers instead of three",
			input: `## Components
- **Service**: Main service component
  - Files: service.go

## File Structure
- service.go

## Technical Decisions
- Keep it simple
`,
			wantComponents: 1,
			wantFiles:      1,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "empty dependencies allowed",
			input: `### Components
- **Core**: Core functionality
  - Files: core.go

### File Structure
- core.go

### Technical Decisions
- No external deps needed
`,
			wantComponents: 1,
			wantDeps:       0,
			wantFiles:      1,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "extra whitespace handling",
			input: `### Components

- **  Parser  **:   Parses things
  - Files:   parser.go  ,  parser_test.go

### File Structure
  -   file1.go
  -   file2.go

### Technical Decisions
  -   Use regex
`,
			wantComponents: 1,
			wantFiles:      2,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "component without bold markers",
			input: `### Components
- Parser: Parses the architecture output
  - Files: parser.go

### File Structure
- parser.go

### Technical Decisions
- Simple approach
`,
			wantComponents: 1,
			wantFiles:      1,
			wantDecisions:  1,
			wantErr:        nil,
		},
		{
			name: "numbered list items",
			input: `### Components
- **Generator**: Generates architecture
  - Files: generator.go

### File Structure
1. internal/arch/generator.go
2. internal/arch/types.go

### Technical Decisions
1. Follow existing patterns
2. Keep it minimal
`,
			wantComponents: 1,
			wantFiles:      2,
			wantDecisions:  2,
			wantErr:        nil,
		},
		{
			name: "asterisk list items",
			input: `### Components
* **Validator**: Validates architecture
  * Files: validator.go

### File Structure
* validator.go
* validator_test.go

### Technical Decisions
* Validate all fields
`,
			wantComponents: 1,
			wantFiles:      2,
			wantDecisions:  1,
			wantErr:        nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			arch, err := parser.Parse(tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, arch)

			assert.Len(t, arch.Components, tt.wantComponents)
			assert.Len(t, arch.Dependencies, tt.wantDeps)
			assert.Len(t, arch.FileStructure, tt.wantFiles)
			assert.Len(t, arch.TechDecisions, tt.wantDecisions)

			if tt.wantTitle != "" {
				assert.Equal(t, tt.wantTitle, arch.Title)
			}

			// Verify RawOutput is preserved
			assert.Equal(t, tt.input, arch.RawOutput)
		})
	}
}

func TestParser_extractComponents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		wantCount  int
		wantFirst  string
		wantFiles  []string
	}{
		{
			name: "bold name with colon",
			input: `### Components
- **Parser**: Parses Claude output
  - Files: parser.go, parser_test.go
`,
			wantCount: 1,
			wantFirst: "Parser",
			wantFiles: []string{"parser.go", "parser_test.go"},
		},
		{
			name: "plain name with colon",
			input: `### Components
- Handler: Handles requests
  - Files: handler.go
`,
			wantCount: 1,
			wantFirst: "Handler",
			wantFiles: []string{"handler.go"},
		},
		{
			name: "space-separated files",
			input: `### Components
- **Builder**: Builds things
  - Files: builder.go builder_test.go
`,
			wantCount: 1,
			wantFirst: "Builder",
			wantFiles: []string{"builder.go", "builder_test.go"},
		},
		{
			name: "no files line",
			input: `### Components
- **Service**: Main service
- **Helper**: Helper functions
`,
			wantCount: 2,
			wantFirst: "Service",
			wantFiles: []string{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			components := parser.extractComponents(tt.input)

			assert.Len(t, components, tt.wantCount)
			if tt.wantCount > 0 {
				assert.Equal(t, tt.wantFirst, components[0].Name)
				assert.Equal(t, tt.wantFiles, components[0].Files)
			}
		})
	}
}

func TestParser_extractListItems(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantCount int
		wantFirst string
	}{
		{
			name: "bullet list",
			input: `### Dependencies
- item1
- item2
- item3
`,
			wantCount: 3,
			wantFirst: "item1",
		},
		{
			name: "numbered list",
			input: `### Dependencies
1. first
2. second
`,
			wantCount: 2,
			wantFirst: "first",
		},
		{
			name: "asterisk list",
			input: `### Dependencies
* one
* two
`,
			wantCount: 2,
			wantFirst: "one",
		},
		{
			name: "empty section",
			input: `### Dependencies

### Next Section
`,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			items := parser.extractListItems(dependenciesPattern, tt.input)

			assert.Len(t, items, tt.wantCount)
			if tt.wantCount > 0 && tt.wantFirst != "" {
				assert.Equal(t, tt.wantFirst, items[0])
			}
		})
	}
}

func TestParser_extractSingleLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "title with colon",
			input: `### Title: My Architecture
Some content
`,
			expected: "My Architecture",
		},
		{
			name: "architecture title variation",
			input: `### Architecture Title: System Design
Content
`,
			expected: "System Design",
		},
		{
			name:     "no match",
			input:    "No title here",
			expected: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			result := parser.extractSingleLine(titlePattern, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParser_extractSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name: "summary section",
			input: `### Summary
This is a multi-line
summary of the architecture.

### Next Section
`,
			contains: "This is a multi-line",
		},
		{
			name: "overview variation",
			input: `### Overview
Overview content here.

### Other
`,
			contains: "Overview content here",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			result := parser.extractSection(summaryPattern, tt.input)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestParser_extractRationale(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{
			name: "rationale section",
			input: `### Rationale
We chose this approach because it provides
better maintainability and testability.

### Next Section
`,
			contains: "better maintainability",
		},
		{
			name: "rationale with single line",
			input: `### Rationale
Simple and effective solution.

### Components
`,
			contains: "Simple and effective",
		},
		{
			name:     "no rationale section",
			input:    "No rationale here",
			contains: "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			result := parser.extractSection(rationalePattern, tt.input)
			if tt.contains != "" {
				assert.Contains(t, result, tt.contains)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}
