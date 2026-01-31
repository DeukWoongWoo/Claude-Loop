package decomposer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"gopkg.in/yaml.v3"
)

// TaskPersistence handles task file operations.
type TaskPersistence interface {
	SaveTask(task *Task, taskDir string) error
	LoadTask(taskID, taskDir string) (*Task, error)
	SaveTaskGraph(graph *TaskGraph, path string) error
	LoadTaskGraph(path string) (*TaskGraph, error)
	ListTasks(taskDir string) ([]string, error)
	TaskPath(taskID, taskDir string) string
}

// FileTaskPersistence implements TaskPersistence using the filesystem.
type FileTaskPersistence struct{}

// NewFileTaskPersistence creates a new FileTaskPersistence.
func NewFileTaskPersistence() *FileTaskPersistence {
	return &FileTaskPersistence{}
}

// SaveTask writes a task to a markdown file with YAML frontmatter.
// Format: .claude/tasks/T001.md
func (p *FileTaskPersistence) SaveTask(task *Task, taskDir string) error {
	if task == nil {
		return &DecomposerError{Phase: "persistence", Message: "cannot save nil task"}
	}

	// Ensure directory exists
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to create task directory",
			Err:     err,
		}
	}

	path := p.TaskPath(task.ID, taskDir)
	content := p.formatTaskMarkdown(task)

	// Write atomically
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, []byte(content), 0600); err != nil {
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to write task file",
			Err:     err,
		}
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to rename task file",
			Err:     err,
		}
	}

	return nil
}

// formatTaskMarkdown creates markdown content with YAML frontmatter.
func (p *FileTaskPersistence) formatTaskMarkdown(task *Task) string {
	var sb strings.Builder

	// YAML frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", task.ID))
	sb.WriteString(fmt.Sprintf("title: %q\n", task.Title))
	sb.WriteString(fmt.Sprintf("status: %s\n", task.Status))

	if len(task.Dependencies) > 0 {
		sb.WriteString("dependencies:\n")
		for _, dep := range task.Dependencies {
			sb.WriteString(fmt.Sprintf("  - %s\n", dep))
		}
	}

	if len(task.Files) > 0 {
		sb.WriteString("files:\n")
		for _, file := range task.Files {
			sb.WriteString(fmt.Sprintf("  - %s\n", file))
		}
	}

	if task.Complexity != "" {
		sb.WriteString(fmt.Sprintf("complexity: %s\n", task.Complexity))
	}

	if task.StartedAt != nil {
		sb.WriteString(fmt.Sprintf("started_at: %s\n", task.StartedAt.Format(time.RFC3339)))
	}
	if task.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("completed_at: %s\n", task.CompletedAt.Format(time.RFC3339)))
	}

	sb.WriteString("---\n\n")

	// Markdown body
	sb.WriteString(fmt.Sprintf("# %s: %s\n\n", task.ID, task.Title))
	sb.WriteString("## Description\n\n")
	sb.WriteString(task.Description)
	sb.WriteString("\n\n")

	if len(task.SuccessCriteria) > 0 {
		sb.WriteString("## Success Criteria\n\n")
		for _, criteria := range task.SuccessCriteria {
			sb.WriteString(fmt.Sprintf("- [ ] %s\n", criteria))
		}
		sb.WriteString("\n")
	}

	if len(task.Files) > 0 {
		sb.WriteString("## Files to Modify\n\n")
		for _, file := range task.Files {
			sb.WriteString(fmt.Sprintf("- `%s`\n", file))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// LoadTask reads a task from a markdown file.
func (p *FileTaskPersistence) LoadTask(taskID, taskDir string) (*Task, error) {
	path := p.TaskPath(taskID, taskDir)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &DecomposerError{Phase: "persistence", Message: "task not found"}
		}
		return nil, &DecomposerError{
			Phase:   "persistence",
			Message: "failed to read task file",
			Err:     err,
		}
	}

	return p.parseTaskMarkdown(string(data))
}

