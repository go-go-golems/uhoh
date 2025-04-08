package steps

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Standard errors
var (
	ErrUserAborted        = errors.New("user aborted")
	ErrStepNotImplemented = errors.New("step not implemented")
)

// Step represents a single step in the wizard.
type Step interface {
	ID() string
	Type() string
	Title() string
	Description() string
	SkipCondition() string
	Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error)
	GetBaseStep() *BaseStep
	// Callback methods return the name of the registered callback (if any)
	BeforeCallback() string
	AfterCallback() string
	ValidationCallback() string
	NavigationCallback() string
}

// BaseStep contains common fields for all step types.
type BaseStep struct {
	StepID                 string `yaml:"id"`
	StepType               string `yaml:"type"`
	StepTitle              string `yaml:"title,omitempty"`
	StepDescription        string `yaml:"description,omitempty"`
	StepSkipCondition      string `yaml:"skip_condition,omitempty"`
	NextStep               string `yaml:"next_step,omitempty"`
	StepBeforeCallback     string `yaml:"before,omitempty"`
	StepAfterCallback      string `yaml:"after,omitempty"`
	StepValidationCallback string `yaml:"validation,omitempty"`
	StepNavigationCallback string `yaml:"navigation,omitempty"`
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

func (bs *BaseStep) Description() string {
	return bs.StepDescription
}

func (bs *BaseStep) SkipCondition() string {
	return bs.StepSkipCondition
}

// Callback implementations for BaseStep
func (bs *BaseStep) BeforeCallback() string {
	return bs.StepBeforeCallback
}

func (bs *BaseStep) AfterCallback() string {
	return bs.StepAfterCallback
}

func (bs *BaseStep) ValidationCallback() string {
	return bs.StepValidationCallback
}

func (bs *BaseStep) NavigationCallback() string {
	return bs.StepNavigationCallback
}

// Placeholder Execute for BaseStep - concrete types should override this.
func (bs *BaseStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	return nil, errors.Errorf("Execute not implemented for step type %s (ID: %s)", bs.Type(), bs.ID())
}

// --- Custom Unmarshalling Logic ---

// UnmarshalYAML implements the yaml.Unmarshaler interface for the Step interface.
// It determines the concrete step type based on the 'type' field and unmarshals
// into the corresponding struct type.
func UnmarshalStepYAML(node *yaml.Node) (Step, error) {
	// 1. Unmarshal into a temporary map to determine the type
	var typeFinder struct {
		Type string `yaml:"type"`
	}
	if err := node.Decode(&typeFinder); err != nil {
		return nil, errors.Wrap(err, "could not decode step type")
	}

	// 2. Based on the type, unmarshal into the specific struct
	var step Step
	switch typeFinder.Type {
	case "form":
		s := &FormStep{}
		err := node.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "could not decode FormStep (ID: %s)", s.ID())
		}
		s.StepType = typeFinder.Type // Ensure type is set
		step = s
	case "decision":
		s := &DecisionStep{}
		err := node.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "could not decode DecisionStep (ID: %s)", s.ID())
		}
		s.StepType = typeFinder.Type
		step = s
	case "action":
		s := &ActionStep{}
		err := node.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "could not decode ActionStep (ID: %s)", s.ID())
		}
		s.StepType = typeFinder.Type
		step = s
	case "info":
		s := &InfoStep{}
		err := node.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "could not decode InfoStep (ID: %s)", s.ID())
		}
		s.StepType = typeFinder.Type
		step = s
	case "summary":
		s := &SummaryStep{}
		err := node.Decode(s)
		if err != nil {
			return nil, errors.Wrapf(err, "could not decode SummaryStep (ID: %s)", s.ID())
		}
		s.StepType = typeFinder.Type
		step = s
	default:
		// Try decoding into BaseStep to get ID for error message if possible
		var base BaseStep
		_ = node.Decode(&base) // Ignore error here, just for context
		return nil, fmt.Errorf("unknown step type '%s' for step ID '%s'", typeFinder.Type, base.ID())
	}

	// Check if the decoded step is nil (could happen with empty YAML sections)
	if step == nil {
		return nil, errors.New("decoded step is nil, check YAML content and step type handlers")
	}

	return step, nil
}

// Custom Unmarshaller for Wizard.Steps
// We need this because []Step is a slice of interfaces.
func (w *WizardSteps) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.SequenceNode {
		return errors.New("steps field must be a sequence")
	}
	var steps []Step
	for _, stepNode := range node.Content {
		step, err := UnmarshalStepYAML(stepNode)
		if err != nil {
			// Try to get context from the node for better error reporting
			var base BaseStep
			_ = stepNode.Decode(&base) // Ignore error, just for ID/type context
			return errors.Wrapf(err, "failed to unmarshal step (approx ID: %s, type from context: %s)", base.ID(), base.Type())
		}
		if step == nil {
			// This case might occur if UnmarshalStepYAML returns nil, nil
			return errors.New("unmarshalled step is nil, check step definitions and UnmarshalStepYAML logic")
		}
		steps = append(steps, step)
	}
	*w = steps
	return nil
}

// Wrapper type for custom unmarshalling of []Step
type WizardSteps []Step
