package verifier

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// CommandExecutor abstracts exec.Command for testing.
type CommandExecutor interface {
	CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd
}

// DefaultExecutor uses the real exec.CommandContext.
type DefaultExecutor struct{}

// CommandContext creates a new exec.Cmd with the given context.
func (e *DefaultExecutor) CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	return exec.CommandContext(ctx, name, args...)
}

// Checker performs a specific type of verification check.
type Checker interface {
	// Type returns the checker type identifier.
	Type() string

	// CanHandle returns true if this checker can handle the given criterion.
	CanHandle(criterion string) bool

	// Check performs the verification and returns the result.
	Check(ctx context.Context, criterion string, workDir string) *CheckResult
}

// FileExistsChecker verifies file existence.
type FileExistsChecker struct{}

// NewFileExistsChecker creates a new FileExistsChecker.
func NewFileExistsChecker() *FileExistsChecker {
	return &FileExistsChecker{}
}

// Type returns the checker type.
func (c *FileExistsChecker) Type() string {
	return "file_exists"
}

// fileExistsPattern matches patterns like "File `path` exists" or "file exists: path".
var fileExistsPattern = regexp.MustCompile(`(?i)(?:file\s+)?[` + "`" + `'"]?([\w/.\\-]+)[` + "`" + `'"]?\s+exists|exists:\s*[` + "`" + `'"]?([\w/.\\-]+)[` + "`" + `'"]?`)

// CanHandle returns true if this checker can handle the criterion.
func (c *FileExistsChecker) CanHandle(criterion string) bool {
	lower := strings.ToLower(criterion)
	return strings.Contains(lower, "exists") &&
		(strings.Contains(lower, "file") || fileExistsPattern.MatchString(criterion))
}

// Check verifies file existence.
func (c *FileExistsChecker) Check(ctx context.Context, criterion string, workDir string) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Criterion:   criterion,
		CheckerType: c.Type(),
	}

	// Extract file path from criterion
	filePath := c.extractFilePath(criterion)
	if filePath == "" {
		result.Passed = false
		result.Error = "could not extract file path from criterion"
		result.Duration = time.Since(start)
		return result
	}

	// Security: Prevent path traversal and absolute paths outside workDir
	if strings.Contains(filePath, "..") {
		result.Passed = false
		result.Error = "path traversal not allowed"
		result.Duration = time.Since(start)
		result.Evidence = &Evidence{
			Type:      EvidenceTypeFileExists,
			Content:   fmt.Sprintf("rejected path: %s", filePath),
			Expected:  filePath,
			Timestamp: time.Now(),
		}
		return result
	}

	// Security: Reject absolute paths when workDir is specified (to prevent arbitrary file access)
	if filepath.IsAbs(filePath) && workDir != "" {
		result.Passed = false
		result.Error = "absolute paths not allowed when workDir is set"
		result.Duration = time.Since(start)
		result.Evidence = &Evidence{
			Type:      EvidenceTypeFileExists,
			Content:   fmt.Sprintf("rejected absolute path: %s", filePath),
			Expected:  filePath,
			Timestamp: time.Now(),
		}
		return result
	}

	// Resolve relative paths
	if !filepath.IsAbs(filePath) && workDir != "" {
		filePath = filepath.Join(workDir, filePath)
	}

	// Check file existence
	info, err := os.Stat(filePath)
	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("file does not exist: %s", filePath)
		result.Evidence = &Evidence{
			Type:      EvidenceTypeFileExists,
			Content:   fmt.Sprintf("os.Stat error: %v", err),
			Expected:  filePath,
			Timestamp: time.Now(),
		}
		return result
	}

	result.Passed = true
	result.Evidence = &Evidence{
		Type:      EvidenceTypeFileExists,
		Content:   fmt.Sprintf("file exists: %s (size: %d bytes)", filePath, info.Size()),
		Expected:  filePath,
		Timestamp: time.Now(),
	}
	return result
}

