package tui

import (
	"context"
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

var (
	colorPrimary   = lipgloss.Color("205")
	colorSecondary = lipgloss.Color("240")
	colorSelected  = lipgloss.Color("151")
	colorCursorBg  = lipgloss.Color("236")
	colorCursorFg  = lipgloss.Color("255")
)

type installFinishedMsg struct {
	installedUrls []string
	err           error
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

type packagesRemovedMsg struct {
	err error
}

type groupsListMsg struct {
	groups []data.Group
	err    error
}

type packagesListMsg struct {
	packages []data.Package
}

type viewMode int
type sortMode int

const (
	packageView viewMode = iota
	groupView
	groupPackageView
)

const (
	sortByLastUsed sortMode = iota
	sortByFrequency
)

type packageSource []data.Package

func (p packageSource) String(i int) string {
	return fmt.Sprintf("%s %s", p[i].Name, p[i].Url)
}
func (p packageSource) Len() int { return len(p) }

type model struct {
	choices  []data.Package
	filtered []data.Package
	selected map[data.Package]struct{}
	queries  *data.Queries
	spinner  spinner.Model
	view     viewMode

	err error

	activeGroup data.Group

	installing    bool
	adding        bool
	searching     bool
	assigning     bool
	creatingGroup bool

	statusMessage string
	sm            sortMode

	inputs     []textinput.Model
	focusIndex int

	searchInput textinput.Model
	groupInput  textinput.Model

	groups []data.Group

	cursorGroup   int
	cursorPackage int

	installFlag bool
	forceFlag   bool
}

func initialModel(q *data.Queries) model {
	packages, err := service.List(q, -1, false)
	if err != nil {
		log.Printf("error retrieving packages: %v", err)
		packages = []data.Package{}
	}

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(colorPrimary)

	inputs := make([]textinput.Model, 3)

	inputs[0] = textinput.New()
	inputs[0].Placeholder = "URL (e.g. github.com/charmbracelet/log)"
	inputs[0].Focus()
	inputs[0].CharLimit = 156
	inputs[0].Width = 50
	inputs[0].PromptStyle = lipgloss.NewStyle().Foreground(colorPrimary)

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
		selected:      make(map[data.Package]struct{}),
		sm:            sortByLastUsed,
		spinner:       s,
		view:          packageView,
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
		groups:        []data.Group{},
	}
}

func (m model) Init() tea.Cmd {
	return fetchGroupsCmd(m.queries)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if key, ok := msg.(tea.KeyMsg); ok {
		if key.String() == "ctrl+c" {
			return m, tea.Quit
		}
	}

	if m.installing {
		switch msg := msg.(type) {
		case spinner.TickMsg:
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case installFinishedMsg, installGroupMsg:
		default:
			return m, nil
		}
	}

	if m.creatingGroup {
		return m.creatingGroupUpdate(msg)
	}
	if m.assigning {
		return m.assigningUpdate(msg)
	}
	if m.adding {
		return m.addingUpdate(msg)
	}
	if m.searching {
		return m.searchingUpdate(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "f":
			m.sm = sortByFrequency
			switch m.view {
			case packageView:
				return m, refreshListCmd(m.queries, m.sm)

			case groupPackageView:
				return m, sortGroupByFreq(context.Background(), m.queries, m.activeGroup.Name)

			default:
				return m, nil
			}

		case "l":
			m.sm = sortByLastUsed
			switch m.view {
			case packageView:
				return m, refreshListCmd(m.queries, m.sm)

			case groupPackageView:
				return m, sortGroupByLastUsed(context.Background(), m.queries, m.activeGroup.Name)
			}
			return m, nil

		case "d":
			if m.view == groupPackageView && len(m.selected) > 0 {
				m.statusMessage = "Removing packages from group..."
				return m, removePackagesFromGroups(m.queries, m.selected, m.activeGroup)
			}

		case "g":
			if m.view == packageView {
				m.view = groupView
				m.cursorGroup = 0
				return m, fetchGroupsCmd(m.queries)
			}

		case "q", "esc":
			switch m.view {
			case groupPackageView:
				m.view = groupView
				m.activeGroup = data.Group{}
				m.cursorPackage = 0
				m.selected = make(map[data.Package]struct{})
				return m, nil

			case groupView:
				m.view = packageView
				m.cursorPackage = 0
				return m, refreshListCmd(m.queries, m.sm)

			default:
				if !m.installing {
					return m, tea.Quit
				}
				m.statusMessage = "Wait for installation to finish or use Ctrl+C to force quit."
			}

		case "up", "k":
			switch m.view {
			case groupView:
				if m.cursorGroup > 0 {
					m.cursorGroup--
				}
			default:
				if m.cursorPackage > 0 {
					m.cursorPackage--
				}

			}

		case "down", "j":
			switch m.view {
			case groupView:
				if m.cursorGroup < len(m.groups)-1 {
					m.cursorGroup++
				}
			default:
				if m.cursorPackage < len(m.filtered)-1 {
					m.cursorPackage++

				}
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
				m.cursorGroup = 0
				return m, fetchGroupsCmd(m.queries)
			}
			m.statusMessage = "Select packages first!"
			return m, nil

		case "/":
			m.searching = true
			m.searchInput.Focus()
			m.searchInput.SetValue("")
			return m, textinput.Blink

		case "x":
			switch m.view {
			case groupView:
				m.statusMessage = "deleting group: " + m.groups[m.cursorGroup].Name
				if err := service.DeleteGroup(context.Background(), m.queries, m.groups[m.cursorGroup]); err != nil {
					m.err = err
					return m, nil
				}

				m.groupInput.Focus()
				return m, nil

			default:
				pkgs := []string{}
				for i := range m.selected {
					pkgs = append(pkgs, i.Name)
				}

				if err := service.DeletePackage(context.Background(), m.queries, pkgs); err != nil {
					m.err = err
					return m, nil
				}

				m.err = nil
				return m, refreshListCmd(m.queries, m.sm)
			}

		case "i":
			switch m.view {
			case groupView:
				m.installing = true
				m.statusMessage = ""
				g := m.groups[m.cursorGroup]
				return m, tea.Batch(installGroupCmd(context.Background(), m.queries, g.Name), m.spinner.Tick)

			default:
				if len(m.selected) > 0 {
					m.installing = true
					m.statusMessage = ""
					pkgs := make([]string, 0, len(m.selected))
					for pkg := range m.selected {
						pkgs = append(pkgs, pkg.Url)
					}
					m.selected = make(map[data.Package]struct{})
					return m, tea.Batch(installPackagesCmd(pkgs), m.spinner.Tick)
				}

			}

		case "enter":
			switch m.view {
			case groupView:
				if len(m.groups) == 0 {
					return m, nil
				}
				m.activeGroup = m.groups[m.cursorGroup]
				m.view = groupPackageView
				m.cursorPackage = 0
				return m, fetchPackagesByGroupCmd(m.queries, m.activeGroup.Name)

			case groupPackageView, packageView:
				if len(m.filtered) > 0 {
					url := m.filtered[m.cursorPackage]
					if _, ok := m.selected[url]; ok {
						delete(m.selected, url)
					} else {
						m.selected[url] = struct{}{}
					}
				}
			}

		case " ":
			switch m.view {
			case groupView:
				if len(m.groups) == 0 {
					return m, nil
				}
				m.activeGroup = m.groups[m.cursorGroup]
				m.view = groupPackageView
				m.cursorPackage = 0
				return m, fetchPackagesByGroupCmd(m.queries, m.activeGroup.Name)

			default:
				if len(m.filtered) > 0 {
					url := m.filtered[m.cursorPackage]
					if _, ok := m.selected[url]; ok {
						delete(m.selected, url)
					} else {
						m.selected[url] = struct{}{}
					}
				}
			}

		}
	case installGroupMsg:
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			m.statusMessage = "Group installed successfully"
		}
		m.installing = false
		return m, nil
	case installFinishedMsg:
		m.installing = false
		if msg.err != nil {
			m.statusMessage = "Error: " + msg.err.Error()
		} else {
			m.statusMessage = "Installation complete!"
			return m, updateStatsCmd(m.queries, msg.installedUrls)
		}
	case sortGroupByFreqMsg:
		if msg.err != nil {
			m.statusMessage = "Error sorting: " + msg.err.Error()
		} else {
			m.choices = msg.pkgs
			m.filtered = m.choices
			m.cursorPackage = 0
			m.statusMessage = "Sorted by frequency"
		}

	case sortGroupByLastUsedMsg:
		if msg.err != nil {
			m.statusMessage = "Error sorting: " + msg.err.Error()
		} else {
			m.choices = msg.pkgs
			m.filtered = m.choices
			m.cursorPackage = 0
			m.statusMessage = "Sorted by last used"
		}

	case packageAddedMsg:
		if msg.err != nil {
			m.statusMessage = "Error adding: " + msg.err.Error()
		} else {
			m.statusMessage = "Package added successfully!"
			return m, refreshListCmd(m.queries, m.sm)
		}

	case packagesRemovedMsg:
		if msg.err != nil {
			m.statusMessage = "Error removing: " + msg.err.Error()
		} else {
			m.statusMessage = "Packages removed from group."
			m.selected = make(map[data.Package]struct{})
			return m, fetchPackagesByGroupCmd(m.queries, m.activeGroup.Name)
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
			m.selected = make(map[data.Package]struct{})
		}

	case groupsListMsg:
		if msg.err != nil {
			m.statusMessage = "Error fetching groups: " + msg.err.Error()
		} else {
			m.groups = msg.groups
		}

	case packagesListMsg:
		m.choices = msg.packages
		if m.searchInput.Value() != "" {
			m.filtered = m.choices
			m.searchInput.Reset()
		} else {
			m.filtered = m.choices
			if m.sm == sortByFrequency {
				m.statusMessage = "sorted by frequency"
			} else {
				m.statusMessage = "sorted by last used"
			}
		}

		if m.cursorPackage >= len(m.filtered) {
			m.cursorPackage = 0
		}

	case spinner.TickMsg:
		if m.installing {
			var spinCmd tea.Cmd
			m.spinner, spinCmd = m.spinner.Update(msg)
			return m, spinCmd
		}
	}

	return m, nil
}

