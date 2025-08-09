# Wizard Step Examples

This directory contains example YAML files demonstrating the different step types available in the Uhoh Wizard DSL. These examples can help you understand how to use various step types in your own wizards.

## Available Examples

1. **info_step_example.yaml** - Demonstrates how to use Info Steps to display information to users.
2. **decision_step_example.yaml** - Shows how Decision Steps can be used to create branching flows based on user choices.
3. **action_step_example.yaml** - Illustrates how Action Steps perform backend operations without direct user input.
4. **summary_step_example.yaml** - Demonstrates how Summary Steps can display collected information for review.
5. **multi_step_example.yaml** - A comprehensive example that combines all step types in a cohesive project setup wizard.
6. **action_callback_example.yaml** - Shows how to use registered action callbacks for executing backend operations.

## Running the Examples

To run any of these examples, use the `uhoh run-wizard` command:

```bash
# Navigate to the repository root
cd /path/to/corporate-headquarters

# Build the uhoh binary (if not already built)
go build ./uhoh/cmd/uhoh

# Run an example
./uhoh run-wizard uhoh/cmd/uhoh/examples/wizard/steps/info_step_example.yaml
```

## Step Types Overview

### Info Step

Info steps display markdown-formatted content to the user. They're useful for welcome screens, instructions, and completion messages.

```yaml
- id: welcome
  type: info
  title: Welcome Screen
  description: Optional subtitle or context
  content: |
    # Markdown content here
    This supports **formatting** and *styles*.
```

### Decision Step

Decision steps present a list of options and store the user's choice in the wizard state. They can also control flow with the `next_step_map`.

```yaml
- id: choose_option
  type: decision
  title: Make a Choice
  description: Choose one of the following options
  target_key: selected_option
  choices:
    - "Option One"
    - "Option Two"
  next_step_map:
    "Option One": step_one
    "Option Two": step_two
```

### Form Step

Form steps collect structured data using various field types (input, select, checkbox, etc.).

```yaml
- id: user_info
  type: form
  title: Enter Information
  form:
    groups:
      - name: Personal Info
        fields:
          - type: input
            key: name
            title: Full Name
          - type: select
            key: language
            title: Preferred Language
            options:
              - label: "English"
                value: "en"
              - label: "Spanish"
                value: "es"
```

### Action Step

Action steps execute operations in the background using registered callback functions.

```yaml
- id: process_data
  type: action
  title: Processing
  description: Please wait while we process your request
  action_type: function
  function_name: processUserData
  arguments:
    validate: true
  output_key: result
```

### Summary Step

Summary steps display collected information for user review.

```yaml
- id: review
  type: summary
  title: Review Information
  description: Please review the entered information
  sections:
    - title: Personal Details
      fields:
        - name
        - email
    - title: Preferences
      fields:
        - language
        - theme
```

## Custom Callbacks

For Action Steps to work in real applications, you need to register callback functions. For examples, check out the wizard documentation and the examples in the `wizard/` directory with callbacks.

## Using Action Callbacks

Action callbacks require registration before running the wizard. Here's an example of how to create a custom command that registers action callbacks:

```go
package main

import (
	"context"
	"fmt"

	"github.com/go-go-golems/uhoh/pkg/wizard"
	"github.com/spf13/cobra"
)

func main() {
	exampleCmd := &cobra.Command{
		Use:   "callback-example",
		Short: "Run the action callback example with registered callbacks",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load the wizard with registered callbacks
			w, err := wizard.LoadWizard(
				"uhoh/cmd/uhoh/examples/wizard/steps/action_callback_example.yaml",
				wizard.WithActionCallback("fetchUserData", fetchUserData),
				wizard.WithActionCallback("processData", processData),
				wizard.WithActionCallback("saveResults", saveResults),
			)
			if err != nil {
				return err
			}

			// Run the wizard
			result, err := w.Run(context.Background(), nil)
			if err != nil {
				return err
			}

			fmt.Println("Wizard completed successfully!")
			fmt.Printf("Operation ID: %v\n", result["operation_id"])
			return nil
		},
	}

	// Execute the command
	if err := exampleCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}

// Action callback implementations
func fetchUserData(ctx context.Context, state map[string]interface{}, args map[string]interface{}) (interface{}, error) {
	userID, _ := args["user_id"].(string)
	includeProfile, _ := args["include_profile"].(bool)

	fmt.Printf("Fetching data for user %s (include profile: %v)\n", userID, includeProfile)

	// Simulate fetching data
	return map[string]interface{}{
		"id": userID,
		"name": "Example User",
		"email": "user@example.com",
		"joined": "2022-06-15",
	}, nil
}

func processData(ctx context.Context, state map[string]interface{}, args map[string]interface{}) (interface{}, error) {
	// Process the user data
	return map[string]interface{}{
		"status": "complete",
		"score": 85,
		"timestamp": "2024-08-08T12:34:56Z",
	}, nil
}

func saveResults(ctx context.Context, state map[string]interface{}, args map[string]interface{}) (interface{}, error) {
	// Save results to database
	return "OP-12345678", nil
}
```

For the examples without registered callbacks, the ActionStep will run in simulation mode, displaying a message that callbacks aren't available but allowing the wizard to continue.
