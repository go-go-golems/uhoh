package shellcmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type streamKind int

const (
	streamStdout streamKind = iota
	streamStderr
)

type status int

const (
	statusInit status = iota
	statusRunning
	statusSucceeded
	statusFailed
)

// Options configures the shell viewer behaviour.
type Options struct {
	WorkDir    string
	Env        []string
	Title      string
	KeepOpen   bool
	ProgramOps []tea.ProgramOption
}

// Result captures the outcome of a shell command run via the viewer.
type Result struct {
	Cmd       string
	ExitCode  int
	Output    string
	Duration  time.Duration
	Err       error
	StartedAt time.Time
	EndedAt   time.Time
}

type appendOutputMsg struct {
	stream streamKind
	text   string
}

type streamClosedMsg struct {
	stream streamKind
}

type streamErrMsg struct {
	stream streamKind
	err    error
}

type commandFinishedMsg struct {
	exitCode int
	err      error
}

type commandStartFailedMsg struct {
	err error
}

type tickMsg struct{}

// Run launches the Bubble Tea viewer and blocks until the command completes or the program exits.
func Run(ctx context.Context, cmdStr string, opts Options) (*Result, error) {
	m := newModel(ctx, cmdStr, opts)
	programOpts := []tea.ProgramOption{tea.WithAltScreen()}
	if ctx != nil {
		programOpts = append(programOpts, tea.WithContext(ctx))
	}
	if len(opts.ProgramOps) > 0 {
		programOpts = append(programOpts, opts.ProgramOps...)
	}

	p := tea.NewProgram(m, programOpts...)
	finalModel, err := p.Run()
	if err != nil {
		return nil, err
	}

	vm, ok := finalModel.(*model)
	if !ok {
		return nil, fmt.Errorf("unexpected final model type %T", finalModel)
	}

	res := &Result{
		Cmd:       cmdStr,
		ExitCode:  vm.exitCode,
		Output:    vm.plainBuffer.String(),
		Duration:  vm.endedAt.Sub(vm.startedAt),
		Err:       vm.err,
		StartedAt: vm.startedAt,
		EndedAt:   vm.endedAt,
	}

	if vm.status == statusSucceeded {
		res.Err = nil
	} else if vm.err == nil && vm.exitCode != 0 {
		res.Err = fmt.Errorf("command exited with code %d", vm.exitCode)
	}

	return res, nil
}

type model struct {
	ctx    context.Context
	cmdStr string
	opts   Options
	cmd    *exec.Cmd

	stdoutReader *bufio.Reader
	stderrReader *bufio.Reader

	viewport viewport.Model
	spinner  spinner.Model

	status status

	startedAt time.Time
	endedAt   time.Time

	exitCode int
	err      error

	stdoutClosed bool
	stderrClosed bool

	plainBuffer strings.Builder
	styledBuf   strings.Builder

	mu sync.Mutex

	keepOpen bool
	title    string
}

func newModel(ctx context.Context, cmdStr string, opts Options) *model {
	vp := viewport.New(0, 0)
	vp.SetContent("")
	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return &model{
		ctx:      ctx,
		cmdStr:   cmdStr,
		opts:     opts,
		viewport: vp,
		spinner:  sp,
		status:   statusInit,
		keepOpen: opts.KeepOpen,
		title:    opts.Title,
	}
}

func (m *model) Init() tea.Cmd {
	if err := m.startCommand(); err != nil {
		return func() tea.Msg {
			return commandStartFailedMsg{err: err}
		}
	}

	cmds := []tea.Cmd{m.spinner.Tick}
	cmds = append(cmds, m.readStdoutCmd(), m.readStderrCmd(), m.waitCmd(), m.tickCmd())
	return tea.Batch(cmds...)
}

