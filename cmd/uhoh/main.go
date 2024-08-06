package main

import (
	"bufio"
	"context"
	_ "embed"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/cmds"
	"github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/errors"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"
)

var version = "dev"
var profiler interface {
	Stop()
}

var rootCmd = &cobra.Command{
	Use:     "uhoh",
	Short:   "uhoh is a tool to help you run forms",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		err := clay.InitLogger()
		cobra.CheckErr(err)

		memProfile, _ := cmd.Flags().GetBool("mem-profile")
		if memProfile {
			log.Info().Msg("Starting memory profiler")
			profiler = profile.Start(profile.MemProfile)

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGHUP)
			go func() {
				for range sigCh {
					log.Info().Msg("Restarting memory profiler")
					profiler.Stop()
					profiler = profile.Start(profile.MemProfile)
				}
			}()
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if profiler != nil {
			log.Info().Msg("Stopping memory profiler")
			profiler.Stop()
		}
	},
}

func ExampleCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "example",
		Short: "Run an example form",
		Run:   runExample,
	}
}

func runExample(cmd *cobra.Command, args []string) {
	yamlData := `
name: Burger Order Form
theme: Charm
groups:
  - name: Burger Selection
    fields:
      - type: select
        key: burger
        title: Choose your burger
        options:
          - label: Charmburger Classic
            value: classic
          - label: Chickwich
            value: chickwich
  - name: Order Details
    fields:
      - type: input
        key: name
        title: What's your name?
        validation:
          - condition: Frank@
            error: Sorry, we don't serve customers named Frank.
      - type: confirm
        key: discount
        title: Would you like 15% off?ggj
`

	var form pkg.Form
	err := yaml.Unmarshal([]byte(yamlData), &form)
	if err != nil {
		log.Fatal().Err(err).Msg("Error parsing YAML")
	}

	values, err := form.Run(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Error running form")
	}

	fmt.Println("Form Results:")
	for key, value := range values {
		fmt.Printf("%s: %v\n", key, value)
	}
}

var runCommandCmd = &cobra.Command{
	Use:   "run-command",
	Short: "Run a command from a file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := handleRunCommand()
		cobra.CheckErr(err)
	},
}

var streamCmd = &cobra.Command{
	Use:   "stream",
	Short: "Stream commands from stdin",
	Run:   runStream,
}

func init() {
	streamCmd.Flags().String("error-behavior", "exit", "Error behavior: ignore, debug, or exit")
}

func runStream(cmd *cobra.Command, args []string) {
	errorBehavior, _ := cmd.Flags().GetString("error-behavior")
	runStreamWithReader(os.Stdin, errorBehavior)
}

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
			time.Sleep(time.Millisecond * 100)
			cancel()
			wg.Wait()
		}

		currentCtx, cancel = context.WithCancel(context.Background())
		wg.Add(1)
		mu.Unlock()

		go func(ctx context.Context, command string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					handleError(fmt.Errorf("%v", r), command, errorBehavior)
				}
			}()
			runStreamCommand(ctx, command, errorBehavior)
		}(currentCtx, accumulatedInput)
	}

	if cancel != nil {
		cancel()
		wg.Wait()
	}

	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error reading input:", err)
	}
}

