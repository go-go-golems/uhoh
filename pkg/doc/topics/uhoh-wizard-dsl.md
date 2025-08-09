---
Title: Uhoh Wizard DSL
Slug: uhoh-wizard-dsl
Short: Define multi-step wizards with conditional flows using the Uhoh Wizard DSL
Topics:
  - dsl
  - uhoh
  - wizard
Commands:
  - help
Flags:
  - dsl
  - topic
  - help
IsTemplate: false
IsTopLevel: true
ShowPerDefault: true
SectionType: GeneralTopic
---

# Comprehensive Guide to Uhoh Wizard DSL

## Introduction

The Uhoh Wizard DSL extends the Form DSL to create multi-step interactive wizards with conditional flows and callbacks. This guide provides a comprehensive overview of the Wizard DSL structure, step types, navigation logic, and callback integration.

## Top-Level Structure

The Wizard DSL defines a sequence of steps, with navigation logic and callbacks between them. The top-level structure includes wizard metadata, global settings, and step definitions.

```yaml
name: string # Required: Wizard name
description: string # Optional: Wizard description
theme: string # Optional: Theme (Charm, Dracula, Catppuccin, Base16, Default)
save_progress: boolean # Optional: Whether to save progress between sessions (default: false)
global_state: # Optional: Global variables accessible across all steps
  key1: value1
  key2: value2
steps: # Required: List of wizard steps
  - id: string # Each step has a unique identifier
    # Step definition (see Step Types section)
```

### Example:

```yaml
name: New Project Setup
description: Configure a new software project
theme: Dracula
save_progress: true
global_state:
  project_type: ""
  advanced_mode: false
steps:
  - id: welcome
    # Step definition here
  - id: project_basics
    # Step definition here
  - id: configuration
    # Step definition here
```

## Step Types

The DSL supports different types of steps, each with specific purposes in the wizard flow. All steps inherit from a base step structure.

### Base Step Structure

All step types share these common properties:

```yaml
id: string # Required: Unique identifier for the step
title: string # Required: Step title shown to user
description: string # Optional: Detailed description of this step
persistent: boolean # Optional: Whether data persists after navigating away (default: true)
skip_condition: string # Optional: Expression that determines if step should be skipped
visible_condition: string # Optional: Expression that determines if step is visible in progress bar
navigation: # Optional: Custom navigation controls
  next_label: string # Optional: Custom label for the next button
  back_label: string # Optional: Custom label for the back button
  show_back: boolean # Optional: Whether to show back button (default: true)
  next_enabled_condition: string # Optional: Expression that enables/disables next button
  back_enabled_condition: string # Optional: Expression that enables/disables back button
callbacks: # Optional: Callbacks to execute at different points
  before: string # Optional: Callback before step is shown
  after: string # Optional: Callback after step is completed
  validation: string # Optional: Custom validation callback
  navigation: string # Optional: Callback to determine next step
```

### Form Step

A form step contains form fields similar to the Uhoh Form DSL:

```yaml
id: string
type: form # Required: Specifies this is a form step
# ... base step properties ...
form: # Required: Form definition
  groups: # Required: Form field groups
    - name: string # Optional: Group name
      fields:# Required:
        List of fields (same as Form DSL)
        # ... fields as defined in Form DSL
```

Note: Wizard form steps also accept a simplified schema for quick forms:

```yaml
id: user-details
type: form
title: Your Details
form:
  fields:
    - name: name
      label: Name
      type: text
    - name: email
      label: Email
      type: email
```

Mapping for the simplified schema:
- `name` → form field `key`
- `label` → form field `title`
- `type`: `text|email|input` → `input`; `confirm|bool` → `confirm`
- All fields are wrapped into a single implicit group

For when to use which, see also: glaze help uhoh-wizards

### Decision Step

A decision step presents the user with multiple paths to choose from:

