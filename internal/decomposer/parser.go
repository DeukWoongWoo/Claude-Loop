package decomposer

import (
	"regexp"
	"strings"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
)

// Task parsing patterns.
var (
	// Pattern: ### Task T001: [Brief Title]
	taskHeaderPattern = regexp.MustCompile(`(?i)###\s*Task\s+(T\d{3}):\s*(.+)`)

	// Field patterns
	descriptionPattern     = regexp.MustCompile(`(?i)^\s*[-*]\s*\*?\*?Description\*?\*?\s*:\s*(.+)$`)
	dependenciesPattern    = regexp.MustCompile(`(?i)^\s*[-*]\s*\*?\*?Dependencies\*?\*?\s*:\s*(.+)$`)
	filesPattern           = regexp.MustCompile(`(?i)^\s*[-*]\s*\*?\*?Files?\*?\*?\s*:\s*(.+)$`)
	complexityPattern      = regexp.MustCompile(`(?i)^\s*[-*]\s*\*?\*?Complexity\*?\*?\s*:\s*(small|medium|large)`)
	successCriteriaPattern = regexp.MustCompile(`(?i)^\s*[-*]\s*\*?\*?Success\s*Criteria\*?\*?\s*:\s*(.+)$`)

	// Task ID reference pattern (e.g., [T001], T002)
	taskIDRefPattern = regexp.MustCompile(`\[?(T\d{3})\]?`)

	// None pattern for dependencies
	nonePattern = regexp.MustCompile(`(?i)^\s*(none|n/?a|-)\s*$`)
)

// Parser extracts structured Task data from Claude's raw output.
type Parser struct{}

// NewParser creates a new Parser.
func NewParser() *Parser {
	return &Parser{}
}

// Parse extracts Tasks from raw output.
func (p *Parser) Parse(rawOutput string) ([]Task, error) {
	// Split by task headers
	taskSections := p.splitByTaskHeaders(rawOutput)

	if len(taskSections) == 0 {
		return nil, ErrParseNoTasks
	}

	var tasks []Task
	for _, section := range taskSections {
		task, err := p.parseTaskSection(section)
		if err != nil {
			continue // Skip malformed tasks
		}
		tasks = append(tasks, *task)
	}

	if len(tasks) == 0 {
		return nil, ErrParseNoTasks
	}

	return tasks, nil
}

// splitByTaskHeaders splits the output into sections, each starting with a task header.
func (p *Parser) splitByTaskHeaders(text string) []string {
	// Find all task header positions
	matches := taskHeaderPattern.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return nil
	}

	var sections []string
	for i, match := range matches {
		start := match[0]
		var end int
		if i+1 < len(matches) {
			end = matches[i+1][0]
		} else {
			end = len(text)
		}
		sections = append(sections, text[start:end])
	}

	return sections
}

// parseTaskSection parses a single task section.
func (p *Parser) parseTaskSection(section string) (*Task, error) {
	lines := strings.Split(section, "\n")

	// Parse header
	headerMatch := taskHeaderPattern.FindStringSubmatch(lines[0])
	if len(headerMatch) < 3 {
		return nil, &DecomposerError{Phase: "parse", Message: "invalid task header"}
	}

	task := &Task{
		Task: planner.Task{
			ID:           headerMatch[1],
			Title:        strings.TrimSpace(headerMatch[2]),
			Status:       planner.TaskStatusPending,
			Dependencies: []string{},
			Files:        []string{},
		},
		SuccessCriteria: []string{},
	}

	// Parse remaining lines
	for _, line := range lines[1:] {
		p.parseTaskLine(line, task)
	}

	return task, nil
}

// parseTaskLine parses a single line and updates the task.
func (p *Parser) parseTaskLine(line string, task *Task) {
	if match := descriptionPattern.FindStringSubmatch(line); len(match) > 1 {
		task.Description = strings.TrimSpace(match[1])
		return
	}

	if match := dependenciesPattern.FindStringSubmatch(line); len(match) > 1 {
		depStr := strings.TrimSpace(match[1])
		if !nonePattern.MatchString(depStr) {
			task.Dependencies = p.extractTaskIDs(depStr)
		}
		return
	}

	if match := filesPattern.FindStringSubmatch(line); len(match) > 1 {
		task.Files = p.extractFileList(match[1])
		return
	}

	if match := complexityPattern.FindStringSubmatch(line); len(match) > 1 {
		task.Complexity = strings.ToLower(strings.TrimSpace(match[1]))
		return
	}

	if match := successCriteriaPattern.FindStringSubmatch(line); len(match) > 1 {
		task.SuccessCriteria = append(task.SuccessCriteria, strings.TrimSpace(match[1]))
		return
	}
}

// extractTaskIDs extracts task IDs from a dependency string.
func (p *Parser) extractTaskIDs(s string) []string {
	matches := taskIDRefPattern.FindAllStringSubmatch(s, -1)
	var ids []string
	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			ids = append(ids, match[1])
			seen[match[1]] = true
		}
	}
	return ids
}

// extractFileList extracts file paths from a comma/space-separated string.
func (p *Parser) extractFileList(s string) []string {
	var files []string
	for _, f := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ' '
	}) {
		f = strings.TrimSpace(f)
		if f != "" && f != "-" {
			files = append(files, f)
		}
	}
	return files
}