func (c *FileExistsChecker) extractFilePath(criterion string) string {
	matches := fileExistsPattern.FindStringSubmatch(criterion)
	if len(matches) > 1 {
		if matches[1] != "" {
			return matches[1]
		}
		if len(matches) > 2 && matches[2] != "" {
			return matches[2]
		}
	}

	// Fallback: look for backtick-quoted paths
	backtickPattern := regexp.MustCompile("[`'\"]([\\w/.\\-]+)[`'\"]")
	matches = backtickPattern.FindStringSubmatch(criterion)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// runCommand executes a command and returns a CheckResult.
// This is a shared helper for BuildChecker and TestChecker.
func runCommand(ctx context.Context, executor CommandExecutor, checkerType, criterion, cmdName string, cmdArgs []string, workDir, operationName string) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Criterion:   criterion,
		CheckerType: checkerType,
	}

	cmd := executor.CommandContext(ctx, cmdName, cmdArgs...)
	if workDir != "" {
		cmd.Dir = workDir
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result.Duration = time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = -1
		}
	}

	result.Evidence = &Evidence{
		Type:       EvidenceTypeCommandOutput,
		Content:    stdout.String() + stderr.String(),
		CommandRun: cmdName + " " + strings.Join(cmdArgs, " "),
		ExitCode:   exitCode,
		Timestamp:  time.Now(),
	}

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("%s failed with exit code %d", operationName, exitCode)
		return result
	}

	result.Passed = true
	return result
}

// BuildChecker verifies build success.
type BuildChecker struct {
	executor CommandExecutor
}

// NewBuildChecker creates a new BuildChecker.
func NewBuildChecker(executor CommandExecutor) *BuildChecker {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &BuildChecker{executor: executor}
}

// Type returns the checker type.
func (c *BuildChecker) Type() string {
	return "build"
}

// buildPattern matches patterns like "build passes", "go build", etc.
var buildPattern = regexp.MustCompile(`(?i)(build\s+passes|go\s+build|make\s+build|npm\s+run\s+build)`)

// CanHandle returns true if this checker can handle the criterion.
func (c *BuildChecker) CanHandle(criterion string) bool {
	return buildPattern.MatchString(criterion)
}

// Check verifies build success.
func (c *BuildChecker) Check(ctx context.Context, criterion string, workDir string) *CheckResult {
	cmdName, cmdArgs := c.extractBuildCommand(criterion)
	return runCommand(ctx, c.executor, c.Type(), criterion, cmdName, cmdArgs, workDir, "build")
}

func (c *BuildChecker) extractBuildCommand(criterion string) (string, []string) {
	lower := strings.ToLower(criterion)

	if strings.Contains(lower, "go build") {
		return "go", []string{"build", "./..."}
	}
	if strings.Contains(lower, "make build") {
		return "make", []string{"build"}
	}
	if strings.Contains(lower, "npm run build") {
		return "npm", []string{"run", "build"}
	}

	// Default to go build for Go projects
	return "go", []string{"build", "./..."}
}

// TestChecker verifies test success.
type TestChecker struct {
	executor CommandExecutor
}

// NewTestChecker creates a new TestChecker.
func NewTestChecker(executor CommandExecutor) *TestChecker {
	if executor == nil {
		executor = &DefaultExecutor{}
	}
	return &TestChecker{executor: executor}
}

// Type returns the checker type.
func (c *TestChecker) Type() string {
	return "test"
}

// testPattern matches patterns like "test passes", "make test", etc.
var testPattern = regexp.MustCompile(`(?i)(tests?\s+pass|make\s+test|go\s+test|npm\s+test)`)

// CanHandle returns true if this checker can handle the criterion.
func (c *TestChecker) CanHandle(criterion string) bool {
	return testPattern.MatchString(criterion)
}

// Check verifies test success.
func (c *TestChecker) Check(ctx context.Context, criterion string, workDir string) *CheckResult {
	cmdName, cmdArgs := c.extractTestCommand(criterion)
	return runCommand(ctx, c.executor, c.Type(), criterion, cmdName, cmdArgs, workDir, "tests")
}