func installGroupCmd(ctx context.Context, queries *data.Queries, groupName string) tea.Cmd {
	return func() tea.Msg {
		return installGroupMsg{
			err: service.InstallGroup(ctx, queries, groupName),
		}
	}
}

type installGroupMsg struct {
	err error
}

func sortGroupByLastUsed(context context.Context, queries *data.Queries, groupName string) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := service.ListPackagesByGroupOrderByLU(context, queries, groupName)
		return sortGroupByLastUsedMsg{
			pkgs: pkgs,
			err:  err,
		}
	}
}

type sortGroupByLastUsedMsg struct {
	pkgs []data.Package
	err  error
}

func sortGroupByFreq(context context.Context, queries *data.Queries, groupName string) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := service.ListPackagesByGroupOrderByFreq(context, queries, groupName)
		return sortGroupByFreqMsg{pkgs, err}
	}
}

type sortGroupByFreqMsg struct {
	pkgs []data.Package
	err  error
}

func removePackagesFromGroups(queries *data.Queries, pkgs map[data.Package]struct{}, group data.Group) tea.Cmd {
	return func() tea.Msg {
		err := service.RemovePackagesFromGroups(context.Background(), queries, pkgs, group.ID)
		return packagesRemovedMsg{err: err}
	}
}

