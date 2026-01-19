package prompt

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Builder is the interface for building enhanced prompts.
type Builder interface {
	// Build constructs an enhanced prompt from the given context.
	Build(ctx BuildContext) (*BuildResult, error)
}

// DefaultBuilder implements the Builder interface.
type DefaultBuilder struct {
	notesLoader NotesLoader
}

// NewBuilder creates a new DefaultBuilder with a FileNotesLoader.
func NewBuilder() *DefaultBuilder {
	return &DefaultBuilder{
		notesLoader: NewFileNotesLoader(),
	}
}

// NewBuilderWithLoader creates a DefaultBuilder with a custom NotesLoader.
// This is primarily useful for testing.
func NewBuilderWithLoader(loader NotesLoader) *DefaultBuilder {
	return &DefaultBuilder{
		notesLoader: loader,
	}
}

// Build constructs an enhanced prompt from the given context.
// The prompt is built in the following order:
// 1. [Conditional] Decision principles (if Principles != nil)
// 2. Workflow context (with CompletionSignal placeholder replaced)
// 3. User prompt
// 4. [Conditional] Notes from previous iteration (if file exists)
// 5. Notes instructions (UPDATE or CREATE)
// 6. Notes guidelines
func (b *DefaultBuilder) Build(ctx BuildContext) (*BuildResult, error) {
	var sb strings.Builder
	result := &BuildResult{}

	// 1. Decision Principles (if principles are loaded)
	if ctx.Principles != nil {
		principlesYAML, err := yaml.Marshal(ctx.Principles)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal principles: %w", err)
		}

		principlesPrompt := strings.ReplaceAll(
			TemplateDecisionPrinciples,
			PlaceholderPrinciplesYAML,
			string(principlesYAML),
		)
		sb.WriteString(principlesPrompt)
		sb.WriteString("\n\n")
		result.PrinciplesInjected = true
	}

	// 2. Workflow Context (with completion signal replaced)
	workflowContext := strings.ReplaceAll(
		TemplateWorkflowContext,
		PlaceholderCompletionSignal,
		ctx.CompletionSignal,
	)
	sb.WriteString(workflowContext)
	sb.WriteString("\n\n")

	// 3. User Prompt
	sb.WriteString(ctx.UserPrompt)
	sb.WriteString("\n\n")

	// 4. Notes from Previous Iteration (if file exists)
	notesContent, notesExists, err := b.notesLoader.Load(ctx.NotesFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", err)
	}

	if notesExists && notesContent != "" {
		notesHeader := strings.ReplaceAll(
			TemplateNotesContext,
			PlaceholderNotesFile,
			ctx.NotesFile,
		)
		sb.WriteString(notesHeader)
		sb.WriteString(notesContent)
		sb.WriteString("\n\n")
		result.NotesIncluded = true
	}

	// 5. Iteration Notes Instructions (only if NotesFile is specified)
	if ctx.NotesFile != "" {
		sb.WriteString(TemplateIterationNotes)

		notesTemplate := TemplateNotesCreateNew
		if notesExists {
			notesTemplate = TemplateNotesUpdateExisting
		}
		notesInstruction := strings.ReplaceAll(notesTemplate, PlaceholderNotesFile, ctx.NotesFile)
		sb.WriteString(notesInstruction)
	}

	// 6. Notes Guidelines (only if NotesFile is specified)
	if ctx.NotesFile != "" {
		sb.WriteString(TemplateNotesGuidelines)
	}

	result.Prompt = sb.String()
	return result, nil
}
