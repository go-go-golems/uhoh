package cmds

import (
	"context"
	"fmt"
	"os"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/uhoh/pkg/wizard"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

type RunWizardSettings struct {
	WizardFile       string                 `glazed.parameter:"wizard-file"`
	InitialState     map[string]string      `glazed.parameter:"initial-state"`
	InitialStateFile map[string]interface{} `glazed.parameter:"initial-state-file"`
}

type RunWizardCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &RunWizardCommand{}

func NewRunWizardCommand() (*RunWizardCommand, error) {
	return &RunWizardCommand{
		CommandDescription: cmds.NewCommandDescription(
			"run-wizard",
			cmds.WithShort("Run a wizard defined in a YAML file"),
			cmds.WithArguments(
				parameters.NewParameterDefinition(
					"wizard-file",
					parameters.ParameterTypeString, // Assuming filename for now
					parameters.WithHelp("Path to the wizard YAML file"),
					parameters.WithRequired(true),
				),
			),
			cmds.WithFlags(
				parameters.NewParameterDefinition(
					"initial-state",
					parameters.ParameterTypeKeyValue,
					parameters.WithHelp("Initial key-value state for the wizard"),
				),
				parameters.NewParameterDefinition(
					"initial-state-file",
					parameters.ParameterTypeObjectFromFile,
					parameters.WithHelp("File containing initial state for the wizard (JSON/YAML)"),
				),
			),
		),
	}, nil
}

func (c *RunWizardCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &RunWizardSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Prepare initial state
	initialState := make(map[string]interface{})

	// Merge key-value state, potentially overwriting file state
	if len(s.InitialState) > 0 {
		log.Debug().Msg("Merging initial state from flags")
		for k, v := range s.InitialState {
			initialState[k] = v // Overwrite if key exists
		}
	}

	if len(s.InitialStateFile) > 0 {
		log.Debug().Msg("Merging initial state from file")
		for k, v := range s.InitialStateFile {
			initialState[k] = v // Overwrite if key exists
		}
	}

	// Load the wizard using LoadWizard, applying the prepared initial state
	// NOTE: LoadWizard itself applies options; we pass the state here to Run.
	wz, err := wizard.LoadWizard(s.WizardFile, wizard.WithInitialState(initialState))
	if err != nil {
		return errors.Wrapf(err, "error loading wizard from file: %s", s.WizardFile)
	}

	// Run the wizard with the combined initial state
	finalState, err := wz.Run(ctx, map[string]interface{}{})
	if err != nil {
		// Don't fatal, return error for main handler
		return errors.Wrap(err, "error running wizard")
	}

	// Print the results
	fmt.Println("\nWizard Results:")
	if len(finalState) == 0 {
		fmt.Println("(No data collected)")
	} else {
		// TODO(manuel, 2024-08-05) Use glazed output processing here?
		// For now, simple print using YAML encoder for better readability
		encoder := yaml.NewEncoder(os.Stdout)
		encoder.SetIndent(2)
		err = encoder.Encode(finalState)
		if err != nil {
			// Log the encoding error but don't necessarily fail the command
			log.Error().Err(err).Msg("Error encoding final state to YAML")
			// Optionally return the error if output is critical:
			// return errors.Wrap(err, "error encoding final state to YAML")
		}
	}

	return nil
}