```yaml
id: string
type: decision # Required: Specifies this is a decision step
# ... base step properties ...
options: # Required: List of options for the user to choose from
  - label: string # Required: Option label
    value: string # Required: Option value (used as key in state)
    description: string # Optional: Detailed description
    icon: string # Optional: Icon identifier
target_key: string # Required: State key to store the selected option
next_step_map: # Optional: Map of option values to next step IDs
  option_value1: step_id1
  option_value2: step_id2
```

### Summary Step

A summary step displays collected information for review:

```yaml
id: string
type: summary # Required: Specifies this is a summary step
# ... base step properties ...
sections: # Required: List of summary sections
  - title: string # Required: Section title
    fields: # Required: List of fields to display
      - key: string # Required: State key to display
        label: string # Optional: Custom label (defaults to field title from original step)
        format: string # Optional: Format string for the value
        condition: string # Optional: Expression that determines if field is shown
editable: boolean # Optional: Whether fields can be edited (default: false)
```

### Action Step

An action step performs operations without user input:

```yaml
id: string
type: action # Required: Specifies this is an action step
# ... base step properties ...
actions: # Required: List of actions to perform
  - type: string # Required: Action type (function, http, etc.)
    name: string # Required: Action name or function identifier
    params: # Optional: Parameters for the action
      param1: value1
      param2: value2
    output_key: string # Optional: State key to store the result
show_progress: boolean # Optional: Whether to show progress indicators (default: true)
auto_proceed: boolean # Optional: Whether to automatically proceed after actions complete (default: true)
error_handling: # Optional: How to handle errors
  on_error: string # Optional: "stop" (default), "continue", or "retry"
  max_retries: integer # Optional: Maximum number of retries
  retry_delay: integer # Optional: Delay between retries in seconds
```

## Navigation and Flow Control

The Wizard DSL provides several ways to control the flow between steps:

### Linear Flow

By default, steps are executed in the order they are defined. A step can specify the next step explicitly:

```yaml
id: step1
# ... step properties ...
next_step: step3 # Skip step2 and go directly to step3
```

### Conditional Navigation

Navigation can be determined by a callback function:

```yaml
id: step1
# ... step properties ...
callbacks:
  navigation: determineNextStep # Function name to call
```

The referenced function should return the ID of the next step to navigate to.

### Branching Based on Conditions

Steps can be skipped based on conditions:

```yaml
id: step2
# ... step properties ...
skip_condition: "state.advanced_mode == false" # Skip if not in advanced mode
```

### Decision-Based Branching

Decision steps allow users to choose different paths:

```yaml
id: choose_path
type: decision
# ... step properties ...
options:
  - label: "Basic Setup"
    value: "basic"
  - label: "Advanced Setup"
    value: "advanced"
target_key: setup_type
next_step_map:
  basic: basic_setup_step
  advanced: advanced_setup_step
```

## Callbacks and Functions

Callbacks are references to functions that are registered programmatically. They allow for custom logic and integrations with external systems.

### Callback Registration

Callbacks are registered programmatically before wizard execution:

```go
wizard.RegisterCallback("validateEmail", func(ctx context.Context, state map[string]interface{}) (interface{}, error) {
    email, ok := state["email"].(string)
    if !ok {
        return nil, errors.New("email not found in state")
    }
    // Validate email
    valid := strings.Contains(email, "@")
    return valid, nil
})
```

### Callback Types

The Wizard DSL supports several types of callbacks:

1. **Before Step Callbacks**: Executed before a step is shown
2. **After Step Callbacks**: Executed after a step is completed
3. **Validation Callbacks**: Custom validation logic
4. **Navigation Callbacks**: Determine the next step dynamically

### Callback Results

Callbacks can return various results that affect wizard behavior:

1. **Boolean**: Used for validation results
2. **String**: Step ID for navigation callbacks
3. **Object**: Data to merge into the wizard state
4. **Error**: Indicates a failure that should be handled

### Example Callback Usage

```yaml
id: contact_info
type: form
# ... step properties ...
callbacks:
  validation: validateContactInfo # Validate the contact info before proceeding
  after: processContactInfo # Process the contact info after step completion
  navigation: determineNextStep # Decide which step to go to next
```

