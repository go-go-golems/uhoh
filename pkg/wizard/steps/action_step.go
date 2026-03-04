package steps

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	ShowProgress *bool                  `yaml:"show_progress,omitempty"`
	ShowComplete *bool                  `yaml:"show_completion,omitempty"`

	// Non-YAML fields
	registry ActionCallbackRegistry // Registry for action callbacks
}

var _ Step = &ActionStep{}

const uiHandledKey = "_uhoh_ui_handled"

// ActionCallbackResult allows callbacks to return structured data and signal that they handled
// UI updates themselves (for example, launching a Bubble Tea program).
type ActionCallbackResult struct {
	Data      interface{}
	UIHandled bool
}

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

	showProgress := boolValue(as.ShowProgress, true)
	showCompletion := boolValue(as.ShowComplete, true)

	if showProgress {
		actionNote := huh.NewNote().
			Title(as.Title()).
			Description(fmt.Sprintf("Executing action: %s\n\nPlease wait...", as.FunctionName))

		go func() {
			_ = actionNote.Run()
		}()

		// Small delay to ensure note is visible before the callback potentially runs its own UI.
		time.Sleep(100 * time.Millisecond)
	}

	var (
		actionResult interface{}
		actionErr    error
		uiHandled    bool
	)

	// Execute the function via the registry if available
	if as.registry != nil {
		log.Debug().Str("stepId", as.ID()).Str("function", as.FunctionName).
			Interface("arguments", as.Arguments).Msg("Executing function via registry")

		rawResult, err := as.registry.ExecuteActionCallback(ctx, as.FunctionName, state, as.Arguments)
		actionResult, uiHandled = interpretActionResult(rawResult)
		actionErr = err
		if actionErr != nil {
			return nil, errors.Wrapf(actionErr, "error executing function %s", as.FunctionName)
		}
	} else {
		// Fallback to simulation for development/testing
		log.Debug().Str("stepId", as.ID()).Str("function", as.FunctionName).
			Msg("No registry available, simulating function execution")

		// Simulate some work
		time.Sleep(2 * time.Second)

		// Set a structured placeholder result
		actionResult = map[string]interface{}{
			"simulated": true,
			"function":  as.FunctionName,
			"message":   fmt.Sprintf("Simulated result from %s (no registry available)", as.FunctionName),
		}
		log.Warn().Str("stepId", as.ID()).Msg("Using simulated action result - no callback registry provided")
	}

	// If we have an output key, store the result
	if as.OutputKey != "" && actionResult != nil {
		stepResult[as.OutputKey] = actionResult
		log.Debug().Str("stepId", as.ID()).Str("outputKey", as.OutputKey).
			Interface("value", actionResult).Msg("Action result stored in state")
	}

	if showCompletion && !uiHandled {
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
	}

	return stepResult, nil
}

func (as *ActionStep) GetBaseStep() *BaseStep {
	return &as.BaseStep
}

func interpretActionResult(result interface{}) (interface{}, bool) {
	if result == nil {
		return nil, false
	}

	switch v := result.(type) {
	case ActionCallbackResult:
		return v.Data, v.UIHandled
	case *ActionCallbackResult:
		if v == nil {
			return nil, false
		}
		return v.Data, v.UIHandled
	case map[string]interface{}:
		if handled, ok := v[uiHandledKey].(bool); ok {
			filtered := make(map[string]interface{}, len(v))
			for k, val := range v {
				if k == uiHandledKey {
					continue
				}
				filtered[k] = val
			}
			return filtered, handled
		}
		return v, false
	default:
		return result, false
	}
}

func boolValue(v *bool, def bool) bool {
	if v == nil {
		return def
	}
	return *v
}

// BuildModel creates a huh.Form with a Note showing progress while the
// async callback runs. The form auto-completes (no user interaction needed)
// because the WizardModel handles advancement on ActionCompletedMsg.
func (as *ActionStep) BuildModel(state map[string]interface{}) (*huh.Form, map[string]interface{}, error) {
	if as.ActionType != "function" {
		return nil, nil, errors.Errorf("unsupported action type: %s", as.ActionType)
	}
	if as.FunctionName == "" {
		return nil, nil, errors.New("function name not specified for function-type action")
	}

	description := fmt.Sprintf("Executing: %s\n\nPlease wait...", as.FunctionName)
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title(as.Title()).
				Description(description).
				Next(true).
				NextLabel(""),
		),
	)

	return form, map[string]interface{}{}, nil
}

// RunAsync returns a tea.Cmd that executes the action callback and sends
// ActionCompletedMsg when done.
func (as *ActionStep) RunAsync(state map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		if as.registry == nil {
			// Simulate work when no registry is available.
			time.Sleep(500 * time.Millisecond)
			return ActionCompletedMsg{
				Result: map[string]interface{}{
					"simulated": true,
					"function":  as.FunctionName,
					"message":   fmt.Sprintf("Simulated result from %s", as.FunctionName),
				},
			}
		}

		rawResult, err := as.registry.ExecuteActionCallback(
			context.Background(), as.FunctionName, state, as.Arguments,
		)
		if err != nil {
			return ActionCompletedMsg{Err: err}
		}

		result, _ := interpretActionResult(rawResult)
		return ActionCompletedMsg{Result: result}
	}
}

// GetOutputKey returns the state key where the action result is stored.
func (as *ActionStep) GetOutputKey() string {
	return as.OutputKey
}

var _ AsyncEmbeddableStep = &ActionStep{}
