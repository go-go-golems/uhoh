# YAML DSL for Form Creation

This document describes a YAML-based Domain Specific Language (DSL) for creating interactive forms. The DSL allows you to define forms with various field types, validation rules, and styling options.

## Top-Level Structure

The top-level structure of a form in the YAML DSL is as follows:

```yaml
name: string  # Optional name for the form
theme: string  # Optional theme (Charm, Dracula, Catppuccin, Base16, Default)
accessible: boolean  # Optional accessibility mode
groups:
  - name: string  # Optional group name
    fields:
      # List of fields (see Field Types section)
```

### Example:

```yaml
name: Customer Feedback Form
theme: Charm
accessible: true
groups:
  - name: Personal Information
    fields:
      # Fields will be listed here
  - name: Feedback
    fields:
      # Fields will be listed here
```

## Field Types

The DSL supports the following field types:

1. `input`: Single-line text input
2. `text`: Multi-line text input
3. `select`: Single-selection from a list of options
4. `multiselect`: Multiple-selection from a list of options
5. `confirm`: Yes/No confirmation

### Common Field Properties

All field types share these common properties:

```yaml
type: string  # Required: input, text, select, multiselect, confirm
key: string   # Required: unique identifier for the field
title: string # Optional: title/prompt for the field
value: any    # Optional: default value
validation:   # Optional: list of validation rules
  - condition: string
    error: string
```

### Field-Specific Properties

#### Input and Text

```yaml
attributes:
  prompt: string     # Optional: custom prompt
  char_limit: integer # Optional: character limit (text only)
```

#### Select and Multiselect

```yaml
options:  # Required: list of options
  - label: string
    value: any
attributes:
  limit: integer  # Optional: selection limit (multiselect only)
  height: integer # Optional: visible height of the selection list
```

#### Confirm

```yaml
attributes:
  affirmative: string # Optional: custom text for "Yes"
  negative: string    # Optional: custom text for "No"
```

## Examples

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

### Accessibility Example

```yaml
name: Accessibility Survey
theme: Default
accessible: true
groups:
  - name: User Experience
    fields:
      - type: select
        key: screen_reader
        title: Which screen reader do you use?
        options:
          - label: JAWS
            value: jaws
          - label: NVDA
            value: nvda
          - label: VoiceOver
            value: voiceover
          - label: Other
            value: other
      - type: text
        key: accessibility_feedback
        title: Please provide any accessibility feedback
        attributes:
          char_limit: 1000
```

## Usage Notes

1. Ensure that each field has a unique `key` within the form.
2. The `theme` option affects the visual appearance of the form. Choose from: Charm, Dracula, Catppuccin, Base16, or Default.
3. Set `accessible: true` to enable a mode optimized for screen readers.
4. The `value` property can be used to set default values for fields.
5. Group your fields logically to improve the user experience.
