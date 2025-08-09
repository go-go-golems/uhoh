package cmds

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/layers"
	"github.com/go-go-golems/glazed/pkg/cmds/parameters"
	"github.com/go-go-golems/uhoh/pkg/cmds" // Use the package alias if needed
	"github.com/pkg/errors"
)

// StreamSettings defines the settings for the stream command.
type StreamSettings struct {
	ErrorBehavior string `glazed.parameter:"error-behavior"`
}

// StreamCommand handles streaming commands from standard input.
type StreamCommand struct {
	*glazed_cmds.CommandDescription
}

var _ glazed_cmds.BareCommand = &StreamCommand{}

// NewStreamCommand creates a new instance of the StreamCommand.
func NewStreamCommand() (*StreamCommand, error) {
	return &StreamCommand{
		CommandDescription: glazed_cmds.NewCommandDescription(
			"stream",
			glazed_cmds.WithShort("Stream commands from stdin"),
			glazed_cmds.WithFlags(
				parameters.NewParameterDefinition(
					"error-behavior",
					parameters.ParameterTypeChoice,
					parameters.WithHelp("Error behavior when processing stream"),
					parameters.WithChoices("ignore", "debug", "exit"),
					parameters.WithDefault("exit"),
				),
			),
		),
	}, nil
}

// Run starts the streaming process.
func (c *StreamCommand) Run(
	ctx context.Context, // Note: The context passed here might not be the primary one used by the stream runner due to its nature.
	parsedLayers *layers.ParsedLayers,
) error {
	s := &StreamSettings{}
	if err := parsedLayers.InitializeStruct(layers.DefaultSlug, s); err != nil {
		return errors.Wrap(err, "failed to initialize settings")
	}

	// Call the helper function that contains the original streaming logic.
	// We pass os.Stdin and the parsed error behavior setting.
	runStreamWithReader(os.Stdin, s.ErrorBehavior)

	// Since runStreamWithReader manages its own lifecycle and potentially exits,
	// reaching here usually means the stream finished without an error needing explicit return.
	return nil
}

// runStreamWithReader processes commands read from the provided reader.
func runStreamWithReader(reader io.Reader, errorBehavior string) {
	scanner := bufio.NewScanner(reader)
	var currentCtx context.Context
	var cancel context.CancelFunc
	var wg sync.WaitGroup
	var mu sync.Mutex
	var accumulatedInput string

	for scanner.Scan() {
		input := scanner.Text()
		accumulatedInput += input + "\n"

		mu.Lock()
		if cancel != nil {
			// Give a brief moment for the previous command to potentially react to cancellation
			time.Sleep(time.Millisecond * 100)
			cancel()  // Cancel the previous context
			wg.Wait() // Wait for the previous goroutine to finish
		}

		// Create a new context for the next command
		currentCtx, cancel = context.WithCancel(context.Background()) // Use Background as parent for independent cancellation
		wg.Add(1)
		mu.Unlock()

		// Launch the command processing in a new goroutine
		go func(ctx context.Context, command string) {
			defer wg.Done()
			defer func() {
				// Recover from potential panics within the command execution
				if r := recover(); r != nil {
					handleError(fmt.Errorf("%v", r), command, errorBehavior)
				}
			}()
			runStreamCommand(ctx, command, errorBehavior)
		}(currentCtx, accumulatedInput)
	}

	// After the loop finishes (scanner is done), cancel the last running command if any
	mu.Lock()
	if cancel != nil {
		cancel()
		wg.Wait()
	}
	mu.Unlock()

	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

// runStreamCommand parses and executes a single command string from the stream.
func runStreamCommand(ctx context.Context, command string, errorBehavior string) {
	// Create a UhohCommandLoader
	loader := &cmds.UhohCommandLoader{}

	// Use the loader to parse the command
	// TODO(manuel, 2024-08-05) Check if we need alias options here? Probably not for streaming.
	// TODO(manuel, 2024-08-05) Pass description options?
	parsedCommands, err := loader.LoadUhohCommandFromReader(
		strings.NewReader(command),
		[]glazed_cmds.CommandDescriptionOption{},
		[]alias.Option{}, // No alias resolution needed for stream, probably
	)
	if err != nil {
		handleError(errors.Wrap(err, "parsing command"), command, errorBehavior)
		return
	}

	if len(parsedCommands) == 0 {
		// Don't treat empty input as an error, just ignore.
		// This can happen with initial empty lines or if the loader returns nothing.
		return
	}

	if len(parsedCommands) != 1 {
		handleError(fmt.Errorf("expected exactly one command, got %d", len(parsedCommands)), command, errorBehavior)
		return
	}

	// Extract the form from the UhohCommand
	uhohCmd, ok := parsedCommands[0].(*cmds.UhohCommand)
	if !ok {
		handleError(fmt.Errorf("unexpected command type: %T", parsedCommands[0]), command, errorBehavior)
		return
	}

	form := uhohCmd.Form

	// Run the form
	values, err := form.Run(ctx)
	if err != nil {
		// Specifically check for context cancellation (which is expected during streaming)
		if errors.Is(err, context.Canceled) || ctx.Err() == context.Canceled {
			fmt.Println("Command cancelled (likely due to new input)")
		} else {
			handleError(errors.Wrap(err, "running form"), command, errorBehavior)
		}
		return
	}

	// Only print results if the command wasn't cancelled and produced output
	if ctx.Err() == nil && len(values) > 0 {
		fmt.Println("Form Results:")
		// TODO(manuel, 2024-08-05) Maybe use YAML or JSON output here for consistency?
		for key, value := range values {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

// handleError handles errors based on the specified behavior.
func handleError(err error, command string, errorBehavior string) {
	switch errorBehavior {
	case "ignore":
		// Do nothing
	case "debug":
		// Print detailed error and the command that caused it
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n---\n%s\n---\nError: %+v\n", command, err) // Use %+v for stack trace with pkg/errors
	case "exit":
		// Print error and exit
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n---\n%s\n---\nError: %+v\nExiting.\n", command, err)
		os.Exit(1)
	default:
		// Default behavior: print error and continue (same as debug but without explicit mention)
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n---\n%s\n---\nError: %+v\n", command, err)
	}
}
