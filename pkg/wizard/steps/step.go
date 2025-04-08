package steps

import (
	"context"

	"github.com/pkg/errors"
)

// Step represents a single step in the wizard.
type Step interface {
	ID() string
	Type() string
	Title() string
	Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error)
	GetBaseStep() *BaseStep
}

// BaseStep contains common fields for all step types.
type BaseStep struct {
	StepID        string `yaml:"id"`
	StepType      string `yaml:"type"`
	StepTitle     string `yaml:"title,omitempty"`
	Description   string `yaml:"description,omitempty"`
	SkipCondition string `yaml:"skip_condition,omitempty"`
	NextStep      string `yaml:"next_step,omitempty"`
}

func (bs *BaseStep) ID() string {
	return bs.StepID
}

func (bs *BaseStep) Type() string {
	return bs.StepType
}

func (bs *BaseStep) Title() string {
	return bs.StepTitle
}

// Placeholder Execute for BaseStep - concrete types should override this.
func (bs *BaseStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	return nil, errors.Errorf("Execute not implemented for step type %s (ID: %s)", bs.Type(), bs.ID())
}
