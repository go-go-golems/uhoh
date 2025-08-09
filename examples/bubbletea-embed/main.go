package main

import (
    "embed"
    "fmt"
    "log"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/huh"
    uhoh "github.com/go-go-golems/uhoh/pkg"
)

//go:embed form.yaml
var formFS embed.FS

type parentModel struct {
    form         *huh.Form
    formValues   map[string]interface{}
    result       map[string]interface{}
    initialized  bool
    done         bool
    err          error
}

func NewParentModel() (parentModel, error) {
    b, err := formFS.ReadFile("form.yaml")
    if err != nil {
        return parentModel{}, err
    }
    f, vals, err := uhoh.BuildBubbleTeaModelFromYAML(b)
    if err != nil {
        return parentModel{}, err
    }
    return parentModel{form: f, formValues: vals}, nil
}

func (m parentModel) Init() tea.Cmd {
    return m.form.Init()
}

func (m parentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    // Let the embedded form handle the message first
    fm, cmd := m.form.Update(msg)
    if f, ok := fm.(*huh.Form); ok {
        m.form = f
    }

    if m.form.State == huh.StateCompleted && !m.done {
        res, err := uhoh.ExtractFinalValues(m.formValues)
        if err != nil {
            m.err = err
        }
        m.result = res
        m.done = true
    }
    return m, cmd
}

func (m parentModel) View() string {
    if m.err != nil {
        return fmt.Sprintf("Error: %v\n", m.err)
    }
    if m.done {
        return fmt.Sprintf("Form completed!\nValues: %v\nPress Ctrl+C to exit.\n", m.result)
    }
    return m.form.View()
}

func main() {
    m, err := NewParentModel()
    if err != nil {
        log.Fatal(err)
    }
    if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
        log.Fatal(err)
    }
}


