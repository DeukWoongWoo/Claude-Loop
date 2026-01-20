// Package claude provides a wrapper for executing the Claude CLI as a subprocess.
package claude

import "encoding/json"

// StreamMessage represents a single line in the stream-json output.
// The claude CLI outputs newline-delimited JSON objects.
type StreamMessage struct {
	Type         string            `json:"type"`                     // "assistant", "user", "result", "system", etc.
	Message      *AssistantMessage `json:"message,omitempty"`        // Present when type="assistant"
	Result       string            `json:"result,omitempty"`         // Present when type="result"
	TotalCostUSD float64           `json:"total_cost_usd,omitempty"` // Present in result
	IsError      bool              `json:"is_error,omitempty"`       // Present in result
	SessionID    string            `json:"session_id,omitempty"`     // Session ID for resume capability
}

// RawMessage is used to handle both assistant and user message formats.
// Both have a "message" field but with different structures.
type RawMessage struct {
	Type    string          `json:"type"`
	Message json.RawMessage `json:"message,omitempty"` // Present for "assistant" and "user" types
	Event   json.RawMessage `json:"event,omitempty"`   // Present for "stream_event" type
	Result  string          `json:"result,omitempty"`  // Present for "result" type

	// Result fields
	TotalCostUSD float64 `json:"total_cost_usd,omitempty"`
	IsError      bool    `json:"is_error,omitempty"`
	SessionID    string  `json:"session_id,omitempty"`
}

// StreamEvent represents a streaming event from --include-partial-messages output.
// These events provide real-time token-level streaming.
type StreamEvent struct {
	Type  string      `json:"type"`            // "content_block_delta", "message_start", etc.
	Index int         `json:"index,omitempty"` // Block index
	Delta *EventDelta `json:"delta,omitempty"` // Delta content for content_block_delta
}

// EventDelta represents the delta content in a stream event.
type EventDelta struct {
	Type string `json:"type"`           // "text_delta"
	Text string `json:"text,omitempty"` // The text fragment
}

// AssistantMessage contains the assistant's response content.
type AssistantMessage struct {
	Content []ContentBlock `json:"content"`
}

// ContentBlock represents a block of content in the response.
type ContentBlock struct {
	Type  string `json:"type"`            // "text", "tool_use", etc.
	Text  string `json:"text,omitempty"`  // Present when type="text"
	ID    string `json:"id,omitempty"`    // Present when type="tool_use"
	Name  string `json:"name,omitempty"`  // Tool name when type="tool_use"
	Input any    `json:"input,omitempty"` // Tool input parameters when type="tool_use"
}

// UserContent represents content in a user message (tool results).
type UserContent struct {
	Type      string `json:"type"`                 // "tool_result"
	ToolUseID string `json:"tool_use_id,omitempty"`
	Content   string `json:"content,omitempty"`
	IsError   bool   `json:"is_error,omitempty"`
}

// UserMessage represents user messages in the stream (tool results).
type UserMessage struct {
	Role    string        `json:"role"`
	Content []UserContent `json:"content"`
}

// ParsedResult represents the final parsed output from a Claude execution.
type ParsedResult struct {
	Output       string          // Concatenated text output from all assistant messages
	ResultText   string          // The final result text from the result message
	TotalCostUSD float64         // Total cost from the result message
	IsError      bool            // Whether the execution resulted in an error
	RawMessages  []StreamMessage // All parsed messages (for debugging)
	SessionID    string          // Session ID for resume capability
}

// SessionResult contains execution result with session info for resume.
// Used for multi-turn interactions like principles collection.
// Note: IsError is not included because errors are returned via the error return value.
type SessionResult struct {
	Output    string  // Concatenated text output from assistant messages
	SessionID string  // Session ID for resume capability
	Cost      float64 // Total cost from the result message
}
