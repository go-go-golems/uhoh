package wizard

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Step represents a single step in the wizard.
type Step interface {
	ID() string
	Type() string
	Title() string
	Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error)
	// TODO(manuel, 2024-08-05) Add methods for SkipCondition, NextStep, etc. as needed
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

// FormStep represents a step that displays an interactive form.
type FormStep struct {
	BaseStep `yaml:",inline"`
	FormData pkg.Form `yaml:"form"` // Reuse the existing Form definition
}

var _ Step = &FormStep{}

// Execute runs the form defined in the step.
func (fs *FormStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	// TODO(manuel, 2024-08-05) Consider passing state into the form for defaults/pre-population
	// TODO(manuel, 2024-08-05) Add step title/description rendering
	fmt.Printf("\n--- Step: %s ---\n", fs.Title())
	if fs.Description != "" {
		fmt.Printf("%s\n", fs.Description)
	}

	// Set theme from wizard if not set in form itself (pass wizard theme down?)
	// if fs.FormData.Theme == "" && w.Theme != "" { // Need Wizard context here
	// 	fs.FormData.Theme = w.Theme
	// }

	// Run the actual form
	formResults, err := fs.FormData.Run(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "error running form step %s", fs.ID())
	}

	// TODO(manuel, 2024-08-05) Define how form results merge into the main wizard state
	// For now, just return the raw form results. The runner will merge them.
	return formResults, nil
}

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

// InfoStep represents a step that displays informational text.
type InfoStep struct {
	BaseStep `yaml:",inline"`
	Content  string `yaml:"content"`
}

var _ Step = &InfoStep{}

func (is *InfoStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n--- Step: %s ---\n", is.Title())
	if is.Description != "" {
		fmt.Printf("%s\n", is.Description)
	}
	fmt.Println(is.Content)
	// Info steps typically don't modify state
	return map[string]interface{}{}, nil
}

// SummaryStep represents a step that displays collected data.
type SummaryStep struct {
	BaseStep `yaml:",inline"`
	Template string `yaml:"template,omitempty"` // Optional Go template to format state
	Editable bool   `yaml:"editable,omitempty"`
	// TODO(manuel, 2024-08-05) Define fields for selecting which state keys to display
}

var _ Step = &SummaryStep{}

func (ss *SummaryStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n--- Step: %s ---\n", ss.Title())
	// TODO(manuel, 2024-08-05) Implement summary display logic (use template, show state)
	fmt.Println("Summary Step (Not Implemented)")
	fmt.Println("Current State:")
	for k, v := range state {
		fmt.Printf("  %s: %v\n", k, v)
	}
	// Summary steps don't modify state unless editable allows going back
	return map[string]interface{}{}, errors.New("summary step not implemented")
}

// Wizard defines the top-level structure for a multi-step wizard.
type Wizard struct {
	Name        string                 `yaml:"name"`
	Description string                 `yaml:"description,omitempty"`
	Steps       []Step                 `yaml:"steps"` // Interface allows different step types
	Theme       string                 `yaml:"theme,omitempty"`
	GlobalState map[string]interface{} `yaml:"global_state,omitempty"`
}

// UnmarshalYAML implements custom YAML unmarshalling for the Wizard struct.
// It specifically handles parsing the 'steps' list into the correct Step interface types.
func (w *Wizard) UnmarshalYAML(node *yaml.Node) error {
	// Use a temporary struct to unmarshal known fields first
	type WizardAlias struct {
		Name        string                 `yaml:"name"`
		Description string                 `yaml:"description,omitempty"`
		Theme       string                 `yaml:"theme,omitempty"`
		GlobalState map[string]interface{} `yaml:"global_state,omitempty"`
		StepsNode   yaml.Node              `yaml:"steps"` // Capture steps node separately
	}

	var alias WizardAlias
	if err := node.Decode(&alias); err != nil {
		// Handle case where 'steps' might be missing or not a sequence initially
		if err.Error() == "yaml: invalid node type for steps" {
			// Try decoding without steps
			type WizardAliasNoSteps struct {
				Name        string                 `yaml:"name"`
				Description string                 `yaml:"description,omitempty"`
				Theme       string                 `yaml:"theme,omitempty"`
				GlobalState map[string]interface{} `yaml:"global_state,omitempty"`
			}
			var aliasNoSteps WizardAliasNoSteps
			if err := node.Decode(&aliasNoSteps); err != nil {
				return errors.Wrap(err, "could not decode wizard YAML (no steps)")
			}
			w.Name = aliasNoSteps.Name
			w.Description = aliasNoSteps.Description
			w.Theme = aliasNoSteps.Theme
			w.GlobalState = aliasNoSteps.GlobalState
			w.Steps = []Step{} // Initialize empty steps slice
			return nil
		}
		return errors.Wrap(err, "could not decode wizard YAML base structure")
	}

	w.Name = alias.Name
	w.Description = alias.Description
	w.Theme = alias.Theme
	w.GlobalState = alias.GlobalState

	if alias.StepsNode.Kind != yaml.SequenceNode {
		// Allow wizards with no steps defined
		if alias.StepsNode.Kind == 0 { // Or check if it's an empty node
			w.Steps = []Step{}
			return nil
		}
		return errors.New("wizard 'steps' field must be a sequence (list)")
	}

	w.Steps = make([]Step, 0, len(alias.StepsNode.Content))

	for i, stepNode := range alias.StepsNode.Content {
		if stepNode.Kind != yaml.MappingNode {
			return errors.Errorf("step %d is not a mapping node", i)
		}

		// Decode into a temporary map to find the 'type'
		var stepMap map[string]interface{}
		if err := stepNode.Decode(&stepMap); err != nil {
			return errors.Wrapf(err, "could not decode step %d into map", i)
		}

		stepType, _ := stepMap["type"].(string)
		if stepType == "" {
			return errors.Errorf("step %d is missing required 'type' field", i)
		}

		var step Step
		var err error

		// Based on type, decode into the appropriate concrete struct
		switch stepType {
		case "form":
			var fs FormStep
			err = stepNode.Decode(&fs)
			step = &fs
		case "decision":
			var ds DecisionStep
			err = stepNode.Decode(&ds)
			step = &ds
		case "action":
			var as ActionStep
			err = stepNode.Decode(&as)
			step = &as
		case "info":
			var is InfoStep
			err = stepNode.Decode(&is)
			step = &is
		case "summary":
			var ss SummaryStep
			err = stepNode.Decode(&ss)
			step = &ss
		default:
			return errors.Errorf("unknown step type '%s' for step %d (ID: %s)", stepType, i, stepMap["id"])
		}

		if err != nil {
			return errors.Wrapf(err, "could not decode step %d (type: %s, ID: %s)", i, stepType, stepMap["id"])
		}

		// Ensure the BaseStep fields are populated even after specific type decoding
		// (might be redundant if `yaml:",inline"` works reliably, but good safety)
		if bs, ok := step.(interface{ GetBaseStep() *BaseStep }); ok {
			if bs.GetBaseStep().StepType == "" { // Check if type wasn't set by inline
				// Manually decode BaseStep fields if needed, though inline should handle it.
				var base BaseStep
				if baseErr := stepNode.Decode(&base); baseErr == nil {
					*bs.GetBaseStep() = base // Overwrite base fields if successful
				}
			}
		}

		w.Steps = append(w.Steps, step)
	}

	return nil
}

