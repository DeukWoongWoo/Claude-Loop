package prd

import (
	"testing"

	"github.com/DeukWoongWoo/claude-loop/internal/planner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	t.Parallel()

	validator := NewValidator()
	assert.NotNil(t, validator)
}

func TestValidator_Validate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		prd      *PRD
		wantErr  bool
		errField string
	}{
		{
			name: "valid PRD",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{"Goal 1"},
					Requirements:    []string{"Req 1"},
					SuccessCriteria: []string{"Criterion 1"},
				},
			},
			wantErr: false,
		},
		{
			name: "valid PRD with multiple items",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{"Goal 1", "Goal 2", "Goal 3"},
					Requirements:    []string{"Req 1", "Req 2"},
					SuccessCriteria: []string{"Criterion 1"},
					Constraints:     []string{"Constraint 1"},
				},
			},
			wantErr: false,
		},
		{
			name:     "nil PRD",
			prd:      nil,
			wantErr:  true,
			errField: "prd",
		},
		{
			name: "no goals",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{},
					Requirements:    []string{"Req 1"},
					SuccessCriteria: []string{"Criterion 1"},
				},
			},
			wantErr:  true,
			errField: "goals",
		},
		{
			name: "nil goals",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           nil,
					Requirements:    []string{"Req 1"},
					SuccessCriteria: []string{"Criterion 1"},
				},
			},
			wantErr:  true,
			errField: "goals",
		},
		{
			name: "no requirements",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{"Goal 1"},
					Requirements:    []string{},
					SuccessCriteria: []string{"Criterion 1"},
				},
			},
			wantErr:  true,
			errField: "requirements",
		},
		{
			name: "no success criteria",
			prd: &PRD{
				PRD: planner.PRD{
					Goals:           []string{"Goal 1"},
					Requirements:    []string{"Req 1"},
					SuccessCriteria: []string{},
				},
			},
			wantErr:  true,
			errField: "success_criteria",
		},
	}

	validator := NewValidator()

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validator.Validate(tt.prd)

			if tt.wantErr {
				require.Error(t, err)
				var ve *ValidationError
				if assert.ErrorAs(t, err, &ve) {
					assert.Equal(t, tt.errField, ve.Field)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	t.Parallel()

	t.Run("nil PRD returns single error", func(t *testing.T) {
		t.Parallel()

		validator := NewValidator()
		errs := validator.ValidateAll(nil)

		assert.Len(t, errs, 1)
		var ve *ValidationError
		require.ErrorAs(t, errs[0], &ve)
		assert.Equal(t, "prd", ve.Field)
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		t.Parallel()

		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{},
				Requirements:    []string{},
				SuccessCriteria: []string{},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(prd)

		assert.Len(t, errs, 3)
	})

	t.Run("empty strings in arrays", func(t *testing.T) {
		t.Parallel()

		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{"Goal 1", ""},
				Requirements:    []string{"", "Req 1"},
				SuccessCriteria: []string{"Criterion 1", ""},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(prd)

		// Should have 3 empty string errors
		assert.Len(t, errs, 3)

		// Verify all are validation errors for empty strings
		for _, err := range errs {
			var ve *ValidationError
			require.ErrorAs(t, err, &ve)
			assert.Contains(t, ve.Message, "empty string at index")
		}
	})

	t.Run("valid PRD returns no errors", func(t *testing.T) {
		t.Parallel()

		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{"Goal 1"},
				Requirements:    []string{"Req 1"},
				SuccessCriteria: []string{"Criterion 1"},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(prd)

		assert.Empty(t, errs)
	})
}

func TestValidator_validateGoals(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty goals", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{Goals: []string{}}}
		err := validator.validateGoals(prd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one goal is required")
	})

	t.Run("with goals", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{Goals: []string{"Goal"}}}
		err := validator.validateGoals(prd)
		assert.NoError(t, err)
	})
}

func TestValidator_validateRequirements(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty requirements", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{Requirements: []string{}}}
		err := validator.validateRequirements(prd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one requirement is required")
	})

	t.Run("with requirements", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{Requirements: []string{"Req"}}}
		err := validator.validateRequirements(prd)
		assert.NoError(t, err)
	})
}

func TestValidator_validateSuccessCriteria(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty criteria", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{SuccessCriteria: []string{}}}
		err := validator.validateSuccessCriteria(prd)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "at least one success criterion is required")
	})

	t.Run("with criteria", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{PRD: planner.PRD{SuccessCriteria: []string{"Criterion"}}}
		err := validator.validateSuccessCriteria(prd)
		assert.NoError(t, err)
	})
}

func TestValidator_validateNoEmptyStrings(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("no empty strings", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{"Goal"},
				Requirements:    []string{"Req"},
				SuccessCriteria: []string{"Criterion"},
			},
		}
		errs := validator.validateNoEmptyStrings(prd)
		assert.Empty(t, errs)
	})

	t.Run("empty string in goals", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{"Goal", ""},
				Requirements:    []string{"Req"},
				SuccessCriteria: []string{"Criterion"},
			},
		}
		errs := validator.validateNoEmptyStrings(prd)
		assert.Len(t, errs, 1)
	})

	t.Run("multiple empty strings at different indices", func(t *testing.T) {
		t.Parallel()
		prd := &PRD{
			PRD: planner.PRD{
				Goals:           []string{"", "Goal", ""},
				Requirements:    []string{"Req"},
				SuccessCriteria: []string{"Criterion"},
			},
		}
		errs := validator.validateNoEmptyStrings(prd)
		assert.Len(t, errs, 2)
	})
}
