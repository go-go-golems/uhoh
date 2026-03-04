package steps

import (
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
