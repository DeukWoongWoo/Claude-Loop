package claude

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"
)

// StreamHandler handles real-time streaming of assistant output.
type StreamHandler interface {
	// OnText is called for each text block from assistant messages.
	// This allows real-time display during execution.
	OnText(text string)
}

// ToolStreamHandler extends StreamHandler with tool-related callbacks.
// Implementations can optionally implement this interface to receive
// tool use and tool result notifications.
type ToolStreamHandler interface {
	StreamHandler
	// OnToolUse is called when the assistant invokes a tool.
	OnToolUse(name string, input string)
	// OnToolResult is called when a tool returns a result.
	OnToolResult(content string, isError bool)
}

// Parser parses the claude stream-json output.
type Parser struct {
	handler StreamHandler

	// receivedStreamEvents tracks if stream_event messages were received during parsing.
	// When true, assistant message text is not streamed to avoid duplicates since
	// stream_events already provide real-time token-level output.
	receivedStreamEvents bool
}

// NewParser creates a new Parser with the given handler.
// If handler is nil, text output is still accumulated but not streamed.
func NewParser(handler StreamHandler) *Parser {
	return &Parser{handler: handler}
}

// Parse reads from the reader and parses stream-json output.
// Returns the final ParsedResult after processing all messages.
// Malformed JSON lines are skipped (matching bash behavior).
func (p *Parser) Parse(r io.Reader) (*ParsedResult, error) {
	scanner := bufio.NewScanner(r)
	// Handle potentially large JSON lines (up to 1MB)
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	result := &ParsedResult{}
	var outputBuilder strings.Builder

	p.receivedStreamEvents = false

	// Check if handler supports tool callbacks
	toolHandler, hasToolHandler := p.handler.(ToolStreamHandler)

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// First, parse to get the type
		var raw RawMessage
		if err := json.Unmarshal([]byte(line), &raw); err != nil {
			// Skip malformed JSON lines (match bash behavior with jq 2>/dev/null)
			continue
		}

		switch raw.Type {
		case "assistant":
			// Parse as assistant message
			var msg StreamMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				continue
			}
			result.RawMessages = append(result.RawMessages, msg)

			if msg.Message != nil {
				for _, block := range msg.Message.Content {
					switch block.Type {
					case "text":
						if block.Text == "" {
							continue
						}
						outputBuilder.WriteString(block.Text)
						// Skip streaming if stream_events already provided real-time output
						if p.receivedStreamEvents || p.handler == nil {
							continue
						}
						p.handler.OnText(block.Text)
					case "tool_use":
						if hasToolHandler && block.Name != "" {
							inputJSON, _ := json.Marshal(block.Input)
							toolHandler.OnToolUse(block.Name, string(inputJSON))
						}
					}
				}
			}

		case "user":
			// Parse user message for tool results
			if hasToolHandler && raw.Message != nil {
				var userMsg UserMessage
				if err := json.Unmarshal(raw.Message, &userMsg); err == nil {
					for _, content := range userMsg.Content {
						if content.Type == "tool_result" {
							toolHandler.OnToolResult(content.Content, content.IsError)
						}
					}
				}
			}

		case "stream_event":
			p.handleStreamEvent(&raw)

		case "result":
			var msg StreamMessage
			if err := json.Unmarshal([]byte(line), &msg); err != nil {
				continue
			}
			result.RawMessages = append(result.RawMessages, msg)
			result.ResultText = msg.Result
			result.TotalCostUSD = msg.TotalCostUSD
			result.IsError = msg.IsError
			result.SessionID = msg.SessionID

		default:
			// Store other message types (system, etc.)
			var msg StreamMessage
			if err := json.Unmarshal([]byte(line), &msg); err == nil {
				result.RawMessages = append(result.RawMessages, msg)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &ParseError{Message: "scanner error", Err: err}
	}

	result.Output = outputBuilder.String()
	return result, nil
}

// handleStreamEvent processes partial message events from --include-partial-messages.
// It extracts text deltas and streams them to the handler for real-time output.
func (p *Parser) handleStreamEvent(raw *RawMessage) {
	if raw.Event == nil {
		return
	}

	var event StreamEvent
	if err := json.Unmarshal(raw.Event, &event); err != nil {
		return
	}

	if event.Type != "content_block_delta" || event.Delta == nil {
		return
	}

	if event.Delta.Type != "text_delta" || event.Delta.Text == "" {
		return
	}

	p.receivedStreamEvents = true
	if p.handler != nil {
		p.handler.OnText(event.Delta.Text)
	}
}

// NoOpHandler is a StreamHandler that does nothing.
// Useful when you don't need real-time streaming.
type NoOpHandler struct{}

func (h *NoOpHandler) OnText(text string) {}
