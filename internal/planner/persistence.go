package planner

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Persistence handles plan file operations.
type Persistence interface {
	Save(plan *Plan, path string) error
	Load(path string) (*Plan, error)
	Exists(path string) bool
	Delete(path string) error
	DefaultPlanPath(planID string) string
}

// FilePersistence implements Persistence using the filesystem.
type FilePersistence struct {
	planDir string
}

// NewFilePersistence creates a new FilePersistence.
func NewFilePersistence(planDir string) *FilePersistence {
	return &FilePersistence{planDir: planDir}
}

// Save writes a plan to a YAML file atomically.
// Uses write-to-temp-then-rename pattern for atomic writes.
func (p *FilePersistence) Save(plan *Plan, path string) error {
	if plan == nil {
		return &PlannerError{
			Phase:   "persistence",
			Message: "cannot save nil plan",
		}
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return &PlannerError{
			Phase:   "persistence",
			Message: "failed to create directory",
			Err:     err,
		}
	}

	// Marshal to YAML
	data, err := yaml.Marshal(plan)
	if err != nil {
		return &PlannerError{
			Phase:   "persistence",
			Message: "failed to marshal plan",
			Err:     err,
		}
	}

	// Write to temp file first (owner read/write only for security)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return &PlannerError{
			Phase:   "persistence",
			Message: "failed to write temp file",
			Err:     err,
		}
	}

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		// Clean up temp file on failure
		_ = os.Remove(tmpPath)
		return &PlannerError{
			Phase:   "persistence",
			Message: "failed to rename temp file",
			Err:     err,
		}
	}

	return nil
}

// Load reads a plan from a YAML file.
func (p *FilePersistence) Load(path string) (*Plan, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrPlanNotFound
		}
		return nil, &PlannerError{
			Phase:   "persistence",
			Message: "failed to read plan file",
			Err:     err,
		}
	}

	var plan Plan
	if err := yaml.Unmarshal(data, &plan); err != nil {
		return nil, &PlannerError{
			Phase:   "persistence",
			Message: "failed to unmarshal plan",
			Err:     err,
		}
	}

	return &plan, nil
}

// Exists checks if a plan file exists.
func (p *FilePersistence) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// Delete removes a plan file.
func (p *FilePersistence) Delete(path string) error {
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return ErrPlanNotFound
		}
		return &PlannerError{
			Phase:   "persistence",
			Message: "failed to delete plan file",
			Err:     err,
		}
	}
	return nil
}

// DefaultPlanPath returns the default path for a plan ID.
// It validates the planID to prevent path traversal attacks.
func (p *FilePersistence) DefaultPlanPath(planID string) string {
	// Reject empty IDs
	if planID == "" {
		return filepath.Join(p.planDir, "invalid-plan-id.yaml")
	}

	// Check for path separators (both Unix and Windows style)
	if strings.ContainsAny(planID, "/\\") {
		return filepath.Join(p.planDir, "invalid-plan-id.yaml")
	}

	// Sanitize using filepath.Base as additional protection
	cleanID := filepath.Base(planID)
	if cleanID == "." || cleanID == ".." || cleanID == "" || cleanID != planID {
		return filepath.Join(p.planDir, "invalid-plan-id.yaml")
	}

	return filepath.Join(p.planDir, cleanID+".yaml")
}

// ListPlans returns all plan IDs in the plan directory.
func (p *FilePersistence) ListPlans() ([]string, error) {
	entries, err := os.ReadDir(p.planDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, &PlannerError{
			Phase:   "persistence",
			Message: "failed to read plan directory",
			Err:     err,
		}
	}

	var planIDs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".tmp") {
			planID := strings.TrimSuffix(name, ".yaml")
			planIDs = append(planIDs, planID)
		}
	}

	return planIDs, nil
}

// PlanDir returns the plan directory.
func (p *FilePersistence) PlanDir() string {
	return p.planDir
}