func runStreamCommand(ctx context.Context, command string, errorBehavior string) {
	// Create a UhohCommandLoader
	loader := &cmds.UhohCommandLoader{}

	// Use the loader to parse the command
	commands, err := loader.LoadUhohCommandFromReader(
		strings.NewReader(command),
		[]glazed_cmds.CommandDescriptionOption{},
		[]alias.Option{},
	)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error parsing command: %v\n", err)
		return
	}

	if len(commands) != 1 {
		handleError(fmt.Errorf("expected exactly one command, got %d", len(commands)), command, errorBehavior)
		return
	}

	// Extract the form from the UhohCommand
	uhohCmd, ok := commands[0].(*cmds.UhohCommand)
	if !ok {
		handleError(fmt.Errorf("unexpected command type"), command, errorBehavior)
		return
	}

	form := uhohCmd.Form

	// Run the form
	values, err := form.Run(ctx)
	if err != nil {
		if err == context.Canceled || ctx.Err() == context.Canceled {
			fmt.Println("Command cancelled")
		} else {
			handleError(err, command, errorBehavior)
		}
		return
	}

	if len(values) > 0 {
		fmt.Println("Form Results:")
		for key, value := range values {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

func handleError(err error, command string, errorBehavior string) {
	switch errorBehavior {
	case "ignore":
		// Do nothing
	case "debug":
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n%s\nError: %v\n", command, err)
	case "exit":
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n%s\nError: %v\n", command, err)
		os.Exit(1)
	default:
		// Default behavior: print error and continue
		_, _ = fmt.Fprintf(os.Stderr, "Error occurred while processing command:\n%s\nError: %v\n", command, err)
	}
}

var testStreamCmd = &cobra.Command{
	Use:   "test-stream",
	Short: "Test stream command with simulated slow input",
	Run:   runTestStream,
}

func init() {
	testStreamCmd.Flags().String("error-behavior", "exit", "Error behavior: ignore, debug, or exit")
}

//go:embed examples/05-snake-care-info.yaml
var snakeCareInfo string

func runTestStream(cmd *cobra.Command, args []string) {
	errorBehavior, _ := cmd.Flags().GetString("error-behavior")
	r, w := io.Pipe()
	defer func(r *io.PipeReader) {
		_ = r.Close()
	}(r)

	go func() {
		defer func(w *io.PipeWriter) {
			_ = w.Close()
		}(w)
		lines := strings.Split(snakeCareInfo, "\n")

		for _, line := range lines {
			time.Sleep(time.Duration(50+rand.Intn(50)) * time.Millisecond)
			_, err := w.Write([]byte(line + "\n"))
			if err != nil {
				log.Error().Err(err).Msg("Error writing to pipe")
				return
			}
		}
	}()

	runStreamWithReader(r, errorBehavior)
}

func main() {
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	helpSystem.SetupCobraRootCommand(rootCmd)

	rootCmd.AddCommand(runCommandCmd)
	rootCmd.AddCommand(ExampleCommand())
	rootCmd.AddCommand(streamCmd)
	rootCmd.AddCommand(testStreamCmd) // Add the new test-stream command

	rootCmd.PersistentFlags().Bool("mem-profile", false, "Enable memory profiling")

	if len(os.Args) >= 3 && os.Args[1] == "run-command" && os.Args[2] != "--help" {
		err := handleRunCommand()
		cobra.CheckErr(err)
		return
	}

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}

func handleRunCommand() error {
	loader := &cmds.UhohCommandLoader{} // Assuming you have a UhohCommandLoader

	fs_, filePath, err := loaders.FileNameToFsFilePath(os.Args[2])
	if err != nil {
		return errors.Wrap(err, "could not get absolute path")
	}

	cmds_, err := loader.LoadCommands(fs_, filePath, []glazed_cmds.CommandDescriptionOption{}, []alias.Option{})
	if err != nil {
		return errors.Wrap(err, "could not load command")
	}

	if len(cmds_) != 1 {
		return errors.Errorf("expected exactly one command, got %d", len(cmds_))
	}

	cobraCommand, err := cli.BuildCobraCommandFromCommand(cmds_[0])
	if err != nil {
		return errors.Wrap(err, "could not build cobra command")
	}

	helpSystem := help.NewHelpSystem()
	err = doc.AddDocToHelpSystem(helpSystem)
	if err != nil {
		return errors.Wrap(err, "could not add doc to help system")
	}

	rootCmd.AddCommand(cobraCommand)
	restArgs := os.Args[3:]
	os.Args = append([]string{os.Args[0], cobraCommand.Use}, restArgs...)

	return rootCmd.Execute()
}
