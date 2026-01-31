package verifier

import (
	"context"
	"fmt"
	"time"
)

// DefaultVerifier implements the Verifier interface.
type DefaultVerifier struct {
	config        *Config
	registry      *CheckerRegistry
	client        ClaudeClient // Optional: for AI verification
	promptBuilder *PromptBuilder
}

// NewVerifier creates a new DefaultVerifier.
// client is optional and only required if EnableAI is true in config.
func NewVerifier(config *Config, client ClaudeClient) *DefaultVerifier {
	if config == nil {
		config = DefaultConfig()
	}

	return &DefaultVerifier{
		config:        config,
		registry:      NewCheckerRegistry(nil),
		client:        client,
		promptBuilder: NewPromptBuilder(),
	}
}

// NewVerifierWithRegistry creates a verifier with a custom registry.
func NewVerifierWithRegistry(config *Config, client ClaudeClient, registry *CheckerRegistry) *DefaultVerifier {
	if config == nil {
		config = DefaultConfig()
	}
	if registry == nil {
		registry = NewCheckerRegistry(nil)
	}

	return &DefaultVerifier{
		config:        config,
		registry:      registry,
		client:        client,
		promptBuilder: NewPromptBuilder(),
	}
}

// Verify verifies a task against its success criteria.
func (v *DefaultVerifier) Verify(ctx context.Context, task *VerificationTask) (*VerificationResult, error) {
	if task == nil {
		return nil, ErrNilTask
	}
	if len(task.SuccessCriteria) == 0 {
		return nil, ErrNoCriteria
	}

	start := time.Now()
	result := &VerificationResult{
		TaskID:    task.TaskID,
		Checks:    make([]CheckResult, 0, len(task.SuccessCriteria)),
		Timestamp: time.Now(),
	}

	workDir := task.WorkDir
	if workDir == "" {
		workDir = v.config.WorkDir
	}

	var totalCost float64

	for _, criterion := range task.SuccessCriteria {
		select {
		case <-ctx.Done():
			return nil, &VerifierError{
				Phase:   "execute",
				Message: "context cancelled",
				Err:     ctx.Err(),
			}
		default:
		}

		checkResult := v.verifyCriterion(ctx, task, criterion, workDir)
		if checkResult.CheckerType == "ai" && v.client != nil {
			// AI verification may have costs associated
			totalCost += v.extractCost(checkResult)
		}
		result.Checks = append(result.Checks, *checkResult)
	}

	result.Duration = time.Since(start)
	result.Cost = totalCost
	result.Passed = result.AllPassed()

	return result, nil
}

// verifyCriterion verifies a single criterion.
func (v *DefaultVerifier) verifyCriterion(ctx context.Context, task *VerificationTask, criterion string, workDir string) *CheckResult {
	// Try to find a built-in checker
	checker := v.registry.FindChecker(criterion)

	if checker != nil {
		// Create per-check context with timeout
		checkCtx := ctx
		if v.config.Timeout > 0 {
			var cancel context.CancelFunc
			checkCtx, cancel = context.WithTimeout(ctx, v.config.Timeout)
			defer cancel()
		}

		return checker.Check(checkCtx, criterion, workDir)
	}

	// No built-in checker found - use AI if enabled
	if v.config.EnableAI && v.client != nil {
		return v.verifyWithAI(ctx, task, criterion)
	}

	// Return unknown check result
	return &CheckResult{
		Criterion:   criterion,
		CheckerType: "unknown",
		Passed:      false,
		Error:       "no suitable checker found and AI verification is disabled",
	}
}

// verifyWithAI uses Claude to verify a criterion.
func (v *DefaultVerifier) verifyWithAI(ctx context.Context, task *VerificationTask, criterion string) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Criterion:   criterion,
		CheckerType: "ai",
	}

	promptResult, err := v.promptBuilder.Build(VerifyContext{
		TaskID:          task.TaskID,
		TaskTitle:       task.Title,
		TaskDescription: task.Description,
		Criterion:       criterion,
		Files:           task.Files,
	})
	if err != nil {
		result.Passed = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		return result
	}

	iterResult, err := v.client.Execute(ctx, promptResult.Prompt)
	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("AI verification failed: %v", err)
		return result
	}

	passed, evidence := ParseVerificationResponse(iterResult.Output)
	result.Passed = passed
	result.Evidence = &Evidence{
		Type:      EvidenceTypeAIAnalysis,
		Content:   evidence,
		Timestamp: time.Now(),
	}

	if !passed {
		result.Error = "AI verification determined criterion not met"
	}

	return result
}

func (v *DefaultVerifier) extractCost(checkResult *CheckResult) float64 {
	// Cost tracking would require storing cost in Evidence or CheckResult
	// For now, return 0 as costs are tracked at the iteration level
	return 0
}

// Config returns the verifier's configuration.
func (v *DefaultVerifier) Config() *Config {
	return v.config
}

// Registry returns the verifier's checker registry.
func (v *DefaultVerifier) Registry() *CheckerRegistry {
	return v.registry
}

// Compile-time interface check.
var _ Verifier = (*DefaultVerifier)(nil)
