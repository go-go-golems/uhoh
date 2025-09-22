# Design: Bubble Tea Shell Output for Onboarding Wizard

## Background & Current Behaviour

The onboarding CLI (see `go-go-mento/go/cmd/onboarding/main.go`) registers a `runCmd` action callback that shells out via `runShell(...)`. The helper executes `bash -lc <cmd>`, captures stdout and stderr into buffers, and optionally tees them directly to the process’ stdout/stderr whenever the `--stream` flag is set. After completion, the callback writes the combined output to a timestamped log file and surfaces the exit code back to the wizard state.

Action steps inside the wizard are driven by `uhoh/pkg/wizard/steps/action_step.go`. Each step spins up a `huh.NewNote()` in a goroutine (“Executing action…”) while the callback runs, then, once the callback exits, shows a second note for completion. No UI bridge exists between the callback logic and the Bubble Tea runtime that powers the wizard. As soon as a callback writes straight to stdout/stderr the `huh` TUI corrupts, because the buffers bypass Bubble Tea’s alt-screen rendering and interleave with the wizard view.

The net effect today:

- Shell commands run to completion, but interactive feedback requires `--stream`, which breaks the layout because the output is *not* rendered through a Bubble Tea model.
- While the first `huh` note is active, the user sees a frozen “Please wait…” screen; the actual command output either appears underneath the TUI (if streaming) or only in the log file afterward.
- There is no per-line progress feedback, no scroll buffer, and no way to abort long-running commands from inside the wizard.

## Problem Statement

We need to render command output (stdout + stderr) inside the wizard itself so users can monitor progress without blowing up the TUI. The design must:

1. Stream output incrementally inside a Bubble Tea component.
2. Preserve the log-file behaviour for later inspection.
3. Handle exit codes and propagate them back to the wizard state.
4. Fit within the existing uhoh wizard/action-step architecture without forking the framework.
5. Leave room for future improvements (cancellation, status badges, etc.).

## High-Level Approach

Introduce a dedicated Bubble Tea model that is responsible for running a shell command and rendering its output. Instead of teeing to `os.Stdout`, the `runCmd` callback will synchronously run this model (via `tea.NewProgram`) and collect the result. The wizard remains in control: the action step calls the callback, the callback blocks until the shell viewer exits, and only then does execution continue.

Key ideas:

- **Encapsulated tea.Model** – Build a `commandViewerModel` struct that starts the process, streams stdout/stderr into an internal buffer, and renders a viewport + status line. Use Charm `viewport` and `spinner` bubbles for UX consistency with other Go Go Golems tooling.
- **Message flow** – Spawn goroutines that read the command’s stdout/stderr pipes line-by-line (or chunk-by-chunk) and emit `appendOutputMsg` messages into the Tea loop. When the process terminates, send a `commandFinishedMsg` carrying the exit code and error (if any).
- **Log persistence** – The model collects the entire transcript in-memory. After the Tea program exits, the callback still writes the log file exactly once using the aggregated buffer, preserving the existing filesystem contract.
- **Config knobs** – Feed working directory, environment (future), and friendly title into the model. Optionally reuse the `--stream` flag to toggle entering the viewer automatically vs. headless execution + static summary.
- **Compatibility with ActionStep** – Because the model is rendered through Bubble Tea, it coexists with `huh`’s runtime. We should suppress the temporary “Executing action…” note when using the new viewer to avoid flicker—see the integration section below.

## Detailed Design

### 1. Bubble Tea Model (`commandViewerModel`)

State fields:

- `cmdStr`, `workDir`, `startTime`
- `buffer` (strings.Builder for log; maybe also a slice for viewport lines)
- `viewport` bubble configured with adaptive width/height (update on `tea.WindowSizeMsg`)
- `spinner` bubble for “running” state
- `status` enum: running / completed / failed
- `exitCode`, `err`

Lifecycle:

- `Init()` returns a `tea.Cmd` that spawns the subprocess and starts reading pipes.
- Two reader goroutines use `bufio.Scanner` to pull stdout/stderr. Each chunk is wrapped in `appendOutputMsg` (containing the text and stream tag) and sent via a channel -> `tea.Cmd` using `tea.Printf` or custom aggregator.
- On process completion, emit `commandFinishedMsg` with exit code, error, and combined transcript.
- `Update` handles:
  - `appendOutputMsg`: append to buffer, append to viewport (apply color/style per stream), scroll to bottom when auto-scroll enabled.
  - `commandFinishedMsg`: stop spinner, freeze viewport, record exit code/error, and return `tea.Quit` if we auto-close on success.
  - Keyboard shortcuts: `ctrl+c` sends kill signal (if we add cancellation), `q` quits after command is done, `pgup/pgdn` scroll viewport, etc.
