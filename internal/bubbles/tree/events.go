package tree

import (
	"fmt"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		if m.searchMode {
			return m.handleSearchKey(msg)
		}
		cmd = m.onKey(msg)
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)

	var statusBarCmd tea.Cmd
	m.statusbar, statusBarCmd = m.statusbar.Update(msg)

	return m, tea.Batch(cmd, tableCmd, statusBarCmd)
}

func (m *Model) onResize(msg tea.WindowSizeMsg) tea.Cmd {
	m.setSize(msg.Width, msg.Height)
	m.table.SetWidth(msg.Width)
	m.table.SetHeight(msg.Height)
	m.SetColumns(m.table.Columns())

	var statusbarCmd tea.Cmd
	m.statusbar, statusbarCmd = m.statusbar.Update(msg)

	return statusbarCmd
}

func (m *Model) onNavUp() {
	m.cursor--
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.onSelectionChange(m.nodesByCursor[m.cursor])
}

func (m *Model) onNavDown() {
	m.cursor++
	if m.cursor >= m.numberOfNodes() {
		m.cursor = m.numberOfNodes() - 1
	}
	m.onSelectionChange(m.nodesByCursor[m.cursor])
}

func (m *Model) onSelectionChange(node *Node) {
	m.statusbar.SetPath(m.pathByNode[node])
}

func (m *Model) onKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.KeyMap.Up):
		m.onNavUp()
	case key.Matches(msg, m.KeyMap.Down):
		m.onNavDown()
	case key.Matches(msg, m.KeyMap.ShowFullHelp):
		fallthrough
	case key.Matches(msg, m.KeyMap.CloseFullHelp):
		m.Help.ShowAll = !m.Help.ShowAll
	case key.Matches(msg, m.KeyMap.Copy):
		//nolint // ignore errors
		clipboard.WriteAll(m.Current().Key)
	case key.Matches(msg, m.KeyMap.Quit):
		return tea.Interrupt
	case key.Matches(msg, m.KeyMap.Search):
		m.searchMode = true
		m.searchQuery = ""
		m.statusbar.SetPath([]string{"Search: type to filter"})
	}
	return nil
}

func (m *Model) handleSearchKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.searchMode = false
		m.applySearchFilter()
		m.statusbar.SetPath([]string{fmt.Sprintf("Showing filtered results for '%s'", m.searchQuery)})
	case "esc":
		m.searchMode = false
		m.searchQuery = ""
		m.filteredNodes = m.nodes
		m.statusbar.SetPath([]string{"Search cleared"})
	default:
		// TODO: this probably needs to deal with backspace and other keys
		m.searchQuery += msg.String()
		m.applySearchFilter()
		m.statusbar.SetPath([]string{fmt.Sprintf("Search: %s", m.searchQuery)})
	}
	return *m, nil
}
