// Package prompt provides prompt building functionality for Claude Code iterations.
package prompt

import "github.com/DeukWoongWoo/claude-loop/internal/config"

// BuildContext contains all context needed to build an enhanced prompt.
type BuildContext struct {
	// UserPrompt is the user's original prompt from -p flag.
	UserPrompt string

	// Principles is the loaded principles configuration (may be nil if not loaded).
	Principles *config.Principles

	// CompletionSignal is the phrase used to signal project completion
	// (e.g., "CONTINUOUS_CLAUDE_PROJECT_COMPLETE").
	CompletionSignal string

	// NotesFile is the path to the notes file (e.g., "SHARED_TASK_NOTES.md").
	NotesFile string

	// Iteration is the current iteration number (1-based).
	Iteration int
}

// BuildResult contains the built prompt and metadata.
type BuildResult struct {
	// Prompt is the complete enhanced prompt ready for Claude.
	Prompt string

	// NotesIncluded indicates whether notes file existed and was included.
	NotesIncluded bool

	// PrinciplesInjected indicates whether principles were injected.
	PrinciplesInjected bool
}
