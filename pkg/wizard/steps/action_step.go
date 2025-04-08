package steps

import (
	"context"

	"github.com/rs/zerolog/log"
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
	log.Debug().Str("stepId", as.ID()).Msgf("--- Step: %s ---", as.Title())
	// TODO(manuel, 2024-08-05) Implement action execution logic (callbacks, API calls)
	log.Warn().Str("stepId", as.ID()).Msg("Action Step (Not Implemented)")
	// Placeholder: return empty results
	stepResult := map[string]interface{}{}
	// Simulate putting a result in the output key
	// if as.OutputKey != "" {
	// 	stepResult[as.OutputKey] = "action_result_placeholder"
	// }
	return stepResult, ErrStepNotImplemented // Use standard error
}

func (as *ActionStep) GetBaseStep() *BaseStep {
	return &as.BaseStep
}
