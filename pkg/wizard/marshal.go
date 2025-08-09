package wizard

import (
	"strings"

	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

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
		if strings.Contains(err.Error(), "cannot unmarshal !!map into yaml.Node") || strings.Contains(err.Error(), "did not find expected key") {
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
			w.Steps = []steps.Step{} // Initialize empty steps slice
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
		if alias.StepsNode.Kind == 0 || alias.StepsNode.Tag == "!!null" { // Check if it's an empty node or null
			w.Steps = []steps.Step{}
			return nil
		}
		return errors.New("wizard 'steps' field must be a sequence (list)")
	}

	w.Steps = make([]steps.Step, 0, len(alias.StepsNode.Content))

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

		var step steps.Step
		var err error

		// Based on type, decode into the appropriate concrete struct
		switch stepType {
		case "form":
			var fs steps.FormStep
			err = stepNode.Decode(&fs)
			step = &fs
		case "decision":
			var ds steps.DecisionStep
			err = stepNode.Decode(&ds)
			step = &ds
		case "action":
			var as steps.ActionStep
			err = stepNode.Decode(&as)
			step = &as
		case "info":
			var is steps.InfoStep
			err = stepNode.Decode(&is)
			step = &is
		case "summary":
			var ss steps.SummaryStep
			err = stepNode.Decode(&ss)
			step = &ss
		default:
			// Attempt to decode into BaseStep to get ID for error message
			var base steps.BaseStep
			_ = stepNode.Decode(&base) // Ignore error here
			return errors.Errorf("unknown step type '%s' for step %d (ID: %s)", stepType, i, base.ID())
		}

		if err != nil {
			// Attempt to decode into BaseStep to get ID for error message
			var base steps.BaseStep
			_ = stepNode.Decode(&base) // Ignore error here
			return errors.Wrapf(err, "could not decode step %d (type: %s, ID: %s)", i, stepType, base.ID())
		}

		// Ensure the BaseStep fields are populated even after specific type decoding
		// This relies on `yaml:",inline"` in the concrete types.
		// Add a check:
		if step.ID() == "" {
			// If ID is still empty after decoding the specific type, something went wrong
			// or the ID field was missing in the YAML for this step.
			// Try decoding BaseStep explicitly to see if it helps (though unlikely if inline failed)
			var base steps.BaseStep
			if baseErr := stepNode.Decode(&base); baseErr == nil && base.ID() != "" {
				// If we could decode BaseStep and got an ID, update the step's BaseStep.
				// This requires the step to implement a setter or have GetBaseStep return a pointer.
				*step.GetBaseStep() = base // Update the embedded BaseStep
				// Note: Removed fmt.Printf warning as it's not needed anymore
			} else {
				// If ID is still empty, return an error as ID is mandatory.
				return errors.Errorf("step %d (type: %s) is missing required 'id' field", i, stepType)
			}
		}

		w.Steps = append(w.Steps, step)
	}

	return nil
}
