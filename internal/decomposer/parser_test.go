package decomposer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParser_Parse_SingleTask(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	input := `### Task T001: Implement types
- **Description**: Create type definitions for the package
- **Dependencies**: none
- **Files**: internal/decomposer/types.go
- **Complexity**: small
`

	tasks, err := parser.Parse(input)
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	task := tasks[0]
	assert.Equal(t, "T001", task.ID)
	assert.Equal(t, "Implement types", task.Title)
	assert.Equal(t, "Create type definitions for the package", task.Description)
	assert.Empty(t, task.Dependencies)
	assert.Equal(t, []string{"internal/decomposer/types.go"}, task.Files)
	assert.Equal(t, "small", task.Complexity)
	assert.Equal(t, "pending", task.Status)
}

func TestParser_Parse_MultipleTasks(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	input := `### Task T001: First task
- **Description**: Do the first thing
- **Dependencies**: none
- **Files**: file1.go
- **Complexity**: small

### Task T002: Second task
- **Description**: Do the second thing
- **Dependencies**: [T001]
- **Files**: file2.go, file3.go
- **Complexity**: medium

### Task T003: Third task
- **Description**: Do the third thing
- **Dependencies**: T001, T002
- **Files**: file4.go
- **Complexity**: large
`

	tasks, err := parser.Parse(input)
	require.NoError(t, err)
	require.Len(t, tasks, 3)

	// First task
	assert.Equal(t, "T001", tasks[0].ID)
	assert.Equal(t, "First task", tasks[0].Title)
	assert.Empty(t, tasks[0].Dependencies)
	assert.Equal(t, "small", tasks[0].Complexity)

	// Second task
	assert.Equal(t, "T002", tasks[1].ID)
	assert.Equal(t, "Second task", tasks[1].Title)
	assert.Equal(t, []string{"T001"}, tasks[1].Dependencies)
	assert.Equal(t, []string{"file2.go", "file3.go"}, tasks[1].Files)
	assert.Equal(t, "medium", tasks[1].Complexity)

	// Third task
	assert.Equal(t, "T003", tasks[2].ID)
	assert.Equal(t, "Third task", tasks[2].Title)
	assert.Equal(t, []string{"T001", "T002"}, tasks[2].Dependencies)
	assert.Equal(t, "large", tasks[2].Complexity)
}

func TestParser_Parse_DependencyVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		expectedDeps []string
	}{
		{
			name: "none keyword",
			input: `### Task T001: Test
- **Dependencies**: none
`,
			expectedDeps: []string{},
		},
		{
			name: "N/A keyword",
			input: `### Task T001: Test
- **Dependencies**: N/A
`,
			expectedDeps: []string{},
		},
		{
			name: "dash keyword",
			input: `### Task T001: Test
- **Dependencies**: -
`,
			expectedDeps: []string{},
		},
		{
			name: "bracketed IDs",
			input: `### Task T001: Test
- **Dependencies**: [T001], [T002]
`,
			expectedDeps: []string{"T001", "T002"},
		},
		{
			name: "unbracketed IDs",
			input: `### Task T001: Test
- **Dependencies**: T001, T002
`,
			expectedDeps: []string{"T001", "T002"},
		},
		{
			name: "mixed format",
			input: `### Task T001: Test
- **Dependencies**: [T001] and T002
`,
			expectedDeps: []string{"T001", "T002"},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			assert.Equal(t, tt.expectedDeps, tasks[0].Dependencies)
		})
	}
}

func TestParser_Parse_FileListVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		input         string
		expectedFiles []string
	}{
		{
			name: "comma separated",
			input: `### Task T001: Test
- **Files**: file1.go, file2.go, file3.go
`,
			expectedFiles: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name: "space separated",
			input: `### Task T001: Test
- **Files**: file1.go file2.go file3.go
`,
			expectedFiles: []string{"file1.go", "file2.go", "file3.go"},
		},
		{
			name: "single file",
			input: `### Task T001: Test
- **Files**: single.go
`,
			expectedFiles: []string{"single.go"},
		},
		{
			name: "with paths",
			input: `### Task T001: Test
- **Files**: internal/decomposer/types.go, internal/decomposer/parser.go
`,
			expectedFiles: []string{"internal/decomposer/types.go", "internal/decomposer/parser.go"},
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			assert.Equal(t, tt.expectedFiles, tasks[0].Files)
		})
	}
}

func TestParser_Parse_EmptyOutput(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	_, err := parser.Parse("")
	assert.Error(t, err)
	assert.Equal(t, ErrParseNoTasks, err)
}

func TestParser_Parse_NoValidTasks(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	input := `Some random text without task headers.
This should not parse to any tasks.
`

	_, err := parser.Parse(input)
	assert.Error(t, err)
	assert.Equal(t, ErrParseNoTasks, err)
}

func TestParser_Parse_HeaderVariations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		input        string
		expectedID   string
		expectedName string
	}{
		{
			name:         "standard format",
			input:        "### Task T001: Standard Task",
			expectedID:   "T001",
			expectedName: "Standard Task",
		},
		{
			name:         "case insensitive",
			input:        "### task T002: Case Test",
			expectedID:   "T002",
			expectedName: "Case Test",
		},
		{
			name:         "extra spaces",
			input:        "###   Task   T003:   Spaced Title",
			expectedID:   "T003",
			expectedName: "Spaced Title",
		},
	}

	parser := NewParser()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tasks, err := parser.Parse(tt.input)
			require.NoError(t, err)
			require.Len(t, tasks, 1)
			assert.Equal(t, tt.expectedID, tasks[0].ID)
			assert.Equal(t, tt.expectedName, tasks[0].Title)
		})
	}
}

func TestParser_Parse_FieldPatternVariations(t *testing.T) {
	t.Parallel()

	// Test that fields can have various formatting (bold, not bold, etc.)
	input := `### Task T001: Test Task
- Description: Plain description
- *Dependencies*: T002
- **Files**: file.go
- Complexity: medium
`

	parser := NewParser()
	tasks, err := parser.Parse(input)
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	assert.Equal(t, "Plain description", tasks[0].Description)
	assert.Equal(t, []string{"T002"}, tasks[0].Dependencies)
	assert.Equal(t, []string{"file.go"}, tasks[0].Files)
	assert.Equal(t, "medium", tasks[0].Complexity)
}

func TestParser_Parse_SuccessCriteria(t *testing.T) {
	t.Parallel()

	input := `### Task T001: Test Task
- **Description**: Test description
- **Success Criteria**: Tests pass with 90%+ coverage
`

	parser := NewParser()
	tasks, err := parser.Parse(input)
	require.NoError(t, err)
	require.Len(t, tasks, 1)

	assert.Equal(t, []string{"Tests pass with 90%+ coverage"}, tasks[0].SuccessCriteria)
}

func TestParser_extractTaskIDs_DeduplicatesDuplicates(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	// Same ID appears multiple times
	ids := parser.extractTaskIDs("[T001], T001, [T001]")
	assert.Equal(t, []string{"T001"}, ids)
}

func TestParser_extractFileList_IgnoresEmpty(t *testing.T) {
	t.Parallel()

	parser := NewParser()

	files := parser.extractFileList("file1.go, , file2.go, -")
	assert.Equal(t, []string{"file1.go", "file2.go"}, files)
}

func TestNewParser(t *testing.T) {
	t.Parallel()

	parser := NewParser()
	assert.NotNil(t, parser)
}
