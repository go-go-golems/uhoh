package cmds

import (
	"fmt"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"io"
	"io/fs"
	"strings"

	"github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/uhoh/pkg"
	"gopkg.in/yaml.v3"
)

type UhohCommandLoader struct{}

func (u *UhohCommandLoader) IsFileSupported(f fs.FS, fileName string) bool {
	return strings.HasSuffix(fileName, ".yaml") || strings.HasSuffix(fileName, ".yml")
}

var _ loaders.CommandLoader = (*UhohCommandLoader)(nil)

// Add this new type
type fieldWithRawAttributes struct {
	Type        string            `yaml:"type"`
	Key         string            `yaml:"key,omitempty"`
	Title       string            `yaml:"title,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Value       interface{}       `yaml:"value,omitempty"`
	Options     []*pkg.Option     `yaml:"options,omitempty"`
	Validation  []*pkg.Validation `yaml:"validation,omitempty"`
	Attributes  yaml.Node         `yaml:"attributes,omitempty"`
}

type UhohCommandDescription struct {
	Name      string                            `yaml:"name"`
	Short     string                            `yaml:"short"`
	Long      string                            `yaml:"long,omitempty"`
	Flags     []*parameters.ParameterDefinition `yaml:"flags,omitempty"`
	Arguments []*parameters.ParameterDefinition `yaml:"arguments,omitempty"`
	Layers    []layers.ParameterLayer           `yaml:"layers,omitempty"`
	Form      struct {
		Name  string `yaml:"name,omitempty"`
		Theme string `yaml:"theme,omitempty"`

		Groups []struct {
			Name   string                   `yaml:"name,omitempty"`
			Fields []fieldWithRawAttributes `yaml:"fields"`
		} `yaml:"groups"`
	} `yaml:"form"`
}

// Modify the LoadUhohCommandFromReader function
func (u *UhohCommandLoader) LoadUhohCommandFromReader(
	s io.Reader,
	options []cmds.CommandDescriptionOption,
	_ []alias.Option,
) ([]cmds.Command, error) {
	yamlContent, err := io.ReadAll(s)
	if err != nil {
		return nil, err
	}

	ucd := UhohCommandDescription{}

	err = yaml.Unmarshal(yamlContent, &ucd)
	if err != nil {
		return nil, err
	}

	form := &pkg.Form{
		Name:   ucd.Form.Name,
		Theme:  ucd.Form.Theme,
		Groups: make([]*pkg.Group, len(ucd.Form.Groups)),
	}

	if len(ucd.Form.Groups) == 0 {
		return nil, fmt.Errorf("no groups found in form %s", ucd.Form.Name)
	}

	// Process the fields and convert the raw attributes to the correct type
	for i, group := range ucd.Form.Groups {
		form.Groups[i] = &pkg.Group{
			Name:   group.Name,
			Fields: make([]*pkg.Field, len(group.Fields)),
		}
		if len(group.Fields) == 0 {
			return nil, fmt.Errorf("no fields found in group %s", group.Name)
		}
		for j, field := range group.Fields {
			processedField, err := processField(field)
			if err != nil {
				return nil, err
			}
			form.Groups[i].Fields[j] = &processedField
		}
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

	uc, err := NewUhohCommand(description, form)
	if err != nil {
		return nil, err
	}

	for _, option := range options {
		option(uc.Description())
	}

	return []cmds.Command{uc}, nil
}

// Add this new function
func processField(field fieldWithRawAttributes) (pkg.Field, error) {
	processedField := pkg.Field{
		Type:        field.Type,
		Key:         field.Key,
		Title:       field.Title,
		Description: field.Description,
		Value:       field.Value,
		Options:     field.Options,
		Validation:  field.Validation,
	}

	switch field.Type {
	case "input":
		var attrs pkg.InputAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.InputAttributes = &attrs
	case "text":
		var attrs pkg.TextAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.TextAttributes = &attrs
	case "select":
		var attrs pkg.SelectAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.SelectAttributes = &attrs
	case "multiselect":
		var attrs pkg.MultiSelectAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.MultiSelectAttributes = &attrs
	case "confirm":
		var attrs pkg.ConfirmAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.ConfirmAttributes = &attrs
	case "note":
		var attrs pkg.NoteAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.NoteAttributes = &attrs
	case "filepicker":
		var attrs pkg.FilePickerAttributes
		if err := field.Attributes.Decode(&attrs); err != nil {
			return pkg.Field{}, err
		}
		processedField.FilePickerAttributes = &attrs
	}

	return processedField, nil
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
		u.LoadUhohCommandFromReader,
		options,
		aliasOptions)
}
