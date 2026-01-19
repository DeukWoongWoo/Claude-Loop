package prompt

// Placeholder constants for template substitution.
const (
	// PlaceholderCompletionSignal is replaced with the actual completion signal phrase.
	PlaceholderCompletionSignal = "COMPLETION_SIGNAL_PLACEHOLDER"

	// PlaceholderPrinciplesYAML is replaced with the YAML-serialized principles.
	PlaceholderPrinciplesYAML = "PRINCIPLES_YAML_PLACEHOLDER"

	// PlaceholderNotesFile is replaced with the notes file path.
	PlaceholderNotesFile = "NOTES_FILE_PLACEHOLDER"
)

// TemplateWorkflowContext provides continuous workflow context.
// Contains: COMPLETION_SIGNAL_PLACEHOLDER
const TemplateWorkflowContext = `## CONTINUOUS WORKFLOW CONTEXT

This is part of a continuous development loop where work happens incrementally across multiple iterations. You might run once, then a human developer might make changes, then you run again, and so on. This could happen daily or on any schedule.

**Important**: You don't need to complete the entire goal in one iteration. Just make meaningful progress on one thing, then leave clear notes for the next iteration (human or AI). Think of it as a relay race where you're passing the baton.

**Project Completion Signal**: If you determine that not just your current task but the ENTIRE project goal is fully complete (nothing more to be done on the overall goal), only include the exact phrase "COMPLETION_SIGNAL_PLACEHOLDER" in your response. Only use this when absolutely certain that the whole project is finished, not just your individual task. We will stop working on this project when multiple developers independently determine that the project is complete.

## PRIMARY GOAL`

// TemplateDecisionPrinciples provides decision-making guidance.
// Contains: PRINCIPLES_YAML_PLACEHOLDER
const TemplateDecisionPrinciples = `## DECISION PRINCIPLES

The following principles guide your autonomous decision-making for this project:

` + "```yaml" + `
PRINCIPLES_YAML_PLACEHOLDER
` + "```" + `

### Decision Protocol (R10: 3-Step Resolution)

When making decisions, apply this protocol:

1. **Compatibility Check**: Are conflicting principles about different dimensions?
   - Scope = breadth (what to build), Trust = depth (quality level) → Compatible
   - Both about breadth (Scope vs UX) → Real conflict

2. **Type Classification**: Constraint vs Objective
   - **Constraints** (must satisfy first): Scope, Cost, Security, Privacy, Time
   - **Objectives** (optimize within constraints): Trust, UX, Innovation, Growth, Quality
   - Rule: Satisfy constraints first → Optimize objectives within constraints

3. **Priority Resolution**: When same-type principles conflict
   - Legal/Regulatory > Security > User Intent > Data Integrity > Quality > Speed > UX

### Response Format for Significant Decisions

When making non-trivial decisions, briefly report:
- **Decision**: What you decided
- **Rationale**: Which principle(s) applied (e.g., "Scope=3 + Trust=7")

### When to Ask (Rare)

Only ask the user when ALL conditions are met:
- Weight difference < 2 between conflicting principles
- R10 3-step resolution doesn't resolve it
- Options are mutually exclusive (can't satisfy both)

**Default behavior**: Decide autonomously and report your reasoning.
`

// TemplateNotesUpdateExisting instructs to update existing notes file.
// Contains: NOTES_FILE_PLACEHOLDER
const TemplateNotesUpdateExisting = "Update the `NOTES_FILE_PLACEHOLDER` file with relevant context for the next iteration. Add new notes and remove outdated information to keep it current and useful."

// TemplateNotesCreateNew instructs to create new notes file.
// Contains: NOTES_FILE_PLACEHOLDER
const TemplateNotesCreateNew = "Create a `NOTES_FILE_PLACEHOLDER` file with relevant context and instructions for the next iteration."

// TemplateNotesGuidelines provides guidelines for notes files.
const TemplateNotesGuidelines = `

This file helps coordinate work across iterations (both human and AI developers). It should:

- Contain relevant context and instructions for the next iteration
- Stay concise and actionable (like a notes file, not a detailed report)
- Help the next developer understand what to do next

The file should NOT include:
- Lists of completed work or full reports
- Information that can be discovered by running tests/coverage
- Unnecessary details`

// TemplateNotesContext wraps notes content from previous iteration.
// Contains: NOTES_FILE_PLACEHOLDER for the file name
const TemplateNotesContext = `## CONTEXT FROM PREVIOUS ITERATION

The following is from ` + "`" + `NOTES_FILE_PLACEHOLDER` + "`" + `, maintained by previous iterations to provide context:

`

// TemplateIterationNotes header for notes instructions.
const TemplateIterationNotes = `## ITERATION NOTES

`

// TemplateReviewerContext provides context for review passes.
const TemplateReviewerContext = `## CODE REVIEW CONTEXT

You are performing a review pass on changes just made by another developer. This is NOT a new feature implementation - you are reviewing and validating existing changes using the instructions given below by the user. Feel free to use git commands to see what changes were made if it's helpful to you.`

// TemplateCIFixContext provides context for CI failure fixes.
const TemplateCIFixContext = `## CI FAILURE FIX CONTEXT

You are analyzing and fixing a CI/CD failure for a pull request.

**Your task:**
1. Inspect the failed CI workflow using the commands below
2. Analyze the error logs to understand what went wrong
3. Make the necessary code changes to fix the issue
4. Stage and commit your changes (they will be pushed to update the PR)

**Commands to inspect CI failures:**
- ` + "`" + `gh run list --status failure --limit 3` + "`" + ` - List recent failed runs
- ` + "`" + `gh run view <RUN_ID> --log-failed` + "`" + ` - View failed job logs (shorter output)
- ` + "`" + `gh run view <RUN_ID> --log` + "`" + ` - View full logs for a specific run

**Important:**
- Focus only on fixing the CI failure, not adding new features
- Make minimal changes necessary to pass CI
- If the failure seems unfixable (e.g., flaky test, infrastructure issue), explain why in your response`
