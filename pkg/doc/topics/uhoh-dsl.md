---
Title: Uhoh Form DSL
Slug: uhoh-dsl
Short: Describe the uhoh DSL for form creation
Topics:
- dsl
- uhoh
Commands:
- help
Flags:
- dsl
- topic
- help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Enhanced Comprehensive Guide to Uhoh YAML DSL for Form Creation

## Introduction

The Uhoh YAML Domain Specific Language (DSL) is designed for creating interactive forms. It allows you to define forms with various field types, validation rules, and styling options. This guide provides a comprehensive overview of the DSL structure, field types, properties, and usage.

## Top-Level Structure

The Uhoh YAML DSL defines forms using a top-level structure that includes an optional form name, theme selection, accessibility mode, and groups of fields. Each group can have a name and contains a list of fields. This structure allows for logical organization of form elements and customization of the form's appearance and behavior.

The top-level structure of a form in the YAML DSL is as follows:

```yaml
name: string  # Optional name for the form
theme: string  # Optional theme (Charm, Dracula, Catppuccin, Base16, Default)
groups:
  - name: string  # Optional group name
    fields:
      # List of fields (see Field Types section)
```

### Example:

```yaml
name: Customer Feedback Form
theme: Charm
groups:
  - name: Personal Information
    fields:
      # Fields will be listed here
  - name: Feedback
    fields:
      # Fields will be listed here
```

## Field Types

The DSL supports seven field types: input (single-line text), text (multi-line text), select (single option from a list), multiselect (multiple options from a list), confirm (yes/no choice), note (informational text), and filepicker (file selection). Each field type has specific properties that allow for customization of its behavior and appearance, providing flexibility in form design.

1. `input`: Single-line text input
2. `text`: Multi-line text input
3. `select`: Single-selection from a list of options
4. `multiselect`: Multiple-selection from a list of options
5. `confirm`: Yes/No confirmation
6. `note`: Informational field
7. `filepicker`: File selection field

## Common Field Properties

All field types share a set of common properties, including the required 'type' and 'key', as well as optional 'title', 'description', 'value', and 'validation' properties. These common properties ensure consistency across different field types and allow for basic configuration of each field, including setting default values and implementing validation rules.

```yaml
type: string  # Required: input, text, select, multiselect, confirm, note, filepicker
key: string   # Required: unique identifier for the field
title: string # Optional: title/prompt for the field
description: string # Optional: description for the field
value: any    # Optional: default value
validation:   # Optional: list of validation rules
  - condition: string
    error: string
```

## Field-Specific Properties

Each field type has unique properties that cater to its specific functionality. These specific properties allow for fine-tuned control over each field's behavior and presentation.

### Input

The input field is used for collecting short, single-line text input from users. It's ideal for information like names, email addresses, or any brief textual data. The input field can be customized with a character limit, placeholder text, and even a special echo mode for password entry.

```yaml
type: input
# ... common properties ...
attributes:
    prompt: string     # Optional: custom prompt
    char_limit: integer # Optional: character limit
    placeholder: string # Optional: placeholder text
    echo_mode: string  # Optional: normal, password, none
```

### Text

The text field is designed for longer, multi-line text input. It's perfect for comments, descriptions, or any content that requires more space. This field type offers options like setting the number of visible lines, showing line numbers, and even specifying an external editor for more complex editing needs.

```yaml
type: text
# ... common properties ...
attributes:
    lines: integer     # Optional: number of lines to show
    char_limit: integer # Optional: character limit
    show_line_numbers: boolean # Optional: whether to show line numbers
    placeholder: string # Optional: placeholder text
    editor: string     # Optional: specify editor command
    editor_args: [string] # Optional: editor command arguments
    editor_extension: string # Optional: file extension for editor
```

### Select

The select field presents users with a list of options from which they can choose one item. It's great for situations where users need to pick from predefined choices, like selecting a country or a product category. The select field can be configured to display inline, have a scrollable height, and even be filterable for easier selection from long lists.

```yaml
type: select
# ... common properties ...
options:  # Required: list of options
  - label: string
    value: any
attributes:
    inline: boolean    # Optional: whether to display inline
    height: integer    # Optional: visible height of the selection list
    filterable: boolean # Optional: whether options are filterable
```

### MultiSelect

The multiselect field is similar to the select field, but it allows users to choose multiple options from the list. This is useful for scenarios like selecting multiple interests, features, or any situation where more than one choice is applicable. It can be configured with a selection limit and made filterable for user convenience.

