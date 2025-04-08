package wizard

import (
	"context"
	"fmt"
	"os"

	"strings"

	"github.com/expr-lang/expr"
	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// WizardFunc defines the signature for custom functions usable in Expr conditions.
// It takes the current wizard state as input.
type WizardFunc func(arguments ...interface{}) (interface{}, error)

// Wizard defines the top-level structure for a multi-step wizard.
type Wizard struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Steps       steps.WizardSteps      `yaml:"steps"` // Custom type for unmarshalling
	Theme       string                 `yaml:"theme,omitempty"`
	GlobalState map[string]interface{} `yaml:"global_state,omitempty"`

	// Non-YAML fields
	customFunctions map[string]WizardFunc
	initialState    map[string]interface{} // Added for external initial state
}

// WizardOption is used to configure a Wizard during creation.
type WizardOption func(*Wizard)

// WithCustomFunction registers a custom function for use in expressions.
func WithCustomFunction(name string, fn WizardFunc) WizardOption {
	return func(w *Wizard) {
		if w.customFunctions == nil {
			w.customFunctions = make(map[string]WizardFunc)
		}
		w.customFunctions[name] = fn
	}
}

func WithCustomFunctions(functions map[string]WizardFunc) WizardOption {
	return func(w *Wizard) {
		for name, fn := range functions {
			w.customFunctions[name] = fn
		}
	}
}

// WithInitialState provides an initial state map to the wizard, merged over global_state.
func WithInitialState(state map[string]interface{}) WizardOption {
	return func(w *Wizard) {
		w.initialState = state
	}
}

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
	for name, fn := range w.customFunctions {
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
	fmt.Printf("=== Starting Wizard: %s ===\n", w.Name)
	if w.Description != "" {
		fmt.Printf("%s\n", w.Description)
	}

	// --- State Management: Initialize state ---
	wizardState := make(map[string]interface{})
	// 1. Load GlobalState from YAML
	if len(w.GlobalState) > 0 { // Check if GlobalState has keys
		fmt.Println("Initializing state with GlobalState (from YAML):")
		for k, v := range w.GlobalState {
			wizardState[k] = v
			fmt.Printf("  - %s: %v\n", k, v) // Log initial value
		}
	} else {
		fmt.Println("No GlobalState defined in YAML.")
	}

	// 2. Merge w.initialState (from YAML) with GlobalState
	if len(w.initialState) > 0 {
		fmt.Println("Merging initialState (from YAML):")
		for k, v := range w.initialState {
			wizardState[k] = v
		}
	}

	// 2. Merge InitialState passed via parameter (overwrites GlobalState)
	if len(initialState) > 0 { // Check if initialState has keys
		fmt.Println("Merging InitialState (from Run argument/CLI):")
		for k, v := range initialState {
			oldValue, exists := wizardState[k]
			wizardState[k] = v
			if exists {
				fmt.Printf("  - %s: %v (overwrites %v)\n", k, v, oldValue)
			} else {
				fmt.Printf("  - %s: %v (added)\n", k, v)
			}
		}
	} else {
		fmt.Println("No additional InitialState provided via Run argument/CLI.")
	}
	fmt.Println("--- Initial State Finalized ---") // Separator
	// --- End State Management ---

	// Basic execution loop - iterates through steps sequentially
	currentStepIndex := 0
	for currentStepIndex < len(w.Steps) {
		step := w.Steps[currentStepIndex]

		// --- Skip Condition Check --- START ---
		skipCond := step.SkipCondition()
		if skipCond != "" {
			fmt.Printf("Checking skip condition for step %s: %s\n", step.ID(), skipCond)
			// Use the method on w to access custom functions
			skip, err := w.evaluateExprCondition(skipCond, wizardState)
			if err != nil {
				// Decide whether to halt or just warn on evaluation error
				fmt.Printf("Warning: Could not evaluate skip condition for step %s: %v. Step will NOT be skipped.\n", step.ID(), err)
				// Optionally, you could return an error here to stop the wizard:
				// return wizardState, errors.Wrapf(err, "error evaluating skip condition for step %s", step.ID())
			} else if skip {
				fmt.Printf("Skipping step %d (ID: %s) due to condition: %s\n", currentStepIndex+1, step.ID(), skipCond)
				currentStepIndex++ // Move to the next step index
				continue           // Skip the rest of the loop for this step
			}
		}
		// --- Skip Condition Check --- END ---

		// TODO(manuel, 2024-08-05) Implement 'before' callback execution

		fmt.Printf("\nExecuting Step %d/%d: ID = %s, Type = %s\n",
			currentStepIndex+1, len(w.Steps), step.ID(), step.Type())

		// --- State Management: Pass state to step ---
		stepResult, err := step.Execute(ctx, wizardState)
		// --- End State Management ---

		if err != nil {
			// Check for specific errors like Abort or NotImplemented
			if errors.Is(err, steps.ErrUserAborted) {
				fmt.Println("Wizard aborted by user.")
				return wizardState, err // Return the specific abort error
			}

			errMsg := err.Error()
			isNotImplemented := errors.Is(err, steps.ErrStepNotImplemented) || // Use defined error
				strings.Contains(errMsg, "not implemented") // Keep broader check for now

			if isNotImplemented {
				fmt.Printf("Warning: Step %s (%s) is not fully implemented. Skipping execution logic.\n", step.ID(), step.Type())
				stepResult = map[string]interface{}{} // Treat as empty result to continue loop
			} else {
				// For other errors, halt execution
				return wizardState, errors.Wrapf(err, "error executing step %d (ID: %s)", currentStepIndex, step.ID())
			}
		}

		// TODO(manuel, 2024-08-05) Implement 'after' callback execution
		// TODO(manuel, 2024-08-05) Implement 'validation' callback/logic

		// --- State Management: Merge step results into state ---
		if stepResult != nil {
			merged := false
			for k, v := range stepResult {
				// Simple overwrite strategy for now
				wizardState[k] = v
				fmt.Printf("State updated: %s = %v\n", k, v) // Debug logging
				merged = true
			}
			if merged {
				fmt.Println("Current Wizard State after merge:")
				for k, v := range wizardState {
					fmt.Printf("  %s: %v\n", k, v)
				}
			}
		}
		// --- End State Management ---

		// Determine the next step
		// TODO(manuel, 2024-08-05) Implement navigation logic (callbacks, next_step_map, next_step)
		// Basic linear progression for now:
		nextStepIndex := currentStepIndex + 1

		// TODO(manuel, 2024-08-05) Add check for explicit next_step field or callback result

		currentStepIndex = nextStepIndex
	}

	fmt.Printf("\n=== Wizard '%s' Finished ===\n", w.Name)
	fmt.Println("Final State:")
	for k, v := range wizardState {
		fmt.Printf("  %s: %v\n", k, v)
	}

	return wizardState, nil
}

