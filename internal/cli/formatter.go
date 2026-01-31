package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ToolFormatter formats tool inputs and results for console output.
type ToolFormatter struct {
	workDir string
}

// NewToolFormatter creates a formatter with current working directory.
func NewToolFormatter() *ToolFormatter {
	wd, _ := os.Getwd()
	return &ToolFormatter{workDir: wd}
}

// NewToolFormatterWithWorkDir creates a formatter with a specific working directory.
// Useful for testing.
func NewToolFormatterWithWorkDir(workDir string) *ToolFormatter {
	return &ToolFormatter{workDir: workDir}
}

// FormatToolUse formats tool input for display.
func (f *ToolFormatter) FormatToolUse(name string, inputJSON string) string {
	var input map[string]interface{}
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return truncateString(inputJSON, 100) // fallback
	}

	switch name {
	case "Write":
		return f.formatWrite(input)
	case "Read":
		return f.formatRead(input)
	case "Edit":
		return f.formatEdit(input)
	case "Bash":
		return f.formatBash(input)
	case "Glob":
		return f.formatGlob(input)
	case "Grep":
		return f.formatGrep(input)
	case "Task":
		return f.formatTask(input)
	case "WebFetch":
		return f.formatWebFetch(input)
	case "WebSearch":
		return f.formatWebSearch(input)
	default:
		return f.formatGeneric(input)
	}
}

// FormatToolResult converts absolute paths to relative.
func (f *ToolFormatter) FormatToolResult(content string) string {
	if f.workDir != "" {
		content = strings.ReplaceAll(content, f.workDir+"/", "./")
	}
	return truncateString(content, 200)
}

func (f *ToolFormatter) formatWrite(input map[string]interface{}) string {
	path := f.relativePath(getString(input, "file_path"))
	content := getString(input, "content")
	preview := truncateString(firstLine(content), 40)
	return fmt.Sprintf("%s <- %q", path, preview)
}

func (f *ToolFormatter) formatRead(input map[string]interface{}) string {
	path := f.relativePath(getString(input, "file_path"))
	if limit := getInt(input, "limit"); limit > 0 {
		offset := getInt(input, "offset")
		return fmt.Sprintf("%s [lines %d-%d]", path, offset, offset+limit)
	}
	return path
}

func (f *ToolFormatter) formatEdit(input map[string]interface{}) string {
	path := f.relativePath(getString(input, "file_path"))
	old := truncateString(firstLine(getString(input, "old_string")), 25)
	newStr := truncateString(firstLine(getString(input, "new_string")), 25)
	return fmt.Sprintf("%s: %q -> %q", path, old, newStr)
}

func (f *ToolFormatter) formatBash(input map[string]interface{}) string {
	if desc := getString(input, "description"); desc != "" {
		return truncateString(desc, 80)
	}
	return truncateString(getString(input, "command"), 80)
}

func (f *ToolFormatter) formatGlob(input map[string]interface{}) string {
	pattern := getString(input, "pattern")
	if path := f.relativePath(getString(input, "path")); path != "" && path != "." {
		return fmt.Sprintf("%s in %s", pattern, path)
	}
	return pattern
}

func (f *ToolFormatter) formatGrep(input map[string]interface{}) string {
	pattern := getString(input, "pattern")
	result := fmt.Sprintf("/%s/", pattern)
	if glob := getString(input, "glob"); glob != "" {
		result += " " + glob
	}
	if path := f.relativePath(getString(input, "path")); path != "" && path != "." {
		result += " in " + path
	}
	return result
}

func (f *ToolFormatter) formatTask(input map[string]interface{}) string {
	if desc := getString(input, "description"); desc != "" {
		return desc
	}
	if prompt := getString(input, "prompt"); prompt != "" {
		return truncateString(firstLine(prompt), 60)
	}
	return f.formatGeneric(input)
}

func (f *ToolFormatter) formatWebFetch(input map[string]interface{}) string {
	url := getString(input, "url")
	return truncateString(url, 80)
}

func (f *ToolFormatter) formatWebSearch(input map[string]interface{}) string {
	query := getString(input, "query")
	return fmt.Sprintf("%q", query)
}

func (f *ToolFormatter) formatGeneric(input map[string]interface{}) string {
	if path := getString(input, "file_path"); path != "" {
		return f.relativePath(path)
	}
	if path := getString(input, "path"); path != "" {
		return f.relativePath(path)
	}
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	return fmt.Sprintf("{%s}", strings.Join(keys, ", "))
}

func (f *ToolFormatter) relativePath(absPath string) string {
	if absPath == "" {
		return ""
	}
	if f.workDir == "" || !filepath.IsAbs(absPath) {
		return filepath.Base(absPath)
	}
	if rel, err := filepath.Rel(f.workDir, absPath); err == nil {
		if !strings.HasPrefix(rel, ".."+string(filepath.Separator)+"..") {
			return rel
		}
	}
	return filepath.Base(absPath) // fallback to filename only
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// getInt safely extracts an int from a map (JSON numbers are float64).
func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}

// firstLine returns the first line of a string.
func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i != -1 {
		return s[:i]
	}
	return s
}

// truncateString truncates a string to maxLen characters, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
