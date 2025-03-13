package cmds

import (
	"context"
	"fmt"

	glazedcmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/uhoh/pkg"
	"gopkg.in/yaml.v2"
)

type UhohCommand struct {
	*glazedcmds.CommandDescription `yaml:",inline"`
	Form                           *pkg.Form `yaml:"form"`
}

var _ glazedcmds.BareCommand = &UhohCommand{}

func NewUhohCommand(
	description *glazedcmds.CommandDescription,
	form *pkg.Form,
) (*UhohCommand, error) {
	// TODO(manuel, 2024-08-04) Probably here we should create a layer that has tweaks for all the fields so that they can be disabled / etc...
	return &UhohCommand{
		CommandDescription: description,
		Form:               form,
	}, nil
}

func (u *UhohCommand) Run(ctx context.Context, parsedLayers *layers.ParsedLayers) error {
	results, err := u.Form.Run(ctx)
	if err != nil {
		return err
	}

	yamlData, err := yaml.Marshal(results)
	if err != nil {
		return err
	}
	fmt.Println(string(yamlData))
	return nil
}