func fetchPackagesByGroupCmd(q *data.Queries, group string) tea.Cmd {
	return func() tea.Msg {
		pkgs, err := service.ListPackagesByGroupOrderByFreq(context.Background(), q, group)
		if err != nil {
			return packagesListMsg{packages: []data.Package{}}
		}
		return packagesListMsg{packages: pkgs}
	}
}

func (m model) searchingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc":
			m.searching = false
			m.searchInput.Blur()
			if m.searchInput.Value() == "" {
				m.filtered = m.choices
			}
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
	m.cursorPackage = 0
	return m, cmd
}

func (m model) addingUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {

		case "tab":
			if m.focusIndex < len(m.inputs)-1 {
				m.focusIndex++
			} else {
				m.focusIndex = 0
			}
			m.updateFocus()
			return m, nil

		case "shift+tab":
			if m.focusIndex > 0 {
				m.focusIndex--
			} else {
				m.focusIndex = len(m.inputs) - 1
			}
			m.updateFocus()
			return m, nil

		case "enter":
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
			m.focusIndex++
			m.updateFocus()
			return m, nil

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

func (m model) assigningUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.cursorGroup > 0 {
				m.cursorGroup--
			}

		case "down", "j":
			if m.cursorGroup < len(m.groups)-1 {
				m.cursorGroup++
			}

		case "enter":
			if len(m.groups) == 0 {
				m.assigning = false
				return m, nil
			}

			group := m.groups[m.cursorGroup]
			m.assigning = false

			pkgs := make([]string, 0, len(m.selected))
			for pkg := range m.selected {
				pkgs = append(pkgs, pkg.Url)
			}
			return m, assignToGroupCmd(m.queries, pkgs, group.Name)

		case "q", "esc":
			m.assigning = false
			return m, nil
		}
	}
	return m, nil
}