## State Management

The wizard maintains a state object that persists data across steps:

### Global State

Global state is accessible throughout the entire wizard:

```yaml
global_state:
  user_id: ""
  session_token: ""
```

### Step State

Each step contributes to the wizard state:

```yaml
id: project_basics
type: form
form:
  groups:
    - fields:
        - type: input
          key: project_name # This creates/updates state.project_name
```

### Accessing State in Expressions

State values can be accessed in conditional expressions:

```yaml
skip_condition: "state.project_type == 'simple'"
```

## Examples

### Project Setup Wizard

```yaml
name: Project Setup Wizard
description: Create and configure a new software project
theme: Dracula
save_progress: true
steps:
  - id: welcome
    type: info
    title: Welcome to the Project Setup Wizard
    description: This wizard will guide you through setting up a new software project.
    content: |
      # Welcome to the Project Setup Wizard

      This wizard will help you:

      - Create a new project structure
      - Configure build settings
      - Set up version control
      - Initialize dependencies

      Click **Next** to begin.

  - id: project_type
    type: decision
    title: Select Project Type
    description: Choose the type of project you want to create
    options:
      - label: "Web Application"
        value: "web"
        description: "A full-stack web application"
      - label: "Library/SDK"
        value: "library"
        description: "A reusable library or SDK"
      - label: "CLI Tool"
        value: "cli"
        description: "A command-line interface tool"
    target_key: project_type
    callbacks:
      after: initializeProjectDefaults

  - id: project_basics
    type: form
    title: Project Basics
    form:
      groups:
        - name: Basic Information
          fields:
            - type: input
              key: project_name
              title: Project Name
              validation:
                - condition: "len(value) < 3"
                  error: Project name must be at least 3 characters
            - type: input
              key: project_description
              title: Description
            - type: select
              key: language
              title: Primary Language
              options:
                - label: "JavaScript"
                  value: "js"
                - label: "Python"
                  value: "py"
                - label: "Go"
                  value: "go"
                - label: "Rust"
                  value: "rs"
            - type: confirm
              key: advanced_setup
              title: Use Advanced Setup?
              value: false
    callbacks:
      after: generateProjectId

  - id: web_config
    type: form
    title: Web Application Configuration
    skip_condition: "state.project_type != 'web'"
    form:
      groups:
        - name: Frontend
          fields:
            - type: select
              key: frontend_framework
              title: Frontend Framework
              options:
                - label: "React"
                  value: "react"
                - label: "Vue"
                  value: "vue"
                - label: "Angular"
                  value: "angular"
        - name: Backend
          fields:
            - type: select
              key: backend_framework
              title: Backend Framework
              options:
                - label: "Express"
                  value: "express"
                - label: "Django"
                  value: "django"
                - label: "Gin"
                  value: "gin"

  - id: advanced_settings
    type: form
    title: Advanced Settings
    skip_condition: "state.advanced_setup == false"
    form:
      groups:
        - name: Build Configuration
          fields:
            - type: multiselect
              key: build_features
              title: Build Features
              options:
                - label: "Minification"
                  value: "minify"
                - label: "Source Maps"
                  value: "sourcemap"
                - label: "TypeScript"
                  value: "typescript"
        - name: Testing
          fields:
            - type: select
              key: test_framework
              title: Test Framework
              options:
                - label: "Jest"
                  value: "jest"
                - label: "Mocha"
                  value: "mocha"
                - label: "Pytest"
                  value: "pytest"

  - id: summary
    type: summary
    title: Project Summary
    description: Review your project configuration
    sections:
      - title: Project Information
        fields:
          - key: project_name
            label: Name
          - key: project_description
            label: Description
          - key: language
            label: Language
          - key: project_type
            label: Project Type
      - title: Configuration
        fields:
          - key: frontend_framework
            label: Frontend Framework
            condition: "state.project_type == 'web'"
          - key: backend_framework
            label: Backend Framework
            condition: "state.project_type == 'web'"
          - key: build_features
            label: Build Features
            condition: "state.advanced_setup == true"
          - key: test_framework
            label: Test Framework
            condition: "state.advanced_setup == true"
    editable: true

  - id: create_project
    type: action
    title: Creating Project
    description: Setting up your project...
    actions:
      - type: function
        name: createProjectDirectory
        params:
          path: "{{state.project_path}}"
      - type: function
        name: initializeProject
        params:
          template: "{{state.project_type}}"
          language: "{{state.language}}"
      - type: function
        name: configureProject
    show_progress: true
    auto_proceed: true
    callbacks:
      after: logProjectCreation

  - id: completion
    type: info
    title: Project Created Successfully
    content: |
      # Congratulations!

      Your project **{{state.project_name}}** has been created successfully.

      ## Next Steps

      1. Navigate to the project directory: `cd {{state.project_path}}`
      2. Initialize version control: `git init`
      3. Install dependencies: `{{state.install_command}}`
      4. Start development server: `{{state.start_command}}`

      Thank you for using the Project Setup Wizard!
    navigation:
      next_label: Finish
      show_back: false
    callbacks:
      after: trackWizardCompletion
```

