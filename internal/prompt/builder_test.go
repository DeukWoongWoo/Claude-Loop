package prompt

import (
	"errors"
	"strings"
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	t.Parallel()

	builder := NewBuilder()
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.notesLoader)
}

func TestNewBuilderWithLoader(t *testing.T) {
	t.Parallel()

	mockLoader := &MockNotesLoader{}
	builder := NewBuilderWithLoader(mockLoader)

	assert.NotNil(t, builder)
	assert.Equal(t, mockLoader, builder.notesLoader)
}

func TestBuilder_Build_BasicPrompt(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Implement a new feature",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "PROJECT_COMPLETE",
		NotesFile:                "SHARED_TASK_NOTES.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Prompt)
	assert.False(t, result.NotesIncluded)
	assert.False(t, result.PrinciplesInjected)

	// Verify user prompt is included
	assert.Contains(t, result.Prompt, "Implement a new feature")

	// Verify workflow context is included with signal replaced
	assert.Contains(t, result.Prompt, "CONTINUOUS WORKFLOW CONTEXT")
	assert.Contains(t, result.Prompt, "PROJECT_COMPLETE")
	assert.NotContains(t, result.Prompt, PlaceholderCompletionSignal)

	// Verify notes create instruction (since notes don't exist)
	assert.Contains(t, result.Prompt, "Create a `SHARED_TASK_NOTES.md` file")
}

func TestBuilder_Build_WithPrinciples(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	principles := &config.Principles{
		Version:   "2.3",
		Preset:    config.PresetStartup,
		CreatedAt: "2026-01-11",
		Layer0: config.Layer0{
			TrustArchitecture: 5,
			CurationModel:     3,
			ScopePhilosophy:   7,
		},
		Layer1: config.Layer1{
			SpeedCorrectness:    8,
			InnovationStability: 4,
		},
	}

	ctx := BuildContext{
		UserPrompt:               "Build the app",
		Principles:               principles,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.True(t, result.PrinciplesInjected)

	// Verify principles are included as YAML
	assert.Contains(t, result.Prompt, "DECISION PRINCIPLES")
	assert.Contains(t, result.Prompt, "version: \"2.3\"")
	assert.Contains(t, result.Prompt, "preset: startup")
	assert.NotContains(t, result.Prompt, PlaceholderPrinciplesYAML)
}

func TestBuilder_Build_WithPrincipleCollection(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Start the project",
		Principles:               nil,
		NeedsPrincipleCollection: true,
		CompletionSignal:         "COMPLETE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.False(t, result.PrinciplesInjected)

	// Verify principle collection prompt is included
	assert.Contains(t, result.Prompt, "PRINCIPLE COLLECTION REQUIRED")
	assert.Contains(t, result.Prompt, "AskUserQuestion")
	assert.Contains(t, result.Prompt, "Project Type")
}

func TestBuilder_Build_PrincipleCollectionOnlyOnFirstIteration(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	// Iteration 2 with NeedsPrincipleCollection=true should NOT include collection prompt
	ctx := BuildContext{
		UserPrompt:               "Continue work",
		Principles:               nil,
		NeedsPrincipleCollection: true,
		CompletionSignal:         "COMPLETE",
		NotesFile:                "notes.md",
		Iteration:                2, // Not first iteration
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	// Principle collection should NOT be included on iteration 2
	assert.NotContains(t, result.Prompt, "PRINCIPLE COLLECTION REQUIRED")
}

func TestBuilder_Build_WithExistingNotes(t *testing.T) {
	t.Parallel()

	notesContent := "# Previous Notes\n\n- Completed task A\n- Working on task B"
	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: notesContent,
		Exists:  true,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Continue the work",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "SHARED_TASK_NOTES.md",
		Iteration:                3,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.True(t, result.NotesIncluded)

	// Verify notes content is included
	assert.Contains(t, result.Prompt, "CONTEXT FROM PREVIOUS ITERATION")
	assert.Contains(t, result.Prompt, notesContent)
	assert.Contains(t, result.Prompt, "SHARED_TASK_NOTES.md")

	// Verify update instruction (not create) since notes exist
	assert.Contains(t, result.Prompt, "Update the `SHARED_TASK_NOTES.md` file")
	assert.NotContains(t, result.Prompt, "Create a `SHARED_TASK_NOTES.md` file")
}

func TestBuilder_Build_WithEmptyNotesFile(t *testing.T) {
	t.Parallel()

	// File exists but is empty
	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  true,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Do something",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	// Empty content should not be included
	assert.False(t, result.NotesIncluded)
	assert.NotContains(t, result.Prompt, "CONTEXT FROM PREVIOUS ITERATION")

	// But update instruction should still be used since file exists
	assert.Contains(t, result.Prompt, "Update the `notes.md` file")
}

func TestBuilder_Build_CompletionSignalReplacement(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	customSignal := "MY_CUSTOM_COMPLETION_SIGNAL_12345"
	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         customSignal,
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.Contains(t, result.Prompt, customSignal)
	assert.NotContains(t, result.Prompt, PlaceholderCompletionSignal)
}

func TestBuilder_Build_NotesFileReplacement(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	customNotesFile := "MY_CUSTOM_NOTES_FILE.md"
	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                customNotesFile,
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.Contains(t, result.Prompt, customNotesFile)
	assert.NotContains(t, result.Prompt, PlaceholderNotesFile)
}

func TestBuilder_Build_ErrorOnNotesLoad(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("permission denied")
	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     expectedErr,
	})

	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to load notes")
}

