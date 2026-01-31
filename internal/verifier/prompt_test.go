package verifier

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPromptBuilder_Build(t *testing.T) {
	builder := NewPromptBuilder()

	tests := []struct {
		name        string
		ctx         VerifyContext
		wantErr     bool
		wantContain []string
	}{
		{
			name: "basic context",
			ctx: VerifyContext{
				TaskID:          "T001",
				TaskTitle:       "Test Task",
				TaskDescription: "A test task description",
				Criterion:       "file `test.go` exists",
				Files:           []string{"test.go"},
			},
			wantErr: false,
			wantContain: []string{
				"T001",
				"Test Task",
				"A test task description",
				"file `test.go` exists",
				"test.go",
				"VERIFICATION_PASS",
				"VERIFICATION_FAIL",
			},
		},
		{
			name: "empty criterion",
			ctx: VerifyContext{
				TaskID:    "T001",
				TaskTitle: "Test",
				Criterion: "",
			},
			wantErr: true,
		},
		{
			name: "no files",
			ctx: VerifyContext{
				TaskID:    "T001",
				TaskTitle: "Test",
				Criterion: "build passes",
				Files:     []string{},
			},
			wantErr: false,
			wantContain: []string{
				"no specific files",
			},
		},
		{
			name: "with previous output",
			ctx: VerifyContext{
				TaskID:         "T001",
				TaskTitle:      "Test",
				Criterion:      "build passes",
				PreviousOutput: "build successful",
			},
			wantErr: false,
			wantContain: []string{
				"Previous Command Output",
				"build successful",
			},
		},
		{
			name: "multiple files",
			ctx: VerifyContext{
				TaskID:    "T001",
				TaskTitle: "Test",
				Criterion: "all files exist",
				Files:     []string{"file1.go", "file2.go", "file3.go"},
			},
			wantErr: false,
			wantContain: []string{
				"file1.go",
				"file2.go",
				"file3.go",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := builder.Build(tt.ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.Prompt)

			for _, contain := range tt.wantContain {
				assert.Contains(t, result.Prompt, contain)
			}
		})
	}
}

func TestPromptBuilder_Build_ErrorIsVerifierError(t *testing.T) {
	builder := NewPromptBuilder()

	_, err := builder.Build(VerifyContext{Criterion: ""})

	require.Error(t, err)
	assert.True(t, IsVerifierError(err))
}

func TestParseVerificationResponse(t *testing.T) {
	tests := []struct {
		name         string
		response     string
		wantPassed   bool
		wantEvidence string
	}{
		{
			name:         "explicit pass",
			response:     "VERIFICATION_PASS: All criteria met",
			wantPassed:   true,
			wantEvidence: "VERIFICATION_PASS: All criteria met",
		},
		{
			name:         "explicit fail",
			response:     "VERIFICATION_FAIL: File not found",
			wantPassed:   false,
			wantEvidence: "VERIFICATION_FAIL: File not found",
		},
		{
			name:         "lowercase pass",
			response:     "verification_pass: done",
			wantPassed:   true,
			wantEvidence: "verification_pass: done",
		},
		{
			name:         "lowercase fail",
			response:     "verification_fail: missing",
			wantPassed:   false,
			wantEvidence: "verification_fail: missing",
		},
		{
			name:         "fallback pass only",
			response:     "The test PASS successfully",
			wantPassed:   true,
			wantEvidence: "The test PASS successfully",
		},
		{
			name:         "both pass and fail - fail wins",
			response:     "PASS but then FAIL",
			wantPassed:   false,
			wantEvidence: "PASS but then FAIL",
		},
		{
			name:         "no keywords - defaults to fail",
			response:     "Something happened",
			wantPassed:   false,
			wantEvidence: "Something happened",
		},
		{
			name:         "empty response",
			response:     "",
			wantPassed:   false,
			wantEvidence: "",
		},
		{
			name:         "multiline response with pass",
			response:     "Analysis complete\nVERIFICATION_PASS\nAll checks passed",
			wantPassed:   true,
			wantEvidence: "Analysis complete\nVERIFICATION_PASS\nAll checks passed",
		},
		{
			name:         "multiline response with fail",
			response:     "Analysis complete\nVERIFICATION_FAIL\nSome checks failed",
			wantPassed:   false,
			wantEvidence: "Analysis complete\nVERIFICATION_FAIL\nSome checks failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			passed, evidence := ParseVerificationResponse(tt.response)
			assert.Equal(t, tt.wantPassed, passed)
			assert.Equal(t, tt.wantEvidence, evidence)
		})
	}
}

func TestTemplateConstants(t *testing.T) {
	// Verify template constants are non-empty
	assert.NotEmpty(t, TemplateVerificationContext)
	assert.NotEmpty(t, TemplateVerificationInstructions)

	// Verify templates contain expected placeholders
	assert.Contains(t, TemplateVerificationContext, "%s")
	assert.Contains(t, TemplateVerificationInstructions, "VERIFICATION_PASS")
	assert.Contains(t, TemplateVerificationInstructions, "VERIFICATION_FAIL")
}

func TestNewPromptBuilder(t *testing.T) {
	builder := NewPromptBuilder()
	assert.NotNil(t, builder)
}

func TestVerifyContext_Fields(t *testing.T) {
	ctx := VerifyContext{
		TaskID:          "T001",
		TaskTitle:       "Title",
		TaskDescription: "Description",
		Criterion:       "Criterion",
		Files:           []string{"file1.go"},
		PreviousOutput:  "Output",
	}

	assert.Equal(t, "T001", ctx.TaskID)
	assert.Equal(t, "Title", ctx.TaskTitle)
	assert.Equal(t, "Description", ctx.TaskDescription)
	assert.Equal(t, "Criterion", ctx.Criterion)
	assert.Len(t, ctx.Files, 1)
	assert.Equal(t, "Output", ctx.PreviousOutput)
}

func TestVerifyPromptResult_Fields(t *testing.T) {
	result := VerifyPromptResult{
		Prompt: "test prompt",
	}

	assert.Equal(t, "test prompt", result.Prompt)
}
