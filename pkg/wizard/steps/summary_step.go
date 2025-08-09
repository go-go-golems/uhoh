package steps

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// SummarySection defines a section in the summary
type SummarySection struct {
	Title  string   `yaml:"title"`
	Fields []string `yaml:"fields"` // List of state keys to display
}

// SummaryStep represents a step that displays collected data.
type SummaryStep struct {
	BaseStep `yaml:",inline"`
	Sections []SummarySection `yaml:"sections"`
	Editable bool             `yaml:"editable,omitempty"`
	Template string           `yaml:"template,omitempty"` // Optional Go template to format state
}

var _ Step = &SummaryStep{}

func (ss *SummaryStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	log.Debug().Str("stepId", ss.ID()).Msgf("--- Step: %s ---", ss.Title())

	// If using template mode, we'll handle that separately
	if ss.Template != "" {
		// TODO: In the future, implement template-based rendering
		log.Warn().Str("stepId", ss.ID()).Msg("Template-based summary not yet implemented")
	}

	// Build a formatted summary text from the sections and state
	var sb strings.Builder

	// If no sections defined, show all state
	if len(ss.Sections) == 0 {
		sb.WriteString("## Current State\n\n")
		for k, v := range state {
			sb.WriteString(fmt.Sprintf("- **%s**: %v\n", k, v))
		}
	} else {
		// Process each defined section
		for _, section := range ss.Sections {
			sb.WriteString(fmt.Sprintf("## %s\n\n", section.Title))

			if len(section.Fields) == 0 {
				sb.WriteString("(No fields defined for this section)\n\n")
				continue
			}

			for _, field := range section.Fields {
				value, exists := state[field]
				if !exists {
					sb.WriteString(fmt.Sprintf("- **%s**: (not set)\n", field))
					continue
				}
				sb.WriteString(fmt.Sprintf("- **%s**: %v\n", field, value))
			}
			sb.WriteString("\n")
		}
	}

	// Create a note to display the summary
	note := huh.NewNote().
		Title(ss.Title()).
		Description(sb.String())

	if ss.Description() != "" {
		// Include the description as part of the note's description
		note = huh.NewNote().
			Title(ss.Title()).
			Description(ss.Description() + "\n\n" + sb.String())
	}

	// Show the note
	err := note.Run()
	if err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return nil, ErrUserAborted
		}
		return nil, errors.Wrap(err, "error displaying summary step")
	}

	// If editable is true, we would need to allow navigation back to edit fields
	// This would require more complex navigation logic in the wizard runner
	if ss.Editable {
		log.Warn().Str("stepId", ss.ID()).Msg("Editable summary not fully implemented yet")
		// TODO: Implement field editing functionality
	}

	// Summary steps don't modify state unless editable with changes
	return map[string]interface{}{}, nil
}

func (ss *SummaryStep) GetBaseStep() *BaseStep {
	return &ss.BaseStep
}
