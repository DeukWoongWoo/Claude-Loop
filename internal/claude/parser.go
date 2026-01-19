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

// Parser parses the claude stream-json output.
type Parser struct {
	handler StreamHandler
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

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg StreamMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Skip malformed JSON lines (match bash behavior with jq 2>/dev/null)
			continue
		}

		result.RawMessages = append(result.RawMessages, msg)

		switch msg.Type {
		case "assistant":
			if msg.Message != nil {
				for _, block := range msg.Message.Content {
					if block.Type == "text" && block.Text != "" {
						outputBuilder.WriteString(block.Text)
						if p.handler != nil {
							p.handler.OnText(block.Text)
						}
					}
				}
			}
		case "result":
			result.ResultText = msg.Result
			result.TotalCostUSD = msg.TotalCostUSD
			result.IsError = msg.IsError
			result.SessionID = msg.SessionID
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, &ParseError{Message: "scanner error", Err: err}
	}

	result.Output = outputBuilder.String()
	return result, nil
}

// NoOpHandler is a StreamHandler that does nothing.
// Useful when you don't need real-time streaming.
type NoOpHandler struct{}

func (h *NoOpHandler) OnText(text string) {}
