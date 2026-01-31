package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewToolFormatter(t *testing.T) {
	f := NewToolFormatter()
	assert.NotNil(t, f)
	assert.NotEmpty(t, f.workDir)
}

func TestNewToolFormatterWithWorkDir(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/test/path")
	assert.Equal(t, "/test/path", f.workDir)
}

func TestFormatToolUse_Write(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic write",
			input:    `{"file_path":"/Users/test/project/test.txt","content":"hello world"}`,
			expected: `test.txt <- "hello world"`,
		},
		{
			name:     "multiline content shows first line",
			input:    `{"file_path":"/Users/test/project/README.md","content":"# Title\n\nContent here"}`,
			expected: `README.md <- "# Title"`,
		},
		{
			name:     "long content truncated",
			input:    `{"file_path":"/Users/test/project/long.txt","content":"This is a very long content that should be truncated after forty chars"}`,
			expected: `long.txt <- "This is a very long content that shou..."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Write", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Read(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic read",
			input:    `{"file_path":"/Users/test/project/test.go"}`,
			expected: "test.go",
		},
		{
			name:     "read with limit",
			input:    `{"file_path":"/Users/test/project/test.go","offset":10,"limit":50}`,
			expected: "test.go [lines 10-60]",
		},
		{
			name:     "read with limit zero offset",
			input:    `{"file_path":"/Users/test/project/test.go","offset":0,"limit":100}`,
			expected: "test.go [lines 0-100]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Read", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Edit(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic edit",
			input:    `{"file_path":"/Users/test/project/test.go","old_string":"old","new_string":"new"}`,
			expected: `test.go: "old" -> "new"`,
		},
		{
			name:     "multiline edit shows first line",
			input:    `{"file_path":"/Users/test/project/test.go","old_string":"func old() {\n\treturn\n}","new_string":"func new() {\n\treturn nil\n}"}`,
			expected: `test.go: "func old() {" -> "func new() {"`,
		},
		{
			name:     "long strings truncated",
			input:    `{"file_path":"/Users/test/project/test.go","old_string":"This is a very long old string that needs truncation","new_string":"This is a very long new string that needs truncation"}`,
			expected: `test.go: "This is a very long ol..." -> "This is a very long ne..."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Edit", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Bash(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with description",
			input:    `{"command":"git status","description":"Check git status"}`,
			expected: "Check git status",
		},
		{
			name:     "without description",
			input:    `{"command":"npm install"}`,
			expected: "npm install",
		},
		{
			name:     "long command truncated",
			input:    `{"command":"find . -name '*.go' -type f | xargs grep -l 'pattern' | head -10 | while read f; do echo $f; done"}`,
			expected: "find . -name '*.go' -type f | xargs grep -l 'pattern' | head -10 | while read...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Bash", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Glob(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pattern only",
			input:    `{"pattern":"**/*.go"}`,
			expected: "**/*.go",
		},
		{
			name:     "with path",
			input:    `{"pattern":"*.go","path":"/Users/test/project/internal"}`,
			expected: "*.go in internal",
		},
		{
			name:     "dot path treated as empty",
			input:    `{"pattern":"*.go","path":"."}`,
			expected: "*.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Glob", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Grep(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pattern only",
			input:    `{"pattern":"func.*Test"}`,
			expected: "/func.*Test/",
		},
		{
			name:     "with glob",
			input:    `{"pattern":"error","glob":"*.go"}`,
			expected: "/error/ *.go",
		},
		{
			name:     "with path",
			input:    `{"pattern":"TODO","path":"/Users/test/project/internal"}`,
			expected: "/TODO/ in internal",
		},
		{
			name:     "with glob and path",
			input:    `{"pattern":"import","glob":"*.go","path":"/Users/test/project/cmd"}`,
			expected: "/import/ *.go in cmd",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Grep", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_Task(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "with description",
			input:    `{"description":"Find error handlers","prompt":"Search for error handling code"}`,
			expected: "Find error handlers",
		},
		{
			name:     "without description",
			input:    `{"prompt":"Search for all files that contain the word error"}`,
			expected: "Search for all files that contain the word error",
		},
		{
			name:     "long prompt truncated",
			input:    `{"prompt":"This is a very long prompt that describes what the task should do and it goes on and on and on"}`,
			expected: "This is a very long prompt that describes what the task s...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse("Task", tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatToolUse_WebFetch(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	input := `{"url":"https://example.com/api/docs","prompt":"Get the API docs"}`
	result := f.FormatToolUse("WebFetch", input)
	assert.Equal(t, "https://example.com/api/docs", result)
}

func TestFormatToolUse_WebSearch(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	input := `{"query":"golang error handling best practices"}`
	result := f.FormatToolUse("WebSearch", input)
	assert.Equal(t, `"golang error handling best practices"`, result)
}

func TestFormatToolUse_Generic(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		toolName string
		input    string
		expected string
	}{
		{
			name:     "with file_path",
			toolName: "UnknownTool",
			input:    `{"file_path":"/Users/test/project/test.txt","other":"value"}`,
			expected: "test.txt",
		},
		{
			name:     "with path",
			toolName: "UnknownTool",
			input:    `{"path":"/Users/test/project/subdir","other":"value"}`,
			expected: "subdir",
		},
		{
			name:     "shows keys",
			toolName: "UnknownTool",
			input:    `{"foo":"bar","baz":"qux"}`,
			expected: "{foo, baz}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolUse(tt.toolName, tt.input)
			// For the keys case, order might vary
			if tt.name == "shows keys" {
				assert.Contains(t, result, "foo")
				assert.Contains(t, result, "baz")
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFormatToolUse_InvalidJSON(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	result := f.FormatToolUse("Write", "not valid json")
	assert.Equal(t, "not valid json", result)

	// Long invalid JSON is truncated
	longInvalid := "this is not valid json and it is very long and needs to be truncated because it exceeds the maximum length"
	result = f.FormatToolUse("Write", longInvalid)
	assert.Equal(t, "this is not valid json and it is very long and needs to be truncated because it exceeds the maxim...", result)
}

func TestFormatToolResult(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "replaces absolute path with relative",
			content:  "File created at /Users/test/project/test.txt",
			expected: "File created at ./test.txt",
		},
		{
			name:     "long content truncated",
			content:  "This is a very long result message that contains a lot of text and needs to be truncated because it exceeds the maximum allowed length of two hundred characters which is quite a lot but this message is even longer than that so it will be cut off",
			expected: "This is a very long result message that contains a lot of text and needs to be truncated because it exceeds the maximum allowed length of two hundred characters which is quite a lot but this messag...",
		},
		{
			name:     "short content unchanged",
			content:  "Success",
			expected: "Success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.FormatToolResult(tt.content)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRelativePath(t *testing.T) {
	f := NewToolFormatterWithWorkDir("/Users/test/project")

	tests := []struct {
		name     string
		absPath  string
		expected string
	}{
		{
			name:     "within project",
			absPath:  "/Users/test/project/internal/cli/root.go",
			expected: "internal/cli/root.go",
		},
		{
			name:     "project root",
			absPath:  "/Users/test/project/main.go",
			expected: "main.go",
		},
		{
			name:     "outside project falls back to basename",
			absPath:  "/totally/different/path/file.go",
			expected: "file.go",
		},
		{
			name:     "empty path",
			absPath:  "",
			expected: "",
		},
		{
			name:     "relative path already",
			absPath:  "internal/cli/root.go",
			expected: "root.go",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.relativePath(tt.absPath)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "short string unchanged",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "exact length unchanged",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "long string truncated",
			input:    "hello world",
			maxLen:   8,
			expected: "hello...",
		},
		{
			name:     "very short maxLen",
			input:    "hello",
			maxLen:   2,
			expected: "he",
		},
		{
			name:     "maxLen 3 exact",
			input:    "hello",
			maxLen:   3,
			expected: "hel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFirstLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "multi line",
			input:    "first line\nsecond line\nthird line",
			expected: "first line",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "just newline",
			input:    "\n",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := firstLine(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"str":    "value",
		"num":    42.0,
		"nested": map[string]interface{}{"key": "val"},
	}

	assert.Equal(t, "value", getString(m, "str"))
	assert.Equal(t, "", getString(m, "num"))
	assert.Equal(t, "", getString(m, "nested"))
	assert.Equal(t, "", getString(m, "missing"))
}

func TestGetInt(t *testing.T) {
	m := map[string]interface{}{
		"num":  42.0,
		"str":  "hello",
		"zero": 0.0,
	}

	assert.Equal(t, 42, getInt(m, "num"))
	assert.Equal(t, 0, getInt(m, "str"))
	assert.Equal(t, 0, getInt(m, "zero"))
	assert.Equal(t, 0, getInt(m, "missing"))
}
