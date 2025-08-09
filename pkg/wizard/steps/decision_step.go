package steps

import (
	"context"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

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
	log.Debug().Str("stepId", ds.ID()).Msgf("--- Step: %s ---", ds.Title())

	if len(ds.Choices) == 0 {
		return nil, errors.New("decision step has no choices defined")
	}

	// Create options for the select field
	options := []huh.Option[string]{}
	for _, choice := range ds.Choices {
		options = append(options, huh.NewOption[string](choice, choice))
	}

	// Initialize the chosen value
	var chosenValue string

	// Display the title and description
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(ds.Title()).
				Description(ds.Description()).
				Options(options...).
				Value(&chosenValue),
		),
	)

	// Run the form
	err := form.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAborted
		}
		return nil, errors.Wrap(err, "error running decision form")
	}

	// Store the result in the state
	stepResult := map[string]interface{}{
		ds.TargetKey: chosenValue,
	}

	// Set the next step if specified in the map
	if ds.NextStepMap != nil {
		if nextStep, ok := ds.NextStepMap[chosenValue]; ok {
			// Store the next step in the base step
			ds.NextStep = nextStep
			log.Debug().Str("stepId", ds.ID()).Str("choice", chosenValue).Str("nextStep", nextStep).Msg("Next step determined from next_step_map")
		}
	}

	log.Debug().Str("stepId", ds.ID()).Str("targetKey", ds.TargetKey).Str("value", chosenValue).Msg("Decision made")
	return stepResult, nil
}

func (ds *DecisionStep) GetBaseStep() *BaseStep {
	return &ds.BaseStep
}