### User Registration Wizard

```yaml
name: User Registration
description: Register a new user account
theme: Catppuccin
steps:
  - id: welcome
    type: info
    title: Create Your Account
    content: |
      # Welcome to Our Service

      Creating an account is quick and easy. We'll need some basic information to get you started.

      Your privacy is important to us. We only collect the information necessary to provide our services.

  - id: account_type
    type: decision
    title: Account Type
    description: Choose the type of account you want to create
    options:
      - label: "Personal"
        value: "personal"
        description: "For individual use"
      - label: "Business"
        value: "business"
        description: "For organizations and teams"
    target_key: account_type

  - id: personal_info
    type: form
    title: Personal Information
    form:
      groups:
        - fields:
            - type: input
              key: full_name
              title: Full Name
              validation:
                - condition: "len(value) < 2"
                  error: Please enter your full name
            - type: input
              key: email
              title: Email Address
              validation:
                - condition: "!strings.Contains(value, '@')"
                  error: Please enter a valid email address
            - type: input
              key: phone
              title: Phone Number (optional)
    callbacks:
      validation: validatePersonalInfo

  - id: business_info
    type: form
    title: Business Information
    skip_condition: "state.account_type != 'business'"
    form:
      groups:
        - fields:
            - type: input
              key: company_name
              title: Company Name
            - type: input
              key: business_email
              title: Business Email
              validation:
                - condition: "!strings.Contains(value, '@')"
                  error: Please enter a valid email address
            - type: input
              key: business_phone
              title: Business Phone
            - type: input
              key: tax_id
              title: Tax ID / VAT Number (optional)

  - id: credentials
    type: form
    title: Create Your Login Credentials
    form:
      groups:
        - fields:
            - type: input
              key: username
              title: Username
              validation:
                - condition: "len(value) < 4"
                  error: Username must be at least 4 characters
            - type: input
              key: password
              title: Password
              attributes:
                echo_mode: password
              validation:
                - condition: "len(value) < 8"
                  error: Password must be at least 8 characters
            - type: input
              key: password_confirm
              title: Confirm Password
              attributes:
                echo_mode: password
              validation:
                - condition: "value != state.password"
                  error: Passwords do not match
    callbacks:
      validation: validateCredentials
      after: checkUsernameAvailability

  - id: verification
    type: action
    title: Verifying Information
    description: Please wait while we verify your information...
    actions:
      - type: function
        name: sendVerificationEmail
        params:
          email: "{{state.email}}"
        output_key: verification_code
    show_progress: true
    auto_proceed: false
    navigation:
      next_label: "Continue to Verification"

  - id: verify_code
    type: form
    title: Email Verification
    description: Please check your email for a verification code
    form:
      groups:
        - fields:
            - type: input
              key: verification_input
              title: Enter Verification Code
              validation:
                - condition: "value != state.verification_code"
                  error: Invalid verification code
    callbacks:
      validation: validateVerificationCode

  - id: preferences
    type: form
    title: Account Preferences
    form:
      groups:
        - name: Notification Preferences
          fields:
            - type: checkbox
              key: email_notifications
              title: Receive Email Notifications
              value: true
            - type: checkbox
              key: sms_notifications
              title: Receive SMS Notifications
              value: false
              visible_condition: "state.phone != ''"
        - name: Privacy Settings
          fields:
            - type: select
              key: privacy_level
              title: Profile Visibility
              options:
                - label: "Public"
                  value: "public"
                - label: "Private"
                  value: "private"
                - label: "Friends Only"
                  value: "friends"

  - id: summary
    type: summary
    title: Review Your Information
    description: Please review your information before creating your account
    sections:
      - title: Account Information
        fields:
          - key: full_name
          - key: email
          - key: phone
            condition: "state.phone != ''"
          - key: company_name
            condition: "state.account_type == 'business'"
          - key: business_email
            condition: "state.account_type == 'business'"
      - title: Account Settings
        fields:
          - key: username
          - key: privacy_level
          - key: email_notifications
            label: Email Notifications
            format: "{{value ? 'Enabled' : 'Disabled'}}"
          - key: sms_notifications
            label: SMS Notifications
            format: "{{value ? 'Enabled' : 'Disabled'}}"
            condition: "state.phone != ''"
    editable: true

  - id: create_account
    type: action
    title: Creating Your Account
    description: Please wait while we set up your account...
    actions:
      - type: function
        name: createUserAccount
        params:
          userData: "{{state}}"
        output_key: user_id
      - type: function
        name: initializeUserProfile
        params:
          user_id: "{{state.user_id}}"
    show_progress: true
    callbacks:
      after: logAccountCreation

  - id: completion
    type: info
    title: Account Created Successfully
    content: |
      # Welcome aboard, {{state.full_name}}!

      Your account has been created successfully. You can now log in using your username and password.

      ## Next Steps

      - Complete your profile
      - Explore our features
      - Connect with others

      [Log In Now](#)
    navigation:
      next_label: "Complete"
      show_back: false
```

