package cmds

import (
	"fmt"

	"github.com/go-go-golems/glazed/pkg/cli"
	glazed_cmds "github.com/go-go-golems/glazed/pkg/cmds"
	"github.com/go-go-golems/glazed/pkg/cmds/alias"
	"github.com/go-go-golems/glazed/pkg/cmds/loaders"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/go-go-golems/uhoh/pkg/cmds"
	"github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewRunCommandCobraCmd creates the cobra command for running commands defined in files.
// This command doesn't fit the glazed.Command pattern well because its primary function
// is to load and execute *another* command.
func NewRunCommandCobraCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run-command [command-file] [args...]",
		Short: "Run a command defined in a YAML file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// The first argument is the command file, the rest are passed down
			err := handleRunCommand(args[0], args[1:])
			cobra.CheckErr(err)
		},
	}
	return cmd
}

// handleRunCommand loads and executes a command defined in a file.
// This function remains largely the same as its previous version in main.go.
func handleRunCommand(commandFile string, commandArgs []string) error {
	loader := &cmds.UhohCommandLoader{}

	// Use the provided commandFile argument
	fs_, filePath, err := loaders.FileNameToFsFilePath(commandFile)
	if err != nil {
		return errors.Wrapf(err, "could not get absolute path for %s", commandFile)
	}

	// TODO(manuel, 2024-08-05) Should we load aliases here? Check how clay does it.
	// Load commands from the specified file
	cmds_, err := loader.LoadCommands(fs_, filePath, []glazed_cmds.CommandDescriptionOption{}, []alias.Option{})
	if err != nil {
		return errors.Wrapf(err, "could not load command from %s", commandFile)
	}

	if len(cmds_) != 1 {
		return errors.Errorf("expected exactly one command in file %s, got %d", commandFile, len(cmds_))
	}

	loadedCmd := cmds_[0]

	// We need a temporary root command to execute the loaded command
	// because the original rootCmd might have already parsed its flags.
	tempRootCmd := &cobra.Command{
		Use:   fmt.Sprintf("uhoh run-command %s", commandFile),
		Short: "Dynamically loaded command runner",
		RunE: func(cmd *cobra.Command, args []string) error {
			// This RunE should not be called directly if setup is correct.
			// The execution happens when tempRootCmd.Execute() is called below.
			return errors.New("temporary root command should not be executed directly")
		},
		// Disable flag parsing for the temporary root itself, as flags belong to the child command.
		DisableFlagParsing: true,
	}

	// Optionally load embedded docs into a help system for future use
	// but we don't wire a custom help command; the framework handles help.
	_ = doc.AddDocToHelpSystem(help.NewHelpSystem())

	// Build the cobra command for the loaded glazed command
	cobraCommand, err := cli.BuildCobraCommandFromCommand(loadedCmd)
	if err != nil {
		return errors.Wrapf(err, "could not build cobra command from %s", commandFile)
	}

	// Add the built command to the temporary root
	tempRootCmd.AddCommand(cobraCommand)

	// Set the arguments for the temporary root command execution.
	// The first arg is the *use* name of the loaded command, followed by the user-provided args.
	argsForExec := append([]string{cobraCommand.Use}, commandArgs...)
	tempRootCmd.SetArgs(argsForExec)

	// Execute the temporary command structure
	// This will parse flags for the loaded command and run its associated Run function.
	return tempRootCmd.Execute()
}
