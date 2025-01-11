package tree

import (
	"github.com/charmbracelet/bubbles/key"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cmd = m.onResize(msg)
	case tea.KeyMsg:
		cmd = m.onKey(msg)
	}

	var tableCmd tea.Cmd
	m.table, tableCmd = m.table.Update(msg)

	return m, tea.Batch(cmd, tableCmd)
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
	m.statusbar.SetPath(node.Path)
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
	}
	return nil
}
