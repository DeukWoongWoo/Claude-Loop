package planner

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlannerError_Error(t *testing.T) {
	t.Parallel()

	t.Run("with wrapped error", func(t *testing.T) {
		t.Parallel()
		wrappedErr := errors.New("underlying error")
		err := &PlannerError{
			Phase:   "persistence",
			Message: "failed to save",
			Err:     wrappedErr,
		}

		assert.Equal(t, "planner persistence: failed to save: underlying error", err.Error())
	})

	t.Run("without wrapped error", func(t *testing.T) {
		t.Parallel()
		err := &PlannerError{
			Phase:   "config",
			Message: "invalid config",
		}

		assert.Equal(t, "planner config: invalid config", err.Error())
	})
}

func TestPlannerError_Unwrap(t *testing.T) {
	t.Parallel()

	t.Run("with wrapped error", func(t *testing.T) {
		t.Parallel()
		wrappedErr := errors.New("underlying error")
		err := &PlannerError{
			Phase:   "run",
			Message: "execution failed",
			Err:     wrappedErr,
		}

		assert.Equal(t, wrappedErr, err.Unwrap())
		assert.True(t, errors.Is(err, wrappedErr))
	})

	t.Run("without wrapped error", func(t *testing.T) {
		t.Parallel()
		err := &PlannerError{
			Phase:   "phase",
			Message: "phase failed",
		}

		assert.Nil(t, err.Unwrap())
	})
}

func TestIsPlannerError(t *testing.T) {
	t.Parallel()

	t.Run("with PlannerError", func(t *testing.T) {
		t.Parallel()
		err := &PlannerError{Phase: "test", Message: "test error"}

		assert.True(t, IsPlannerError(err))
	})

	t.Run("with wrapped PlannerError", func(t *testing.T) {
		t.Parallel()
		plannerErr := &PlannerError{Phase: "test", Message: "test error"}
		wrappedErr := errors.Join(errors.New("wrapper"), plannerErr)

		assert.True(t, IsPlannerError(wrappedErr))
	})

	t.Run("with non-PlannerError", func(t *testing.T) {
		t.Parallel()
		err := errors.New("regular error")

		assert.False(t, IsPlannerError(err))
	})

	t.Run("with nil error", func(t *testing.T) {
		t.Parallel()

		assert.False(t, IsPlannerError(nil))
	})
}

func TestPredefinedErrors(t *testing.T) {
	t.Parallel()

	t.Run("ErrPlanNotFound", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrPlanNotFound)
		assert.Equal(t, "persistence", ErrPlanNotFound.Phase)
		assert.Equal(t, "plan not found", ErrPlanNotFound.Message)
		assert.True(t, IsPlannerError(ErrPlanNotFound))
	})

	t.Run("ErrInvalidPlanState", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrInvalidPlanState)
		assert.Equal(t, "run", ErrInvalidPlanState.Phase)
		assert.Equal(t, "invalid plan state", ErrInvalidPlanState.Message)
		assert.True(t, IsPlannerError(ErrInvalidPlanState))
	})

	t.Run("ErrPhaseNotFound", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrPhaseNotFound)
		assert.Equal(t, "phase", ErrPhaseNotFound.Phase)
		assert.Equal(t, "phase not found", ErrPhaseNotFound.Message)
		assert.True(t, IsPlannerError(ErrPhaseNotFound))
	})

	t.Run("ErrPhaseFailed", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrPhaseFailed)
		assert.Equal(t, "phase", ErrPhaseFailed.Phase)
		assert.Equal(t, "phase execution failed", ErrPhaseFailed.Message)
		assert.True(t, IsPlannerError(ErrPhaseFailed))
	})

	t.Run("ErrInvalidPlanID", func(t *testing.T) {
		t.Parallel()
		require.NotNil(t, ErrInvalidPlanID)
		assert.Equal(t, "persistence", ErrInvalidPlanID.Phase)
		assert.Equal(t, "invalid plan ID", ErrInvalidPlanID.Message)
		assert.True(t, IsPlannerError(ErrInvalidPlanID))
	})
}

func TestPlannerError_ErrorsAs(t *testing.T) {
	t.Parallel()

	err := &PlannerError{
		Phase:   "persistence",
		Message: "load failed",
		Err:     errors.New("file not found"),
	}

	var pe *PlannerError
	require.True(t, errors.As(err, &pe))
	assert.Equal(t, "persistence", pe.Phase)
	assert.Equal(t, "load failed", pe.Message)
}
