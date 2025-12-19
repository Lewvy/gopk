package tui

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lewvy/gopk/cmd/internal/data"
	"github.com/lewvy/gopk/cmd/internal/service"
)

type installFinishedMsg struct {
	err error
}

type model struct {
	choices       []data.Package
	cursor        int
	selected      map[int]string
	q             *data.Queries
	spinner       spinner.Model
	installing    bool
	statusMessage string
	// err      error
}

func (m model) View() string {
	var s strings.Builder
	s.WriteString("What should we buy at the market?\n\n")
	if m.installing {
		s := fmt.Sprintf("\n %s Installing packages...\n\n", m.spinner.View())

		// Add this check here so the warning shows up while spinning
		if m.statusMessage != "" {
			s += lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMessage) + "\n"
		}
		return s
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

	s.WriteString("\nPress q to quit.\n")

	return s.String()
}

func initialModel(q *data.Queries) model {

	packages, err := service.List(q, -1, false)
	if err != nil {
		log.Fatalf("error retreiving packages: %q", err)
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		choices:  packages,
		selected: make(map[int]string),
		spinner:  s,
		q:        q,
	}
}
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if !m.installing {
				return m, tea.Quit
			}
			m.statusMessage = "Wait for the packages to be installed or press ctrl+c to force quit"

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}
		case "i":

			if len(m.selected) > 0 {
				m.installing = true
				pkgs := []string{}
				for _, pkg := range m.selected {
					pkgs = append(pkgs, pkg)
				}
				return m, tea.Batch(installPackagesCmd(pkgs, m.q), m.spinner.Tick)
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

	case spinner.TickMsg:
		if m.installing {
			var spinCmd tea.Cmd
			m.spinner, spinCmd = m.spinner.Update(msg)
			return m, spinCmd
		}

	}

	return m, nil
}

func installPackagesCmd(pkgs []string, q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		err := service.Get(pkgs, true, q)

		return installFinishedMsg{err: err}
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func Start(q *data.Queries) error {
	p := tea.NewProgram(initialModel(q))
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil

}
