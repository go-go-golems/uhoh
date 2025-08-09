---
Title: Use Uhoh Programmatically
Slug: uhoh-programmatic
Short: Call Uhoh from Go using a YAML string or a constructed DSL object
Topics:
- uhoh
- dsl
- api
- go
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Use Uhoh Programmatically

## Overview

Uhoh can be embedded in your Go programs. You can run a form by unmarshaling a YAML string into the DSL `Form` type or by constructing the DSL object directly. When executed, a form returns a `map[string]interface{}` with all collected values.

## Import

```go
import (
    "context"
    "fmt"

    uhoh "github.com/go-go-golems/uhoh/pkg"
    "gopkg.in/yaml.v3"
)
```

## Option A — From a YAML string (DSL form only)

Use this when your YAML contains the top-level `name`, `theme`, and `groups` keys of the form DSL.

```go
func runFormFromYAML(ctx context.Context, yamlSrc string) error {
    var f uhoh.Form
    if err := yaml.Unmarshal([]byte(yamlSrc), &f); err != nil {
        return err
    }

    values, err := f.Run(ctx)
    if err != nil {
        return err
    }

    // Use the collected values
    fmt.Println("values:", values)
    return nil
}
```

Example YAML string:

```yaml
name: Contact Form
groups:
  - name: Contact
    fields:
      - type: input
        key: name
        title: Your Name
      - type: input
        key: email
        title: Email
```

## Option B — From a YAML command file (CLI command format)

If your YAML follows the command format (`name`, `short`, `form: {...}`), you can load a ready-to-run command using the provided loader.

```go
import (
    glazedcmds "github.com/go-go-golems/glazed/pkg/cmds"
    "github.com/go-go-golems/uhoh/pkg/cmds"
    "io"
    "strings"
)

func runCommandYAML(ctx context.Context, yamlSrc string) error {
    r := io.NopCloser(strings.NewReader(yamlSrc))
    defer r.Close()

    // The loader expects an fs.FS + entry name. For a single reader, call the helper directly.
    loader := &cmds.UhohCommandLoader{}

    // Use the low-level reader API to parse one command from YAML
    cs, err := loader.LoadUhohCommandFromReader(
        strings.NewReader(yamlSrc),
        []glazedcmds.CommandDescriptionOption{},
        nil,
    )
    if err != nil {
        return err
    }

    if len(cs) != 1 {
        return fmt.Errorf("expected 1 command, got %d", len(cs))
    }

    // Run as a bare command (prints YAML of results to stdout)
    c := cs[0]
    return c.Run(ctx, nil)
}
```

Example command YAML:

```yaml
name: contact
short: Simple contact form
type: uhoh
form:
  name: Contact Form
  groups:
    - name: Contact
      fields:
        - type: input
          key: name
          title: Your Name
        - type: input
          key: email
          title: Email
```

## Option C — Construct the DSL object in Go

Create the form programmatically using the DSL structs and run it.

```go
func runConstructedForm(ctx context.Context) error {
    f := &uhoh.Form{
        Name:  "Survey",
        Theme: "Charm",
        Groups: []*uhoh.Group{
            {
                Name: "Basics",
                Fields: []*uhoh.Field{
                    { Type: "input", Key: "name", Title: "Your Name" },
                    { Type: "confirm", Key: "agree", Title: "Do you agree?" },
                },
            },
        },
    }

    values, err := f.Run(ctx)
    if err != nil {
        return err
    }
    fmt.Println("values:", values)
    return nil
}
```

## Notes and tips

- Themes: Supported values are `Charm`, `Dracula`, `Catppuccin`, `Base16`, and `Default`.
- Results: `Form.Run` returns a `map[string]interface{}` (strings, bools, and slices, keyed by field `key`).
- Validation: The `validation` field exists in the DSL, but custom validation callbacks are not implemented yet in the current codebase.
- File picker: When using `filepicker`, set `current_directory` and allowed types as needed.
- CLI printing: The built-in command implementation prints the resulting values as YAML.

## Get the DSL guide content programmatically

Sometimes you need the rendered Markdown for the DSL guide at runtime. Here are two practical approaches.

### Option 1 — Read from the repository (simple)

```go
import (
    "fmt"
    "os"
)

func readDslGuideFromRepo() (string, error) {
    b, err := os.ReadFile("pkg/doc/topics/uhoh-dsl.md")
    if err != nil {
        return "", err
    }
    return string(b), nil
}
```

This works when your program runs from the repo or includes that file alongside your binary.

### Option 2 — Use the embedded docs helper

Use a helper that loads the embedded Uhoh docs and returns the DSL guide markdown.

```go
import uhohdoc "github.com/go-go-golems/uhoh/pkg/doc"

func loadDslGuideViaHelpSystem() (string, error) {
    return uhohdoc.GetUhohDSLDocumentation()
}
```

If your app exposes the help command via Cobra, you can also shell out and capture the output of `uhoh help uhoh-dsl`.

```go
import (
    "bytes"
    "os/exec"
)

func readDslGuideViaCli() (string, error) {
    cmd := exec.Command("uhoh", "help", "uhoh-dsl")
    var out bytes.Buffer
    cmd.Stdout = &out
    cmd.Stderr = &out
    if err := cmd.Run(); err != nil {
        return "", err
    }
    return out.String(), nil
}
```

## See also

- glaze help how-to-write-good-documentation-pages
- glaze help commands-reference
