package prd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 3, config.MaxRetries)
	assert.True(t, config.ValidateOutput)
}

func TestConfig_IsEnabled(t *testing.T) {
	t.Parallel()

	t.Run("nil config", func(t *testing.T) {
		t.Parallel()
		var c *Config
		assert.False(t, c.IsEnabled())
	})

	t.Run("non-nil config", func(t *testing.T) {
		t.Parallel()
		c := &Config{}
		assert.True(t, c.IsEnabled())
	})

	t.Run("default config", func(t *testing.T) {
		t.Parallel()
		c := DefaultConfig()
		assert.True(t, c.IsEnabled())
	})
}

func TestPRD_EmbeddedFields(t *testing.T) {
	t.Parallel()

	prd := &PRD{}

	// Verify embedded planner.PRD fields are accessible
	prd.Goals = []string{"Goal 1", "Goal 2"}
	prd.Requirements = []string{"Req 1"}
	prd.Constraints = []string{"Constraint 1"}
	prd.SuccessCriteria = []string{"Criterion 1"}
	prd.RawOutput = "raw output"

	// Extended fields
	prd.ID = "prd-001"
	prd.Title = "Test PRD"
	prd.Summary = "Summary text"
	prd.DetailedGoals = []string{"Detailed 1"}
	prd.OutOfScope = []string{"Out 1", "Out 2"}

	assert.Equal(t, []string{"Goal 1", "Goal 2"}, prd.Goals)
	assert.Equal(t, "Req 1", prd.Requirements[0])
	assert.Equal(t, "raw output", prd.RawOutput)
	assert.Equal(t, "prd-001", prd.ID)
	assert.Equal(t, "Test PRD", prd.Title)
	assert.Len(t, prd.OutOfScope, 2)
}

func TestResult(t *testing.T) {
	t.Parallel()

	prd := &PRD{ID: "test"}
	result := &Result{
		PRD:      prd,
		Cost:     0.05,
		Duration: 1000000000, // 1 second in nanoseconds
	}

	assert.Equal(t, prd, result.PRD)
	assert.Equal(t, 0.05, result.Cost)
	assert.Equal(t, int64(1000000000), int64(result.Duration))
}
