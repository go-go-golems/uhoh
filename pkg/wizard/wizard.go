package wizard

import (
	"context"
	"fmt"
	"os"

	"strings"

	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Wizard defines the top-level structure for a multi-step wizard.
type Wizard struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Steps       []steps.Step           `yaml:"steps"` // Interface allows different step types
	Theme       string                 `yaml:"theme,omitempty"`
	GlobalState map[string]interface{} `yaml:"global_state,omitempty"`
}

// Run executes the wizard steps sequentially.
func (w *Wizard) Run(ctx context.Context) (map[string]interface{}, error) {
	fmt.Printf("=== Starting Wizard: %s ===\n", w.Name)
	if w.Description != "" {
		fmt.Printf("%s\n", w.Description)
	}

	// --- State Management: Initialize state ---
	wizardState := make(map[string]interface{})
	if w.GlobalState != nil {
		for k, v := range w.GlobalState {
			wizardState[k] = v
		}
		fmt.Println("Initialized with Global State:")
		for k, v := range wizardState {
			fmt.Printf("  %s: %v\n", k, v)
		}
	}
	// --- End State Management ---

	// Basic execution loop - iterates through steps sequentially
	currentStepIndex := 0
	for currentStepIndex < len(w.Steps) {
		step := w.Steps[currentStepIndex]

		// TODO(manuel, 2024-08-05) Implement skip_condition check using @Expr
		// TODO(manuel, 2024-08-05) Implement 'before' callback execution

		fmt.Printf("\nExecuting Step %d/%d: ID = %s, Type = %s\n",
			currentStepIndex+1, len(w.Steps), step.ID(), step.Type())

		// --- State Management: Pass state to step ---
		stepResult, err := step.Execute(ctx, wizardState)
		// --- End State Management ---

		if err != nil {
			// Check for specific "not implemented" errors to allow continuing
			errMsg := err.Error()
			isNotImplemented := errors.Is(err, errors.New("step not implemented")) ||
				strings.Contains(errMsg, "not implemented") // Broader check

			if isNotImplemented {
				fmt.Printf("Warning: Step %s (%s) is not fully implemented. Skipping.\n", step.ID(), step.Type())
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

// LoadWizard loads a Wizard definition from a YAML file.
func LoadWizard(filePath string) (*Wizard, error) {
	yamlData, err := os.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not read wizard file %s", filePath)
	}

	var wizard Wizard
	err = yaml.Unmarshal(yamlData, &wizard)
	if err != nil {
		// Try to provide more context on YAML parsing errors
		// Note: This is a basic check; a full YAML validator would be more robust.
		var attempt map[string]interface{}
		if yaml.Unmarshal(yamlData, &attempt) != nil {
			return nil, errors.Wrap(err, "could not unmarshal wizard YAML (likely syntax error)")
		}
		// If basic map unmarshal works, the error is likely in the structure/types
		return nil, errors.Wrap(err, "could not unmarshal wizard YAML (check structure/types)")

	}

	// Post-unmarshal validation (optional but good)
	stepIDs := make(map[string]bool)
	for i, step := range wizard.Steps {
		if step == nil {
			return nil, errors.Errorf("step %d loaded as nil, check YAML structure and unmarshalling logic", i)
		}
		if step.Type() == "" {
			// This might happen if a step type is incorrect or structure is wrong
			return nil, errors.Errorf("step %d loaded with empty type (ID: %s), check YAML structure and unmarshalling logic", i, step.ID())
		}
		if step.ID() == "" {
			return nil, errors.Errorf("step %d (type: %s) is missing required 'id' field", i, step.Type())
		}
		if _, exists := stepIDs[step.ID()]; exists {
			return nil, errors.Errorf("duplicate step ID found: %s", step.ID())
		}
		stepIDs[step.ID()] = true
		// Could add more checks, e.g., required fields per type
		switch s := step.(type) {
		case *steps.FormStep:
			if s.FormData.Groups == nil { // Basic check for form structure
				fmt.Printf("Warning: Form step '%s' has nil Groups. Ensure 'form' key is correctly indented in YAML.  \n", s.ID())
			}
			// Add checks for other types if necessary
		}

	}

	return &wizard, nil
}
