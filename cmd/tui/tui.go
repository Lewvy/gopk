package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/lewvy/gopk/cmd/internal/service"
)

type installFinishedMsg struct {
	err error
}

type packageAddedMsg struct {
	err error
}

type packagesListMsg struct {
	packages []data.Package
}

type model struct {
	choices  []data.Package
	cursor   int
	selected map[int]string
	queries  *data.Queries
	spinner  spinner.Model

	installing bool
	adding     bool

	statusMessage string

	inputs     []textinput.Model
	focusIndex int

	installFlag bool
	forceFlag   bool
}

func initialModel(q *data.Queries) model {
	packages, err := service.List(q, -1, false)
	if err != nil {
		log.Fatalf("error retrieving packages: %q", err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "URL (e.g. github.com/charmbracelet/log)"
	inputs[0].Focus()
	inputs[0].CharLimit = 156
	inputs[0].Width = 50
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Name (optional, Enter to skip)"
	inputs[1].CharLimit = 50
	inputs[1].Width = 50

	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Version (optional, Enter to skip)"
	inputs[2].CharLimit = 20
	inputs[2].Width = 50

	return model{
		choices:     packages,
		selected:    make(map[int]string),
		spinner:     s,
		inputs:      inputs,
		focusIndex:  0,
		installFlag: false,
		adding:      false,
		queries:     q,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.adding {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "tab", "enter":

				if m.focusIndex < len(m.inputs)-1 {
					m.focusIndex++
					m.updateFocus()
					return m, nil
				}

				if m.focusIndex == len(m.inputs)-1 {
					url := m.inputs[0].Value()
					name := m.inputs[1].Value()
					version := m.inputs[2].Value()

					if url == "" {
						return m, nil
					}

					m.adding = false
					m.statusMessage = "Adding " + url + "..."

					m.resetForm()

					return m, addPackageCmd(m.queries, url, name, version, m.installFlag, m.forceFlag)
				}

			case "up", "shift+tab":
				if m.focusIndex > 0 {
					m.focusIndex--
					m.updateFocus()
					return m, nil
				}

			case "ctrl+g":
				m.installFlag = !m.installFlag
				return m, nil

			case "ctrl+f":
				m.forceFlag = !m.forceFlag
				return m, nil

			case "esc":
				m.adding = false
				m.resetForm()
				return m, nil

			}
		}

		cmds := make([]tea.Cmd, len(m.inputs))
		for i := range m.inputs {
			m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
		}

		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "esc", "q":
			if !m.installing {
				return m, tea.Quit
			}
			m.statusMessage = "Wait for installation to finish or use Ctrl+C to force quit."

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "a":
			m.adding = true
			m.resetForm()
			return m, textinput.Blink

		case "i":
			if len(m.selected) > 0 {
				m.installing = true
				m.statusMessage = ""
				pkgs := []string{}
				for _, pkg := range m.selected {
					pkgs = append(pkgs, pkg)
				}
				return m, tea.Batch(installPackagesCmd(pkgs, m.queries), m.spinner.Tick)
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = m.choices[m.cursor].Name
			}
		}

	case installFinishedMsg:
		m.installing = false
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			m.statusMessage = "Installation complete!"
		}
		return m, nil

	case packageAddedMsg:
		if msg.err != nil {
			m.statusMessage = "Error adding: " + msg.err.Error()
		} else {
			m.statusMessage = "Package added successfully!"
			return m, refreshListCmd(m.queries)
		}

	case packagesListMsg:
		m.choices = msg.packages
		return m, nil

	case spinner.TickMsg:
		if m.installing {
			var spinCmd tea.Cmd
			m.spinner, spinCmd = m.spinner.Update(msg)
			return m, spinCmd
		}
	}

	return m, nil
}

func (m model) View() string {
	var s strings.Builder

	if m.adding {
		s.WriteString("Add New Package\n\n")

		for i := range m.inputs {
			s.WriteString(m.inputs[i].View())
			if i < len(m.inputs)-1 {
				s.WriteRune('\n')
			}
		}

		installCheck := "[ ]"
		if m.installFlag {
			installCheck = "[x]"
		}

		forceCheck := "[ ]"
		if m.forceFlag {
			forceCheck = "[x]"
		}

		fmt.Fprintf(&s, "\n\n%s Install immediately (ctrl+g)", installCheck)
		fmt.Fprintf(&s, "\n%s Force update (ctrl+f)", forceCheck)
		s.WriteString("\n\n(esc to cancel, enter to next/submit)")
		return s.String()
	}

	s.WriteString("GOPK\n\n")
	if m.installing {
		fmt.Fprintf(&s, " %s Installing packages...\n\n", m.spinner.View())
		if m.statusMessage != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMessage) + "\n")
		}
		return s.String()
	}

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		fmt.Fprintf(&s, "%s [%s] %s %s %v\n", cursor, checked, choice.Name, choice.Url, choice.Freq.Int64)
	}

	s.WriteString("\nPress 'a' to add, 'i' to install, 'q' to quit.\n")

	if m.statusMessage != "" && !m.installing {
		s.WriteString("\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMessage) + "\n")
	}

	return s.String()
}

func (m *model) updateFocus() {
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIndex {
			m.inputs[i].Focus()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
		} else {
			m.inputs[i].Blur()
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		}
	}
}

func (m *model) resetForm() {
	for i := range m.inputs {
		m.inputs[i].Reset()
	}
	m.focusIndex = 0
	m.updateFocus()
}

func installPackagesCmd(pkgs []string, q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		err := service.Get(pkgs, true, q)
		return installFinishedMsg{err: err}
	}
}

func addPackageCmd(q *data.Queries, url, name, version string, install, force bool) tea.Cmd {
	return func() tea.Msg {
		err := service.Add(url, name, version, install, force, true, q)
		return packageAddedMsg{err: err}
	}
}

func refreshListCmd(q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		packages, _ := service.List(q, -1, false)
		return packagesListMsg{packages: packages}
	}
}

func Start(q *data.Queries) error {
	p := tea.NewProgram(initialModel(q))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