```yaml
type: multiselect
# ... common properties ...
options:  # Required: list of options
  - label: string
    value: any
attributes:
    limit: integer     # Optional: selection limit
    height: integer    # Optional: visible height of the selection list
    filterable: boolean # Optional: whether options are filterable
```

### Confirm

The confirm field is a simple yes/no or true/false input. It's perfect for getting user agreement on terms, confirming actions, or any binary choice. The text for the affirmative and negative options can be customized to fit the specific context of the question.

```yaml
type: confirm
# ... common properties ...
attributes:
    affirmative: string # Optional: custom text for "Yes"
    negative: string    # Optional: custom text for "No"
```

### Note

The note field is not an input field, but rather a way to display information to the user within the form. It's ideal for providing instructions, explanations, or any additional context that might help users fill out the form correctly. The note can be configured with a custom height and an optional "Next" button for multi-step forms.

```yaml
type: note
# ... common properties ...
attributes:
    height: integer    # Optional: height of the note field
    show_next_button: boolean # Optional: whether to show a "Next" button
    next_label: string # Optional: custom label for the "Next" button
```

### FilePicker

The filepicker field provides a way for users to select files from their system. It's essential for forms that require file uploads or selection. This field type offers extensive customization options, including specifying allowed file types, showing or hiding certain file attributes, and controlling directory access.

```yaml
type: filepicker
# ... common properties ...
attributes:
    current_directory: string # Optional: initial directory
    show_hidden: boolean # Optional: whether to show hidden files
    show_size: boolean   # Optional: whether to show file sizes
    show_permissions: boolean # Optional: whether to show file permissions
    file_allowed: boolean # Optional: whether to allow file selection
    dir_allowed: boolean  # Optional: whether to allow directory selection
    allowed_types: [string] # Optional: list of allowed file extensions
    height: integer    # Optional: visible height of the file list
```

## Examples

The DSL supports creation of various form types, from simple contact forms to complex product order forms and file upload interfaces. Examples demonstrate how to combine different field types, set validation rules, and utilize field-specific properties to create functional and user-friendly forms. These examples serve as practical guides for implementing the DSL in real-world scenarios.

### Simple Contact Form

```yaml
name: Contact Form
theme: Default
groups:
  - name: Contact Information
    fields:
      - type: input
        key: name
        title: Your Name
      - type: input
        key: email
        title: Email Address
        validation:
          - condition: "!string.contains(value, '@')"
            error: Please enter a valid email address
      - type: text
        key: message
        title: Your Message
        attributes:
            char_limit: 500
            lines: 5
```

### Product Order Form

```yaml
name: Product Order
theme: Charm
groups:
  - name: Product Selection
    fields:
      - type: select
        key: product
        title: Choose a Product
        options:
          - label: Basic Widget
            value: basic
          - label: Premium Widget
            value: premium
          - label: Deluxe Widget
            value: deluxe
        attributes:
            filterable: true
      - type: input
        key: quantity
        title: Quantity
        validation:
          - condition: "parseInt(value) <= 0"
            error: Please enter a positive number
  - name: Additional Options
    fields:
      - type: multiselect
        key: addons
        title: Select Add-ons
        options:
          - label: Extended Warranty
            value: warranty
          - label: Gift Wrapping
            value: giftwrap
          - label: Express Shipping
            value: express
        attributes:
            limit: 2
      - type: confirm
        key: terms
        title: Do you accept the terms and conditions?
        attributes:
            affirmative: I Accept
            negative: I Do Not Accept
```

### File Upload Form

```yaml
name: File Upload
theme: Default
groups:
  - name: File Selection
    fields:
      - type: filepicker
        key: upload_file
        title: Select a file to upload
        attributes:
            current_directory: "/home/user/documents"
            allowed_types: [".pdf", ".doc", ".docx"]
            show_hidden: false
            show_size: true
            file_allowed: true
            dir_allowed: false
      - type: note
        title: Upload Instructions
        description: "Please select a PDF or Word document to upload. Maximum file size is 10MB."
        attributes:
            show_next_button: true
            next_label: Proceed to Upload
```

## Command Format for CLI Execution

The Uhoh DSL supports a command format that allows forms to be executed directly from the command line interface (CLI). This format includes additional top-level fields that provide metadata for the command.

The command format structure is as follows:

```yaml
name: string       # Required: Unique identifier for the command, kebab case lower case. This name will be used for the command on the CLI.
short: string      # Required: Short description of the command
form:              # Required: Form definition (see Top-Level Structure)
  # ... form definition ...
```

This command format allows you to execute forms from the CLI using the following syntax:

```bash
uhoh run-command ui.yaml [options]
```
