package cmds

import (
	"context"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/uhoh/pkg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

// ExampleSettings represents the settings for the example command.
// Currently, it doesn't have any specific flags or arguments,
// but it's here for future extension or consistency.
type ExampleSettings struct{}

// ExampleCommand demonstrates running a simple hardcoded form.
type ExampleCommand struct {
	*cmds.CommandDescription
}

var _ cmds.BareCommand = &ExampleCommand{}

// NewExampleCommand creates a new instance of the ExampleCommand.
func NewExampleCommand() (*ExampleCommand, error) {
	return &ExampleCommand{
		CommandDescription: cmds.NewCommandDescription(
			"example",
			cmds.WithShort("Run an example form"),
			// No flags or arguments for this simple example
		),
	}, nil
}

// Run executes the example command.
func (c *ExampleCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	// The settings struct is currently empty, but we initialize it for consistency.
	s := &ExampleSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Hardcoded YAML for the example form
	yamlData := `
name: Burger Order Form
theme: Charm
groups:
  - name: Burger Selection
    fields:
      - type: select
        key: burger
        title: Choose your burger
        options:
          - label: Charmburger Classic
            value: classic
          - label: Chickwich
            value: chickwich
  - name: Order Details
    fields:
      - type: input
        key: name
        title: What's your name?
      - type: confirm
        key: discount
        title: Would you like 15% off?ggj
`

	var form pkg.Form
	err := yaml.Unmarshal([]byte(yamlData), &form)
	if err != nil {
		// Use log for internal errors, return error for CLI handling
		log.Error().Err(err).Msg("Error parsing example form YAML")
		return errors.Wrap(err, "error parsing example form YAML")
	}

	values, err := form.Run(ctx)
	if err != nil {
		// Don't fatal, return error
		return errors.Wrap(err, "error running example form")
	}

	fmt.Println("Form Results:")
	for key, value := range values {
		fmt.Printf("%s: %v\n", key, value)
	}

	return nil
}
