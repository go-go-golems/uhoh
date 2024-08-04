package cmds

import (
	"context"
	"fmt"

	glazedcmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/uhoh/pkg"
	"gopkg.in/yaml.v2"
)

type UhohCommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	Layers    []layers.ParameterLayer           `yaml:"layers,omitempty"`

	Form *pkg.Form `yaml:"form"`
}

type UhohCommand struct {
	*glazedcmds.CommandDescription `yaml:",inline"`
	Form                           *pkg.Form `yaml:"form"`
}

var _ glazedcmds.BareCommand = &UhohCommand{}

func NewUhohCommand(
	description *glazedcmds.CommandDescription,
	form *pkg.Form,
) (*UhohCommand, error) {
	return &UhohCommand{
		CommandDescription: description,
		Form:               form,
	}, nil
}

func (u *UhohCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	// Run the form
	results, err := u.Form.Run()
	if err != nil {
		return err
	}

	// Print out results as yaml
	yamlData, err := yaml.Marshal(results)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlData))
	return nil
}
