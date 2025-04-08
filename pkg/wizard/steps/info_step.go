package steps

import (
	"context"
	"fmt"
)

// InfoStep represents a step that displays informational text.
type InfoStep struct {
	BaseStep `yaml:",inline"`
	Content  string `yaml:"content"`
}

var _ Step = &InfoStep{}

func (is *InfoStep) Execute(ctx context.Context, state map[string]interface{}) (map[string]interface{}, error) {
	fmt.Printf("\n--- Step: %s ---\n", is.Title())
	if is.Description() != "" {
		fmt.Printf("%s\n", is.Description())
	}
	fmt.Println(is.Content)
	// Info steps typically don't modify state
	return map[string]interface{}{}, nil
}

func (is *InfoStep) GetBaseStep() *BaseStep {
	return &is.BaseStep
}
