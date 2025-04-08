package steps

import (
	"context"

	"github.com/rs/zerolog/log"
)

// InfoStep represents a step that displays informational text.
type InfoStep struct {
	BaseStep `yaml:",inline"`
	Content  string `yaml:"content"`
}

var _ Step = &InfoStep{}

func (is *InfoStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	log.Debug().Str("stepId", is.ID()).Msgf("--- Step: %s ---", is.Title())
	if is.Description() != "" {
		log.Debug().Str("stepId", is.ID()).Msg(is.Description())
	}
	log.Debug().Str("stepId", is.ID()).Msg(is.Content)
	// Debug steps typically don't modify state
	return map[string]interface{}{}, nil
}

func (is *InfoStep) GetBaseStep() *BaseStep {
	return &is.BaseStep
}
