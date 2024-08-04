package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	clay "github.com/go-go-golems/clay/pkg"
	"github.com/go-go-golems/glazed/pkg/help"
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

	values, err := form.Run()
	if err != nil {
		log.Fatal().Err(err).Msg("Error running form")
	}

	fmt.Println("Form Results:")
	for key, value := range values {
		fmt.Printf("%s: %v\n", key, value)
	}
}

func main() {
	helpSystem := help.NewHelpSystem()
	err := doc.AddDocToHelpSystem(helpSystem)
	cobra.CheckErr(err)

	helpSystem.SetupCobraRootCommand(rootCmd)

	rootCmd.AddCommand(ExampleCommand())

	rootCmd.PersistentFlags().Bool("mem-profile", false, "Enable memory profiling")

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