func TestBuilder_Build_FullIntegration(t *testing.T) {
	t.Parallel()

	notesContent := "# Task Notes\n\nPrevious work completed."
	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: notesContent,
		Exists:  true,
		Err:     nil,
	})

	principles := &config.Principles{
		Version:   "2.3",
		Preset:    config.PresetEnterprise,
		CreatedAt: "2026-01-12",
		Layer0: config.Layer0{
			TrustArchitecture: 8,
			PrivacyPosture:    9,
		},
		Layer1: config.Layer1{
			SpeedCorrectness: 7,
			SecurityPosture:  9,
		},
	}

	ctx := BuildContext{
		UserPrompt:               "Implement authentication",
		Principles:               principles,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "AUTH_COMPLETE",
		NotesFile:                "AUTH_NOTES.md",
		Iteration:                5,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.True(t, result.PrinciplesInjected)
	assert.True(t, result.NotesIncluded)

	// Verify all sections are present in order
	prompt := result.Prompt

	// Decision principles should come first (no principle collection needed)
	principlesIdx := strings.Index(prompt, "DECISION PRINCIPLES")
	workflowIdx := strings.Index(prompt, "CONTINUOUS WORKFLOW CONTEXT")
	userPromptIdx := strings.Index(prompt, "Implement authentication")
	notesContextIdx := strings.Index(prompt, "CONTEXT FROM PREVIOUS ITERATION")
	iterationNotesIdx := strings.Index(prompt, "ITERATION NOTES")

	assert.True(t, principlesIdx < workflowIdx, "principles should come before workflow")
	assert.True(t, workflowIdx < userPromptIdx, "workflow should come before user prompt")
	assert.True(t, userPromptIdx < notesContextIdx, "user prompt should come before notes context")
	assert.True(t, notesContextIdx < iterationNotesIdx, "notes context should come before iteration notes")

	// Verify specific content
	assert.Contains(t, prompt, "preset: enterprise")
	assert.Contains(t, prompt, "AUTH_COMPLETE")
	assert.Contains(t, prompt, "AUTH_NOTES.md")
	assert.Contains(t, prompt, notesContent)
}

func TestBuilder_Build_PromptOrdering(t *testing.T) {
	t.Parallel()

	// Test ordering when principle collection is needed
	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "notes",
		Exists:  true,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "User prompt here",
		Principles:               nil,
		NeedsPrincipleCollection: true,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)

	// Principle collection should come first
	collectionIdx := strings.Index(result.Prompt, "PRINCIPLE COLLECTION REQUIRED")
	workflowIdx := strings.Index(result.Prompt, "CONTINUOUS WORKFLOW CONTEXT")
	userPromptIdx := strings.Index(result.Prompt, "User prompt here")

	assert.True(t, collectionIdx < workflowIdx, "collection should come before workflow")
	assert.True(t, workflowIdx < userPromptIdx, "workflow should come before user prompt")
}

func TestBuilder_Build_NotesGuidelines(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	// Notes guidelines should always be included
	assert.Contains(t, result.Prompt, "coordinate work across iterations")
	assert.Contains(t, result.Prompt, "concise and actionable")
}

func TestBuilderInterface(t *testing.T) {
	t.Parallel()

	// Verify that DefaultBuilder implements the Builder interface
	var _ Builder = &DefaultBuilder{}
}

func TestBuilder_Build_EmptyUserPrompt(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	// Should still work with empty user prompt
	result, err := builder.Build(ctx)

	require.NoError(t, err)
	assert.NotEmpty(t, result.Prompt)
	assert.Contains(t, result.Prompt, "CONTINUOUS WORKFLOW CONTEXT")
}

func TestBuilder_Build_EmptyCompletionSignal(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "",
		NotesFile:                "notes.md",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	// Placeholder should be replaced with empty string
	assert.NotContains(t, result.Prompt, PlaceholderCompletionSignal)
}

func TestBuilder_Build_EmptyNotesFile(t *testing.T) {
	t.Parallel()

	builder := NewBuilderWithLoader(&MockNotesLoader{
		Content: "",
		Exists:  false,
		Err:     nil,
	})

	ctx := BuildContext{
		UserPrompt:               "Test",
		Principles:               nil,
		NeedsPrincipleCollection: false,
		CompletionSignal:         "DONE",
		NotesFile:                "",
		Iteration:                1,
	}

	result, err := builder.Build(ctx)

	require.NoError(t, err)
	// Placeholder should be replaced with empty string
	assert.NotContains(t, result.Prompt, PlaceholderNotesFile)
}