// parseTaskMarkdown parses markdown with YAML frontmatter.
func (p *FileTaskPersistence) parseTaskMarkdown(content string) (*Task, error) {
	// Split frontmatter and body
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return nil, &DecomposerError{Phase: "persistence", Message: "invalid task file format"}
	}

	frontmatter := strings.TrimSpace(parts[1])

	// Parse YAML frontmatter
	var taskData struct {
		ID           string     `yaml:"id"`
		Title        string     `yaml:"title"`
		Status       string     `yaml:"status"`
		Dependencies []string   `yaml:"dependencies"`
		Files        []string   `yaml:"files"`
		Complexity   string     `yaml:"complexity"`
		StartedAt    *time.Time `yaml:"started_at"`
		CompletedAt  *time.Time `yaml:"completed_at"`
	}

	if err := yaml.Unmarshal([]byte(frontmatter), &taskData); err != nil {
		return nil, &DecomposerError{
			Phase:   "persistence",
			Message: "failed to parse task frontmatter",
			Err:     err,
		}
	}

	// Extract description from body
	body := strings.TrimSpace(parts[2])
	description := p.extractDescription(body)

	task := &Task{
		Task: planner.Task{
			ID:           taskData.ID,
			Title:        taskData.Title,
			Description:  description,
			Status:       taskData.Status,
			Dependencies: taskData.Dependencies,
			Files:        taskData.Files,
		},
		Complexity:  taskData.Complexity,
		StartedAt:   taskData.StartedAt,
		CompletedAt: taskData.CompletedAt,
	}

	return task, nil
}

// extractDescription extracts the description from markdown body.
func (p *FileTaskPersistence) extractDescription(body string) string {
	lines := strings.Split(body, "\n")
	var descLines []string
	inDescription := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## Description") {
			inDescription = true
			continue
		}

		if inDescription {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			descLines = append(descLines, line)
		}
	}

	return strings.TrimSpace(strings.Join(descLines, "\n"))
}

// SaveTaskGraph writes a TaskGraph to a YAML file.
func (p *FileTaskPersistence) SaveTaskGraph(graph *TaskGraph, path string) error {
	if graph == nil {
		return &DecomposerError{Phase: "persistence", Message: "cannot save nil task graph"}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to create directory",
			Err:     err,
		}
	}

	data, err := yaml.Marshal(graph)
	if err != nil {
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to marshal task graph",
			Err:     err,
		}
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to write task graph file",
			Err:     err,
		}
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return &DecomposerError{
			Phase:   "persistence",
			Message: "failed to rename task graph file",
			Err:     err,
		}
	}

	return nil
}

// LoadTaskGraph reads a TaskGraph from a YAML file.
func (p *FileTaskPersistence) LoadTaskGraph(path string) (*TaskGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &DecomposerError{Phase: "persistence", Message: "task graph not found"}
		}
		return nil, &DecomposerError{
			Phase:   "persistence",
			Message: "failed to read task graph file",
			Err:     err,
		}
	}

	var graph TaskGraph
	if err := yaml.Unmarshal(data, &graph); err != nil {
		return nil, &DecomposerError{
			Phase:   "persistence",
			Message: "failed to unmarshal task graph",
			Err:     err,
		}
	}

	return &graph, nil
}

// ListTasks returns all task IDs in the task directory.
func (p *FileTaskPersistence) ListTasks(taskDir string) ([]string, error) {
	entries, err := os.ReadDir(taskDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, &DecomposerError{
			Phase:   "persistence",
			Message: "failed to read task directory",
			Err:     err,
		}
	}

	var taskIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".tmp") {
			taskID := strings.TrimSuffix(name, ".md")
			taskIDs = append(taskIDs, taskID)
		}
	}

	return taskIDs, nil
}

// TaskPath returns the path for a task ID.
func (p *FileTaskPersistence) TaskPath(taskID, taskDir string) string {
	// Validate taskID to prevent path traversal
	if strings.ContainsAny(taskID, "/\\") || taskID == "." || taskID == ".." {
		return filepath.Join(taskDir, "invalid-task-id.md")
	}
	return filepath.Join(taskDir, taskID+".md")
}

// Compile-time interface check.
var _ TaskPersistence = (*FileTaskPersistence)(nil)
