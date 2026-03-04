package steps

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

// EmbeddableStep extends Step with the ability to build a non-blocking
// huh.Form that can be embedded in a parent Bubble Tea model (e.g., a
// FormOverlay). Steps that implement this interface can participate in
// the async WizardModel instead of blocking with Run().
type EmbeddableStep interface {
	Step
	// BuildModel creates a huh.Form for this step. The returned values map
	// contains pointers to the form's bound variables; call
	// pkg.ExtractFinalValues(values) after the form reaches StateCompleted
	// to get the final results.
	BuildModel(state map[string]interface{}) (*huh.Form, map[string]interface{}, error)
}

// ActionCompletedMsg is sent when an async action step's callback finishes.
type ActionCompletedMsg struct {
	Result interface{}
	Err    error
}

// AsyncEmbeddableStep extends EmbeddableStep with async work that runs
// alongside the form (e.g., a callback that executes while a spinner is shown).
// The WizardModel batches the form init with RunAsync and waits for
// ActionCompletedMsg before advancing.
type AsyncEmbeddableStep interface {
	EmbeddableStep
	// RunAsync returns a tea.Cmd that performs async work and sends
	// ActionCompletedMsg when done.
	RunAsync(state map[string]interface{}) tea.Cmd
	// OutputKey returns the state key where the action result should be stored.
	GetOutputKey() string
}