func (c *TestChecker) extractTestCommand(criterion string) (string, []string) {
	lower := strings.ToLower(criterion)

	if strings.Contains(lower, "go test") {
		return "go", []string{"test", "./..."}
	}
	if strings.Contains(lower, "make test") {
		return "make", []string{"test"}
	}
	if strings.Contains(lower, "npm test") {
		return "npm", []string{"test"}
	}

	// Default to make test
	return "make", []string{"test"}
}

// ContentMatchChecker verifies content patterns in files.
type ContentMatchChecker struct{}

// NewContentMatchChecker creates a new ContentMatchChecker.
func NewContentMatchChecker() *ContentMatchChecker {
	return &ContentMatchChecker{}
}

// Type returns the checker type.
func (c *ContentMatchChecker) Type() string {
	return "content_match"
}

// contentPattern matches patterns like "contains", "matches", etc.
var contentPattern = regexp.MustCompile(`(?i)(contains|matches|includes)`)

// CanHandle returns true if this checker can handle the criterion.
func (c *ContentMatchChecker) CanHandle(criterion string) bool {
	lower := strings.ToLower(criterion)
	// Must have content-related keywords but not be a simple file exists check
	return contentPattern.MatchString(criterion) && !strings.Contains(lower, "file exists")
}

// Check verifies content pattern match.
func (c *ContentMatchChecker) Check(ctx context.Context, criterion string, workDir string) *CheckResult {
	start := time.Now()
	result := &CheckResult{
		Criterion:   criterion,
		CheckerType: c.Type(),
	}

	// Extract file and pattern from criterion
	filePath, pattern := c.extractFileAndPattern(criterion)
	if filePath == "" || pattern == "" {
		result.Passed = false
		result.Error = "could not extract file path and pattern from criterion"
		result.Duration = time.Since(start)
		return result
	}

	// Security: Prevent path traversal
	if strings.Contains(filePath, "..") {
		result.Passed = false
		result.Error = "path traversal not allowed"
		result.Duration = time.Since(start)
		return result
	}

	// Security: Reject absolute paths when workDir is specified
	if filepath.IsAbs(filePath) && workDir != "" {
		result.Passed = false
		result.Error = "absolute paths not allowed when workDir is set"
		result.Duration = time.Since(start)
		return result
	}

	// Resolve relative paths
	if !filepath.IsAbs(filePath) && workDir != "" {
		filePath = filepath.Join(workDir, filePath)
	}

	content, err := os.ReadFile(filePath)
	result.Duration = time.Since(start)

	if err != nil {
		result.Passed = false
		result.Error = fmt.Sprintf("failed to read file: %v", err)
		return result
	}

	// Truncate content to prevent memory issues and accidental secret exposure
	const maxContentSize = 10 * 1024 // 10KB limit
	contentStr := string(content)
	if len(contentStr) > maxContentSize {
		contentStr = contentStr[:maxContentSize] + "\n... [truncated]"
	}

	matched := strings.Contains(string(content), pattern)
	result.Evidence = &Evidence{
		Type:      EvidenceTypeFileContent,
		Content:   contentStr,
		Expected:  pattern,
		Timestamp: time.Now(),
	}

	if !matched {
		result.Passed = false
		result.Error = fmt.Sprintf("pattern '%s' not found in file", pattern)
		return result
	}

	result.Passed = true
	return result
}

func (c *ContentMatchChecker) extractFileAndPattern(criterion string) (string, string) {
	// Pattern: "file `path` contains `pattern`"
	filePatternRegex := regexp.MustCompile("[`'\"]([\\w/.\\-]+)[`'\"]\\s+contains\\s+[`'\"]([^`'\"]+)[`'\"]")
	matches := filePatternRegex.FindStringSubmatch(criterion)
	if len(matches) >= 3 {
		return matches[1], matches[2]
	}
	return "", ""
}
