package planner

import (
	"fmt"
	"strings"
)

// TemplatePlanningConstraints provides shared constraints for all planning phases.
const TemplatePlanningConstraints = `## PLANNING CONSTRAINTS

When creating plans and decomposing tasks:
1. Be specific and actionable - each task should be implementable in a single session
2. Reference specific files and functions where possible
3. Consider dependencies between tasks
4. Estimate complexity (small/medium/large) for each task
5. Identify potential risks and blockers
6. NEVER guess or assume - base every decision on verified facts

## OUTPUT FORMAT

Respond with structured output that can be parsed. Use markdown headers and lists consistently.
`

// TemplatePRDPhase is the prompt template for PRD generation.
const TemplatePRDPhase = `## PRODUCT REQUIREMENTS DOCUMENT (PRD) PHASE

You are creating a PRD for the following user request:

%s

Analyze the request and produce a structured PRD with:

### Goals
What are we trying to achieve? List 2-5 clear goals.

### Requirements
What specific features/changes are needed? List each requirement.

### Constraints
What limitations should we respect? (existing code patterns, dependencies, etc.)

### Success Criteria
How do we know when we're done? List measurable criteria.

` + TemplatePlanningConstraints

// TemplateArchitecturePhase is the prompt template for architecture design.
const TemplateArchitecturePhase = `## ARCHITECTURE DESIGN PHASE

Based on the following PRD:

%s

Design the technical architecture:

### Components
What modules/packages are needed? For each component:
- Name
- Description
- Files to create/modify

### Dependencies
External libraries or internal imports needed.

### File Structure
What files need to be created or modified?

### Technical Decisions
Key choices and their rationale.

` + TemplatePlanningConstraints

// TemplateTasksPhase is the prompt template for task decomposition.
const TemplateTasksPhase = `## TASK DECOMPOSITION PHASE

Based on the following architecture:

%s

Decompose into executable tasks:
1. Each task should be completable in one session
2. Tasks must have clear dependencies
3. Include file paths and specific changes
4. Order tasks by dependency (independent tasks first)

Output tasks in this format:

### Task T001: [Brief Title]
- **Description**: What to do
- **Dependencies**: [T000] or none
- **Files**: [list of files to modify]
- **Complexity**: small/medium/large

` + TemplatePlanningConstraints

// PromptBuilder builds prompts for planning phases.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// BuildPRDPrompt constructs the PRD phase prompt.
func (b *PromptBuilder) BuildPRDPrompt(userPrompt string) string {
	return fmt.Sprintf(TemplatePRDPhase, userPrompt)
}

// BuildArchitecturePrompt constructs the architecture phase prompt.
func (b *PromptBuilder) BuildArchitecturePrompt(prd *PRD) (string, error) {
	if prd == nil {
		return "", &PlannerError{
			Phase:   "prompt",
			Message: "PRD is required for architecture prompt",
		}
	}

	prdSummary := formatPRDSummary(prd)
	return fmt.Sprintf(TemplateArchitecturePhase, prdSummary), nil
}

// BuildTasksPrompt constructs the task decomposition prompt.
func (b *PromptBuilder) BuildTasksPrompt(arch *Architecture) (string, error) {
	if arch == nil {
		return "", &PlannerError{
			Phase:   "prompt",
			Message: "Architecture is required for tasks prompt",
		}
	}

	archSummary := formatArchitectureSummary(arch)
	return fmt.Sprintf(TemplateTasksPhase, archSummary), nil
}

// formatPRDSummary formats a PRD into a readable summary for prompts.
func formatPRDSummary(prd *PRD) string {
	var sb strings.Builder

	sb.WriteString("### Goals\n")
	for _, goal := range prd.Goals {
		sb.WriteString(fmt.Sprintf("- %s\n", goal))
	}

	sb.WriteString("\n### Requirements\n")
	for _, req := range prd.Requirements {
		sb.WriteString(fmt.Sprintf("- %s\n", req))
	}

	sb.WriteString("\n### Constraints\n")
	for _, constraint := range prd.Constraints {
		sb.WriteString(fmt.Sprintf("- %s\n", constraint))
	}

	sb.WriteString("\n### Success Criteria\n")
	for _, criteria := range prd.SuccessCriteria {
		sb.WriteString(fmt.Sprintf("- %s\n", criteria))
	}

	return sb.String()
}

// formatArchitectureSummary formats an Architecture into a readable summary for prompts.
func formatArchitectureSummary(arch *Architecture) string {
	var sb strings.Builder

	sb.WriteString("### Components\n")
	for _, comp := range arch.Components {
		sb.WriteString(fmt.Sprintf("- **%s**: %s\n", comp.Name, comp.Description))
		if len(comp.Files) > 0 {
			sb.WriteString(fmt.Sprintf("  - Files: %s\n", strings.Join(comp.Files, ", ")))
		}
	}

	sb.WriteString("\n### Dependencies\n")
	for _, dep := range arch.Dependencies {
		sb.WriteString(fmt.Sprintf("- %s\n", dep))
	}

	sb.WriteString("\n### File Structure\n")
	for _, file := range arch.FileStructure {
		sb.WriteString(fmt.Sprintf("- %s\n", file))
	}

	sb.WriteString("\n### Technical Decisions\n")
	for _, decision := range arch.TechDecisions {
		sb.WriteString(fmt.Sprintf("- %s\n", decision))
	}

	return sb.String()
}
