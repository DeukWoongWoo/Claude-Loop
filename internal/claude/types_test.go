package claude

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStreamMessage_UnmarshalAssistant(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"Hello, world!"}]}}`

	var msg StreamMessage
	err := json.Unmarshal([]byte(input), &msg)
	require.NoError(t, err)

	assert.Equal(t, "assistant", msg.Type)
	require.NotNil(t, msg.Message)
	require.Len(t, msg.Message.Content, 1)
	assert.Equal(t, "text", msg.Message.Content[0].Type)
	assert.Equal(t, "Hello, world!", msg.Message.Content[0].Text)
}

func TestStreamMessage_UnmarshalResult(t *testing.T) {
	input := `{"type":"result","result":"Task completed successfully.","total_cost_usd":0.0523,"is_error":false}`

	var msg StreamMessage
	err := json.Unmarshal([]byte(input), &msg)
	require.NoError(t, err)

	assert.Equal(t, "result", msg.Type)
	assert.Equal(t, "Task completed successfully.", msg.Result)
	assert.InDelta(t, 0.0523, msg.TotalCostUSD, 0.0001)
	assert.False(t, msg.IsError)
}

func TestStreamMessage_UnmarshalResultError(t *testing.T) {
	input := `{"type":"result","result":"Something went wrong","total_cost_usd":0.01,"is_error":true}`

	var msg StreamMessage
	err := json.Unmarshal([]byte(input), &msg)
	require.NoError(t, err)

	assert.Equal(t, "result", msg.Type)
	assert.Equal(t, "Something went wrong", msg.Result)
	assert.True(t, msg.IsError)
}

func TestStreamMessage_UnmarshalMultipleContentBlocks(t *testing.T) {
	input := `{"type":"assistant","message":{"content":[{"type":"text","text":"First"},{"type":"tool_use","id":"123"},{"type":"text","text":"Second"}]}}`

	var msg StreamMessage
	err := json.Unmarshal([]byte(input), &msg)
	require.NoError(t, err)

	require.NotNil(t, msg.Message)
	require.Len(t, msg.Message.Content, 3)

	assert.Equal(t, "text", msg.Message.Content[0].Type)
	assert.Equal(t, "First", msg.Message.Content[0].Text)

	assert.Equal(t, "tool_use", msg.Message.Content[1].Type)
	assert.Empty(t, msg.Message.Content[1].Text)

	assert.Equal(t, "text", msg.Message.Content[2].Type)
	assert.Equal(t, "Second", msg.Message.Content[2].Text)
}

func TestParsedResult_ZeroValue(t *testing.T) {
	var result ParsedResult

	assert.Empty(t, result.Output)
	assert.Empty(t, result.ResultText)
	assert.Zero(t, result.TotalCostUSD)
	assert.False(t, result.IsError)
	assert.Nil(t, result.RawMessages)
}
