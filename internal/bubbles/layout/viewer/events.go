package viewer

import (
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var viewerCmd tea.Cmd
	m.viewer, viewerCmd = m.viewer.Update(msg)

	return m, tea.Batch(viewerCmd)
}
