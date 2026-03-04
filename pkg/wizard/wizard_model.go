package wizard

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	uhoh "github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/wizard/steps"
)

// WizardCompletedMsg is sent when all wizard steps have been completed.
type WizardCompletedMsg struct {
	State map[string]interface{}
}

// WizardAbortedMsg is sent when the wizard is cancelled.
type WizardAbortedMsg struct{}

// stepAdvanceMsg is an internal message to trigger advancing to the next step.
type stepAdvanceMsg struct{}

// WizardModel is a non-blocking tea.Model that drives wizard steps through
// the Bubble Tea update loop. Each step that implements EmbeddableStep
// produces a huh.Form which is driven via Update/View.
type WizardModel struct {
	wizard    *Wizard
	stepIndex int
	state     map[string]interface{}

	// Current step's form and values.
	form   *huh.Form
	values map[string]interface{}

	// Metadata for the current step.
	currentStep steps.EmbeddableStep
}

// NewWizardModel creates a WizardModel from an existing Wizard.
// The wizard's steps must implement EmbeddableStep.
func NewWizardModel(w *Wizard) *WizardModel {
	return &WizardModel{
		wizard:    w,
		stepIndex: -1, // Will advance to 0 on Init
		state:     make(map[string]interface{}),
	}
}

// Init starts the wizard by advancing to the first step.
func (m WizardModel) Init() tea.Cmd {
	return m.advanceToNextStep()
}

// Update handles messages for the current step's form.
func (m WizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case stepAdvanceMsg:
		cmd := m.advanceToNextStep()
		return m, cmd
	}

	if m.form == nil {
		return m, nil
	}

	// Delegate to the inner form.
	model, cmd := m.form.Update(msg)
	if f, ok := model.(*huh.Form); ok {
		m.form = f
	}

	// Check for form completion.
	if m.form != nil {
		switch m.form.State {
		case huh.StateNormal:
			// Form still active, nothing to do.
		case huh.StateCompleted:
			completeCmd := m.handleStepCompletion()
			return m, tea.Batch(cmd, completeCmd)
		case huh.StateAborted:
			return m, func() tea.Msg { return WizardAbortedMsg{} }
		}
	}

	return m, cmd
}

// View renders the current step's form.
func (m WizardModel) View() string {
	if m.form == nil {
		return ""
	}
	return m.form.View()
}

// State returns the current wizard state.
func (m *WizardModel) State() map[string]interface{} {
	return m.state
}

// StepIndex returns the current step index (zero-based).
func (m *WizardModel) StepIndex() int {
	return m.stepIndex
}

// StepCount returns the total number of steps.
func (m *WizardModel) StepCount() int {
	return len(m.wizard.Steps)
}

// advanceToNextStep moves to the next step and builds its form.
func (m *WizardModel) advanceToNextStep() tea.Cmd {
	m.stepIndex++

	// Check if wizard is complete.
	if m.stepIndex >= len(m.wizard.Steps) {
		finalState := make(map[string]interface{})
		for k, v := range m.state {
			finalState[k] = v
		}
		return func() tea.Msg {
			return WizardCompletedMsg{State: finalState}
		}
	}

	step := m.wizard.Steps[m.stepIndex]

	embeddable, ok := step.(steps.EmbeddableStep)
	if !ok {
		// Skip non-embeddable steps.
		return m.advanceToNextStep()
	}

	m.currentStep = embeddable

	form, values, err := embeddable.BuildModel(m.state)
	if err != nil || form == nil {
		// Skip steps that fail to build.
		return m.advanceToNextStep()
	}

	m.form = form
	m.values = values

	return m.form.Init()
}

// handleStepCompletion extracts values from the completed step and advances.
func (m *WizardModel) handleStepCompletion() tea.Cmd {
	if m.values != nil {
		finalValues, err := uhoh.ExtractFinalValues(m.values)
		if err == nil {
			for k, v := range finalValues {
				m.state[k] = v
			}
		}
	}

	// Handle decision step navigation.
	if ds, ok := m.currentStep.(*steps.DecisionStep); ok {
		if ds.NextStepMap != nil {
			if targetKey, exists := m.state[ds.TargetKey]; exists {
				if targetStr, ok := targetKey.(string); ok {
					if nextStep, ok := ds.NextStepMap[targetStr]; ok {
						ds.GetBaseStep().NextStep = nextStep
					}
				}
			}
		}
	}

	m.form = nil
	m.values = nil
	m.currentStep = nil

	return func() tea.Msg { return stepAdvanceMsg{} }
}