// LoadWizard loads a Wizard definition from a YAML file and applies options.
func LoadWizard(filePath string, opts ...WizardOption) (*Wizard, error) {
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read wizard file %s", filePath)
	}

	var wizard Wizard
	err = yaml.Unmarshal(yamlData, &wizard)
	if err != nil {
		// Try to provide more context on YAML parsing errors
		var attempt map[string]interface{}
		if yaml.Unmarshal(yamlData, &attempt) != nil {
			return nil, errors.Wrap(err, "could not unmarshal wizard YAML (likely syntax error)")
		}
		// If basic map unmarshal works, the error is likely in the structure/types
		return nil, errors.Wrap(err, "could not unmarshal wizard YAML (check structure/types)")
	}

	// Apply functional options *after* unmarshalling
	for _, opt := range opts {
		opt(&wizard)
	}

	// Post-unmarshal validation
	stepIDs := make(map[string]bool)
	for i, step := range wizard.Steps {
		if step == nil {
			// This check should be less necessary with the custom unmarshaller, but keep as a safeguard
			return nil, errors.Errorf("step %d loaded as nil, check YAML structure and UnmarshalStepYAML function", i)
		}
		stepID := step.ID()
		if stepID == "" {
			// The custom unmarshaller should ideally catch steps without IDs earlier
			return nil, errors.Errorf("step %d (type: %s) is missing required 'id' field", i, step.Type())
		}
		if _, exists := stepIDs[stepID]; exists {
			return nil, errors.Errorf("duplicate step ID found: %s", stepID)
		}
		stepIDs[stepID] = true

		// Validate required fields per type (example for FormStep)
		switch s := step.(type) {
		case *steps.FormStep:
			// Check if the form data structure itself is present, not a specific key.
			// A more robust check might ensure Groups is not nil/empty.
			if s.FormData.Groups == nil { // Corrected check
				// This might indicate an empty 'form:' key or incorrect indentation in YAML.
				// Consider if this should be an error or just a warning.
				fmt.Printf("Warning: Form step '%s' has nil FormData.Groups. Check YAML structure.\n", stepID)
				// return nil, errors.Errorf("form step '%s' has missing or invalid form definition", stepID)
			}
		case *steps.DecisionStep:
			if s.TargetKey == "" {
				return nil, errors.Errorf("decision step '%s' is missing required 'target_key' field", stepID)
			}
			if len(s.Choices) == 0 {
				return nil, errors.Errorf("decision step '%s' must have at least one 'choice'", stepID)
			}
			// Add similar validation for ActionStep, InfoStep, SummaryStep as needed
		}
	}

	return &wizard, nil
}