func (m model) creatingGroupUpdate(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	var cmd tea.Cmd
	m.groupInput, cmd = m.groupInput.Update(msg)
	return m, cmd
}

func (m model) View() string {
	var s strings.Builder

	if m.creatingGroup {
		return m.createGroupView()
	}
	if m.adding {
		return m.addPackageView()
	}
	if m.assigning {
		return m.assignGroupView()
	}

	switch m.view {
	case packageView:
		s.WriteString(m.packageView("GOPK MANAGER"))
	case groupPackageView:
		s.WriteString(m.packageView("Group: " + m.activeGroup.Name))
	case groupView:
		s.WriteString(m.groupListView())
	}

	if m.installing {
		s.WriteString("\n")
		s.WriteString(m.installingView())
	}

	if m.searching {
		s.WriteString("\n")
		s.WriteString(m.searchingView())
	}

	if m.statusMessage != "" {
		s.WriteString("\n")
		// Updated to use global colorPrimary
		s.WriteString(
			lipgloss.NewStyle().
				Foreground(colorPrimary).
				Render(m.statusMessage),
		)
		s.WriteRune('\n')
	}

	help := m.helpText()
	if help != "" {
		s.WriteString("\n")
		// Updated to use global colorSecondary
		s.WriteString(
			lipgloss.NewStyle().
				Foreground(colorSecondary).
				Render(help),
		)
		s.WriteRune('\n')
	}

	return s.String()
}

func (m model) groupListView() string {
	var s strings.Builder

	s.WriteString("Groups\n\n")

	if len(m.groups) == 0 {
		s.WriteString("No groups found. Press 'c' to create one.\n")
	} else {
		for i, g := range m.groups {
			cursor := " "
			if i == m.cursorGroup {
				cursor = ">"
			}
			fmt.Fprintf(&s, "%s %s\n", cursor, g.Name)
		}
	}

	return s.String()
}

func (m model) installingView() string {
	var s strings.Builder
	fmt.Fprintf(&s, " %s Installing packages...\n\n", m.spinner.View())
	return s.String()
}

func (m model) searchingView() string {
	var s strings.Builder
	s.WriteString("\n" + m.searchInput.View())
	return s.String()
}

func (m model) addPackageView() string {
	var s strings.Builder
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

func (m model) assignGroupView() string {
	var s strings.Builder
	s.WriteString("Assign to Group\n\n")
	if len(m.groups) == 0 {
		s.WriteString("No groups found. Press 'c' to create one.")
	} else {
		for i, group := range m.groups {
			cursor := " "
			if m.cursorGroup == i {
				cursor = ">"
			}
			fmt.Fprintf(&s, "%s %s\n", cursor, group.Name)
		}
	}
	s.WriteString("\n(esc to cancel, enter to assign)")
	return s.String()
}

func (m model) createGroupView() string {
	var s strings.Builder
	s.WriteString("Create New Group\n\n")
	s.WriteString(m.groupInput.View())
	s.WriteString("\n\n(esc to cancel, enter to create)")
	return s.String()
}

func (m model) packageView(title string) string {
	var s strings.Builder

	s.WriteString(title)
	s.WriteString("\n\n")

	statusStyle := lipgloss.NewStyle().Width(6).PaddingRight(1)
	nameStyle := lipgloss.NewStyle().
		Width(30).
		PaddingRight(2).
		Foreground(colorPrimary). // Global color
		Bold(true)

	urlStyle := lipgloss.NewStyle().
		Width(50).
		PaddingRight(2).
		Foreground(colorSecondary) // Global color

	freqStyle := lipgloss.NewStyle().
		Width(6).
		Align(lipgloss.Right).
		Foreground(colorSecondary) // Global color

	header := lipgloss.JoinHorizontal(
		lipgloss.Left,
		statusStyle.Render(""),
		nameStyle.Render("PACKAGE"),
		urlStyle.Render("REPOSITORY URL"),
		freqStyle.Render("FREQ"),
	)

	s.WriteString(
		lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(colorSecondary). // Global color
			Render(header),
	)
	s.WriteRune('\n')

	if len(m.filtered) == 0 {
		s.WriteString(
			lipgloss.NewStyle().
				Foreground(colorSecondary). // Global color
				Render("  No packages found."),
		)
		s.WriteRune('\n')
		return s.String()
	}

	for i, pkg := range m.filtered {
		cursor := " "
		if m.cursorPackage == i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[pkg]; ok {
			checked = "x"
		}

		status := fmt.Sprintf("%s [%s]", cursor, checked)

		url := pkg.Url
		if len(url) > 48 {
			url = url[:45] + "..."
		}

		row := lipgloss.JoinHorizontal(
			lipgloss.Left,
			statusStyle.Render(status),
			nameStyle.Render(pkg.Name),
			urlStyle.Render(url),
			freqStyle.Render(fmt.Sprintf("%d", pkg.Freq.Int64)),
		)

		rowStyle := lipgloss.NewStyle().Width(94)

		if _, ok := m.selected[pkg]; ok {
			rowStyle = rowStyle.Foreground(colorSelected) // Global color
		}

		if m.cursorPackage == i {
			rowStyle = rowStyle.
				Background(colorCursorBg). // Global color
				Foreground(colorCursorFg)  // Global color
		}

		row = rowStyle.Render(row)

		s.WriteString(row)
		s.WriteRune('\n')

	}
	return s.String()
}

func updateStatsCmd(q *data.Queries, urls []string) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()
		for _, u := range urls {
			if err := q.UpdatePackageUsage(ctx, u); err != nil {
				return err
			}
		}
		return nil
	}
}

