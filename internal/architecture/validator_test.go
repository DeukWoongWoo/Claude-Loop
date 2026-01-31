package architecture

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
		arch     *Architecture
		wantErr  bool
		errField string
	}{
		{
			name: "valid architecture",
			arch: &Architecture{
				Architecture: planner.Architecture{
					Components: []planner.Component{
						{Name: "Parser", Description: "Parses output", Files: []string{"parser.go"}},
					},
					FileStructure: []string{"parser.go"},
					TechDecisions: []string{"Use regex"},
				},
			},
			wantErr: false,
		},
		{
			name:     "nil architecture",
			arch:     nil,
			wantErr:  true,
			errField: "architecture",
		},
		{
			name: "no components",
			arch: &Architecture{
				Architecture: planner.Architecture{
					Components:    []planner.Component{},
					FileStructure: []string{"file.go"},
					TechDecisions: []string{"Decision"},
				},
			},
			wantErr:  true,
			errField: "components",
		},
		{
			name: "no file structure",
			arch: &Architecture{
				Architecture: planner.Architecture{
					Components: []planner.Component{
						{Name: "Test", Description: "Test", Files: nil},
					},
					FileStructure: []string{},
					TechDecisions: []string{"Decision"},
				},
			},
			wantErr:  true,
			errField: "file_structure",
		},
		{
			name: "no tech decisions",
			arch: &Architecture{
				Architecture: planner.Architecture{
					Components: []planner.Component{
						{Name: "Test", Description: "Test", Files: nil},
					},
					FileStructure: []string{"file.go"},
					TechDecisions: []string{},
				},
			},
			wantErr:  true,
			errField: "tech_decisions",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			validator := NewValidator()
			err := validator.Validate(tt.arch)

			if tt.wantErr {
				require.Error(t, err)
				var ve *ValidationError
				if assert.ErrorAs(t, err, &ve) {
					assert.Equal(t, tt.errField, ve.Field)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	t.Parallel()

	t.Run("nil architecture returns single error", func(t *testing.T) {
		t.Parallel()

		validator := NewValidator()
		errs := validator.ValidateAll(nil)

		assert.Len(t, errs, 1)
		var ve *ValidationError
		if assert.ErrorAs(t, errs[0], &ve) {
			assert.Equal(t, "architecture", ve.Field)
		}
	})

	t.Run("multiple validation errors", func(t *testing.T) {
		t.Parallel()

		arch := &Architecture{
			Architecture: planner.Architecture{
				Components:    []planner.Component{},
				FileStructure: []string{},
				TechDecisions: []string{},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(arch)

		// Should have at least 3 errors: components, file_structure, tech_decisions
		assert.GreaterOrEqual(t, len(errs), 3)

		fields := make(map[string]bool)
		for _, err := range errs {
			var ve *ValidationError
			if assert.ErrorAs(t, err, &ve) {
				fields[ve.Field] = true
			}
		}

		assert.True(t, fields["components"])
		assert.True(t, fields["file_structure"])
		assert.True(t, fields["tech_decisions"])
	})

	t.Run("valid architecture returns empty slice", func(t *testing.T) {
		t.Parallel()

		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "Test", Description: "Test component", Files: []string{"test.go"}},
				},
				FileStructure: []string{"test.go"},
				TechDecisions: []string{"Simple decision"},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(arch)

		assert.Empty(t, errs)
	})

	t.Run("component with empty name", func(t *testing.T) {
		t.Parallel()

		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "", Description: "Test", Files: nil},
				},
				FileStructure: []string{"file.go"},
				TechDecisions: []string{"Decision"},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(arch)

		// Should have error for empty name
		found := false
		for _, err := range errs {
			var ve *ValidationError
			if assert.ErrorAs(t, err, &ve) {
				if ve.Field == "components" && ve.Message != "" {
					found = true
				}
			}
		}
		assert.True(t, found, "should have validation error for empty component name")
	})

	t.Run("component with empty description", func(t *testing.T) {
		t.Parallel()

		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "Test", Description: "", Files: nil},
				},
				FileStructure: []string{"file.go"},
				TechDecisions: []string{"Decision"},
			},
		}

		validator := NewValidator()
		errs := validator.ValidateAll(arch)

		// Should have error for empty description
		found := false
		for _, err := range errs {
			var ve *ValidationError
			if assert.ErrorAs(t, err, &ve) {
				if ve.Field == "components" && ve.Message != "" {
					found = true
				}
			}
		}
		assert.True(t, found, "should have validation error for empty component description")
	})
}

