package claude

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockStreamHandler collects all text for testing.
type mockStreamHandler struct {
	texts []string
}

func (h *mockStreamHandler) OnText(text string) {
	h.texts = append(h.texts, text)
}

func TestParser_ParseSingleAssistant(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, world!"}]}}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Hello, world!", result.Output)
	assert.Len(t, result.RawMessages, 1)
}

func TestParser_ParseMultipleAssistants(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":" world"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"!"}]}}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Hello world!", result.Output)
	assert.Len(t, result.RawMessages, 3)
}

func TestParser_ParseWithResult(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Done."}]}}
{"type":"result","result":"Task completed","total_cost_usd":0.0523,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Done.", result.Output)
	assert.Equal(t, "Task completed", result.ResultText)
	assert.InDelta(t, 0.0523, result.TotalCostUSD, 0.0001)
	assert.False(t, result.IsError)
}

func TestParser_ParseErrorResult(t *testing.T) {
	input := `{"type":"result","result":"Something went wrong","total_cost_usd":0.01,"is_error":true}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.True(t, result.IsError)
	assert.Equal(t, "Something went wrong", result.ResultText)
}

func TestParser_SkipsMalformedJSON(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"First"}]}}
this is not json
{"type":"assistant","message":{"content":[{"type":"text","text":"Second"}]}}
{broken json}
{"type":"result","result":"Done","total_cost_usd":0.05,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "FirstSecond", result.Output)
	assert.Equal(t, "Done", result.ResultText)
	// Only valid JSON messages are kept
	assert.Len(t, result.RawMessages, 3)
}

func TestParser_SkipsEmptyLines(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}

{"type":"assistant","message":{"content":[{"type":"text","text":"World"}]}}

`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "HelloWorld", result.Output)
	assert.Len(t, result.RawMessages, 2)
}

func TestParser_StreamHandler(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"First"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Second"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockStreamHandler{}
	parser := NewParser(handler)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, []string{"First", "Second"}, handler.texts)
	assert.Equal(t, "FirstSecond", result.Output)
}

func TestParser_MultipleContentBlocks(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Part1"},{"type":"tool_use"},{"type":"text","text":"Part2"}]}}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Part1Part2", result.Output)
}

func TestParser_EmptyInput(t *testing.T) {
	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(""))

	require.NoError(t, err)
	assert.Empty(t, result.Output)
	assert.Empty(t, result.ResultText)
	assert.Zero(t, result.TotalCostUSD)
	assert.False(t, result.IsError)
	assert.Empty(t, result.RawMessages)
}

func TestParser_NoTextContent(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Empty(t, result.Output)
	assert.Equal(t, "Done", result.ResultText)
}

