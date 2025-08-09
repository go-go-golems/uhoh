package wizard

import (
	"context"
	"os"

	"github.com/expr-lang/expr"
	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ExprFunc defines the signature for custom functions usable in Expr conditions.
// It takes the current wizard state as input.
type ExprFunc func(arguments ...interface{}) (interface{}, error)

// WizardCallbackFunc defines the signature for callbacks executed during the wizard lifecycle.
// It receives the context and the current wizard state.
// It can return:
// - result: Arbitrary data, often used by 'action' steps or stored by 'validation'.
// - nextStepID: A pointer to a string indicating the ID of the next step to jump to (used by 'navigation' callbacks). nil means default flow.
// - error: An error if the callback fails.
type WizardCallbackFunc func(ctx context.Context, state map[string]interface{}) (result interface{}, nextStepID *string, err error)

// ActionCallbackFunc defines the signature for callbacks specifically designed for action steps.
// It receives the context, the current wizard state, and a map of arguments from the action step.
// It returns a result that can be stored in the action's output key.
type ActionCallbackFunc func(ctx context.Context, state map[string]interface{}, args map[string]interface{}) (interface{}, error)

// Wizard defines the top-level structure for a multi-step wizard.
type Wizard struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Steps       steps.WizardSteps      `yaml:"steps"` // Custom type for unmarshalling
	Theme       string                 `yaml:"theme,omitempty"`
	GlobalState map[string]interface{} `yaml:"global_state,omitempty"`

	// Non-YAML fields
	exprFunctions   map[string]ExprFunc // Renamed from customFunctions
	callbacks       map[string]WizardCallbackFunc
	actionCallbacks map[string]ActionCallbackFunc // New field for action-specific callbacks
	initialState    map[string]interface{}        // Added for external initial state
}

// WizardOption is used to configure a Wizard during creation.
type WizardOption func(*Wizard)

// WithExprFunction registers a custom function for use in expressions.
func WithExprFunction(name string, fn ExprFunc) WizardOption {
	return func(w *Wizard) {
		if w.exprFunctions == nil {
			w.exprFunctions = make(map[string]ExprFunc)
		}
		w.exprFunctions[name] = fn
	}
}

// WithExprFunctions registers multiple custom functions for use in expressions.
func WithExprFunctions(functions map[string]ExprFunc) WizardOption {
	return func(w *Wizard) {
		if w.exprFunctions == nil {
			w.exprFunctions = make(map[string]ExprFunc)
		}
		for name, fn := range functions {
			w.exprFunctions[name] = fn
		}
	}
}

// WithCallback registers a callback function for use at different lifecycle points.
func WithCallback(name string, fn WizardCallbackFunc) WizardOption {
	return func(w *Wizard) {
		if w.callbacks == nil {
			w.callbacks = make(map[string]WizardCallbackFunc)
		}
		w.callbacks[name] = fn
	}
}

// WithCallbacks registers multiple callback functions.
func WithCallbacks(callbacks map[string]WizardCallbackFunc) WizardOption {
	return func(w *Wizard) {
		if w.callbacks == nil {
			w.callbacks = make(map[string]WizardCallbackFunc)
		}
		for name, fn := range callbacks {
			w.callbacks[name] = fn
		}
	}
}

// WithInitialState provides an initial state map to the wizard, merged over global_state.
func WithInitialState(state map[string]interface{}) WizardOption {
	return func(w *Wizard) {
		w.initialState = state
	}
}

// WithActionCallback registers a callback function specifically for action steps.
func WithActionCallback(name string, fn ActionCallbackFunc) WizardOption {
	return func(w *Wizard) {
		if w.actionCallbacks == nil {
			w.actionCallbacks = make(map[string]ActionCallbackFunc)
		}
		w.actionCallbacks[name] = fn
	}
}

// WithActionCallbacks registers multiple action callback functions.
func WithActionCallbacks(callbacks map[string]ActionCallbackFunc) WizardOption {
	return func(w *Wizard) {
		if w.actionCallbacks == nil {
			w.actionCallbacks = make(map[string]ActionCallbackFunc)
		}
		for name, fn := range callbacks {
			w.actionCallbacks[name] = fn
		}
	}
}

// Make sure Wizard implements ActionCallbackRegistry
var _ steps.ActionCallbackRegistry = &Wizard{}

