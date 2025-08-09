package steps

import (
	"context"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
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
	log.Debug().Str("stepId", fs.ID()).Msgf("--- Step: %s ---", fs.Title())
	if fs.Description() != "" {
		log.Debug().Str("stepId", fs.ID()).Msg(fs.Description())
	}

	// Run the actual form
	log.Debug().Str("stepId", fs.ID()).Msg("Running form")
	formResults, err := fs.FormData.Run(ctx)
	if err != nil {
		// Check if the error is ErrUserAborted from the form runner
		if errors.Is(err, ErrUserAborted) {
			log.Debug().Str("stepId", fs.ID()).Msg("Form aborted by user")
			return nil, ErrUserAborted // Propagate standard wizard abort error
		}
		return nil, errors.Wrapf(err, "error running form step %s", fs.ID())
	}
	log.Debug().Str("stepId", fs.ID()).Interface("formResults", formResults).Msg("Form completed")

	// TODO(manuel, 2024-08-05) Define how form results merge into the main wizard state
	// For now, just return the raw form results. The runner will merge them.
	return formResults, nil
}

func (fs *FormStep) GetBaseStep() *BaseStep {
	return &fs.BaseStep
}
