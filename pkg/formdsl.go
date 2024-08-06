package pkg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
)

type Form struct {
	Name   string   `yaml:"name,omitempty"`
	Theme  string   `yaml:"theme,omitempty"`
	Groups []*Group `yaml:"groups"`
}

type Group struct {
	Name   string   `yaml:"name,omitempty"`
	Fields []*Field `yaml:"fields"`
}

type Field struct {
	Type                  string        `yaml:"type"`
	Key                   string        `yaml:"key,omitempty"`
	Title                 string        `yaml:"title,omitempty"`
	Description           string        `yaml:"description,omitempty"`
	Value                 interface{}   `yaml:"value,omitempty"`
	Options               []*Option     `yaml:"options,omitempty"`
	Validation            []*Validation `yaml:"validation,omitempty"`
	InputAttributes       *InputAttributes
	TextAttributes        *TextAttributes
	SelectAttributes      *SelectAttributes
	MultiSelectAttributes *MultiSelectAttributes
	ConfirmAttributes     *ConfirmAttributes
	NoteAttributes        *NoteAttributes
	FilePickerAttributes  *FilePickerAttributes
}

type Option struct {
	Label string      `yaml:"label"`
	Value interface{} `yaml:"value"`
}

type Validation struct {
	Condition string `yaml:"condition"`
	Error     string `yaml:"error"`
}

type InputAttributes struct {
	Prompt      string `yaml:"prompt,omitempty"`
	CharLimit   int    `yaml:"char_limit,omitempty"`
	Placeholder string `yaml:"placeholder,omitempty"`
	EchoMode    string `yaml:"echo_mode,omitempty"`
}

type TextAttributes struct {
	Lines           int      `yaml:"lines,omitempty"`
	CharLimit       int      `yaml:"char_limit,omitempty"`
	ShowLineNumbers bool     `yaml:"show_line_numbers,omitempty"`
	Placeholder     string   `yaml:"placeholder,omitempty"`
	Editor          string   `yaml:"editor,omitempty"`
	EditorArgs      []string `yaml:"editor_args,omitempty"`
	EditorExtension string   `yaml:"editor_extension,omitempty"`
}

type SelectAttributes struct {
	Inline     bool `yaml:"inline,omitempty"`
	Height     int  `yaml:"height,omitempty"`
	Filterable bool `yaml:"filterable,omitempty"`
}

type MultiSelectAttributes struct {
	Limit      int  `yaml:"limit,omitempty"`
	Height     int  `yaml:"height,omitempty"`
	Filterable bool `yaml:"filterable,omitempty"`
}

type ConfirmAttributes struct {
	Affirmative string `yaml:"affirmative,omitempty"`
	Negative    string `yaml:"negative,omitempty"`
}

type NoteAttributes struct {
	Height         int    `yaml:"height,omitempty"`
	ShowNextButton bool   `yaml:"show_next_button,omitempty"`
	NextLabel      string `yaml:"next_label,omitempty"`
}

type FilePickerAttributes struct {
	CurrentDirectory string   `yaml:"current_directory,omitempty"`
	ShowHidden       bool     `yaml:"show_hidden,omitempty"`
	ShowSize         bool     `yaml:"show_size,omitempty"`
	ShowPermissions  bool     `yaml:"show_permissions,omitempty"`
	FileAllowed      bool     `yaml:"file_allowed,omitempty"`
	DirAllowed       bool     `yaml:"dir_allowed,omitempty"`
	AllowedTypes     []string `yaml:"allowed_types,omitempty"`
	Height           int      `yaml:"height,omitempty"`
}

type FieldWithValidation interface {
	Validate(func(string) error) huh.Field
}

func addValidation(field huh.Field, validations []*Validation) (huh.Field, error) {
	return nil, errors.New("not implemented")
}

