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
	"github.com/sahilm/fuzzy"
)

type installFinishedMsg struct {
	err error
}

type packageAddedMsg struct {
	err error
}

type groupCreatedMsg struct {
	name string
	err  error
}

type groupAssignedMsg struct {
	group string
	count int
	err   error
}

type groupsListMsg struct {
	groups []string
}

type packagesListMsg struct {
	packages []data.Package
}

type packageSource []data.Package

func (p packageSource) String(i int) string { return p[i].Name }
func (p packageSource) Len() int            { return len(p) }

type model struct {
	choices  []data.Package
	filtered []data.Package
	cursor   int
	selected map[int]string
	queries  *data.Queries
	spinner  spinner.Model

	installing    bool
	adding        bool
	searching     bool
	assigning     bool
	creatingGroup bool
	debugFlag     bool

	statusMessage string

	inputs     []textinput.Model
	focusIndex int

	searchInput textinput.Model
	groupInput  textinput.Model

	groups      []string
	groupCursor int

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

	si := textinput.New()
	si.Placeholder = "Search packages..."
	si.CharLimit = 50
	si.Width = 30
	si.Prompt = "/"

	gi := textinput.New()
	gi.Placeholder = "New Group Name"
	gi.CharLimit = 30
	gi.Width = 30
	gi.Prompt = "Name: "

	return model{
		choices:       packages,
		filtered:      packages,
		selected:      make(map[int]string),
		spinner:       s,
		inputs:        inputs,
		searchInput:   si,
		groupInput:    gi,
		focusIndex:    0,
		installFlag:   false,
		adding:        false,
		searching:     false,
		assigning:     false,
		creatingGroup: false,
		queries:       q,
		groups:        []string{},
	}
}

