package steps

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// DecisionStep represents a step where the user makes a choice.
type DecisionStep struct {
	BaseStep    `yaml:",inline"`
	TargetKey   string            `yaml:"target_key"`
	Choices     []string          `yaml:"choices"` // Simplified for now
	NextStepMap map[string]string `yaml:"next_step_map,omitempty"`
	// TODO(manuel, 2024-08-05) Add support for more complex choice objects (value, label)
}

var _ Step = &DecisionStep{}

func (ds *DecisionStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n--- Step: %s ---\n", ds.Title())
	// TODO(manuel, 2024-08-05) Implement decision logic (present choices, get input)
	fmt.Println("Decision Step (Not Implemented)")
	// Placeholder: return empty results for now
	stepResult := map[string]interface{}{}
	// Simulate setting the target key based on a hypothetical choice
	// stepResult[ds.TargetKey] = "dummy_choice"
	return stepResult, errors.New("decision step not implemented")
}

func (ds *DecisionStep) GetBaseStep() *BaseStep {
	return &ds.BaseStep
}
