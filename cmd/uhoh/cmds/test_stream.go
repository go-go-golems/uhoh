package cmds

import (
	"context"
	_ "embed"
	"io"
	"math/rand"
	"strings"
	"time"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

//go:embed examples/05-snake-care-info.yaml
var snakeCareInfo string

// TestStreamSettings holds settings for the test-stream command.
type TestStreamSettings struct {
	ErrorBehavior string `glazed.parameter:"error-behavior"`
}

// TestStreamCommand tests the streaming functionality with simulated slow input.
type TestStreamCommand struct {
	*glazed_cmds.CommandDescription
}

var _ glazed_cmds.BareCommand = &TestStreamCommand{}

// NewTestStreamCommand creates a new TestStreamCommand.
func NewTestStreamCommand() (*TestStreamCommand, error) {
	return &TestStreamCommand{
		CommandDescription: glazed_cmds.NewCommandDescription(
			"test-stream",
			glazed_cmds.WithShort("Test stream command with simulated slow input"),
			glazed_cmds.WithFlags(
				parameters.NewParameterDefinition(
					"error-behavior",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Error behavior during test stream"),
					parameters.WithChoices("ignore", "debug", "exit"),
					parameters.WithDefault("exit"),
				),
			),
		),
	}, nil
}

// Run executes the test stream simulation.
func (c *TestStreamCommand) Run(
	ctx context.Context,
	parsedLayers *layers.ParsedLayers,
) error {
	s := &TestStreamSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	r, w := io.Pipe()

	// Goroutine to simulate slow writing to the pipe
	go func() {
		defer func() {
			// Ensure the writer is closed even if errors occur during writing
			if err := w.Close(); err != nil {
				log.Error().Err(err).Msg("Error closing pipe writer in test-stream")
			}
		}()

		lines := strings.Split(snakeCareInfo, "\n")
		for _, line := range lines {
			// Introduce random delay to simulate typing or slow stream
			delay := time.Duration(50+rand.Intn(50)) * time.Millisecond
			time.Sleep(delay)

			_, err := w.Write([]byte(line + "\n"))
			if err != nil {
				// Log the error and stop writing if the pipe breaks (e.g., reader closed)
				log.Error().Err(err).Msg("Error writing to pipe in test-stream")
				return
			}
		}
	}()

	// Run the stream reader, consuming from the pipe
	// This function will block until the pipe writer is closed or an error occurs.
	runStreamWithReader(r, s.ErrorBehavior)

	// Close the reader end of the pipe after runStreamWithReader finishes.
	// This is important to clean up resources, though runStreamWithReader might have already handled reader closure implicitly.
	if err := r.Close(); err != nil {
		log.Warn().Err(err).Msg("Error closing pipe reader in test-stream (might already be closed)")
	}

	// Similar to StreamCommand, reaching here usually means success.
	return nil
}