func (m model) Init() tea.Cmd {
	return fetchGroupsCmd(m.queries)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	if m.creatingGroup {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter":
				name := m.groupInput.Value()
				if name != "" {
					m.creatingGroup = false
					m.groupInput.Reset()
					m.statusMessage = "Creating group " + name + "..."
					return m, createGroupCmd(m.queries, name)
				}
				m.creatingGroup = false
				return m, nil
			case "esc":
				m.creatingGroup = false
				m.groupInput.Reset()
				return m, nil
			}
		}
		m.groupInput, cmd = m.groupInput.Update(msg)
		return m, cmd
	}

	if m.assigning {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.groupCursor > 0 {
					m.groupCursor--
				}
			case "down", "j":
				if m.groupCursor < len(m.groups)-1 {
					m.groupCursor++
				}
			case "enter":
				if len(m.groups) == 0 {
					m.assigning = false
					return m, nil
				}
				group := m.groups[m.groupCursor]
				m.assigning = false

				pkgs := []string{}
				for _, p := range m.selected {
					pkgs = append(pkgs, p)
				}

				return m, assignToGroupCmd(m.queries, pkgs, group)

			case "esc":
				m.assigning = false
				return m, nil
			}
		}
		return m, nil
	}

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
					return m, addPackageCmd(m.queries, url, name, version, m.installFlag, m.forceFlag, m.debugFlag)
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

	if m.searching {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "enter", "esc":
				m.searching = false
				m.searchInput.Blur()
				return m, nil
			}
		}

		m.searchInput, cmd = m.searchInput.Update(msg)

		query := m.searchInput.Value()
		if query == "" {
			m.filtered = m.choices
		} else {
			matches := fuzzy.FindFrom(query, packageSource(m.choices))
			var results []data.Package
			for _, match := range matches {
				results = append(results, m.choices[match.Index])
			}
			m.filtered = results
		}
		m.cursor = 0
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			if !m.installing {
				return m, tea.Quit
			}
			m.statusMessage = "Wait for installation to finish or use Ctrl+C to force quit."

		case "d":
			m.debugFlag = !m.debugFlag
			m.statusMessage = fmt.Sprintf("Debug Mode: %v", m.debugFlag)
			return m, nil

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}

		case "+":
			m.adding = true
			m.resetForm()
			return m, textinput.Blink

		case "c":
			m.creatingGroup = true
			m.groupInput.Reset()
			m.groupInput.Focus()
			return m, textinput.Blink

		case "a":
			if len(m.selected) > 0 {
				m.assigning = true
				m.groupCursor = 0
				return m, fetchGroupsCmd(m.queries)
			}
			m.statusMessage = "Select packages first!"
			return m, nil

		case "/":
			m.searching = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink

		case "i":
			if len(m.selected) > 0 {
				m.installing = true
				m.statusMessage = ""
				pkgs := []string{}
				for _, pkg := range m.selected {
					pkgs = append(pkgs, pkg)
				}
				return m, tea.Batch(installPackagesCmd(pkgs, m.debugFlag, m.queries), m.spinner.Tick)
			}

		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = m.filtered[m.cursor].Name
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

	case groupCreatedMsg:
		if msg.err != nil {
			m.statusMessage = "Error creating group: " + msg.err.Error()
		} else {
			m.statusMessage = "Group '" + msg.name + "' created!"
			return m, fetchGroupsCmd(m.queries)
		}

	case groupAssignedMsg:
		if msg.err != nil {
			m.statusMessage = "Error assigning: " + msg.err.Error()
		} else {
			m.statusMessage = fmt.Sprintf("Assigned %d packages to '%s'", msg.count, msg.group)
			m.selected = make(map[int]string)
		}

	case groupsListMsg:
		m.groups = msg.groups
		return m, nil

	case packagesListMsg:
		m.choices = msg.packages
		m.filtered = msg.packages
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

	if m.creatingGroup {
		s.WriteString("Create New Group\n\n")
		s.WriteString(m.groupInput.View())
		s.WriteString("\n\n(esc to cancel, enter to create)")
		return s.String()
	}

	if m.assigning {
		s.WriteString("Assign to Group\n\n")
		if len(m.groups) == 0 {
			s.WriteString("No groups found. Press 'c' to create one.")
		} else {
			for i, group := range m.groups {
				cursor := " "
				if m.groupCursor == i {
					cursor = ">"
				}
				fmt.Fprintf(&s, "%s %s\n", cursor, group)
			}
		}
		s.WriteString("\n(esc to cancel, enter to assign)")
		return s.String()
	}

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

	s.WriteString("GOPK MANAGER\n\n")

	if m.installing {
		fmt.Fprintf(&s, " %s Installing packages...\n\n", m.spinner.View())
		if m.statusMessage != "" {
			s.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Render(m.statusMessage) + "\n")
		}
		return s.String()
	}

	statusStyle := lipgloss.NewStyle().Width(6).PaddingRight(1)
	nameStyle := lipgloss.NewStyle().Width(30).PaddingRight(2).Foreground(lipgloss.Color("205")).Bold(true)
	urlStyle := lipgloss.NewStyle().Width(50).PaddingRight(2).Foreground(lipgloss.Color("240"))
	freqStyle := lipgloss.NewStyle().Width(6).Align(lipgloss.Right).Foreground(lipgloss.Color("240"))

	header := lipgloss.JoinHorizontal(lipgloss.Left,
		statusStyle.Render(""),
		nameStyle.Render("PACKAGE"),
		urlStyle.Render("REPOSITORY URL"),
		freqStyle.Render("FREQ"),
	)

	headerStr := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("240")).
		Render(header)

	s.WriteString(headerStr + "\n")

	for i, choice := range m.filtered {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[i]; ok {
			checked = "x"
		}

		status := fmt.Sprintf("%s [%s]", cursor, checked)

		url := choice.Url
		if len(url) > 48 {
			url = url[:45] + "..."
		}

		row := lipgloss.JoinHorizontal(lipgloss.Left,
			statusStyle.Render(status),
			nameStyle.Render(choice.Name),
			urlStyle.Render(url),
			freqStyle.Render(fmt.Sprintf("%d", choice.Freq.Int64)),
		)

		if m.cursor == i {
			row = lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(row)
		}

		s.WriteString(row + "\n")
	}

	if m.searching {
		s.WriteString("\n" + m.searchInput.View())
	} else {
		helpText := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("\nPress '/' search, '+' add, 'a' assign, 'c' create group, 'i' install, 'q' quit.\n")
		s.WriteString(helpText)
	}

	if m.statusMessage != "" && !m.installing && !m.searching {
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
	m.installFlag = false
	m.forceFlag = false
	m.updateFocus()
}

func installPackagesCmd(pkgs []string, debug bool, q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		err := service.Get(pkgs, q)
		return installFinishedMsg{err: err}
	}
}

func addPackageCmd(q *data.Queries, url, name, version string, install, force, debug bool) tea.Cmd {
	return func() tea.Msg {
		err := service.Add(url, name, version, install, force, q)
		return packageAddedMsg{err: err}
	}
}

func createGroupCmd(q *data.Queries, name string) tea.Cmd {
	return func() tea.Msg {
		// err := service.CreateGroup(q, name)
		return groupCreatedMsg{name: name, err: nil}
	}
}

func assignToGroupCmd(q *data.Queries, pkgs []string, group string) tea.Cmd {
	return func() tea.Msg {
		// err := service.AssignToGroup(q, pkgs, group)
		return groupAssignedMsg{group: group, count: len(pkgs), err: nil}
	}
}

func fetchGroupsCmd(q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		// groups, err := service.ListGroups(q)
		groups := []string{"work", "personal", "dev-tools"}
		return groupsListMsg{groups: groups}
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
