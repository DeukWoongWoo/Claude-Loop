package verifier

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVerificationResult_AllPassed(t *testing.T) {
	tests := []struct {
		name   string
		checks []CheckResult
		want   bool
	}{
		{
			name:   "empty checks returns false",
			checks: []CheckResult{},
			want:   false,
		},
		{
			name: "all passed",
			checks: []CheckResult{
				{Criterion: "c1", Passed: true},
				{Criterion: "c2", Passed: true},
			},
			want: true,
		},
		{
			name: "one failed",
			checks: []CheckResult{
				{Criterion: "c1", Passed: true},
				{Criterion: "c2", Passed: false},
			},
			want: false,
		},
		{
			name: "all failed",
			checks: []CheckResult{
				{Criterion: "c1", Passed: false},
				{Criterion: "c2", Passed: false},
			},
			want: false,
		},
		{
			name: "single passed",
			checks: []CheckResult{
				{Criterion: "c1", Passed: true},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VerificationResult{Checks: tt.checks}
			assert.Equal(t, tt.want, r.AllPassed())
		})
	}
}

func TestVerificationResult_FailedChecks(t *testing.T) {
	tests := []struct {
		name      string
		checks    []CheckResult
		wantCount int
	}{
		{
			name:      "empty checks",
			checks:    []CheckResult{},
			wantCount: 0,
		},
		{
			name: "no failures",
			checks: []CheckResult{
				{Criterion: "c1", Passed: true},
				{Criterion: "c2", Passed: true},
			},
			wantCount: 0,
		},
		{
			name: "one failure",
			checks: []CheckResult{
				{Criterion: "c1", Passed: true},
				{Criterion: "c2", Passed: false},
			},
			wantCount: 1,
		},
		{
			name: "all failures",
			checks: []CheckResult{
				{Criterion: "c1", Passed: false},
				{Criterion: "c2", Passed: false},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &VerificationResult{Checks: tt.checks}
			failed := r.FailedChecks()
			assert.Len(t, failed, tt.wantCount)

			// Verify all returned checks are actually failed
			for _, check := range failed {
				assert.False(t, check.Passed)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	require.NotNil(t, config)
	assert.Equal(t, VerificationLevelStandard, config.Level)
	assert.Equal(t, 1, config.MaxRetries)
	assert.Equal(t, 5*time.Minute, config.Timeout)
	assert.False(t, config.EnableAI)
	assert.True(t, config.CaptureOutput)
	assert.Empty(t, config.WorkDir)
}

func TestConfig_IsEnabled(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "nil config",
			config: nil,
			want:   false,
		},
		{
			name:   "default config",
			config: DefaultConfig(),
			want:   true,
		},
		{
			name:   "empty config",
			config: &Config{},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.config.IsEnabled())
		})
	}
}

func TestEvidenceType_Constants(t *testing.T) {
	// Verify evidence type constants are correctly defined
	assert.Equal(t, EvidenceType("command_output"), EvidenceTypeCommandOutput)
	assert.Equal(t, EvidenceType("file_content"), EvidenceTypeFileContent)
	assert.Equal(t, EvidenceType("file_exists"), EvidenceTypeFileExists)
	assert.Equal(t, EvidenceType("ai_analysis"), EvidenceTypeAIAnalysis)
}

func TestVerificationLevel_Constants(t *testing.T) {
	// Verify verification level constants are correctly defined
	assert.Equal(t, VerificationLevel("basic"), VerificationLevelBasic)
	assert.Equal(t, VerificationLevel("standard"), VerificationLevelStandard)
	assert.Equal(t, VerificationLevel("strict"), VerificationLevelStrict)
}

func TestVerificationTask_Fields(t *testing.T) {
	task := &VerificationTask{
		TaskID:          "T001",
		Title:           "Test Task",
		Description:     "Description",
		SuccessCriteria: []string{"criterion1", "criterion2"},
		Files:           []string{"file1.go", "file2.go"},
		WorkDir:         "/tmp/workdir",
	}

	assert.Equal(t, "T001", task.TaskID)
	assert.Equal(t, "Test Task", task.Title)
	assert.Equal(t, "Description", task.Description)
	assert.Len(t, task.SuccessCriteria, 2)
	assert.Len(t, task.Files, 2)
	assert.Equal(t, "/tmp/workdir", task.WorkDir)
}

func TestCheckResult_Fields(t *testing.T) {
	evidence := &Evidence{
		Type:       EvidenceTypeFileExists,
		Content:    "file exists",
		Expected:   "test.go",
		Timestamp:  time.Now(),
		CommandRun: "",
		ExitCode:   0,
	}

	check := &CheckResult{
		Criterion:   "file exists",
		CheckerType: "file_exists",
		Passed:      true,
		Evidence:    evidence,
		Duration:    100 * time.Millisecond,
		Error:       "",
	}

	assert.Equal(t, "file exists", check.Criterion)
	assert.Equal(t, "file_exists", check.CheckerType)
	assert.True(t, check.Passed)
	assert.NotNil(t, check.Evidence)
	assert.Equal(t, 100*time.Millisecond, check.Duration)
	assert.Empty(t, check.Error)
}

func TestEvidence_Fields(t *testing.T) {
	now := time.Now()
	evidence := &Evidence{
		Type:       EvidenceTypeCommandOutput,
		Content:    "build successful",
		Expected:   "",
		Timestamp:  now,
		CommandRun: "go build ./...",
		ExitCode:   0,
	}

	assert.Equal(t, EvidenceTypeCommandOutput, evidence.Type)
	assert.Equal(t, "build successful", evidence.Content)
	assert.Empty(t, evidence.Expected)
	assert.Equal(t, now, evidence.Timestamp)
	assert.Equal(t, "go build ./...", evidence.CommandRun)
	assert.Equal(t, 0, evidence.ExitCode)
}
