package pkg

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
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
	// Create a map to store pointers to the input values
	values := make(map[string]interface{})

	// Create huh Form groups
	var huhGroups []*huh.Group

	// Iterate through groups and fields to build the huh Form
	for _, group := range f.Groups {
		huhFields := make([]huh.Field, 0, len(group.Fields))

		for _, field := range group.Fields {
			// Create the target variable and store its pointer in the map FIRST
			// Initialize with default or YAML value
			switch field.Type {
			case "input", "text", "select", "filepicker":
				var strValue string
				if field.Value != nil {
					// Attempt type assertion, handle potential errors if YAML value isn't a string
					if val, ok := field.Value.(string); ok {
						strValue = val
					} else {
						// Handle non-string default value for string fields - convert or error?
						// For now, convert using fmt.Sprintf, might need refinement
						strValue = fmt.Sprintf("%v", field.Value)
						// Consider logging a warning here
					}
				}
				values[field.Key] = &strValue // Store pointer
			case "multiselect":
				var strSliceValue []string
				if field.Value != nil {
					// Attempt type assertion for slice of interfaces, then convert each to string
					if valSlice, ok := field.Value.([]interface{}); ok {
						for _, item := range valSlice {
							strSliceValue = append(strSliceValue, fmt.Sprintf("%v", item))
						}
					} else if valStrSlice, ok := field.Value.([]string); ok {
						// Handle case where YAML already provided []string
						strSliceValue = valStrSlice
					} else {
						// Handle other unexpected types for multiselect default
						// For now, initialize empty or log warning
						fmt.Printf("Warning: Unexpected type for multiselect default value: %T\n", field.Value)
					}
				}
				values[field.Key] = &strSliceValue // Store pointer
			case "confirm":
				var boolValue bool
				if field.Value != nil {
					if val, ok := field.Value.(bool); ok {
						boolValue = val
					} else {
						// Handle non-bool default value for confirm field - maybe parse string?
						// For now, default to false if type is wrong
						fmt.Printf("Warning: Unexpected type for confirm default value: %T\n", field.Value)
					}
				}
				values[field.Key] = &boolValue // Store pointer
			case "note":
				// Notes don't store values, but add key to map to avoid nil checks later?
				values[field.Key] = nil // Or maybe an empty struct?
			default:
				return nil, fmt.Errorf("unsupported field type for value initialization: %s", field.Type)
			}

			// Create the appropriate huh field based on the type, passing the stored pointer
			var huhField huh.Field
			switch field.Type {
			case "input":
				// value := values[field.Key].(string) // Old way
				input := huh.NewInput().
					Title(field.Title).
					Value(values[field.Key].(*string)) // Pass stored pointer
				if field.InputAttributes != nil {
					// ... (rest of input attribute settings remain the same)
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
				// values[field.Key] = &value // REMOVE THIS

			case "text":
				// value := values[field.Key].(string) // Old way
				text := huh.NewText().
					Title(field.Title).
					Value(values[field.Key].(*string)) // Pass stored pointer
				if field.TextAttributes != nil {
					// ... (rest of text attribute settings remain the same)
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
				// values[field.Key] = &value // REMOVE THIS

			case "select":
				// value := values[field.Key].(string) // Old way
				select_ := huh.NewSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(values[field.Key].(*string)) // Pass stored pointer
				if field.SelectAttributes != nil {
					// ... (rest of select attribute settings remain the same)
					select_ = select_.Inline(field.SelectAttributes.Inline)
					if field.SelectAttributes.Height > 0 {
						select_ = select_.Height(field.SelectAttributes.Height)
					}
					select_ = select_.Filtering(field.SelectAttributes.Filterable)
				}
				huhField = select_
				// values[field.Key] = &value // REMOVE THIS

			case "multiselect":
				// value := values[field.Key].([]string) // Old way
				multiSelect := huh.NewMultiSelect[string]().
					Title(field.Title).
					Options(createOptions(field.Options)...).
					Value(values[field.Key].(*[]string)) // Pass stored pointer
				if field.MultiSelectAttributes != nil {
					// ... (rest of multiselect attribute settings remain the same)
					if field.MultiSelectAttributes.Limit > 0 {
						multiSelect = multiSelect.Limit(field.MultiSelectAttributes.Limit)
					}
					if field.MultiSelectAttributes.Height > 0 {
						multiSelect = multiSelect.Height(field.MultiSelectAttributes.Height)
					}
					multiSelect = multiSelect.Filterable(field.MultiSelectAttributes.Filterable)
				}
				huhField = multiSelect
				// values[field.Key] = &value // REMOVE THIS

			case "confirm":
				// value := values[field.Key].(bool) // Old way
				confirm := huh.NewConfirm().
					Title(field.Title).
					Value(values[field.Key].(*bool)) // Pass stored pointer
				if field.ConfirmAttributes != nil {
					// ... (rest of confirm attribute settings remain the same)
					if field.ConfirmAttributes.Affirmative != "" {
						confirm = confirm.Affirmative(field.ConfirmAttributes.Affirmative)
					}
					if field.ConfirmAttributes.Negative != "" {
						confirm = confirm.Negative(field.ConfirmAttributes.Negative)
					}
				}
				huhField = confirm
				// values[field.Key] = &value // REMOVE THIS

			case "note":
				note := huh.NewNote().
					Title(field.Title).
					Description(field.Description)
				if field.NoteAttributes != nil {
					// ... (rest of note attribute settings remain the same)
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
				// No value to store or remove

			case "filepicker":
				// value := values[field.Key].(string) // Old way
				filePicker := huh.NewFilePicker().
					Title(field.Title).
					Value(values[field.Key].(*string)) // Pass stored pointer
				if field.FilePickerAttributes != nil {
					// ... (rest of filepicker attribute settings remain the same)
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
				// values[field.Key] = &value // REMOVE THIS

			default:
				// This case should ideally be caught by the initialization switch above
				return nil, fmt.Errorf("unsupported field type during huh field creation: %s", field.Type)
			}

			// Add validation if specified
			// ... (validation logic remains the same) ...
			if len(field.Validation) > 0 {
				var err error
				huhField, err = addValidation(huhField, field.Validation)
				if err != nil {
					// validation is not implemented yet
					// return nil, fmt.Errorf("validation error for field %s: %w", field.Key, err)
					fmt.Printf("Warning: Validation not yet implemented for field %s\n", field.Key)
				}
			}

			huhFields = append(huhFields, huhField)
		} // End field loop

		if len(huhFields) == 0 {
			// Allow groups with no fields (e.g., for layout purposes if huh supports it?)
			// Or return error? Let's allow it for now.
			// fmt.Printf("Warning: Group '%s' has no fields.\n", group.Name)
			continue // Skip creating an empty huh.Group
		}

		// Create the huh Group
		huhGroup := huh.NewGroup(huhFields...)
		// TODO: Add Group Title/Name handling if huh supports it directly on groups
		huhGroups = append(huhGroups, huhGroup)
	} // End group loop

	// Check if there are any groups to run
	if len(huhGroups) == 0 {
		fmt.Println("Warning: No interactive fields found in the form.")
		return make(map[string]interface{}), nil // Return empty results if no groups/fields
	}

	// Create the huh Form
	huhForm := huh.NewForm(huhGroups...)

	// Set the theme if specified
	// ... (theme logic remains the same) ...
	if f.Theme != "" {
		theme, err := getTheme(f.Theme)
		if err != nil {
			return nil, err
		}
		huhForm = huhForm.WithTheme(theme)
	}

	fmt.Println("--- Running Form ---") // Debug statement
	// Run the form
	err := huhForm.RunWithContext(ctx)
	if err != nil {
		// Check for specific errors like Abort
		if errors.Is(err, huh.ErrUserAborted) {
			fmt.Println("Form aborted by user.")
			// Return a specific error or nil with partial results?
			// For now, return the error.
			return nil, errors.Wrap(err, "form aborted")
		}
		return nil, errors.Wrap(err, "error running huh form")
	}
	fmt.Println("--- Form Finished ---") // Debug statement

	// Extract final values from the pointers in the map
	finalValues := make(map[string]interface{})
	for key, valuePtr := range values {
		if valuePtr == nil { // Handle note fields or potential issues
			continue
		}
		switch p := valuePtr.(type) {
		case *string:
			finalValues[key] = *p
		case *[]string:
			// Ensure we don't return nil slice if it was empty
			if *p == nil {
				finalValues[key] = []string{}
			} else {
				finalValues[key] = *p
			}
		case *bool:
			finalValues[key] = *p
		default:
			// Should not happen with current types, but handle defensively
			return nil, fmt.Errorf("unexpected pointer type in results map for key '%s': %T", key, p)
		}
	}

	return finalValues, nil
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