## Command Format for CLI Execution

The Wizard DSL supports a command format that allows wizards to be executed directly from the command line interface (CLI).

```yaml
name: string # Required: Unique identifier for the command
short: string # Required: Short description of the command
wizard:# Required:
  Wizard definition
  # ... wizard definition ...
```

This command format allows you to execute wizards from the CLI using the following syntax:

```bash
uhoh run-wizard wizard.yaml [options]
```

## Expr Language Integration

The Wizard DSL utilizes the `@Expr` language (from `expr-lang/expr`) for conditional logic and dynamic values within various fields like `skip_condition`, `visible_condition`, `next_enabled_condition`, validation `condition`s, and potentially within template strings.

### Accessing State

You can access the wizard's state within expressions using the `state` prefix. For example:

```yaml
skip_condition: "state.advanced_mode == false"
visible_condition: "state.user_role == 'admin'"
```

### Using Functions

`@Expr` allows registering and using custom functions. Built-in functions like `len()` are also available:

```yaml
skip_condition: "len(state.selected_items) == 0"
next_enabled_condition: "isValid(state.email)" # Assuming isValid is registered
validation:
  - condition: "len(value) < 3"
```

### Template Strings

Template strings allow embedding state values directly into fields like `title` or `description`. The exact syntax (e.g., `{{ }}`) might depend on the implementation layer that processes the DSL before passing conditions to `@Expr`.

```yaml
title: "Welcome, {{state.user_name}}"
description: "You selected {{len(state.selected_items)}} items."
format: "{{value ? 'Enabled' : 'Disabled'}}" # Ternary operator example
```

Refer to the official `@Expr` documentation for the full syntax and available features.

## Conclusion

The Uhoh Wizard DSL provides a powerful, flexible system for creating multi-step interactive wizards. By combining form elements with conditional logic (using `@Expr`), callbacks, and dynamic navigation, it enables the creation of sophisticated guided workflows for users.
