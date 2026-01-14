package council

import (
	"regexp"
	"strings"
)

// ConflictPatterns defines the regex patterns for detecting unresolved conflicts.
// Matches bash implementation (claude_loop.sh lines 1203-1206).
var ConflictPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)PRINCIPLE_CONFLICT_UNRESOLVED`),
	regexp.MustCompile(`(?i)cannot\s+resolve.*principle`),
	regexp.MustCompile(`(?i)conflicting\s+principles.*unresolved`),
}

// DecisionPatterns for extracting decision info from output.
var (
	DecisionPattern  = regexp.MustCompile(`\*\*Decision\*\*:\s*([^*\n]+)`)
	RationalePattern = regexp.MustCompile(`\*\*Rationale\*\*:\s*([^*\n]+)`)
)

// ConflictDetector checks output for principle conflicts.
type ConflictDetector struct{}

// NewConflictDetector creates a new ConflictDetector.
func NewConflictDetector() *ConflictDetector {
	return &ConflictDetector{}
}

// Detect checks if the output contains any conflict patterns.
func (d *ConflictDetector) Detect(output string) bool {
	for _, pattern := range ConflictPatterns {
		if pattern.MatchString(output) {
			return true
		}
	}
	return false
}

// ExtractDecision extracts decision and rationale from output.
func (d *ConflictDetector) ExtractDecision(output string) (decision, rationale string) {
	if matches := DecisionPattern.FindStringSubmatch(output); len(matches) > 1 {
		decision = strings.TrimSpace(matches[1])
	}
	if matches := RationalePattern.FindStringSubmatch(output); len(matches) > 1 {
		rationale = strings.TrimSpace(matches[1])
	}
	return
}
