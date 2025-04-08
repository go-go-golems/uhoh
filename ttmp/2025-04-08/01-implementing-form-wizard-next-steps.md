# Uhoh Wizard Feature: Overview and Next Steps

## 1. Purpose and Scope

The `uhoh` tool initially provided functionality to define and run single-page interactive forms using a YAML-based Domain Specific Language (DSL), specified in `uhoh/pkg/doc/topics/uhoh-dsl.md` and implemented in `uhoh/pkg/formdsl.go`.

The goal of the **Wizard Feature** is to extend this capability by introducing a new DSL (`uhoh/pkg/doc/topics/uhoh-wizard-dsl.md`) for defining **multi-step, potentially conditional wizards**. These wizards allow for more complex guided workflows, involving multiple steps (including different types like forms, decisions, actions), conditional navigation between steps, and the ability to execute custom Go functions (callbacks) at various points in the flow.

The ultimate aim is to have a robust system where users can define complex wizards in YAML and execute them via the `uhoh run-wizard` command.

## 2. Current Status (As of 2024-08-04)

A very basic foundation for the wizard feature has been implemented:

- **DSL Specification:** The target DSL is defined in `uhoh/pkg/doc/topics/uhoh-wizard-dsl.md`. This document outlines the structure, step types, callbacks, state management, and `@Expr` integration.
- **Core Structs:** Basic Go structs (`Wizard`, `Step`) are defined in `uhoh/pkg/wizard/wizard.go`. Currently, `Step` is just a `map[string]interface{}`, lacking proper type handling.
- **Basic Runner:** A minimal `Run` method exists in `wizard.go`. **Crucially, it only supports wizards with exactly one step, and that step must be of type `form`**. It works by extracting the form definition, marshalling/unmarshalling it into a `pkg.Form`, and then executing the existing `pkg.Form.Run` method.
- **Example:** A simple single-step wizard example exists at `uhoh/pkg/wizard/examples/single-step-form.yaml`.
- **CLI Command:** A new command `uhoh run-wizard <file>` has been added to `uhoh/cmd/uhoh/main.go` to load and run a wizard YAML file using the basic `Run` method.

**In summary:** We have the DSL design, a placeholder implementation that runs only the simplest possible case (single form step), and the CLI entry point. The core multi-step logic, different step types, callbacks, state management, and conditional execution are **not yet implemented**.

## 3. Key Concepts & Design (Target)

Refer to `uhoh/pkg/doc/topics/uhoh-wizard-dsl.md` for the complete design. Key elements include:

- **Wizard:** Top-level definition containing metadata, global state, and an ordered list of steps.
- **Steps:** Defined by `id`, `type`, `title`, etc. Supported types:
  - `form`: Reuses the existing `pkg.Form` logic.
  - `decision`: Presents choices to the user, potentially altering flow.
  - `summary`: Displays collected data for review.
  - `action`: Executes backend logic (e.g., API calls via callbacks) without direct user input.
  - `info`: Displays informational text.
- **State:** A map (`map[string]interface{}`) holding data collected across steps (`global_state` + step outputs).
- **Navigation:** Control flow via:
  - Default linear progression.
  - `skip_condition`: Skip steps based on state (`@Expr`).
  - `next_step`: Explicitly jump to a step ID.
  - `next_step_map`: Branching within `decision` steps.
  - `navigation` callback: Programmatically determine the next step.
- **Callbacks:** Named Go functions (registered programmatically) referenced in the YAML. Executed at specific points (`before`, `after`, `validation`, `navigation`) with access to the current wizard state.
- **@Expr Integration:** Use the `expr-lang/expr` library to evaluate conditions (`skip_condition`, `visible_condition`, validation conditions) based on the wizard state.

## 4. Next Steps (Implementation Roadmap)

The core task is to implement the logic described in the DSL specification within `uhoh/pkg/wizard/wizard.go` and potentially new files within that package.

1.  **[x] Implement Step Execution Loop:**

    - Modify `wizard.Run` to iterate through the `w.Steps` list.
    - Maintain a pointer to the current step index/ID.
    - Handle basic linear navigation (current -> next).

2.  **[x] Refactor Step Handling:**

    - Define Go structs for each step type (`FormStep`, `DecisionStep`, `ActionStep`, `InfoStep`, `SummaryStep`) inheriting common fields (ID, Title, etc.).
    - Implement custom YAML unmarshalling for the `Wizard.Steps` field to correctly parse the `type` field and unmarshal into the appropriate Go struct type. This avoids the current `map[string]interface{}` approach.

3.  **[x] Implement State Management:**

    - Initialize the wizard state map (potentially pre-populating with `global_state` from YAML).
    - Pass the state map to steps that need it.
    - Define how step outputs (e.g., form results, action results) update the state map.

4.  **[x] Integrate @Expr:**

    - Add `github.com/expr-lang/expr` as a dependency.
    - Create helper functions to evaluate `@Expr` conditions (`skip_condition`, `visible_condition`, etc.) against the current wizard state map.
    - Use this evaluation within the step execution loop for conditional logic.

5.  **[x] Implement Navigation Logic:**

    - Handle `skip_condition` evaluation before executing a step.
    - Determine the _next_ step based on (in order of precedence): `navigation` callback result, `next_step_map` (for `decision` steps), explicit `next_step` field, or default linear progression.

6.  **[ ] Implement Callback System:**

    - Design a mechanism to register Go functions (callbacks) by name (e.g., a `map[string]WizardCallbackFunc`).
    - Define the `WizardCallbackFunc` signature (e.g., `func(ctx context.Context, state map[string]interface{}) (result interface{}, nextStepID *string, err error)` - adapt as needed).
    - Call registered callbacks at appropriate times (`before`, `after`, `validation`, `navigation`) within the step lifecycle, passing the state and handling results/errors.

7.  **[ ] Implement Remaining Step Types:**

    - **Decision Step:** Implement logic to present options and store the result in `target_key`, using `next_step_map` for navigation.
    - **Action Step:** Implement logic to execute specified actions (initially focusing on `type: function` using the callback system). Handle `output_key` to store results in state.
    - **Info Step:** Display content.
    - **Summary Step:** Display formatted data from the state. Consider how `editable: true` would work (might require navigating back).

8.  **[ ] Improve Error Handling:** Add more robust error handling throughout the `Run` loop.

9.  **[ ] Add Tests:** Implement unit tests for step execution, state management, callbacks, and navigation logic. Add integration tests using example wizard YAML files.

10. **[ ] (Optional) Implement Progress Saving:** If needed, implement the `save_progress` feature.

## 5. Key Resources

- **Wizard DSL Specification:** `uhoh/pkg/doc/topics/uhoh-wizard-dsl.md` (Source of Truth for requirements)
- **Wizard Implementation:** `uhoh/pkg/wizard/wizard.go` (Main implementation file)
- **Form DSL Implementation:** `uhoh/pkg/formdsl.go` (Reused by `form` steps)
- **CLI Entrypoint:** `uhoh/cmd/uhoh/main.go` (Contains `run-wizard` command)
- **Example Wizard:** `uhoh/pkg/wizard/examples/single-step-form.yaml`

## 6. Getting Started (Current State)

1.  Build the binary:
    ```bash
    go build ./cmd/uhoh
    ```
2.  Run the basic single-step example:
    ```bash
    ./uhoh run-wizard uhoh/pkg/wizard/examples/single-step-form.yaml
    ```
    _(Note: This will only execute the single form step defined in the example)_.

Start by tackling the **Step Execution Loop** and **Refactor Step Handling** as foundational steps.
