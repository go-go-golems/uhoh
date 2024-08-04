# UHOH - EASY TUI FORMS AND WIZARDS

Uhoh is a tool and framework to create and run interactive TUI (Text User Interface) forms and wizards based on YAML files. It provides a simple and declarative way to define complex forms with various field types, validation rules, and styling options.

The main abstraction presented is the concept of a form "command", based around the [glazed](https://github.com/go-go-golems/glazed) command abstraction.

Using this as a central abstraction, we can easily create interactive forms and wizards while being able to run them easily on the terminal.

## Installation

To install the `uhoh` command line tool with homebrew, run:

```bash
brew tap go-go-golems/go-go-go
brew install go-go-golems/go-go-go/uhoh
```

To install the `uhoh` command using apt-get, run:

```bash
echo "deb [trusted=yes] https://apt.fury.io/go-go-golems/ /" >> /etc/apt/sources.list.d/fury.list
apt-get update
apt-get install uhoh
```

To install using `yum`, run:

```bash
echo "
[fury]
name=Gemfury Private Repo
baseurl=https://yum.fury.io/go-go-golems/
enabled=1
gpgcheck=0
" >> /etc/yum.repos.d/fury.repo
yum install uhoh
```

To install using `go get`, run:

```bash
go get -u github.com/go-go-golems/uhoh/cmd/uhoh
```

Finally, install by downloading the binaries straight from [github](https://github.com/go-go-golems/uhoh/releases).

## Usage

Uhoh uses YAML files to define forms and wizards. Here's a simple example of a form definition:

```yaml
name: Simple Contact Form
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

You can run this form using the `uhoh` command:

```bash
uhoh run-command simple-contact-form.yaml
```

## YAML DSL for Form Creation

The Uhoh YAML DSL (Domain Specific Language) allows you to define forms with various field types, validation rules, and styling options. Here's a brief overview of the structure:

```yaml
name: Form Name
theme: Theme Name
groups:
  - name: Group Name
    fields:
      - type: field_type
        key: unique_key
        title: Field Title
        # Additional field-specific properties
```

Supported field types include:
- `input`: Single-line text input
- `text`: Multi-line text input
- `select`: Single-selection from a list of options
- `multiselect`: Multiple-selection from a list of options
- `confirm`: Yes/No confirmation
- `note`: Informational field
- `filepicker`: File selection field

See the [documentation for the DSL](pkg/doc/topics/uhoh-dsl.md) or by running:

```bash
uhoh help uhoh-dsl
```

## Examples

Here's an example of a more complex form for a snake information system:

```
name: snake-info
short: Snake Information Form
form:
  name: Snake Information Form
  theme: Default
  groups:
    - name: Basic Information
      fields:
        - type: input
          key: snake_name
          title: Snake Name
          attributes:
            placeholder: e.g., Slytherin
        - type: input
          key: species
          title: Species
          attributes:
            placeholder: e.g., Python regius
        - type: select
          key: venom_status
          title: Venom Status
          options:
            - label: Non-venomous
              value: non_venomous
            - label: Mildly venomous
              value: mildly_venomous
            - label: Highly venomous
              value: highly_venomous
```

## Contributing

This is GO GO GOLEMS playground, and GO GO GOLEMS don't accept contributions. 
The structure of the project may change as we go forward, but
the core concept of a declarative form structure will stay the same,
and as such, you should be reasonably safe writing YAMLs to be used with uhoh.