func TestParser_OtherMessageTypes(t *testing.T) {
	// Unknown message types (e.g., "init") are parsed but ignored for text output
	input := `{"type":"init"}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Hello", result.Output)
	assert.Len(t, result.RawMessages, 3) // All valid JSON messages are captured
}

func TestNoOpHandler(t *testing.T) {
	handler := &NoOpHandler{}
	// Should not panic
	handler.OnText("test")
}

func TestParser_ParseWithSessionID(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Done."}]}}
{"type":"result","result":"Task completed","total_cost_usd":0.05,"is_error":false,"session_id":"abc-123-def"}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Done.", result.Output)
	assert.Equal(t, "abc-123-def", result.SessionID)
	assert.InDelta(t, 0.05, result.TotalCostUSD, 0.0001)
	assert.False(t, result.IsError)
}

func TestParser_ParseWithoutSessionID(t *testing.T) {
	input := `{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Empty(t, result.SessionID)
}

// mockToolStreamHandler implements ToolStreamHandler for testing.
type mockToolStreamHandler struct {
	texts       []string
	toolUses    []struct{ name, input string }
	toolResults []struct {
		content string
		isError bool
	}
}

func (h *mockToolStreamHandler) OnText(text string) {
	h.texts = append(h.texts, text)
}

func (h *mockToolStreamHandler) OnToolUse(name string, input string) {
	h.toolUses = append(h.toolUses, struct{ name, input string }{name, input})
}

func (h *mockToolStreamHandler) OnToolResult(content string, isError bool) {
	h.toolResults = append(h.toolResults, struct {
		content string
		isError bool
	}{content, isError})
}

func TestParser_StreamEvent(t *testing.T) {
	// Test that stream_event with content_block_delta is handled correctly
	input := `{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}}
{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":" World"}}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockStreamHandler{}
	parser := NewParser(handler)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, []string{"Hello", " World"}, handler.texts)
	// Output should be empty since stream_events don't contribute to Output
	assert.Empty(t, result.Output)
}

func TestParser_StreamEventPreventsDoubleCallback(t *testing.T) {
	// When stream_event is received first, assistant text should NOT trigger OnText
	// This prevents duplicate output when --include-partial-messages is used
	input := `{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello World"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockStreamHandler{}
	parser := NewParser(handler)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	// Should only see "Hello" from stream_event, not "Hello World" from assistant
	assert.Equal(t, []string{"Hello"}, handler.texts)
	// Output should still contain full text from assistant message
	assert.Equal(t, "Hello World", result.Output)
}

func TestParser_StreamEventWithoutHandler(t *testing.T) {
	// stream_event should not cause errors when handler is nil
	input := `{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"Hello"}}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Hello World"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	parser := NewParser(nil)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, "Hello World", result.Output)
}

func TestParser_StreamEventIgnoresNonTextDelta(t *testing.T) {
	// Non-text_delta events should be ignored
	input := `{"type":"stream_event","event":{"type":"message_start"}}
{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"input_json_delta","partial_json":"{}"}}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Result"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockStreamHandler{}
	parser := NewParser(handler)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	// Should see "Result" from assistant since no text_delta was received
	assert.Equal(t, []string{"Result"}, handler.texts)
	assert.Equal(t, "Result", result.Output)
}

func TestParser_ToolStreamHandler_ToolUse(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tool_1","name":"Bash","input":{"command":"ls -la"}}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockToolStreamHandler{}
	parser := NewParser(handler)
	_, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Len(t, handler.toolUses, 1)
	assert.Equal(t, "Bash", handler.toolUses[0].name)
	assert.Contains(t, handler.toolUses[0].input, "ls -la")
}

func TestParser_ToolStreamHandler_ToolResult(t *testing.T) {
	input := `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"file1.txt\nfile2.txt","is_error":false}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockToolStreamHandler{}
	parser := NewParser(handler)
	_, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Len(t, handler.toolResults, 1)
	assert.Equal(t, "file1.txt\nfile2.txt", handler.toolResults[0].content)
	assert.False(t, handler.toolResults[0].isError)
}

func TestParser_ToolStreamHandler_ToolResultError(t *testing.T) {
	input := `{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"command not found","is_error":true}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockToolStreamHandler{}
	parser := NewParser(handler)
	_, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Len(t, handler.toolResults, 1)
	assert.Equal(t, "command not found", handler.toolResults[0].content)
	assert.True(t, handler.toolResults[0].isError)
}

func TestParser_BasicHandlerIgnoresToolCallbacks(t *testing.T) {
	// When using basic StreamHandler (not ToolStreamHandler),
	// tool_use and tool_result should not cause errors
	input := `{"type":"assistant","message":{"content":[{"type":"tool_use","id":"tool_1","name":"Bash","input":{"command":"ls"}}]}}
{"type":"user","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"tool_1","content":"result"}]}}
{"type":"assistant","message":{"content":[{"type":"text","text":"Done"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	handler := &mockStreamHandler{} // basic handler, not ToolStreamHandler
	parser := NewParser(handler)
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	assert.Equal(t, []string{"Done"}, handler.texts)
	assert.Equal(t, "Done", result.Output)
}

func TestParser_ParserReusable(t *testing.T) {
	// Verify that parser can be reused and receivedStreamEvents is reset
	handler := &mockStreamHandler{}
	parser := NewParser(handler)

	// First parse: with stream_events
	input1 := `{"type":"stream_event","event":{"type":"content_block_delta","index":0,"delta":{"type":"text_delta","text":"First"}}}
{"type":"assistant","message":{"content":[{"type":"text","text":"First Complete"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	result1, err := parser.Parse(strings.NewReader(input1))
	require.NoError(t, err)
	assert.Equal(t, []string{"First"}, handler.texts)
	assert.Equal(t, "First Complete", result1.Output)

	// Clear handler for second parse
	handler.texts = nil

	// Second parse: without stream_events (should stream from assistant)
	input2 := `{"type":"assistant","message":{"content":[{"type":"text","text":"Second"}]}}
{"type":"result","result":"Done","total_cost_usd":0.01,"is_error":false}`

	result2, err := parser.Parse(strings.NewReader(input2))
	require.NoError(t, err)
	// Should see "Second" because receivedStreamEvents was reset
	assert.Equal(t, []string{"Second"}, handler.texts)
	assert.Equal(t, "Second", result2.Output)
}