- `View` renders header (command + cwd + elapsed + status), viewport body, footer hints.

### 2. Running the Model from `runCmd`

Inside `runCmd` (currently lines `136-180`):

1. Build the model with the resolved command + options.
2. Run it: `program := tea.NewProgram(model, tea.WithAltScreen())` then `resultModel, err := program.Run()`.
3. Extract the transcript + exit metadata from the returned model (expose an accessor or type-assert the final model).
4. If the Tea program itself errors, treat as command failure and bubble up `err`.
5. Persist the transcript to the log file (honouring custom path overrides).
6. Return `{cmd, exit_code, log_file}` as before.

This replaces `runShell` entirely; we can delete it or refactor it into utility pieces (e.g. `startCommandProcess` shared between the model and logger). The existing `--stream` flag becomes redundant; instead, consider:

- `--no-viewer`: skip Bubble Tea and fall back to headless execution (for CI scripts).
- `--keep-open`: keep viewer open after success until user presses a key.

Those switches can be wired later; the default behaviour should launch the viewer.

### 3. Adjusting `ActionStep` UX

The dual-note pattern in `ActionStep.Execute` (`uhoh/pkg/wizard/steps/action_step.go:56-110`) conflicts with our dedicated viewer. Options:

- Minimal change: have the callback signal that it fully handled UI. We could return a sentinel error (e.g., `ErrActionUIHandled`) or set a known key in the step result (`ui_handled: true`). The step checks this before showing the completion note.
- Cleaner change (preferred): extend `ActionStep` with a `ShowStatusNotes` boolean defaulting to true. For the wizard YAML we set it to false on steps that call `runCmd`. Implementation: new field + guard around the note logic.

Either approach keeps the scope inside the `uhoh` repo, but because `go-go-mento` drives the wizard we can start with the sentinel key (no YAML changes) and follow up with a formal option in uhoh if needed.

### 4. Error Handling & Logging

- If the command exits non-zero, mark status as failed, leave the viewer open with the exit code highlighted, and return an `error` from the callback (preserving current semantics).
- Always write the log file, even when the Tea program errors, so we retain the partial transcript.
- When the viewer is aborted (future feature), return a context cancellation error so the wizard can decide whether to retry or abort the flow.

### 5. Testing & Validation Strategy

- **Manual**: run the onboarding wizard locally and verify the viewer renders output for fast & slow commands, handles terminal resize, and preserves logs.
- **Unit-ish**: factor the pipe-reading logic into pure functions and drive with fake readers (in `*_test.go`). Verify multi-stream ordering and that log file contents match appended messages.
- **Regression**: ensure `--log-file` still works by pointing to a temp path and checking file contents.

## Migration Steps

1. Implement the `commandViewerModel` and supporting helpers under a new internal package (`go-go-mento/go/cmd/onboarding/internal/tea/shellviewer`, for example).
2. Replace `runShell` usage in `runCmd` with the Bubble Tea program run. Remove the `teeToStdout` parameter/flag or repurpose as discussed.
3. Update logging logic to operate on the viewer’s transcript.
4. Teach `ActionStep` to skip the default notes when the callback reports `ui_handled` (temporary shim) or add a core feature in uhoh to configure note behaviour.
5. Verify wizard YAML doesn’t require changes; if we add `ShowStatusNotes`, update relevant steps accordingly.
6. Document the new viewer behaviour in onboarding docs / README.

## Risks & Open Questions

- **Concurrency with huh runtime** – Running a nested Tea program while a huh note goroutine is active might still clash. Suppressing the note is important; otherwise, we need to ensure the note’s program exits before launching ours (e.g., by not starting it when UI is handled externally).
- **Terminal capabilities** – Some users might prefer plain logs (non-interactive). Provide an escape hatch to disable the viewer.
- **Command cancellation** – Not part of initial scope, but design should leave hooks (store `*exec.Cmd` inside the model for `ctrl+c`).
- **Log growth** – For very chatty commands the in-memory buffer might be large; consider streaming to file concurrently while keeping a bounded viewport buffer (e.g., last N lines) to avoid OOM.

---

This design keeps all user-visible rendering inside Bubble Tea, aligns with the `uhoh` philosophy of declarative TUIs, and confines changes to the onboarding command plus a small extension in `uhoh`’s action step plumbing.