func (m *model) updateFocus() {
	for i := 0; i < len(m.inputs); i++ {
		if i == m.focusIndex {
			m.inputs[i].Focus()
			// Updated to use global colorPrimary
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(colorPrimary)
		} else {
			m.inputs[i].Blur()
			// Updated to use global colorSecondary
			m.inputs[i].PromptStyle = lipgloss.NewStyle().Foreground(colorSecondary)
		}
	}
}

func (m model) helpText() string {
	switch m.view {

	case packageView:
		return "/: search	g: group   +: add   a: assign to group   c: create group   i: install   q: quit"

	case groupView:
		return "space/enter: open	i: install   c: create	esc/q: back"

	case groupPackageView:
		return "space: select   i: install   d: remove from group  esc/q: back"

	default:
		return ""
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

func installPackagesCmd(pkgs []string) tea.Cmd {
	return func() tea.Msg {
		err := service.GetFromUrl(pkgs)
		return installFinishedMsg{
			err:           err,
			installedUrls: pkgs,
		}
	}
}

func addPackageCmd(q *data.Queries, url, name, version string, install, force bool) tea.Cmd {
	return func() tea.Msg {
		err := service.Add(url, name, version, install, force, q)
		return packageAddedMsg{err: err}
	}
}

func createGroupCmd(q *data.Queries, name string) tea.Cmd {
	return func() tea.Msg {
		err := service.CreateGroup(q, name)
		return groupCreatedMsg{name: name, err: err}
	}
}

func assignToGroupCmd(q *data.Queries, pkgs []string, group string) tea.Cmd {
	return func() tea.Msg {
		err := service.AssignToGroup(q, pkgs, group)
		return groupAssignedMsg{group: group, count: len(pkgs), err: err}
	}
}

func fetchGroupsCmd(q *data.Queries) tea.Cmd {
	return func() tea.Msg {
		groups, err := service.ListGroups(q)
		return groupsListMsg{groups: groups, err: err}
	}
}

func refreshListCmd(q *data.Queries, mode sortMode) tea.Cmd {
	return func() tea.Msg {
		var pkgs []data.Package
		var err error

		switch mode {
		case sortByFrequency:
			pkgs, err = q.ListPackagesByFrequency(context.Background(), -1)
		case sortByLastUsed:
			pkgs, err = q.ListPackagesByLastUsed(context.Background(), -1)
		default:
			pkgs, err = service.List(q, -1, false)
		}

		if err != nil {
			log.Printf("error refreshing list: %v", err)
			return packagesListMsg{packages: []data.Package{}}
		}
		return packagesListMsg{packages: pkgs}
	}
}

func Start(q *data.Queries) error {
	p := tea.NewProgram(
		initialModel(q),
		tea.WithAltScreen(),
	)
	if _, err := p.Run(); err != nil {
		return err
	}
	return nil
}
