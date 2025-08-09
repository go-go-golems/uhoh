package main

import (
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/spf13/cobra"

	clay "github.com/go-go-golems/clay/pkg"
	clay_repositories "github.com/go-go-golems/clay/pkg/cmds/repositories"
	"github.com/go-go-golems/glazed/pkg/cli"
	"github.com/go-go-golems/glazed/pkg/cmds/logging"
	"github.com/go-go-golems/glazed/pkg/help"
	"github.com/pkg/profile"
	"github.com/rs/zerolog/log"

	// Adjust import path for commands
	app_cmds "github.com/go-go-golems/uhoh/cmd/uhoh/cmds"
)

var version = "dev"
var profiler interface {
	Stop()
}

var rootCmd = &cobra.Command{
	Use:     "uhoh",
	Short:   "uhoh is a tool to help you run forms and wizards",
	Version: version,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		err := logging.InitLoggerFromViper()
		if err != nil {
			return err
		}

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
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if profiler != nil {
			log.Info().Msg("Stopping memory profiler")
			profiler.Stop()
		}
	},
}

func main() {
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	err = clay.InitViper("uhoh", rootCmd)
	cobra.CheckErr(err)

	// Instantiate and build commands from the cmds package
	exampleCmd, err := app_cmds.NewExampleCommand()
	cobra.CheckErr(err)
	cobraExampleCmd, err := cli.BuildCobraCommandFromBareCommand(exampleCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraExampleCmd)

	streamCmd, err := app_cmds.NewStreamCommand()
	cobra.CheckErr(err)
	cobraStreamCmd, err := cli.BuildCobraCommandFromBareCommand(streamCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraStreamCmd)

	testStreamCmd, err := app_cmds.NewTestStreamCommand()
	cobra.CheckErr(err)
	cobraTestStreamCmd, err := cli.BuildCobraCommandFromBareCommand(testStreamCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraTestStreamCmd)

	runWizardCmd, err := app_cmds.NewRunWizardCommand()
	cobra.CheckErr(err)
	cobraRunWizardCmd, err := cli.BuildCobraCommandFromBareCommand(runWizardCmd)
	cobra.CheckErr(err)
	rootCmd.AddCommand(cobraRunWizardCmd)

	// Add the dynamic run-command from the cmds package
	runCmdCobra := app_cmds.NewRunCommandCobraCmd()
	rootCmd.AddCommand(runCmdCobra)

	// Add clay repositories command group
	rootCmd.AddCommand(clay_repositories.NewRepositoriesGroupCommand())

	// Add persistent flags
	rootCmd.PersistentFlags().Bool("mem-profile", false, "Enable memory profiling")

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