// Run executes the form and returns a map of the input values and an error if any
func (f *Form) Run(ctx context.Context) (map[string]interface{}, error) {
	// Create a map to store the input values
	values := make(map[string]interface{})

	// Create huh Form groups
	var huhGroups []*huh.Group

	// Iterate through groups and fields to build the huh Form
	for _, group := range f.Groups {
		huhFields := make([]huh.Field, 0, len(group.Fields))

		for _, field := range group.Fields {
			// Initialize the value in the map
			if field.Value != nil {
				values[field.Key] = field.Value
			} else {
				values[field.Key] = getDefaultValue(field.Type)
			}

			// Create the appropriate huh field based on the type
			var huhField huh.Field
			switch field.Type {
			case "input":
				value := values[field.Key].(string)
				input := huh.NewInput().
					Title(field.Title).
					Value(&value)
				if field.InputAttributes != nil {
					if field.InputAttributes.Prompt != "" {
						input = input.Prompt(field.InputAttributes.Prompt)
					}
					if field.InputAttributes.CharLimit > 0 {
						input = input.CharLimit(field.InputAttributes.CharLimit)
					}
					if field.InputAttributes.Placeholder != "" {
						input = input.Placeholder(field.InputAttributes.Placeholder)
					}
					if field.InputAttributes.EchoMode != "" {
						switch field.InputAttributes.EchoMode {
						case "password":
							input = input.EchoMode(huh.EchoModePassword)
						case "none":
							input = input.EchoMode(huh.EchoModeNone)
						}
					}
				}
				huhField = input
				values[field.Key] = &value
			case "text":
				value := values[field.Key].(string)
				text := huh.NewText().
					Title(field.Title).
					Value(&value)
				if field.TextAttributes != nil {
					if field.TextAttributes.Lines > 0 {
						text = text.Lines(field.TextAttributes.Lines)
					}
					if field.TextAttributes.CharLimit > 0 {
						text = text.CharLimit(field.TextAttributes.CharLimit)
					}
					text = text.ShowLineNumbers(field.TextAttributes.ShowLineNumbers)
					if field.TextAttributes.Placeholder != "" {
						text = text.Placeholder(field.TextAttributes.Placeholder)
					}
					if field.TextAttributes.Editor != "" {
						text = text.Editor(field.TextAttributes.Editor)
					}
					if len(field.TextAttributes.EditorArgs) > 0 {
						text = text.EditorExtension(strings.Join(field.TextAttributes.EditorArgs, " "))
					}
					if field.TextAttributes.EditorExtension != "" {
						text = text.EditorExtension(field.TextAttributes.EditorExtension)
					}
				}
				huhField = text
				values[field.Key] = &value

			case "select":
				value := values[field.Key].(string)
				select_ := huh.NewSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(&value)
				if field.SelectAttributes != nil {
					select_ = select_.Inline(field.SelectAttributes.Inline)
					if field.SelectAttributes.Height > 0 {
						select_ = select_.Height(field.SelectAttributes.Height)
					}
					select_ = select_.Filtering(field.SelectAttributes.Filterable)
				}
				huhField = select_
				values[field.Key] = &value

			case "multiselect":
				value := values[field.Key].([]string)
				multiSelect := huh.NewMultiSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(&value)
				if field.MultiSelectAttributes != nil {
					if field.MultiSelectAttributes.Limit > 0 {
						multiSelect = multiSelect.Limit(field.MultiSelectAttributes.Limit)
					}
					if field.MultiSelectAttributes.Height > 0 {
						multiSelect = multiSelect.Height(field.MultiSelectAttributes.Height)
					}
					multiSelect = multiSelect.Filterable(field.MultiSelectAttributes.Filterable)
				}
				huhField = multiSelect
				values[field.Key] = &value

			case "confirm":
				value := values[field.Key].(bool)
				confirm := huh.NewConfirm().
					Title(field.Title).
					Value(&value)
				if field.ConfirmAttributes != nil {
					if field.ConfirmAttributes.Affirmative != "" {
						confirm = confirm.Affirmative(field.ConfirmAttributes.Affirmative)
					}
					if field.ConfirmAttributes.Negative != "" {
						confirm = confirm.Negative(field.ConfirmAttributes.Negative)
					}
				}
				huhField = confirm
				values[field.Key] = &value

			case "note":
				note := huh.NewNote().
					Title(field.Title).
					Description(field.Description)
				if field.NoteAttributes != nil {
					if field.NoteAttributes.Height > 0 {
						note = note.Height(field.NoteAttributes.Height)
					}
					if field.NoteAttributes.ShowNextButton {
						note = note.Next(true)
						if field.NoteAttributes.NextLabel != "" {
							note = note.NextLabel(field.NoteAttributes.NextLabel)
						}
					}
				}
				huhField = note

			case "filepicker":
				value := values[field.Key].(string)
				filePicker := huh.NewFilePicker().
					Title(field.Title).
					Value(&value)
				if field.FilePickerAttributes != nil {
					if field.FilePickerAttributes.CurrentDirectory != "" {
						filePicker = filePicker.CurrentDirectory(field.FilePickerAttributes.CurrentDirectory)
					}
					filePicker = filePicker.
						ShowHidden(field.FilePickerAttributes.ShowHidden).
						ShowSize(field.FilePickerAttributes.ShowSize).
						ShowPermissions(field.FilePickerAttributes.ShowPermissions).
						FileAllowed(field.FilePickerAttributes.FileAllowed).
						DirAllowed(field.FilePickerAttributes.DirAllowed)
					if len(field.FilePickerAttributes.AllowedTypes) > 0 {
						filePicker = filePicker.AllowedTypes(field.FilePickerAttributes.AllowedTypes)
					}
					if field.FilePickerAttributes.Height > 0 {
						filePicker = filePicker.Height(field.FilePickerAttributes.Height)
					}
				}
				huhField = filePicker
				values[field.Key] = &value

			default:
				return nil, fmt.Errorf("unsupported field type: %s", field.Type)
			}

			// Add validation if specified
			if len(field.Validation) > 0 {
				var err error
				huhField, err = addValidation(huhField, field.Validation)
				if err != nil {
					return nil, fmt.Errorf("validation error for field %s: %w", field.Key, err)
				}
			}

			huhFields = append(huhFields, huhField)
		}

		if len(huhFields) == 0 {
			return nil, fmt.Errorf("no fields found in group %s", group.Name)
		}

		// Create the huh Group
		huhGroup := huh.NewGroup(huhFields...)
		huhGroups = append(huhGroups, huhGroup)
	}

	// Create the huh Form
	huhForm := huh.NewForm(huhGroups...)

	// Set the theme if specified
	if f.Theme != "" {
		theme, err := getTheme(f.Theme)
		if err != nil {
			return nil, err
		}
		huhForm = huhForm.WithTheme(theme)
	}

	// Run the form
	err := huhForm.RunWithContext(ctx)
	if err != nil {
		return nil, err
	}

	// Extract final values from the map
	finalValues := make(map[string]interface{})
	for key, value := range values {
		switch v := value.(type) {
		case *string:
			finalValues[key] = *v
		case *[]string:
			finalValues[key] = *v
		case *bool:
			finalValues[key] = *v
		default:
			finalValues[key] = v
		}
	}

	return finalValues, nil
}

// Helper function to get the default value based on field type
func getDefaultValue(fieldType string) interface{} {
	switch fieldType {
	case "input", "text":
		return ""
	case "select":
		return ""
	case "multiselect":
		return []string{}
	case "confirm":
		return false
	case "note":
		return ""
	case "filepicker":
		return ""
	default:
		return nil
	}
}

// Helper function to create huh options from our Option structs
func createOptions(options []*Option) []huh.Option[string] {
	var huhOptions []huh.Option[string]
	for _, opt := range options {
		huhOptions = append(huhOptions, huh.NewOption(opt.Label, opt.Value.(string)))
	}
	return huhOptions
}

// Helper function to get the huh theme based on the theme name
func getTheme(themeName string) (*huh.Theme, error) {
	switch themeName {
	case "Charm":
		return huh.ThemeCharm(), nil
	case "Dracula":
		return huh.ThemeDracula(), nil
	case "Catppuccin":
		return huh.ThemeCatppuccin(), nil
	case "Base16":
		return huh.ThemeBase16(), nil
	case "Default":
		return huh.ThemeBase(), nil
	default:
		return nil, fmt.Errorf("unsupported theme: %s", themeName)
	}
}
