package steps

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
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

	// Build the content, potentially combining the description and content
	displayContent := is.Content

	if is.Description() != "" {
		// Add the description as a header if it exists
		displayContent = fmt.Sprintf("%s\n\n%s", is.Description(), is.Content)
	}

	// Create a note component to display the information
	note := huh.NewNote().
		Title(is.Title()).
		Description(displayContent)

	// Show the note
	err := note.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAborted
		}
		return nil, errors.Wrap(err, "error displaying info step")
	}

	// Info steps typically don't modify state
	return map[string]interface{}{}, nil
}

func (is *InfoStep) GetBaseStep() *BaseStep {
	return &is.BaseStep
}