// Helper method to get BaseStep for reflection/assertion if needed later
func (fs *FormStep) GetBaseStep() *BaseStep     { return &fs.BaseStep }
func (ds *DecisionStep) GetBaseStep() *BaseStep { return &ds.BaseStep }
func (as *ActionStep) GetBaseStep() *BaseStep   { return &as.BaseStep }
func (is *InfoStep) GetBaseStep() *BaseStep     { return &is.BaseStep }
func (ss *SummaryStep) GetBaseStep() *BaseStep  { return &ss.BaseStep }

// Run executes the wizard steps sequentially.
func (w *Wizard) Run(ctx context.Context) (map[string]interface{}, error) {
	fmt.Printf("=== Starting Wizard: %s ===\n", w.Name)
	if w.Description != "" {
		fmt.Printf("%s\n", w.Description)
	}

	// Initialize state with global state if provided
	wizardState := make(map[string]interface{})
	if w.GlobalState != nil {
		for k, v := range w.GlobalState {
			wizardState[k] = v
		}
	}

	// Basic execution loop - iterates through steps sequentially
	currentStepIndex := 0
	for currentStepIndex < len(w.Steps) {
		step := w.Steps[currentStepIndex]

		// TODO(manuel, 2024-08-05) Implement skip_condition check using @Expr
		// TODO(manuel, 2024-08-05) Implement 'before' callback execution

		fmt.Printf("\nExecuting Step %d/%d: ID = %s, Type = %s\n",
			currentStepIndex+1, len(w.Steps), step.ID(), step.Type())

		stepResult, err := step.Execute(ctx, wizardState)
		if err != nil {
			// Allow steps to signal unimplemented state without halting everything
			// TODO(manuel, 2024-08-05): Improve this error handling / skipping logic
			if errors.Is(err, errors.New("step not implemented")) || errors.Is(err, errors.New("decision step not implemented")) || errors.Is(err, errors.New("action step not implemented")) || errors.Is(err, errors.New("summary step not implemented")) {
				fmt.Printf("Warning: Step %s (%s) is not fully implemented. Skipping.\n", step.ID(), step.Type())
				stepResult = map[string]interface{}{} // Treat as empty result
			} else {
				return wizardState, errors.Wrapf(err, "error executing step %d (ID: %s)", currentStepIndex, step.ID())
			}
		}

		// TODO(manuel, 2024-08-05) Implement 'after' callback execution
		// TODO(manuel, 2024-08-05) Implement 'validation' callback/logic

		// Merge step results into the main wizard state
		if stepResult != nil {
			for k, v := range stepResult {
				// TODO(manuel, 2024-08-05) Define merging strategy (overwrite, deep merge?)
				// Simple overwrite for now
				wizardState[k] = v
				fmt.Printf("State updated: %s = %v\n", k, v) // Debug logging
			}
		}

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
		return nil, errors.Wrap(err, "could not unmarshal wizard YAML")
	}

	// Post-unmarshal validation (optional but good)
	for i, step := range wizard.Steps {
		if step.Type() == "" {
			// This indicates a potential issue in unmarshalling or the YAML structure
			return nil, errors.Errorf("step %d loaded with empty type, check YAML structure and unmarshalling logic", i)
		}
		// Could add more checks, e.g., required fields per type
	}

	return &wizard, nil
}
