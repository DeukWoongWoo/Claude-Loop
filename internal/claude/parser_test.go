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
