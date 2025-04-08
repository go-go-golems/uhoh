package steps

import (
	"context"

	"github.com/rs/zerolog/log"
)

// SummaryStep represents a step that displays collected data.
type SummaryStep struct {
	BaseStep `yaml:",inline"`
	Template string `yaml:"template,omitempty"` // Optional Go template to format state
	Editable bool   `yaml:"editable,omitempty"`
	// TODO(manuel, 2024-08-05) Define fields for selecting which state keys to display
}

var _ Step = &SummaryStep{}

func (ss *SummaryStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	log.Debug().Str("stepId", ss.ID()).Msgf("--- Step: %s ---", ss.Title())
	// TODO(manuel, 2024-08-05) Implement summary display logic (use template, show state)
	log.Warn().Str("stepId", ss.ID()).Msg("Summary Step (Not Implemented)")
	log.Debug().Str("stepId", ss.ID()).Interface("currentState", state).Msg("Current State:")

	// Summary steps don't modify state unless editable allows going back
	return map[string]interface{}{}, ErrStepNotImplemented // Use standard error
}

func (ss *SummaryStep) GetBaseStep() *BaseStep {
	return &ss.BaseStep
}
