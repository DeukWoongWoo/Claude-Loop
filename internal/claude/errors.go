package claude

import "fmt"

// ClaudeError represents an error from the Claude CLI execution.
type ClaudeError struct {
	Message    string // Error description
	ResultText string // Error text from JSON response (if available)
	Stderr     string // Stderr output for context
	Err        error  // Underlying error
}

func (e *ClaudeError) Error() string {
	if e.ResultText != "" {
		return fmt.Sprintf("claude: %s: %s", e.Message, e.ResultText)
	}
	if e.Err != nil {
		return fmt.Sprintf("claude: %s: %v", e.Message, e.Err)
	}
	return fmt.Sprintf("claude: %s", e.Message)
}

func (e *ClaudeError) Unwrap() error {
	return e.Err
}

// ParseError represents a JSON parsing error.
type ParseError struct {
	Message string // Error description
	Line    string // The line that failed to parse (if available)
	Err     error  // Underlying error
}

func (e *ParseError) Error() string {
	switch {
	case e.Line != "" && e.Err != nil:
		return fmt.Sprintf("parse error: %s: %v (line: %q)", e.Message, e.Err, truncateLine(e.Line, 100))
	case e.Line != "":
		return fmt.Sprintf("parse error: %s (line: %q)", e.Message, truncateLine(e.Line, 100))
	case e.Err != nil:
		return fmt.Sprintf("parse error: %s: %v", e.Message, e.Err)
	default:
		return fmt.Sprintf("parse error: %s", e.Message)
	}
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// truncateLine truncates a line to maxLen characters for display.
func truncateLine(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