func (m *model) startCommand() error {
	m.status = statusRunning
	m.startedAt = time.Now()

	cmd := exec.CommandContext(m.ctx, "bash", "-lc", m.cmdStr)
	if m.opts.WorkDir != "" {
		cmd.Dir = m.opts.WorkDir
	}
	if len(m.opts.Env) > 0 {
		cmd.Env = append(os.Environ(), m.opts.Env...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	m.cmd = cmd
	m.stdoutReader = bufio.NewReader(stdout)
	m.stderrReader = bufio.NewReader(stderr)
	return nil
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case commandStartFailedMsg:
		m.status = statusFailed
		m.err = msg.err
		m.exitCode = exitStatusFromError(msg.err)
		m.endedAt = time.Now()
		return m, tea.Quit

	case appendOutputMsg:
		m.appendOutput(msg)
		switch msg.stream {
		case streamStdout:
			if m.stdoutReader != nil {
				return m, m.readStdoutCmd()
			}
		case streamStderr:
			if m.stderrReader != nil {
				return m, m.readStderrCmd()
			}
		}
		return m, nil

	case streamClosedMsg:
		if msg.stream == streamStdout {
			m.stdoutReader = nil
			m.stdoutClosed = true
		} else {
			m.stderrReader = nil
			m.stderrClosed = true
		}
		return m, nil

	case streamErrMsg:
		if msg.err != nil && !errors.Is(msg.err, io.EOF) && m.err == nil {
			m.err = msg.err
		}
		switch msg.stream {
		case streamStdout:
			return m, m.readStdoutCmd()
		case streamStderr:
			return m, m.readStderrCmd()
		}
		return m, nil

	case commandFinishedMsg:
		m.exitCode = msg.exitCode
		if msg.err != nil {
			m.err = msg.err
		}
		if msg.err == nil && msg.exitCode == 0 {
			m.status = statusSucceeded
		} else {
			m.status = statusFailed
		}
		m.endedAt = time.Now()
		if !m.keepOpen {
			return m, tea.Quit
		}
		return m, nil

	case tickMsg:
		if m.status == statusRunning {
			return m, m.tickCmd()
		}
		return m, nil

	case tea.WindowSizeMsg:
		width := msg.Width
		height := msg.Height
		if height < 5 {
			height = 5
		}
		m.viewport.Width = width
		m.viewport.Height = height - 4
		return m, nil

	case tea.KeyMsg:
		if m.status == statusRunning {
			switch msg.String() {
			case "ctrl+c":
				if m.cmd != nil && m.cmd.Process != nil {
					_ = m.cmd.Process.Signal(os.Interrupt)
				}
				return m, nil
			}
		} else {
			switch msg.String() {
			case "q", "enter", "esc", "ctrl+c":
				return m, tea.Quit
			}
		}
		return m, nil

	case spinner.TickMsg:
		if m.status == statusRunning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m *model) readStdoutCmd() tea.Cmd {
	if m.stdoutReader == nil {
		return nil
	}
	reader := m.stdoutReader
	return func() tea.Msg {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			return appendOutputMsg{stream: streamStdout, text: line}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return streamClosedMsg{stream: streamStdout}
			}
			return streamErrMsg{stream: streamStdout, err: err}
		}
		return streamClosedMsg{stream: streamStdout}
	}
}

func (m *model) readStderrCmd() tea.Cmd {
	if m.stderrReader == nil {
		return nil
	}
	reader := m.stderrReader
	return func() tea.Msg {
		line, err := reader.ReadString('\n')
		if len(line) > 0 {
			return appendOutputMsg{stream: streamStderr, text: line}
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return streamClosedMsg{stream: streamStderr}
			}
			return streamErrMsg{stream: streamStderr, err: err}
		}
		return streamClosedMsg{stream: streamStderr}
	}
}

func (m *model) waitCmd() tea.Cmd {
	if m.cmd == nil {
		return nil
	}
	cmd := m.cmd
	return func() tea.Msg {
		err := cmd.Wait()
		return commandFinishedMsg{
			exitCode: exitStatusFromError(err),
			err:      err,
		}
	}
}

func (m *model) tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg { return tickMsg{} })
}

func (m *model) View() string {
	headerTitle := m.title
	if headerTitle == "" {
		headerTitle = fmt.Sprintf("Running: %s", m.cmdStr)
	}
	header := lipgloss.NewStyle().Bold(true).Render(headerTitle)

	statusLine := m.renderStatusLine()
	body := m.viewport.View()

	footer := m.renderFooter()

	return lipgloss.JoinVertical(lipgloss.Left, header, statusLine, body, footer)
}

func (m *model) appendOutput(msg appendOutputMsg) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.plainBuffer.WriteString(msg.text)

	var styledText string
	switch msg.stream {
	case streamStdout:
		styledText = msg.text
	case streamStderr:
		styledText = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Render(msg.text)
	}

	m.styledBuf.WriteString(styledText)
	m.viewport.SetContent(m.styledBuf.String())
	m.viewport.GotoBottom()
}

func (m *model) renderStatusLine() string {
	elapsed := time.Since(m.startedAt)
	switch m.status {
	case statusInit:
		return lipgloss.NewStyle().Render("Preparing command…")
	case statusRunning:
		return lipgloss.NewStyle().Render(fmt.Sprintf("%s  elapsed: %s", m.spinner.View(), elapsed.Truncate(time.Second)))
	case statusSucceeded:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render(fmt.Sprintf("✔ completed in %s", m.endedAt.Sub(m.startedAt).Truncate(time.Millisecond)))
	case statusFailed:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("✖ failed (exit %d) in %s", m.exitCode, m.endedAt.Sub(m.startedAt).Truncate(time.Millisecond)))
	}
	return ""
}

func (m *model) renderFooter() string {
	switch m.status {
	case statusInit, statusRunning:
		return "Press Ctrl+C to interrupt."
	case statusSucceeded, statusFailed:
		return "Press q or Enter to close."
	}
	return ""
}

func exitStatusFromError(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if status, ok := exitErr.Sys().(interface{ ExitStatus() int }); ok {
			return status.ExitStatus()
		}
	}
	return 1
}
