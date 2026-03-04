package main

import (
	"embed"
	"fmt"
	"log"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-go-golems/uhoh/pkg/wizard"
)

//go:embed wizard.yaml
var wizardFS embed.FS

// appModel wraps a WizardModel in a parent Bubble Tea model that handles
// wizard lifecycle messages and provides surrounding UI (header/footer).
type appModel struct {
	wizard *wizard.WizardModel
	result map[string]interface{}
	err    error
	done   bool
}

func newAppModel() (appModel, error) {
	b, err := wizardFS.ReadFile("wizard.yaml")
	if err != nil {
		return appModel{}, err
	}
	w, err := wizard.LoadWizardFromBytes(b)
	if err != nil {
		return appModel{}, err
	}
	wm := wizard.NewWizardModel(w)
	return appModel{wizard: wm}, nil
}

func (m appModel) Init() tea.Cmd {
	return m.wizard.Init()
}

func (m appModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case wizard.WizardCompletedMsg:
		m.result = msg.State
		m.done = true
		return m, nil

	case wizard.WizardAbortedMsg:
		m.done = true
		return m, tea.Quit
	}

	// Delegate to the wizard model.
	updated, cmd := m.wizard.Update(msg)
	if wm, ok := updated.(wizard.WizardModel); ok {
		m.wizard = &wm
	}

	return m, cmd
}

func (m appModel) View() string {
	var sb strings.Builder

	step := m.wizard.StepIndex() + 1
	total := m.wizard.StepCount()
	sb.WriteString(fmt.Sprintf("=== Embedded Wizard Demo (Step %d/%d) ===\n\n", step, total))

	if m.err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", m.err))
		return sb.String()
	}

	if m.done {
		if m.result != nil {
			sb.WriteString("Wizard completed!\n\n")
			for k, v := range m.result {
				sb.WriteString(fmt.Sprintf("  %s: %v\n", k, v))
			}
			sb.WriteString("\nPress ctrl+c to exit.\n")
		} else {
			sb.WriteString("Wizard was cancelled.\n")
		}
		return sb.String()
	}

	sb.WriteString(m.wizard.View())

	return sb.String()
}

func main() {
	m, err := newAppModel()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