// evaluateExprCondition evaluates a condition string against the wizard state,
// including any registered custom functions.
func (w *Wizard) evaluateExprCondition(condition string, state map[string]interface{}) (bool, error) {
	if condition == "" {
		return false, nil // No condition means don't skip/evaluate
	}

	// Prepare Expr options
	opts := []expr.Option{
		expr.Env(state),
		expr.AsBool(),
	}

	// Add custom functions
	for name, fn := range w.exprFunctions {
		opts = append(opts, expr.Function(name, fn))
	}

	program, err := expr.Compile(condition, opts...)
	if err != nil {
		return false, errors.Wrapf(err, "failed to compile condition: %s", condition)
	}

	result, err := expr.Run(program, state)
	if err != nil {
		// It's often useful to know *what* failed to evaluate, e.g. missing variable
		return false, errors.Wrapf(err, "failed to run condition: %s", condition)
	}

	boolResult, ok := result.(bool)
	if !ok {
		return false, errors.Errorf("condition did not return a boolean: %s (returned %T)", condition, result)
	}

	return boolResult, nil
}

// Run executes the wizard steps sequentially.
// It now accepts an initial state map that overrides/merges with the global state.
func (w *Wizard) Run(ctx context.Context, initialState map[string]interface{}) (map[string]interface{}, error) {
	logger := log.With().Str("wizardName", w.Name).Logger()
	logger.Debug().Msg("Starting Wizard")
	if w.Description != "" {
		logger.Debug().Msg(w.Description)
	}

	// --- State Management: Initialize state ---
	wizardState := make(map[string]interface{})
	// 1. Load GlobalState from YAML
	if len(w.GlobalState) > 0 { // Check if GlobalState has keys
		logger.Debug().Interface("globalState", w.GlobalState).Msg("Initializing state with GlobalState (from YAML)")
		for k, v := range w.GlobalState {
			wizardState[k] = v
		}
	} else {
		logger.Debug().Msg("No GlobalState defined in YAML.")
	}

	// 2. Merge w.initialState (from YAML) with GlobalState
	if len(w.initialState) > 0 {
		logger.Debug().Interface("initialStateYAML", w.initialState).Msg("Merging initialState (from YAML)")
		for k, v := range w.initialState {
			wizardState[k] = v
		}
	}

	// 2. Merge InitialState passed via parameter (overwrites GlobalState)
	if len(initialState) > 0 { // Check if initialState has keys
		logger.Debug().Interface("initialStateArg", initialState).Msg("Merging InitialState (from Run argument/CLI)")
		for k, v := range initialState {
			_, exists := wizardState[k]
			wizardState[k] = v
			logger.Debug().Str("key", k).Interface("value", v).Bool("overwritten", exists).Msg("Merged initial state value")
		}
	} else {
		logger.Debug().Msg("No additional InitialState provided via Run argument/CLI.")
	}
	logger.Debug().Interface("finalInitialState", wizardState).Msg("Initial State Finalized")
	// --- End State Management ---

	// Set up ActionStep registries
	for i := range w.Steps {
		if actionStep, ok := w.Steps[i].(*steps.ActionStep); ok {
			actionStep.SetCallbackRegistry(w)
			logger.Debug().Str("stepId", actionStep.ID()).Msg("Registered action callbacks for step")
		}
	}

	// Basic execution loop - iterates through steps sequentially
	currentStepIndex := 0
	for currentStepIndex < len(w.Steps) {
		step := w.Steps[currentStepIndex]
		stepID := step.ID()
		stepType := step.Type()
		stepLogger := logger.With().
			Str("stepId", stepID).
			Str("stepType", stepType).
			Int("stepIndex", currentStepIndex).
			Int("totalSteps", len(w.Steps)).
			Int("currentStepIndex", currentStepIndex).
			Logger()

		nextStepIDOverride := new(string) // Used to capture navigation overrides from callbacks
		*nextStepIDOverride = ""          // Initialize empty, means no override

		// --- Skip Condition Check --- START ---
		skipCond := step.SkipCondition()
		if skipCond != "" {
			stepLogger.Debug().Str("condition", skipCond).Msg("Checking skip condition")
			skip, err := w.evaluateExprCondition(skipCond, wizardState)
			if err != nil {
				stepLogger.Warn().Err(err).Str("condition", skipCond).Msg("Could not evaluate skip condition, step will NOT be skipped")
				// Optionally return error: return wizardState, errors.Wrapf(err, "error evaluating skip condition for step %s", stepID)
			} else if skip {
				stepLogger.Debug().Str("condition", skipCond).Msg("Skipping step due to condition")
				currentStepIndex++
				continue
			}
		}
		// --- Skip Condition Check --- END ---

		// --- Before Callback --- START ---
		if beforeCallbackName := step.BeforeCallback(); beforeCallbackName != "" {
			callback, found := w.callbacks[beforeCallbackName]
			if !found {
				stepLogger.Warn().Str("callbackName", beforeCallbackName).Msg("'before' callback not registered, skipping")
			} else {
				stepLogger.Debug().Str("callbackName", beforeCallbackName).Msg("Executing 'before' callback")
				_, _, err := callback(ctx, wizardState)
				if err != nil {
					return wizardState, errors.Wrapf(err, "'before' callback '%s' for step '%s' failed", beforeCallbackName, stepID)
				}
				// Optionally update state based on callback result? TBD
			}
		}
		// --- Before Callback --- END ---

		stepLogger.Debug().Msgf("Executing Step %d/%d", currentStepIndex+1, len(w.Steps))

		// --- State Management: Pass state to step --- // TODO(manuel, 2024-08-06) Pass logger too?
		stepResult, err := step.Execute(ctx, wizardState)
		// --- End State Management --- //

		if err != nil {
			// Check for specific errors like Abort or NotImplemented
			if errors.Is(err, steps.ErrUserAborted) {
				stepLogger.Debug().Msg("Wizard aborted by user")
				return wizardState, err // Return the specific abort error
			}

			if errors.Is(err, steps.ErrStepNotImplemented) {
				stepLogger.Warn().Msg("Step is not fully implemented. Skipping execution logic.")
				stepResult = map[string]interface{}{} // Treat as empty result to continue loop
			} else {
				// For other errors, halt execution
				stepLogger.Error().Err(err).Msg("Error executing step")
				return wizardState, errors.Wrapf(err, "error executing step %d (ID: %s)", currentStepIndex, stepID)
			}
		}

		// --- After Callback --- START ---
		if afterCallbackName := step.AfterCallback(); afterCallbackName != "" {
			callback, found := w.callbacks[afterCallbackName]
			if !found {
				stepLogger.Warn().Str("callbackName", afterCallbackName).Msg("'after' callback not registered, skipping")
			} else {
				stepLogger.Debug().Str("callbackName", afterCallbackName).Msg("Executing 'after' callback")
				_, _, err := callback(ctx, wizardState)
				if err != nil {
					return wizardState, errors.Wrapf(err, "'after' callback '%s' for step '%s' failed", afterCallbackName, stepID)
				}
				// TODO(manuel, 2024-08-06) Decide how 'after' callback results affect state.
			}
		}
		// --- After Callback --- END ---

		// --- State Management: Merge step results into state ---
		if stepResult != nil {
			merged := false
			for k, v := range stepResult {
				wizardState[k] = v
				stepLogger.Debug().Str("key", k).Interface("value", v).Msg("State updated")
				merged = true
			}
			if merged {
				stepLogger.Debug().Interface("newState", wizardState).Msg("Current Wizard State after merge")
			}
		}
		// --- End State Management ---

		// --- Validation Callback --- START ---
		if validationCallbackName := step.ValidationCallback(); validationCallbackName != "" {
			callback, found := w.callbacks[validationCallbackName]
			if !found {
				stepLogger.Warn().Str("callbackName", validationCallbackName).Msg("'validation' callback not registered, skipping")
			} else {
				stepLogger.Debug().Str("callbackName", validationCallbackName).Msg("Executing 'validation' callback")
				_, _, err := callback(ctx, wizardState)
				if err != nil {
					// Validation failure should likely halt the process or trigger remediation (TBD)
					return wizardState, errors.Wrapf(err, "'validation' callback '%s' for step '%s' failed", validationCallbackName, stepID)
				}
				stepLogger.Debug().Str("callbackName", validationCallbackName).Msg("Validation callback completed successfully")
			}
		}
		// --- Validation Callback --- END ---

		// Determine the next step
		var nextStepIndex int

		// --- Navigation Callback --- START ---
		if navigationCallbackName := step.NavigationCallback(); navigationCallbackName != "" {
			callback, found := w.callbacks[navigationCallbackName]
			if !found {
				stepLogger.Warn().Str("callbackName", navigationCallbackName).Msg("'navigation' callback not registered, using default navigation")
			} else {
				stepLogger.Debug().Str("callbackName", navigationCallbackName).Msg("Executing 'navigation' callback")
				_, nextStepIDPtr, err := callback(ctx, wizardState)
				if err != nil {
					return wizardState, errors.Wrapf(err, "'navigation' callback '%s' for step '%s' failed", navigationCallbackName, stepID)
				}
				if nextStepIDPtr != nil {
					*nextStepIDOverride = *nextStepIDPtr // Capture the override
					stepLogger.Debug().Str("callbackName", navigationCallbackName).Str("nextStepId", *nextStepIDOverride).Msg("Navigation callback requests jump to step")
				} else {
					stepLogger.Debug().Str("callbackName", navigationCallbackName).Msg("Navigation callback did not specify a next step, using default flow")
				}
			}
		}
		// --- Navigation Callback --- END ---

		// --- Navigation Logic --- START ---
		if *nextStepIDOverride != "" {
			foundIndex := -1
			for i, s := range w.Steps {
				if s.ID() == *nextStepIDOverride {
					foundIndex = i
					break
				}
			}
			if foundIndex == -1 {
				err := errors.Errorf("navigation callback requested jump to non-existent step ID: '%s' from step '%s'", *nextStepIDOverride, stepID)
				stepLogger.Error().Err(err).Str("requestedStepId", *nextStepIDOverride).Msg("Invalid navigation target")
				return wizardState, err
			}
			nextStepIndex = foundIndex
			stepLogger.Debug().Int("nextStepIndex", nextStepIndex).Str("nextStepId", *nextStepIDOverride).Msg("Navigating based on callback override")
		} else {
			// TODO(manuel, 2024-08-06) Add logic for next_step_map (decision) and next_step field here
			// Default linear progression if no override or specific field
			nextStepIndex = currentStepIndex + 1
			if nextStepIndex < len(w.Steps) {
				stepLogger.Debug().Int("nextStepIndex", nextStepIndex).Str("nextStepId", w.Steps[nextStepIndex].ID()).Msg("Navigating linearly to next step")
			} else {
				stepLogger.Debug().Msg("Reached end of steps")
			}
		}
		// --- Navigation Logic --- END ---

		currentStepIndex = nextStepIndex
	}

	logger.Debug().Interface("finalState", wizardState).Msg("Wizard Finished")

	return wizardState, nil
}

