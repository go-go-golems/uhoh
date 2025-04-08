package steps

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
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
	fmt.Printf("\n--- Step: %s ---\n", ss.Title())
	// TODO(manuel, 2024-08-05) Implement summary display logic (use template, show state)
	fmt.Println("Summary Step (Not Implemented)")
	fmt.Println("Current State:")
	for k, v := range state {
		fmt.Printf("  %s: %v\n", k, v)
	}
	// Summary steps don't modify state unless editable allows going back
	return map[string]interface{}{}, errors.New("summary step not implemented")
}

func (ss *SummaryStep) GetBaseStep() *BaseStep {
	return &ss.BaseStep
}
