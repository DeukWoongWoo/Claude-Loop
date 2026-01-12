package prompt

import (
	"fmt"
	"strings"
)

// CIFixBuilder builds prompts for CI failure fixes.
type CIFixBuilder struct{}

// NewCIFixBuilder creates a new CIFixBuilder.
func NewCIFixBuilder() *CIFixBuilder {
	return &CIFixBuilder{}
}

// CIFailureInfo contains information about a CI failure.
// This mirrors github.CIFailureInfo but is defined here to avoid import cycles.
type CIFailureInfo struct {
	RunID        string   // Workflow run ID
	WorkflowName string   // Name of the workflow
	JobName      string   // Failed job name
	FailedSteps  []string // Names of failed steps
	ErrorLogs    string   // Error log output
	URL          string   // Link to the failed run
}

// CIFixContext contains context for building a CI fix prompt.
type CIFixContext struct {
	// FailureInfo contains CI failure details.
	FailureInfo *CIFailureInfo

	// PRNumber is the PR being fixed.
	PRNumber int

	// BranchName is the current branch.
	BranchName string

	// Attempt is the current attempt number (1-based).
	Attempt int

	// MaxAttempts is the maximum attempts allowed.
	MaxAttempts int
}

// Build constructs a CI fix prompt.
func (b *CIFixBuilder) Build(ctx CIFixContext) (*BuildResult, error) {
	if ctx.FailureInfo == nil {
		return nil, fmt.Errorf("FailureInfo is required")
	}

	var sb strings.Builder

	// Add CI fix context template
	fmt.Fprintf(&sb, "%s\n\n", TemplateCIFixContext)

	// Add failure details
	b.writeFailureDetails(&sb, ctx.FailureInfo)

	// Add attempt info if this is a retry
	if ctx.Attempt > 1 {
		fmt.Fprintf(&sb, "**Note**: This is attempt %d of %d. Previous fix attempts did not resolve the issue.\n\n",
			ctx.Attempt, ctx.MaxAttempts)
	}

	// Add error logs
	fmt.Fprintf(&sb, "## Error Logs\n\n```\n%s\n```\n\n", ctx.FailureInfo.ErrorLogs)

	// Add instructions
	b.writeInstructions(&sb, ctx.FailureInfo.RunID)

	return &BuildResult{
		Prompt: sb.String(),
	}, nil
}

// writeFailureDetails writes the failure details section.
func (b *CIFixBuilder) writeFailureDetails(sb *strings.Builder, info *CIFailureInfo) {
	fmt.Fprintf(sb, "## Failure Details\n\n")
	fmt.Fprintf(sb, "- **Run ID**: %s\n", info.RunID)
	fmt.Fprintf(sb, "- **Workflow**: %s\n", info.WorkflowName)
	if info.JobName != "" {
		fmt.Fprintf(sb, "- **Failed Job**: %s\n", info.JobName)
	}
	if len(info.FailedSteps) > 0 {
		fmt.Fprintf(sb, "- **Failed Steps**: %s\n", strings.Join(info.FailedSteps, ", "))
	}
	if info.URL != "" {
		fmt.Fprintf(sb, "- **Run URL**: %s\n", info.URL)
	}
	sb.WriteString("\n")
}

// writeInstructions writes the fix instructions section.
func (b *CIFixBuilder) writeInstructions(sb *strings.Builder, runID string) {
	sb.WriteString("## Instructions\n\n")
	sb.WriteString("1. Analyze the error logs above to understand what failed\n")
	fmt.Fprintf(sb, "2. Use `gh run view %s --log` if you need more context\n", runID)
	sb.WriteString("3. Make the minimal code changes necessary to fix the failure\n")
	sb.WriteString("4. Stage and commit your changes (they will be pushed automatically)\n")
}
