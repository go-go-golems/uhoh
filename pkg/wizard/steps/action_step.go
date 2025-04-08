package steps

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// ActionStep represents a step that performs a backend action.
type ActionStep struct {
	BaseStep     `yaml:",inline"`
	ActionType   string                 `yaml:"action_type"`             // e.g., "function", "api_call"
	FunctionName string                 `yaml:"function_name,omitempty"` // For action_type: function
	Arguments    map[string]interface{} `yaml:"arguments,omitempty"`
	OutputKey    string                 `yaml:"output_key,omitempty"`
	// TODO(manuel, 2024-08-05) Add fields for api_call type
}

var _ Step = &ActionStep{}

func (as *ActionStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n--- Step: %s ---\n", as.Title())
	// TODO(manuel, 2024-08-05) Implement action execution logic (callbacks, API calls)
	fmt.Println("Action Step (Not Implemented)")
	// Placeholder: return empty results
	stepResult := map[string]interface{}{}
	// Simulate putting a result in the output key
	// if as.OutputKey != "" {
	// 	stepResult[as.OutputKey] = "action_result_placeholder"
	// }
	return stepResult, errors.New("action step not implemented")
}

func (as *ActionStep) GetBaseStep() *BaseStep {
	return &as.BaseStep
}