// LoadWizard loads a Wizard definition from a YAML file and applies options.
func LoadWizard(filePath string, opts ...WizardOption) (*Wizard, error) {
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read wizard file %s", filePath)
	}

	var wizard Wizard
	log.Debug().Str("filePath", filePath).Int("bytes", len(yamlData)).Msg("Attempting to unmarshal wizard YAML")
	err = yaml.Unmarshal(yamlData, &wizard)
	if err != nil {
		log.Error().Err(err).Str("filePath", filePath).Msg("Failed to unmarshal wizard YAML")
		// Try to provide more context on YAML parsing errors
		var attempt map[string]interface{}
		if yaml.Unmarshal(yamlData, &attempt) != nil {
			// If even basic map unmarshal fails, it's likely a syntax error
			return nil, errors.Wrap(err, "could not unmarshal wizard YAML (likely syntax error)")
		}
		// If basic map unmarshal works, the error is likely in the structure/types (caught by custom unmarshaler)
		return nil, errors.Wrap(err, "could not unmarshal wizard YAML (check structure/types, possibly caught by custom step unmarshaler)")
	}

	// Apply functional options *after* unmarshalling
	for _, opt := range opts {
		opt(&wizard)
	}

	// Post-unmarshal validation (mostly done by custom unmarshaller now)
	stepIDs := make(map[string]bool)
	for i, step := range wizard.Steps {
		if step == nil {
			// Should be caught by the custom unmarshaller, but defensive check
			return nil, errors.Errorf("step %d loaded as nil, check YAML structure and UnmarshalStepYAML function", i)
		}
		stepID := step.ID()
		if stepID == "" {
			// Should be caught by the custom unmarshaller
			return nil, errors.Errorf("step %d (type: %s) is missing required 'id' field", i, step.Type())
		}
		if _, exists := stepIDs[stepID]; exists {
			return nil, errors.Errorf("duplicate step ID found: %s", stepID)
		}
		stepIDs[stepID] = true

		// Remove the type switch validation here; it's handled by the custom unmarshaller
		// and caused linter errors due to signature mismatches during refactoring.
		// The custom unmarshaller provides more specific error messages if decoding fails.
	}

	log.Debug().Str("filePath", filePath).Str("wizardName", wizard.Name).Int("stepCount", len(wizard.Steps)).Msg("Wizard loaded successfully")
	return &wizard, nil
}

// ExecuteActionCallback looks up and executes an action callback by name.
// It returns the result of the callback execution or an error if the callback
// is not found or fails.
func (w *Wizard) ExecuteActionCallback(ctx context.Context, callbackName string, state map[string]interface{}, args map[string]interface{}) (interface{}, error) {
	if w.actionCallbacks == nil {
		return nil, errors.Errorf("no action callbacks registered")
	}

	callback, found := w.actionCallbacks[callbackName]
	if !found {
		return nil, errors.Errorf("action callback '%s' not registered", callbackName)
	}

	return callback(ctx, state, args)
}
