package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// ActionCallbackFunc defines the signature for an action callback function that can be executed by an ActionStep.
// This should match the ActionCallbackFunc in the wizard package.
type ActionCallbackFunc func(ctx context.Context, state map[string]interface{}, args map[string]interface{}) (interface{}, error)

// ActionCallbackRegistry defines an interface for accessing action callbacks.
// This allows the ActionStep to execute callbacks without directly depending on the Wizard type.
type ActionCallbackRegistry interface {
	ExecuteActionCallback(ctx context.Context, callbackName string, state map[string]interface{}, args map[string]interface{}) (interface{}, error)
}

// ActionStep represents a step that performs a backend action.
type ActionStep struct {
	BaseStep     `yaml:",inline"`
	ActionType   string                 `yaml:"action_type"`             // e.g., "function", "api_call"
	FunctionName string                 `yaml:"function_name,omitempty"` // For action_type: function
	Arguments    map[string]interface{} `yaml:"arguments,omitempty"`
	OutputKey    string                 `yaml:"output_key,omitempty"`

	// Non-YAML fields
	registry ActionCallbackRegistry // Registry for action callbacks
}

var _ Step = &ActionStep{}

// SetCallbackRegistry sets the callback registry for this action step.
func (as *ActionStep) SetCallbackRegistry(registry ActionCallbackRegistry) {
	as.registry = registry
}

func (as *ActionStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	log.Debug().Str("stepId", as.ID()).Msgf("--- Step: %s ---", as.Title())

	stepResult := map[string]interface{}{}

	// Currently only support "function" type actions
	if as.ActionType != "function" {
		return nil, errors.Errorf("unsupported action type: %s", as.ActionType)
	}

	if as.FunctionName == "" {
		return nil, errors.New("function name not specified for function-type action")
	}

	// Show a note that we're executing an action
	actionNote := huh.NewNote().
		Title(as.Title()).
		Description(fmt.Sprintf("Executing action: %s\n\nPlease wait...", as.FunctionName))

	// Display the note but don't wait for completion
	go func() {
		_ = actionNote.Run()
	}()

	// Small delay to ensure note is visible
	time.Sleep(100 * time.Millisecond)

	var actionResult interface{}
	var actionErr error

	// Execute the function via the registry if available
	if as.registry != nil {
		log.Debug().Str("stepId", as.ID()).Str("function", as.FunctionName).
			Interface("arguments", as.Arguments).Msg("Executing function via registry")

		actionResult, actionErr = as.registry.ExecuteActionCallback(ctx, as.FunctionName, state, as.Arguments)
		if actionErr != nil {
			return nil, errors.Wrapf(actionErr, "error executing function %s", as.FunctionName)
		}
	} else {
		// Fallback to simulation for development/testing
		log.Debug().Str("stepId", as.ID()).Str("function", as.FunctionName).
			Msg("No registry available, simulating function execution")

		// Simulate some work
		time.Sleep(2 * time.Second)

		// Set a placeholder result
		actionResult = fmt.Sprintf("Simulated result from %s (no registry available)", as.FunctionName)
		log.Warn().Str("stepId", as.ID()).Msg("Using simulated action result - no callback registry provided")
	}

	// If we have an output key, store the result
	if as.OutputKey != "" && actionResult != nil {
		stepResult[as.OutputKey] = actionResult
		log.Debug().Str("stepId", as.ID()).Str("outputKey", as.OutputKey).
			Interface("value", actionResult).Msg("Action result stored in state")
	}

	// Show a confirmation message after the action completes
	confirmation := huh.NewNote().
		Title("Action Complete").
		Description(fmt.Sprintf("Action '%s' completed successfully.", as.FunctionName))

	err := confirmation.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAborted
		}
		return nil, errors.Wrap(err, "error showing completion message")
	}

	return stepResult, nil
}

func (as *ActionStep) GetBaseStep() *BaseStep {
	return &as.BaseStep
}
