package cmds

import (
	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"gopkg.in/yaml.v2"
	"io"
	"io/fs"
	"strings"
)

type UhohCommandLoader struct{}

func (u *UhohCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	return strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")
}

var _ loaders.CommandLoader = (*UhohCommandLoader)(nil)

func (u *UhohCommandLoader) loadUhohCommandFromReader(
	s io.Reader,
	options []cmds.CommandDescriptionOption,
	_ []alias.Option,
) ([]cmds.Command, error) {
	yamlContent, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}

	ucd := &UhohCommandDescription{}
	err = yaml.Unmarshal(yamlContent, ucd)
	if err != nil {
		return nil, err
	}

	options_ := []cmds.CommandDescriptionOption{
		cmds.WithShort(ucd.Short),
		cmds.WithLong(ucd.Long),
		cmds.WithFlags(ucd.Flags...),
		cmds.WithArguments(ucd.Arguments...),
		cmds.WithLayersList(ucd.Layers...),
	}

	description := cmds.NewCommandDescription(
		ucd.Name,
		options_...,
	)

	uc, err := NewUhohCommand(description, ucd.Form)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		option(uc.Description())
	}

	return []cmds.Command{uc}, nil
}

func (u *UhohCommandLoader) LoadCommands(
	f fs.FS, entryName string,
	options []cmds.CommandDescriptionOption,
	aliasOptions []alias.Option,
) ([]cmds.Command, error) {
	r, err := f.Open(entryName)
	if err != nil {
		return nil, err
	}
	defer func(r fs.File) {
		_ = r.Close()
	}(r)
	return loaders.LoadCommandOrAliasFromReader(
		r,
		u.loadUhohCommandFromReader,
		options,
		aliasOptions)
}
