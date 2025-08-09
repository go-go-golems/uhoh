# Integrating Uhoh Forms into Standard Bubble Tea Applications

Date: 2025-08-09

## Purpose and scope

This document explains how to embed Uhoh YAML-defined forms as Bubble Tea models inside a larger TUI application. We add a small, focused API to construct a `huh.Form` (which is a `tea.Model`) from the Uhoh DSL without running it, and a helper to extract final values when the form completes under the parent app’s update loop. This enables composition with existing Bubble Tea state machines, routing, and layout.

## What we added

- New public helpers in `uhoh/pkg/formdsl.go`:
  - `(*Form).BuildBubbleTeaModel() (*huh.Form, map[string]interface{}, error)`
  - `BuildBubbleTeaModelFromYAML(src []byte) (*huh.Form, map[string]interface{}, error)`
  - `ExtractFinalValues(values map[string]interface{}) (map[string]interface{}, error)`

These build a fully-configured `*huh.Form` and return an internal `values` map that keeps pointers to bound variables. You run the returned `huh.Form` inside your Bubble Tea app. Once it reaches `StateCompleted`, call `ExtractFinalValues(values)` to get the final plain `map[string]any`.

## Why this works (Huh + Tea)

`huh.Form` already implements `tea.Model` and is designed to be embedded (see `huh/README.md` Bubble Tea example). Uhoh’s DSL compiles to a `huh.Form`, so exposing a builder that returns the underlying form is sufficient for integration.

## API reference (added)

File: `uhoh/pkg/formdsl.go`

1) `(*Form).BuildBubbleTeaModel()`
   - Builds the internal value pointers, groups and fields, theme, and returns a `*huh.Form` and `values` map.
   - Does not run the form.

2) `BuildBubbleTeaModelFromYAML(src []byte)`
   - Convenience to `yaml.Unmarshal` a Uhoh `Form` then call `BuildBubbleTeaModel()`.

3) `ExtractFinalValues(values)`
   - Converts the internal pointer map into a plain `map[string]any` after the form has completed.

## Usage examples

### A. Embed a YAML-defined form in your Bubble Tea model

```go
type parentModel struct {
    form      *huh.Form
    formVals  map[string]interface{}
    done      bool
    result    map[string]interface{}
}

func NewParentModel(yamlBytes []byte) (parentModel, error) {
    form, vals, err := uhoh.BuildBubbleTeaModelFromYAML(yamlBytes)
    if err != nil {
        return parentModel{}, err
    }
    return parentModel{form: form, formVals: vals}, nil
}

func (m parentModel) Init() tea.Cmd { return m.form.Init() }

func (m parentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    fm, cmd := m.form.Update(msg)
    if f, ok := fm.(*huh.Form); ok {
        m.form = f
    }
    if m.form.State == huh.StateCompleted && !m.done {
        res, err := uhoh.ExtractFinalValues(m.formVals)
        if err == nil {
            m.result = res
            m.done = true
        }
    }
    return m, cmd
}

func (m parentModel) View() string {
    if m.done {
        return fmt.Sprintf("Done! Values: %v\n", m.result)
    }
    return m.form.View()
}
```

Notes:
- The parent model owns the `huh.Form` as a child `tea.Model`.
- `ExtractFinalValues` should be called only when `m.form.State == huh.StateCompleted`.
- You can switch views or push to another route once done.

### B. Build from a pre-parsed DSL `Form`

```go
var f uhoh.Form // filled programmatically or via yaml.Unmarshal
form, vals, err := f.BuildBubbleTeaModel()
```

### C. Themed and accessible forms

Uhoh passes `Form.Theme` to `huh`. You can also enable accessible mode at the caller level by wiring environment-based flags and passing them to your parent app’s runtime (see `huh/README.md` for `WithAccessible`). If needed, we can add an `Accessible` boolean to Uhoh’s DSL later.

## Edge cases and limitations

- Validation: `addValidation` is currently a stub; warning logs are emitted when validations are present. This does not block embedding; fields still work. Future work: implement the predicate compilation.
- `note` fields have no values; they’re tracked with `nil` in `values` and skipped by `ExtractFinalValues`.
- Type assertions for options assume string values in `createOptions`. If you plan to use non-string underlying values for selects, we should extend the DSL or mapping logic accordingly.

## How this integrates with Wizards

- Wizards (`uhoh/pkg/wizard`) execute steps sequentially using `Form.Run(ctx)`. For embedding in Tea, you generally embed a single form. If you need multi-step flows inside a parent app, wrap a stepper model around repeated `BuildBubbleTeaModel()` constructions, or consider an adapter that returns a `tea.Model` for a whole wizard in the future.

## Key source references

```1:120:uhoh/pkg/formdsl.go
type Form struct {
    Name   string   `yaml:"name,omitempty"`
    Theme  string   `yaml:"theme,omitempty"`
    Groups []*Group `yaml:"groups"`
}
```

```110:173:uhoh/pkg/formdsl.go
// BuildBubbleTeaModel ...
func (f *Form) BuildBubbleTeaModel() (*huh.Form, map[string]interface{}, error) { /* ... */ }
```

```376:385:uhoh/pkg/formdsl.go
func BuildBubbleTeaModelFromYAML(src []byte) (*huh.Form, map[string]interface{}, error) { /* ... */ }
```

```349:374:uhoh/pkg/formdsl.go
func ExtractFinalValues(values map[string]interface{}) (map[string]interface{}, error) { /* ... */ }
```

## Next steps

- Implement validation plumbing in `addValidation`.
- Optional DSL flags for `accessible` and per-form behaviors (timeouts, next labels at group level).
- Add a small `examples/bubbletea-embed` under `uhoh` to demonstrate routing and layout with a parent app.


