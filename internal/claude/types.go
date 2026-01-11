// Package claude provides a wrapper for executing the Claude CLI as a subprocess.
package claude

// StreamMessage represents a single line in the stream-json output.
// The claude CLI outputs newline-delimited JSON objects.
type StreamMessage struct {
	Type         string            `json:"type"`                    // "assistant", "result", "system", etc.
	Message      *AssistantMessage `json:"message,omitempty"`       // Present when type="assistant"
	Result       string            `json:"result,omitempty"`        // Present when type="result"
	TotalCostUSD float64           `json:"total_cost_usd,omitempty"` // Present in result
	IsError      bool              `json:"is_error,omitempty"`      // Present in result
}

// AssistantMessage contains the assistant's response content.
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a block of content in the response.
type ContentBlock struct {
	Type string `json:"type"`           // "text", "tool_use", etc.
	Text string `json:"text,omitempty"` // Present when type="text"
}

// ParsedResult represents the final parsed output from a Claude execution.
type ParsedResult struct {
	Output       string          // Concatenated text output from all assistant messages
	ResultText   string          // The final result text from the result message
	TotalCostUSD float64         // Total cost from the result message
	IsError      bool            // Whether the execution resulted in an error
	RawMessages  []StreamMessage // All parsed messages (for debugging)
}