func TestValidator_validateComponents(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty components", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{Components: []planner.Component{}},
		}
		err := validator.validateComponents(arch)
		require.Error(t, err)
	})

	t.Run("with components", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{{Name: "Test", Description: "Test"}},
			},
		}
		err := validator.validateComponents(arch)
		require.NoError(t, err)
	})
}

func TestValidator_validateFileStructure(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty file structure", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{FileStructure: []string{}},
		}
		err := validator.validateFileStructure(arch)
		require.Error(t, err)
	})

	t.Run("with file structure", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{FileStructure: []string{"file.go"}},
		}
		err := validator.validateFileStructure(arch)
		require.NoError(t, err)
	})
}

func TestValidator_validateTechDecisions(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("empty tech decisions", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{TechDecisions: []string{}},
		}
		err := validator.validateTechDecisions(arch)
		require.Error(t, err)
	})

	t.Run("with tech decisions", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{TechDecisions: []string{"Decision"}},
		}
		err := validator.validateTechDecisions(arch)
		require.NoError(t, err)
	})
}

func TestValidator_validateComponentsComplete(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("all components complete", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "A", Description: "Description A"},
					{Name: "B", Description: "Description B"},
				},
			},
		}
		errs := validator.validateComponentsComplete(arch)
		assert.Empty(t, errs)
	})

	t.Run("component missing name", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "", Description: "Description"},
				},
			},
		}
		errs := validator.validateComponentsComplete(arch)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "empty name")
	})

	t.Run("component missing description", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "Test", Description: ""},
				},
			},
		}
		errs := validator.validateComponentsComplete(arch)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "empty description")
	})

	t.Run("multiple incomplete components", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Components: []planner.Component{
					{Name: "", Description: ""},
					{Name: "Valid", Description: ""},
				},
			},
		}
		errs := validator.validateComponentsComplete(arch)
		assert.Len(t, errs, 3) // 2 for first component (name+desc), 1 for second (desc)
	})
}

func TestValidator_validateNoEmptyStrings(t *testing.T) {
	t.Parallel()

	validator := NewValidator()

	t.Run("all valid strings", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				FileStructure: []string{"file.go"},
				TechDecisions: []string{"decision"},
				Dependencies:  []string{"dep"},
			},
		}
		errs := validator.validateNoEmptyStrings(arch)
		assert.Empty(t, errs)
	})

	t.Run("empty file structure entry", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				FileStructure: []string{"valid.go", ""},
			},
		}
		errs := validator.validateNoEmptyStrings(arch)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "file_structure")
	})

	t.Run("empty tech decision entry", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				TechDecisions: []string{"", "valid"},
			},
		}
		errs := validator.validateNoEmptyStrings(arch)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "tech_decisions")
	})

	t.Run("empty dependency entry", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				Dependencies: []string{"valid", ""},
			},
		}
		errs := validator.validateNoEmptyStrings(arch)
		assert.Len(t, errs, 1)
		assert.Contains(t, errs[0].Error(), "dependencies")
	})

	t.Run("multiple empty strings across fields", func(t *testing.T) {
		t.Parallel()
		arch := &Architecture{
			Architecture: planner.Architecture{
				FileStructure: []string{"", "valid.go"},
				TechDecisions: []string{"valid", ""},
				Dependencies:  []string{""},
			},
		}
		errs := validator.validateNoEmptyStrings(arch)
		assert.Len(t, errs, 3)
	})
}
