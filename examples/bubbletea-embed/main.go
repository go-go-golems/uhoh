package main

import (
	"embed"
	"fmt"
	"log"
	"math/rand"
	"time"

	bubblespinner "github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	uhoh "github.com/go-go-golems/uhoh/pkg"
)

//go:embed form.yaml
var formFS embed.FS

type parentModel struct {
	form       *huh.Form
	formValues map[string]interface{}
	result     map[string]interface{}
	err        error

	info infoModel

	focused string // "form" or "info"

	spin     bubblespinner.Model
	spinning bool
	width    int
	height   int
}

type infoModel struct {
	width  int
	height int
	count  int
	last   string
}

func (i infoModel) Init() tea.Cmd { return nil }

func (i infoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		i.width = m.Width
		i.height = m.Height
	case tea.KeyMsg:
		switch m.String() {
		case "j":
			i.count++
			i.last = "j pressed"
		case "k":
			i.count--
			i.last = "k pressed"
		default:
			i.last = fmt.Sprintf("key: %s", m.String())
		}
	}
	return i, nil
}

func (i infoModel) View() string {
	return fmt.Sprintf("Info Panel\nCount: %d\nLast: %s\n(Use j/k, Tab to switch, q to quit)\n", i.count, i.last)
}

type toggleSpinnerMsg struct{}

func scheduleNextSpinnerToggle() tea.Cmd {
	// Random delay between 1s and 3s
	d := time.Duration(1000+rand.Intn(2000)) * time.Millisecond
	return tea.Tick(d, func(time.Time) tea.Msg { return toggleSpinnerMsg{} })
}

func NewParentModel() (parentModel, error) {
	b, err := formFS.ReadFile("form.yaml")
	if err != nil {
		return parentModel{}, err
	}
	f, vals, err := uhoh.BuildBubbleTeaModelFromYAML(b)
	if err != nil {
		return parentModel{}, err
	}
	sp := bubblespinner.New()
	sp.Spinner = bubblespinner.Line
	return parentModel{
		form:       f,
		formValues: vals,
		info:       infoModel{},
		focused:    "form",
		spin:       sp,
	}, nil
}

func (m parentModel) Init() tea.Cmd {
	rand.Seed(time.Now().UnixNano())
	return tea.Batch(
		m.form.Init(),
		m.info.Init(),
		scheduleNextSpinnerToggle(),
	)
}

func (m parentModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch t := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = t.Width, t.Height
		// forward size to children
		if fm, c := m.form.Update(t); c != nil || fm != nil {
			if f, ok := fm.(*huh.Form); ok {
				m.form = f
			}
			cmds = append(cmds, c)
		}
		if im, c := m.info.Update(t); c != nil || im != nil {
			if ii, ok := im.(infoModel); ok {
				m.info = ii
			}
			cmds = append(cmds, c)
		}

	case tea.KeyMsg:
		switch t.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			if m.focused == "form" {
				m.focused = "info"
			} else {
				m.focused = "form"
			}
			// no cmd
		case "f1":
			m.focused = "form"
		case "f2":
			m.focused = "info"
		}

	case toggleSpinnerMsg:
		// toggle state and schedule next change
		m.spinning = !m.spinning
		if m.spinning {
			// kick off spinner ticking
			cmds = append(cmds, m.spin.Tick)
		}
		cmds = append(cmds, scheduleNextSpinnerToggle())
	}

	// Always update spinner with all messages to progress animation
	var sc tea.Cmd
	m.spin, sc = m.spin.Update(msg)
	if sc != nil {
		cmds = append(cmds, sc)
	}

	// Forward to focused child for non-size messages (size already forwarded)
	if _, isSize := msg.(tea.WindowSizeMsg); !isSize {
		switch m.focused {
		case "form":
			fm, c := m.form.Update(msg)
			if f, ok := fm.(*huh.Form); ok {
				m.form = f
			}
			if c != nil {
				cmds = append(cmds, c)
			}
		case "info":
			im, c := m.info.Update(msg)
			if ii, ok := im.(infoModel); ok {
				m.info = ii
			}
			if c != nil {
				cmds = append(cmds, c)
			}
		}
	}

	// Detect form completion to extract results
	if m.form.State == huh.StateCompleted && m.result == nil {
		res, err := uhoh.ExtractFinalValues(m.formValues)
		if err != nil {
			m.err = err
		}
		m.result = res
	}

	return m, tea.Batch(cmds...)
}

func (m parentModel) View() string {
	header := fmt.Sprintf("[F1] Form  [F2] Info  [Tab] Switch  [q/Ctrl+C] Quit  Spinner:%v\n",
		func() string {
			if m.spinning {
				return m.spin.View()
			} else {
				return "-"
			}
		}())

	if m.err != nil {
		return header + fmt.Sprintf("Error: %v\n", m.err)
	}

	body := ""
	switch m.focused {
	case "form":
		if m.result != nil {
			body = fmt.Sprintf("Form completed! Values: %v\n(Press Tab to view Info panel, q to quit)\n", m.result)
		} else {
			body = m.form.View()
		}
	case "info":
		body = m.info.View()
	}
	return header + body
}

func main() {
	m, err := NewParentModel()
	if err != nil {
		log.Fatal(err)
	}
	if _, err := tea.NewProgram(m, tea.WithAltScreen()).Run(); err != nil {
		log.Fatal(err)
	}
}
