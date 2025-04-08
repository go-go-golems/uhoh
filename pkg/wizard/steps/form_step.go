package steps

import (
	"context"
	"fmt"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/pkg/errors"
)

// FormStep represents a step that displays an interactive form.
type FormStep struct {
	BaseStep `yaml:",inline"`
	FormData pkg.Form `yaml:"form"` // Reuse the existing Form definition
}

var _ Step = &FormStep{}

// Execute runs the form defined in the step.
func (fs *FormStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	// TODO(manuel, 2024-08-05) Consider passing state into the form for defaults/pre-population
	// TODO(manuel, 2024-08-05) Add step title/description rendering
	fmt.Printf("\n--- Step: %s ---\n", fs.Title())
	if fs.Description != "" {
		fmt.Printf("%s\n", fs.Description)
	}

	// Run the actual form
	formResults, err := fs.FormData.Run(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "error running form step %s", fs.ID())
	}

	// TODO(manuel, 2024-08-05) Define how form results merge into the main wizard state
	// For now, just return the raw form results. The runner will merge them.
	return formResults, nil
}

func (fs *FormStep) GetBaseStep() *BaseStep {
	return &fs.BaseStep
}
