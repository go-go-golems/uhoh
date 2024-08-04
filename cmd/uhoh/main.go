package main

import (
	"fmt"
	"log"

	"github.com/go-go-golems/uhoh/pkg"
	"github.com/go-go-golems/uhoh/pkg/doc"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/go-go-golems/glazed/pkg/help"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "uhoh",
	Short:   "uhoh is a tool to help you run forms",
	Version: version,
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
		log.Fatalf("Error parsing YAML: %v", err)
	}

	values, err := form.Run()
	if err != nil {
		log.Fatalf("Error running form: %v", err)
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

	err = rootCmd.Execute()
	cobra.CheckErr(err)
}
