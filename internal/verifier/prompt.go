package verifier

import (
	"fmt"
	"strings"
)

// PromptBuilder builds prompts for AI-based verification.
type PromptBuilder struct{}

// NewPromptBuilder creates a new PromptBuilder.
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

// VerifyContext contains inputs for building a verification prompt.
type VerifyContext struct {
	TaskID          string
	TaskTitle       string
	TaskDescription string
	Criterion       string
	Files           []string
	PreviousOutput  string // Output from command execution (if any)
}

// VerifyPromptResult contains the built prompt.
type VerifyPromptResult struct {
	Prompt string
}

// TemplateVerificationContext provides context for AI verification.
const TemplateVerificationContext = `## TASK VERIFICATION CONTEXT

You are verifying whether a task has been completed successfully.

**Task ID**: %s
**Task Title**: %s
**Task Description**:
%s

**Success Criterion to Verify**:
%s

**Files Associated with Task**:
%s

`

// TemplateVerificationInstructions provides instructions for AI verification.
const TemplateVerificationInstructions = `## VERIFICATION INSTRUCTIONS

Please verify if the success criterion above has been met.

1. Examine the relevant files and code
2. Run any necessary commands to verify functionality
3. Provide a clear PASS or FAIL verdict

**Response Format**:
- Start with either "VERIFICATION_PASS" or "VERIFICATION_FAIL"
- Provide brief evidence supporting your verdict
- If FAIL, explain what is missing or incorrect

**Important**: Be strict in your verification. The criterion must be fully met, not partially.
`

// Build constructs a verification prompt.
func (b *PromptBuilder) Build(ctx VerifyContext) (*VerifyPromptResult, error) {
	if ctx.Criterion == "" {
		return nil, &VerifierError{Phase: "prompt", Message: "criterion is required"}
	}

	filesStr := "- (no specific files)"
	if len(ctx.Files) > 0 {
		filesStr = "- " + strings.Join(ctx.Files, "\n- ")
	}

	prompt := fmt.Sprintf(TemplateVerificationContext,
		ctx.TaskID,
		ctx.TaskTitle,
		ctx.TaskDescription,
		ctx.Criterion,
		filesStr,
	)

	if ctx.PreviousOutput != "" {
		prompt += fmt.Sprintf(`**Previous Command Output**:
%s

`, "```\n"+ctx.PreviousOutput+"\n```")
	}

	prompt += TemplateVerificationInstructions

	return &VerifyPromptResult{Prompt: prompt}, nil
}

// ParseVerificationResponse extracts pass/fail from AI response.
func ParseVerificationResponse(response string) (passed bool, evidence string) {
	upper := strings.ToUpper(response)

	if strings.Contains(upper, "VERIFICATION_PASS") {
		return true, response
	}
	if strings.Contains(upper, "VERIFICATION_FAIL") {
		return false, response
	}

	// Fallback heuristics
	if strings.Contains(upper, "PASS") && !strings.Contains(upper, "FAIL") {
		return true, response
	}

	return false, response
}
