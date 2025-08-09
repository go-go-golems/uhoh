---
Title: Uhoh Wizards — Format and Usage
Slug: uhoh-wizards
Short: Define multi-step interactive flows (wizards) and run them from the CLI or Go
Topics:
- uhoh
- wizards
- dsl
- cli
- go
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Uhoh Wizards — Format and Usage

## Overview

Wizards let you compose multi-step interactive flows that can include forms, informational notes, branching decisions, actions, and a final summary. You define a wizard in YAML and execute it from the CLI or programmatically from Go. The wizard accumulates state across steps and returns a final map of values.

## Quickstart

Run one of the bundled examples and follow the prompts.

```bash
# Simple multi-step example
uhoh run-wizard ./cmd/uhoh/examples/wizard/simple-multi-step.yaml

# Or a basic single-form wizard
uhoh run-wizard ./cmd/uhoh/examples/wizard/basic-form.yaml
```

Pass initial state to pre-fill values or drive conditions:

```bash
uhoh run-wizard ./cmd/uhoh/examples/wizard/conditional-steps.yaml \
  --initial-state speed=fast featureX=true
```

Expected behavior:
- Steps render one after another in the terminal
- State is accumulated across steps
- Final state is printed as YAML

## Wizard format (DSL)

A wizard file declares steps of various types. Each step can read and write state. Conditions allow skipping or branching.

For the full specification and many examples, see:

- glaze help uhoh-wizard-dsl

Minimal example:

```yaml
name: Simple Wizard
steps:
  - type: form
    title: Your info
    form:
      groups:
        - fields:
            - type: input
              key: name
              title: Your name
  - type: summary
    title: Summary
    keys:
      - name
```

Key step types you can use:
- `form`: renders a form using the Uhoh form DSL
- `info`: shows text, optional next label
- `decision`: choose a path based on a question or expression
- `action`: run a callback (e.g., compute derived fields)
- `summary`: display selected state at the end

See rich examples under the repository samples:
- [cmd/uhoh/examples/wizard/](file:///home/manuel/workspaces/2025-08-03/use-inference-api-for-pinocchio/uhoh/cmd/uhoh/examples/wizard)

## Running a wizard from the CLI

Use the built-in command `run-wizard` and pass the YAML file path. You can inject initial state from flags or a file.

```bash
uhoh run-wizard ./cmd/uhoh/examples/wizard/simple-multi-step.yaml
```

With initial state:

```bash
# Key-value pairs
uhoh run-wizard ./wizard.yaml --initial-state name=Alice env=prod

# From a JSON/YAML file
uhoh run-wizard ./wizard.yaml --initial-state-file ./state.yaml
```

What happens:
- The wizard is loaded with `wizard.LoadWizard`
- It runs interactively, updating a shared state map
- On completion, the final state is printed as YAML

Implementation reference:
- [`RunWizardCommand`](file:///home/manuel/workspaces/2025-08-03/use-inference-api-for-pinocchio/uhoh/cmd/uhoh/cmds/run_wizard.go#L29-L118)

## Form step schemas

Form steps accept two schemas:

1) Full Uhoh Form DSL (recommended for advanced features)

```yaml
steps:
  - id: user-details
    type: form
    title: Your Details
    form:
      groups:
        - fields:
            - type: input
              key: name
              title: Name
              # ... full DSL supports: description, options, validation, attributes, etc.
            - type: input
              key: email
              title: Email
```

2) Simplified schema under `form.fields` (convenient for small wizards)

```yaml
steps:
  - id: user-details
    type: form
    title: Your Details
    form:
      fields:
        - name: name
          label: Name
          type: text
        - name: email
          label: Email
          type: email
```

Mapping rules for the simplified schema:
- `name` maps to the form field `key`
- `label` maps to the form field `title`
- `type` maps as follows:
  - `text`, `email`, `input` → `input`
  - `confirm`, `bool` → `confirm`
- All simplified fields are wrapped into a single implicit group

Use the full DSL when you need advanced features like validation expressions, selections, multi-selects, file pickers, or detailed attributes.

## Running a wizard from Go

Use the `wizard` package to load and run a wizard programmatically.

```go
import (
    "context"
    "fmt"

    "github.com/go-go-golems/uhoh/pkg/wizard"
)

func runWizard(ctx context.Context, path string) error {
    wz, err := wizard.LoadWizard(path, wizard.WithInitialState(map[string]interface{}{
        "name": "Alice",
    }))
    if err != nil {
        return err
    }

    final, err := wz.Run(ctx, map[string]interface{}{})
    if err != nil {
        return err
    }
    fmt.Println(final)
    return nil
}
```

## Tips

- Start small: one or two `form` steps and a final `summary`.
- Use `decision` to branch based on previous answers.
- Reuse forms by keeping them as separate YAML and embedding under `form:` blocks in steps.
- Keep keys stable; the final state is a map keyed by field `key`.

## See also

- glaze help uhoh-dsl
- glaze help uhoh-wizard-dsl
